package chatapi

import (
	"github.com/gorilla/websocket"
	"org.freedom/go-chat-server/bootstrap"
	"sync"
	"time"
)

type channel struct {
	m             sync.RWMutex
	id            string
	name          string
	isPublic      bool
	isSelf        bool
	isP2P         bool
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

func (c channel) HasPeer(user *User) bool {
	for _, peer := range c.peers {
		if peer == user {
			return true
		}
	}
	return false
}

type channels struct {
	m   sync.RWMutex
	chs map[string]*channel
}

func (c *channels) Add(attrs newChannelAttributes) *channel {
	c.m.Lock()
	defer c.m.Unlock()
	id := RandomString(32)

	ch := &channel{
		id:       id,
		name:     attrs.name,
		isPublic: attrs.isPublic,
		isP2P:    attrs.isP2P,
		isSelf:   attrs.isSelf,
		messages: make([]channelMessageJSON, 0, 0),
	}

	c.chs[id] = ch

	return ch
}

func (c channels) Get(channelId string) (ch *channel, ok bool) {
	ch, ok = c.chs[channelId]
	return
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

	for id, channel := range u.channels {
		result.Channels[id] = channelJSON{
			Name:     channel.name,
			IsPublic: channel.isPublic,
			IsSelf:   channel.isSelf,
		}
	}
	return result
}

func (u *User) FindOrCreateP2PChannel(peerName string) (ch *channel, result bool) {
	peer, ok := users.Get(peerName)
	if ok {
		u.m.RLock()
		for _, channel := range u.channels {
			if channel.isP2P && len(channel.peers) == 2 && channel.HasPeer(peer) {
				u.m.RUnlock()
				return channel, true
			}
		}
		u.m.RUnlock()

		ch := createChannelConnectPeers(newChannelAttributes{
			isP2P:    true,
			isPublic: false,
			peers:    []*User{u, peer},
		})
		return ch, true
	}
	return nil, false
}

type usersList struct {
	m     sync.RWMutex
	users map[string]*User
}

func (ul *usersList) Get(name string) (*User, bool) {
	if ul.users == nil {
		return nil, false
	}
	return ul.users[name], true
}

func (ul *usersList) LoadStoreUser(name string) (result *User, exists bool) {
	ul.m.Lock()
	defer ul.m.Unlock()

	if ul.users == nil {
		ul.users = make(map[string]*User)
	}

	result, exists = ul.users[name]

	if !exists {
		result = &User{
			name:     name,
			channels: make(map[string]*channel, 0),
		}
		ul.users[name] = result
	}
	return
}

func (u *User) AddConn(conn *websocket.Conn) {
	u.m.Lock()
	u.conns = append(u.conns, conn)
	u.m.Unlock()
}

func (u *User) ConnectChannel(ch *channel) {
	u.m.Lock()
	defer u.m.Unlock()

	_, exists := u.channels[ch.id]

	if !exists {
		u.channels[ch.id] = ch
		ch.AddPeer(u)
	}
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
			Conn:      conn,
			IsControl: false,
			Jsonable:  data,
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
		bootstrap.NetworkMessagesChannel <- bootstrap.NetworkMessage{
			Conn:      conn.(*websocket.Conn),
			IsControl: false,
			Jsonable:  data,
		}
		return true
	})
}

func (ul *usersList) GetOnlineUsers() UsersJSON {
	ul.m.RLock()
	defer ul.m.RUnlock()
	var users = UsersJSON{Users: make(map[string]userJSON)}
	for name, user := range ul.users {
		users.Users[name] = userJSON{Online: len(user.conns) > 0}
	}
	return users
}

type channelJSON struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"isPublic"`
	IsSelf   bool   `json:"isSelf"`
	IsP2P    bool   `json:"isP2P"`
	Peer     string `json:"peer"`
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
	Time   int64  `json:"time"`
	Text   string `json:"text"`
	Sender string `json:"sender"`
}

type newChannelMessageJSON struct {
	Message   channelMessageJSON `json:"message"`
	ChannelId string             `json:"channelId"`
}

type userJSON struct {
	Online bool `json:"online"`
}

type UsersJSON struct {
	Users map[string]userJSON `json:"users"`
}

func (c *channel) AppendMessage(text, sender string) channelMessageJSON {
	var newMessage = channelMessageJSON{
		Text:   text,
		Time:   time.Now().Unix(),
		Sender: sender,
	}

	c.messagesMutex.Lock()
	defer c.messagesMutex.Unlock()

	c.messages = append(c.messages, newMessage)
	return newMessage
}

func (c *channel) SendPeersChannelList() {
	c.m.RLock()
	defer c.m.RUnlock()
	for _, user := range c.peers {
		channels := user.GetChannels()
		user.SendMessage(&channels)
	}
}

type clientChannelAttributes struct {
	channelId   string
	channelName string
	isPublic    bool
	isP2P       bool
	peers       []string
}

type newChannelAttributes struct {
	name     string
	isPublic bool
	isP2P    bool
	isSelf   bool
	peers    []*User
}
