package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type CommandListener func(conn *websocket.Conn, data interface{}) interface{}

type ConnectionsMap map[*websocket.Conn]struct{}

type UserSocketConnections struct {
	Mutex       sync.RWMutex
	Connections ConnectionsMap
}

type connectionsByUser struct {
	mutex             sync.RWMutex
	socketConnections map[string]*UserSocketConnections
}

func (c *connectionsByUser) AddUserConn(conn *websocket.Conn, name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	conns, ok := c.socketConnections[name]
	if ok {
		conns.Mutex.Lock()
		conns.Connections[conn] = struct{}{}
		conns.Mutex.Unlock()
	} else {
		conns = &UserSocketConnections{
			Connections: make(ConnectionsMap),
		}
		conns.Connections[conn] = struct{}{}
		c.socketConnections[name] = conns
	}
}

func (c *connectionsByUser) WriteMessageToAll(jsonable interface{}) {
	jsonValue, err := json.Marshal(jsonable)
	if err != nil {
		return
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, user := range c.socketConnections {
		user.Mutex.RLock()
		for conn := range user.Connections {
			_ = conn.WriteMessage(websocket.TextMessage, jsonValue)
		}
		user.Mutex.RUnlock()
	}
}

func (c *connectionsByUser) GetConnectedUsersStatus() map[string]int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var users = make(map[string]int)
	for userName, connections := range c.socketConnections {
		connections.Mutex.RLock()
		users[userName] = len(connections.Connections)
		connections.Mutex.RUnlock()
	}

	return users
}

var ConnectionsByUser = connectionsByUser{
	socketConnections: make(map[string]*UserSocketConnections),
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

			ConnectionsByUser.mutex.RLock()
			for userName, userConns := range ConnectionsByUser.socketConnections {
				userConns.Mutex.Lock()
				for conn := range userConns.Connections {
					err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(time.Second*10))
					if err != nil {
						_ = conn.Close()
						delete(userConns.Connections, conn)
						UserConnections.Delete(conn)
						fmt.Printf("Disconnecting %v\n", userName)
					}
				}
				userConns.Mutex.Unlock()
			}
			ConnectionsByUser.mutex.RUnlock()
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
