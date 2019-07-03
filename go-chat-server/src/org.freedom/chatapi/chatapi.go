package chatapi

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"org.freedom/bootstrap"
	"sync"
	"time"
)

var wsHandlers = bootstrap.HttpHandler{
	ApiHandlers: map[string]bootstrap.ApiHandler{
		"get": wsHandler,
	},
}

type ChannelsList struct {
	mutex    sync.RWMutex
	channels map[string]*channelPeers
}

type channelJSON struct {
	IsCommon bool `json:"is_common"`
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
	isCommon bool
	peers    []string
}

type channelsMessagesMap map[string][]channelMessageJSON

type channelMessagesHistory struct {
	mutex    sync.RWMutex
	messages channelsMessagesMap
}

var channelsList = ChannelsList{
	channels: make(map[string]*channelPeers),
}

var channelMessages = channelMessagesHistory{
	messages: make(channelsMessagesMap, 0),
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

func (cl *ChannelsList) AddChannel(name string, isCommon bool) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	_, exists := cl.channels[name]
	if !exists {
		cl.channels[name] = &channelPeers{isCommon: isCommon}
	}
}

func Setup() {
	bootstrap.AddEndPoints("/ws", &wsHandlers)
	bootstrap.AddCommandListener("SET_USERNAME", commandSetUserName)
	bootstrap.AddCommandListener("GET_CHANNELS", commandListChannels)
	bootstrap.AddCommandListener("GET_CHANNEL_MESSAGES", commandListChannelMessages)
	bootstrap.AddCommandListener("POST_MESSAGE", commandStoreUserMessage)
	bootstrap.AddCommandListener("CREATE_CHANNEL", commandCreateChannel)
	//bootstrap.AddCommandListener("LIST_USERS", commandListUsers)
	channelsList.AddChannel("general", true)
	channelsList.AddChannel("news", true)
	go checkActiveConnections()
}

func wsHandler(r *http.Request) (status int, response *[]byte, e error) {
	var body = []byte("PONG")
	return http.StatusOK, &body, nil
}

func checkActiveConnections() {
	step := 0
	var usersListUpdated bool
	for {
		_ = <-time.After(time.Second * 1)
		step++
		if step > 30 {
			step = 0
			if bootstrap.OsSignal != nil {
				break
			}

			usersListUpdated = false

			bootstrap.ConnectionsByUser.Mutex.RLock()
			for userName, userConns := range bootstrap.ConnectionsByUser.SocketConnections {
				userConns.Mutex.Lock()
				for conn := range userConns.Connections {
					err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(time.Second*10))
					if err != nil {
						_ = conn.Close()
						delete(userConns.Connections, conn)
						bootstrap.UserConnections.Delete(conn)
						fmt.Printf("Disconnecting %v\n", userName)
						usersListUpdated = true
					}
				}
				userConns.Mutex.Unlock()
			}
			bootstrap.ConnectionsByUser.Mutex.RUnlock()

			if usersListUpdated {
				go dispatchUsersList()
			}
		}
	}
}
