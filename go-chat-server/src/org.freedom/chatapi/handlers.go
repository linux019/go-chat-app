package chatapi

import (
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
)

var commandSetUserName bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	if result {
		bootstrap.UserConnections[conn] = name
	}
	return commandListChannels(conn, nil)
}

var commandListChannels bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	channelsList.mutex.Lock()
	defer channelsList.mutex.Unlock()

	names := make([]string, 0)
	for name := range channelsList.channels {
		names = append(names, name)
	}

	return channelsJSON{
		Channels: &names,
	}
}

var commandListChannelMessages bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	channel, success := data.(string)
	if !success {
		return nil
	}
	messages, ok := channelMessages[channel]

	if ok {
		return messagesJSON{
			Messages: &messages,
		}

	}

	return nil
}

var commandStoreUserMessage bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	/*name := r.FormValue("name")
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
	}*/
	return nil
}