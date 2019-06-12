package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var ConnectionPool struct {
	connectionCounter uint64
	connections       sync.Map //connId:websocket.Conn
}

var UserConnections = make(map[*websocket.Conn]string)

type CommandListener func(conn *websocket.Conn, data interface{}) interface{}

var serverCommandListeners = make(map[string]CommandListener)

func AddCommandListener(command string, f CommandListener) {
	serverCommandListeners[command] = f
}

func CheckKeepAliveSockets() {
	step := 0
	for {
		_ = <-time.After(time.Second * 1)
		step++
		if step > 30 {
			step = 0
			if OsSignal != nil {
				break
			}
			ConnectionPool.connections.Range(func(connId, value interface{}) bool {
				conn := value.(*websocket.Conn)
				fmt.Printf("PING %v\n", connId)
				err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(time.Second*10))
				if err != nil {
					_ = conn.Close()
					ConnectionPool.connections.Delete(connId)
					delete(UserConnections, conn)
					fmt.Printf("Disconnecting %v\n", connId)
				}
				return true
			})
		}
	}
}

func ReadSocket(conn *websocket.Conn) {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if (messageType == websocket.BinaryMessage || messageType == websocket.TextMessage) && data != nil {
			var v interface{}
			err := json.Unmarshal(data, &v)
			if err == nil {
				jsonMap, success := v.(map[string]interface{})
				if jsonMap != nil && success {
					command, result := jsonMap["command"]
					if !result {
						continue
					}
					stringCommand, result := command.(string)
					if !result {
						continue
					}
					commandData, result := jsonMap["data"]
					if !result {
						continue
					}

					cmdHandler, result := serverCommandListeners[stringCommand]
					if result {
						response := cmdHandler(conn, commandData)
						if response != nil {
							jsonValue, err := json.Marshal(response)
							if err == nil {
								_ = conn.WriteMessage(websocket.TextMessage, jsonValue)
							}
						}
					}
				}
			}
		}
	}
}
