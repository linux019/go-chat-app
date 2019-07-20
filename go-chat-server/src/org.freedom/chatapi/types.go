package chatapi

import (
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
	"org.freedom/constants"
	"sync"
	"time"
)

type channel struct {
	m             sync.RWMutex
	IsPublic      bool
	peers         []*User
	messages      []channelMessageJSON
	messagesMutex sync.RWMutex
}

func (c *channel) AddPeer(u *User) {
	c.m.Lock()
	defer c.m.Unlock()
	for _, peer := range c.peers {
		if peer == u {
			return
		}
	}
	c.peers = append(c.peers, u)
}

type channels struct {
	m   sync.RWMutex
	chs map[string]*channel
}

func (c *channels) Add(publicity bool, creator *User, name string) *channel {
	c.m.Lock()
	defer c.m.Unlock()
	ch := &channel{IsPublic: publicity}

	if !publicity && creator == nil {
		panic("Private channels must have owner")
	}

	if publicity {
		users.m.Lock()
		defer users.m.Unlock()

		for _, user := range users.users {
			ch.AddPeer(user)
		}
	} else {
		creator.AddPrivateChannels(name)
	}

	c.chs[name] = ch

	return ch
}

type User struct {
	name     string
	m        sync.RWMutex
	conns    []*websocket.Conn
	channels map[string]*channel
}

func (u *User) GetChannels() ChannelsJSON {
	u.m.RLock()
	defer u.m.RUnlock()
	result := ChannelsJSON{
		Channels: make(map[string]channelJSON),
	}

	for name, channel := range u.channels {
		result.Channels[name] = channelJSON{
			IsPublic: channel.IsPublic,
		}
	}
	return result
}

type usersList struct {
	m     sync.RWMutex
	users map[string]*User
}

func (ul *usersList) LoadStoreUser(name string) *User {
	ul.m.Lock()
	defer ul.m.Unlock()
	var result *User

	if ul.users == nil {
		ul.users = make(map[string]*User)
	}
	result, ok := ul.users[name]

	if !ok {
		result = &User{
			name:     name,
			channels: make(map[string]*channel),
		}
		ul.users[name] = result
	}

	return result
}

func (u *User) AddConn(conn *websocket.Conn) {
	u.m.Lock()
	u.conns = append(u.conns, conn)
	u.m.Unlock()
}

func (u *User) AddPublicChannels() {
	allChannelsList.m.RLock()
	defer allChannelsList.m.RUnlock()
	for _, channelName := range constants.PublicChannels {
		channel, ok := allChannelsList.chs[channelName]
		if ok {
			u.m.Lock()
			u.channels[channelName] = channel
			u.m.Unlock()
			channel.AddPeer(u)
		}
	}
}

func (u *User) AddPrivateChannels(name string) {
	allChannelsList.m.Lock()
	defer allChannelsList.m.Unlock()
	ch := channel{
		IsPublic: false,
	}
	ch.AddPeer(u)
	allChannelsList.chs[u.name+":"+name] = &ch
	u.channels[name] = &ch
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

func (u *User) SendMessage(data interface{}) {
	u.m.RLock()
	for _, conn := range u.conns {
		bootstrap.NetworkMessagesChannel <- bootstrap.NetworkMessage{
			Conn:     conn,
			Jsonable: data,
		}
	}
	u.m.RUnlock()
}

type userSocketConnection struct {
	m               sync.RWMutex
	connMap         sync.Map
	sendOnlineUsers *debounceDataWriter
}

type debounceDataWriter struct {
	dataCh chan []interface{}
	acc    interface{}
}

func (d *debounceDataWriter) Write(line []interface{}) {
	go func() {
		d.dataCh <- line
	}()
}

func createDebouncedWriter(d time.Duration, callback func(data ...interface{})) *debounceDataWriter {
	dwr := &debounceDataWriter{
		dataCh: make(chan []interface{}),
	}

	go func() {
		t := time.NewTimer(d)
		t.Stop()

		for {
			select {
			case dwr.acc = <-dwr.dataCh:
				t.Reset(d)
			case <-t.C:
				callback(dwr.acc)
			}
		}
	}()

	return dwr
}

func (c *userSocketConnection) Store(conn *websocket.Conn, user *User) {
	c.connMap.Store(conn, user)
}

func (c *userSocketConnection) Get(conn *websocket.Conn) (*User, bool) {
	user, ok := c.connMap.Load(conn)
	if ok {
		return user.(*User), true
	}
	return nil, false
}

func (c *userSocketConnection) DispatchToAll(data interface{}) {
	c.m.RLock()
	defer c.m.RUnlock()
	c.connMap.Range(func(conn, user interface{}) bool {
		bootstrap.NetworkMessagesChannel <- bootstrap.NetworkMessage{Conn: conn.(*websocket.Conn), Jsonable: data}
		return true
	})
}

func (ul *usersList) GetOnlineUsers() UsersJSON {
	ul.m.RLock()
	defer ul.m.RUnlock()
	var users = UsersJSON{Users:make(map[string]userJSON)}
	for name, user := range ul.users {
		users.Users[name] = userJSON{Online: len(user.conns) > 0}
	}
	return users
}

type channelJSON struct {
	IsPublic bool `json:"isPublic"`
}

type ChannelsJSON struct {
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

type UsersJSON struct {
	Users map[string]userJSON `json:"users"`
}

func (c *channel) AppendMessage(text, sender string) *channelMessageJSON {
	var newMessage = channelMessageJSON{
		Message: text,
		Time:    time.Now().Unix(),
		Sender:  sender,
	}

	c.messagesMutex.Lock()
	defer c.messagesMutex.Unlock()

	c.messages = append(c.messages, newMessage)
	return &newMessage
}

func (c *channel) SendPeersChannelList() {
	c.m.RLock()
	defer c.m.RUnlock()
	for _, user := range c.peers {
		channels := user.GetChannels()
		user.SendMessage(&channels)
	}
}
