package chatapi

import (
	"net/http"
	"org.freedom/bootstrap"
	"sync"
)

var wsHandlers = bootstrap.HttpHandler{
	ApiHandlers: map[string]bootstrap.ApiHandler{
		"get": wsHandler,
	},
}

type ChannelsList struct {
	mutex    sync.Mutex
	channels map[string]struct{}
}

type channelsJSON struct {
	Channels *[]string `json:"channels"`
}

type messagesJSON struct {
	Messages *[]channelMessage `json:"messages"`
}

var channelsList = ChannelsList{
	channels: map[string]struct{}{
		"general": {},
		"news":    {},
	},
}

type channelMessage struct {
	Time    int64  `json:"time"`
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

type channelsMessagesMap map[string][]channelMessage
type channelMessagesHistory struct {
	mutex    sync.Mutex
	messages channelsMessagesMap
}

var channelMessages = channelMessagesHistory{
	messages: make(channelsMessagesMap, 0),
}

func Setup() {
	bootstrap.AddEndPoints("/ws", &wsHandlers)
	bootstrap.AddCommandListener("SET_USERNAME", commandSetUserName)
	bootstrap.AddCommandListener("GET_CHANNELS", commandListChannels)
	bootstrap.AddCommandListener("GET_CHANNEL_MESSAGES", commandListChannelMessages)
	bootstrap.AddCommandListener("POST_MESSAGE", commandStoreUserMessage)
	bootstrap.AddCommandListener("CREATE_CHANNEL", commandCreateChannel)
}

func wsHandler(r *http.Request) (status int, response *[]byte, e error) {
	var body = []byte("PONG")
	return http.StatusOK, &body, nil
}

func dispatchMessage() {

}
