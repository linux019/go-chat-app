package chatapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"org.freedom/bootstrap"
	"sync"
	"time"
)

//var channelsHandlers = bootstrap.HttpHandler{
//	ApiHandlers: map[string]bootstrap.ApiHandler{
//		"get":  listChannels,
//		"post": addChannel,
//	},
//}

//var messagesHandlers = bootstrap.HttpHandler{
//	ApiHandlers: map[string]bootstrap.ApiHandler{
//		"get":  getChannelHistory,
//		"post": storeMessage,
//	},
//}

var wsHandlers = bootstrap.HttpHandler{
	ApiHandlers: map[string]bootstrap.ApiHandler{
		"get": wsHandler,
	},
}

func Setup() {
	//bootstrap.AddEndPoints("/channels", &channelsHandlers)
	//bootstrap.AddEndPoints("/messages", &messagesHandlers)
	bootstrap.AddEndPoints("/ws", &wsHandlers)
}

type Channels struct {
	mutex    sync.Mutex
	channels map[string]bool
}

type channelsJSON struct {
	Channels *[]string `json:"channels"`
}

var channelsList = Channels{
	channels: map[string]bool{"general": true},
}

type channelMessage struct {
	Time    int64  `json:"time"`
	Message string `json:"message"`
}

//type Messages []channelMessage

var chatMessagesHistory = make(map[string][]channelMessage)

func listChannels(r *http.Request) (status int, response *[]byte, e error) {
	channelsList.mutex.Lock()
	defer channelsList.mutex.Unlock()

	names := make([]string, 0)
	for name := range channelsList.channels {
		names = append(names, name)
	}
	channelsResponse := channelsJSON{
		Channels: &names,
	}
	body, _ := json.Marshal(&channelsResponse)

	return http.StatusOK, &body, nil
}

func addChannel(r *http.Request) (status int, response *[]byte, e error) {
	name := r.FormValue("name")
	if len(name) > 0 && len(name) < 255 {
		channelsList.mutex.Lock()
		channelsList.channels[name] = true
		channelsList.mutex.Unlock()
		return listChannels(r)
	}
	return http.StatusBadRequest, nil, nil
}

func storeMessage(r *http.Request) (status int, response *[]byte, e error) {
	name := r.FormValue("name")
	text := r.FormValue("text")
	if len(name) > 0 && len(text) > 0 {
		channelsList.mutex.Lock()
		defer channelsList.mutex.Unlock()
		_, ok := channelsList.channels[name]

		if ok {
			var newMessage = channelMessage{Message: text, Time: time.Now().Unix()}
			channelHistory, exists := chatMessagesHistory[name]
			if !exists {
				fmt.Println(channelHistory)
				channelHistory = make([]channelMessage, 0, 1)
			}
			channelHistory = append(channelHistory, newMessage)
			chatMessagesHistory[name] = channelHistory
			return http.StatusOK, nil, nil
		}
	}
	return http.StatusBadRequest, nil, nil
}

func getChannelHistory(r *http.Request) (status int, response *[]byte, e error) {
	channel, ok := r.URL.Query()["channel"]
	if !ok {
		return http.StatusBadRequest, nil, nil
	}
	messages, ok := chatMessagesHistory[channel[0]]

	if !ok {
		return http.StatusNotFound, nil, nil
	}

	body, _ := json.Marshal(&messages)
	return http.StatusOK, &body, nil
}

func wsHandler(r *http.Request) (status int, response *[]byte, e error) {
	var body = []byte("PONG")
	return http.StatusOK, &body, nil
}
