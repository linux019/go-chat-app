package chatapi

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

//-----------
type channelData struct {
	IsPublic bool
	messages []channelMessageJSON
}

type User struct {
	m        sync.RWMutex
	conns    []*websocket.Conn
	channels map[string]*channelData
}

type usersList struct {
	m     sync.Mutex
	users map[string]User
}

func (ul *usersList) LoadStoreUser(name string) *User {
	ul.m.Lock()
	defer ul.m.Unlock()

	if ul.users == nil {
		ul.users = make(map[string]User)
	}
	_, ok := ul.users[name]

	if !ok {
		ul.users[name] = User{}
	}
	result, _ := ul.users[name]

	return &result
}

func (u *User) AddConn(conn *websocket.Conn) {
	u.m.Lock()
	u.conns = append(u.conns, conn)
	u.m.Unlock()
}

func (u *User) RemoveConn(conn *websocket.Conn) {
	go func() {
		u.m.Lock()
		for i, c := range u.conns {
			if c == conn {
				copy(u.conns[i:], u.conns[i+1:])
				u.conns = u.conns[:len(u.conns)-1]
				break
			}
		}
		u.m.Unlock()
		_ = conn.Close()
	}()
}

type userSocketConnection struct {
	connMap sync.Map
	//ch      chan interface{}
}

/*type debouncedWriter struct {
	incoming chan []interface{}
	b        bytes.Buffer
}

func (d *debouncedWriter) Write(line []interface{}) (int, error) {
	go func() { d.incoming <- line }()
	return len(line), nil
}

func DebouncedWriter(w io.Writer, d time.Duration) io.Writer {
	dwr := &debouncedWriter{
		incoming: make(chan []interface{}),
	}

	go func() {
		t := time.NewTimer(d)
		t.Stop()

		for {
			select {
			case line := <-dwr.incoming:
				dwr.b.Write(line)
				t.Reset(d)
			case <-t.C:
				w.Write(dwr.b.Bytes())
			}
		}
	}()


	return dwr
}
*/
func (c *userSocketConnection) Store(conn *websocket.Conn, user *User) {
	c.connMap.Store(conn, user)
}

func (c *userSocketConnection) DispatchToAll(data interface{}) {
	//c.ch <- data
}

type ChannelsList struct {
	mutex    sync.RWMutex
	channels map[string]*channelPeers
}

type channelJSON struct {
	IsPublic bool `json:"isPublic"`
}

type channelsJSON struct {
	Channels map[string]channelJSON `json:"channels"`
}

type messagesJSON struct {
	Messages *[]channelMessageJSON `json:"messages"`
}

type messageJSON struct {
	ChannelName string             `json:"channelName"`
	Message     channelMessageJSON `json:"message"`
}

type channelMessageJSON struct {
	Time    int64  `json:"time"`
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

type userJSON struct {
	Online bool `json:"online"`
}

type usersJSON struct {
	Users map[string]userJSON `json:"users"`
}

type channelPeers struct {
	mutex    sync.RWMutex
	IsPublic bool
	peers    []string
}

type channelsMessagesMap map[string][]channelMessageJSON

type channelMessagesHistory struct {
	mutex    sync.RWMutex
	messages channelsMessagesMap
}

func (c *channelMessagesHistory) AppendMessage(channelName, text, sender string) *channelMessageJSON {
	var newMessage = channelMessageJSON{
		Message: text,
		Time:    time.Now().Unix(),
		Sender:  sender,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	channelsMessagesArray, _ := c.messages[channelName]
	if channelsMessagesArray == nil {
		channelsMessagesArray = make([]channelMessageJSON, 0, 1)
	}
	channelsMessagesArray = append(channelsMessagesArray, newMessage)
	channelMessages.messages[channelName] = channelsMessagesArray
	return &newMessage
}

func (cl *ChannelsList) AddChannel(name string, IsPublic bool) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	_, exists := cl.channels[name]
	if !exists {
		cl.channels[name] = &channelPeers{IsPublic: IsPublic}
	}
}
