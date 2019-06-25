package chatapi

import (
	"fmt"
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
)

var commandSetUserName bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	if result {
		userData, exists := bootstrap.ConnectionsByUser[name]
		if exists {
			userData.Mutex.Lock()
			userData.Connections[conn] = struct{}{}
			userData.Mutex.Unlock()
		} else {
			userData = &bootstrap.UserSocketConnections{
				Connections: make(bootstrap.ConnectionsMap),
			}
			userData.Connections[conn] = struct{}{}
			bootstrap.ConnectionsByUser[name] = userData
		}
		bootstrap.UserConnections.StoreConnection(conn, name)
	}
	return commandListChannels(conn, nil)
}

var commandListChannels bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	channelsList.mutex.Lock()
	defer channelsList.mutex.Unlock()

	channels := make(map[string]channelJSON)
	for name, attributes := range channelsList.channels {
		channels[name] = channelJSON{
			IsPublic: attributes.isPublic,
		}
	}

	return channelsJSON{
		Channels: channels,
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
				var user, exists = bootstrap.UserConnections.LoadConnection(conn)
				if exists {
					channelMessages.AppendMessage(channelName, message.(string), user)
					return commandListChannelMessages(conn, channelName)
				}

			}
		}
	}

	return nil
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)

	if result {
		channelsList.AddChannel(name, true)
	}

	return commandListChannels(conn, nil)
}
