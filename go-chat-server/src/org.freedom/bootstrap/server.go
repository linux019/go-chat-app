package bootstrap

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)

type CommandListener func(conn *websocket.Conn, data interface{}) interface{}

type ConnectionsMap map[*websocket.Conn]struct{}

type UserSocketConnections struct {
	Mutex       sync.RWMutex
	Connections ConnectionsMap
}

type connectionsByUser struct {
	Mutex             sync.RWMutex
	SocketConnections map[string]*UserSocketConnections
}

func (c *connectionsByUser) AddUserConn(conn *websocket.Conn, name string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	conns, ok := c.SocketConnections[name]
	if ok {
		conns.Mutex.Lock()
		conns.Connections[conn] = struct{}{}
		conns.Mutex.Unlock()
	} else {
		conns = &UserSocketConnections{
			Connections: make(ConnectionsMap),
		}
		conns.Connections[conn] = struct{}{}
		c.SocketConnections[name] = conns
	}
}

func (c *connectionsByUser) WriteMessageToAll(jsonable interface{}) {
	jsonValue, err := json.Marshal(jsonable)
	if err != nil {
		return
	}
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	for _, user := range c.SocketConnections {
		user.Mutex.RLock()
		for conn := range user.Connections {
			_ = conn.WriteMessage(websocket.TextMessage, jsonValue)
		}
		user.Mutex.RUnlock()
	}
}

func (c *connectionsByUser) GetConnectedUsersStatus() map[string]int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	var users = make(map[string]int)
	for userName, connections := range c.SocketConnections {
		connections.Mutex.RLock()
		users[userName] = len(connections.Connections)
		connections.Mutex.RUnlock()
	}

	return users
}

var ConnectionsByUser = connectionsByUser{
	SocketConnections: make(map[string]*UserSocketConnections),
}

type UserConnection struct {
	sync.Map
}

func (uc *UserConnection) StoreConnection(conn *websocket.Conn, userName string) (name string, loaded bool) {
	v, loaded := uc.LoadOrStore(conn, userName)
	if loaded {
		name = v.(string)
	}
	return
}

func (uc *UserConnection) LoadConnection(conn *websocket.Conn) (name string, ok bool) {
	v, ok := uc.Load(conn)
	if ok {
		name = v.(string)
	}
	return
}

var UserConnections = UserConnection{}
var serverCommandListeners = make(map[string]CommandListener)

func AddCommandListener(command string, f CommandListener) {
	serverCommandListeners[command] = f
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
