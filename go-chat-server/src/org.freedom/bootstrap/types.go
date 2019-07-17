package bootstrap

import (
	"container/list"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"org.freedom/constants"
	"sync"
	"sync/atomic"
	"time"
)

type void struct{}

type ApiHandler = func(r *http.Request) (status int, response *[]byte, e error)

type HttpHandler struct {
	ApiHandlers map[string]ApiHandler
}

type MaintenanceRoutine struct {
	blocker atomic.Value
	signals []chan void
}

func (m *MaintenanceRoutine) StartFunc(f func(signalChannel <-chan void, args ...interface{})) {
	ch := make(chan void)
	m.signals = append(m.signals, ch)
	m.blocker.Store(false)
	go f(ch)
}

func (m *MaintenanceRoutine) TerminateAll() {
	if m.blocker.Load() == false {
		m.blocker.Store(true)
		defer m.blocker.Store(false)
		for _, ch := range m.signals {
			ch <- void{}
			close(ch)
		}
	}
}

type pendingConnection struct {
	conn *websocket.Conn
	time int64
}

type pendingConnectionsType struct {
	conn list.List
	m    sync.RWMutex
}

func (pc *pendingConnectionsType) AddConnection(conn *websocket.Conn) {
	pc.m.Lock()
	defer pc.m.Unlock()
	pc.conn.PushBack(pendingConnection{
		conn: conn,
		time: time.Now().Unix(),
	})
}

func (pc *pendingConnectionsType) GetConnCount() int {
	pc.m.RLock()
	defer pc.m.RUnlock()
	return pc.conn.Len()
}

func (pc *pendingConnectionsType) RemoveConn(conn *websocket.Conn) {
	pc.m.Lock()
	defer pc.m.Unlock()
	for connIter := pc.conn.Front(); connIter != nil; connIter = conn.Next() {
		if connIter.Value.(pendingConnection).conn == conn {
			pc.conn.Remove(connIter)
			break
		}
	}
}

func (pc *pendingConnectionsType) CheckPendingConnections(signalChannel <-chan void, args ...interface{}) {
	var timeout <-chan time.Time

	for {
		timeout = time.After(time.Second * constants.MaxAuthorizationTime)

		select {

		case <-signalChannel:
			break

		case <-timeout:
			pc.m.Lock()

			currentTime := time.Now().Unix() - constants.MaxAuthorizationTime

			for conn := pc.conn.Front(); conn != nil; conn = conn.Next() {
				if conn.Value.(pendingConnection).time > currentTime {
					break
				}
				_ = conn.Value.(pendingConnection).conn.Close()
				fmt.Println("Closing conn ", conn.Value.(pendingConnection).conn.RemoteAddr())
				pc.conn.Remove(conn)
			}

			pc.m.Unlock()
		}
	}
}

func (pc *pendingConnectionsType) Init() {
	pc.conn.Init()
}
