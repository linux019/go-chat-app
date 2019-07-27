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

//func (c *connectionsByUser) AddUserConn(conn *websocket.Conn, pUser *User) {
	//c.Mutex.Lock()
	//defer c.Mutex.Unlock()
	//conns, ok := c.SocketConnections[name]
	//if ok {
	//	conns.Mutex.Lock()
	//	conns.Connections[conn] = connectionData{}
	//	conns.Mutex.Unlock()
	//} else {
	//	conns = &UserSocketConnections{
	//		Connections: make(ConnectionsMap),
	//	}
	//	conns.Connections[conn] = connectionData{}
	//	c.SocketConnections[name] = conns
	//}
//}



//type UserConnection struct {
//	sync.Map
//}
//
//func (uc *UserConnection) StoreConnection(conn *websocket.Conn, userName string) (name string, loaded bool) {
//	v, loaded := uc.LoadOrStore(conn, userName)
//	if loaded {
//		name = v.(string)
//	}
//	return
//}
//
//func (uc *UserConnection) LoadConnection(conn *websocket.Conn) (name string, ok bool) {
//	v, ok := uc.Load(conn)
//	if ok {
//		name = v.(string)
//	}
//	return
//}
//
//var UserConnections = UserConnection{}
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
							_ = conn.WriteJSON(response)
						}
					}
				}
			}
		}
	}
}
