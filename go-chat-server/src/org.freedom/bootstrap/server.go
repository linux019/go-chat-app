package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"reflect"
	"sync"
	"time"
)

var ConnectionPool struct {
	//mutex             sync.Mutex
	connectionCounter uint64
	connections       sync.Map
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
			if err == nil && reflect.ValueOf(v).Kind() == reflect.Map {
				command, exists := v.(map[string]interface{})["command"]
				if !exists {
					continue
				}
				commandData, exists := v.(map[string]interface{})["data"]
				if exists {
					fmt.Println(command, commandData)
				}
			}
		}
	}
}
