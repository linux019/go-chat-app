package chatapi

import (
	"fmt"
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
	"time"
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
	messages, ok := channelMessages.messages[channel]

	if ok {
		return messagesJSON{
			Messages: &messages,
		}

	}

	return nil
}

var commandStoreUserMessage bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	valueMap, success := data.(map[string]interface{})
	if !success {
		return nil
	}
	channel, exists := valueMap["channel"]
	if exists {
		channelName, success := channel.(string)
		if !success {
			return nil
		}
		message, exists := valueMap["message"]
		fmt.Println(message, exists)

		if len(channelName) > 0 && len(message.(string)) > 0 {
			_, exists := channelsList.channels[channelName]

			if exists {
				var user, _ = bootstrap.UserConnections[conn]
				if exists {
					var newMessage = channelMessage{
						Message: message.(string),
						Time:    time.Now().Unix(),
						Sender:  user,
					}
					channelMessages.mutex.Lock()
					defer channelMessages.mutex.Unlock()

					channelsMessagesArray, _ := channelMessages.messages[channelName]
					if channelsMessagesArray == nil {
						channelsMessagesArray = make([]channelMessage, 0, 1)
					}
					channelsMessagesArray = append(channelsMessagesArray, newMessage)
					channelMessages.messages[channelName] = channelsMessagesArray
					return commandListChannelMessages(conn, channelName)
				}

			}
		}
	}

	return nil
}
