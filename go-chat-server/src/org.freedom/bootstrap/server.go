package bootstrap

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var ConnectionPool struct {
	//mutex             sync.Mutex
	connectionCounter uint64
	connections       sync.Map
}

func CheckKeepAliveSockets() {
	step := 0;
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
