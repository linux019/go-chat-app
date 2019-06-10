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
	connections       sync.Map
}

var serverCommandListeners = make(map[string]func(data interface{}) interface{})

func AddCommandListener(command string, f func(data interface{}) interface{}) {
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
						cmdHandler(commandData)
					}
				}
			}
		}
	}
}
