package bootstrap

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)

type CommandListener func(conn *websocket.Conn, data interface{}) interface{}

type connectionData struct {
	M sync.Mutex
}

type ConnectionsMap map[*websocket.Conn]connectionData

type UserSocketConnections struct {
	Mutex       sync.RWMutex
	Connections ConnectionsMap
}

var serverCommandListeners = make(map[string]CommandListener)

func AddCommandListener(command string, f CommandListener) {
	serverCommandListeners[command] = f
}

func readSocket(conn *websocket.Conn) {
	for {
		messageType, data, err := conn.ReadMessage()

		if err != nil {
			break
		}

		if (messageType == websocket.BinaryMessage || messageType == websocket.TextMessage) && data != nil {
			var v interface{}
			if json.Unmarshal(data, &v) == nil {
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

					if cmdHandler, result := serverCommandListeners[stringCommand]; result {
						response := cmdHandler(conn, commandData)
						if response != nil {
							NetworkMessagesChannel <- NetworkMessage{
								Conn:      conn,
								IsControl: false,
								Jsonable:  response,
							}
						}
					}
				}
			}
		}
	}
}
