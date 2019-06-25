package chatapi

import (
	"encoding/json"
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
			IsCommon: attributes.isCommon,
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
			channelsList.mutex.Lock()
			defer channelsList.mutex.Unlock()
			channelData, exists := channelsList.channels[channelName]

			if exists {
				var user, exists = bootstrap.UserConnections.LoadConnection(conn)
				if exists {
					message := channelMessages.AppendMessage(channelName, message.(string), user)
					go dispatchChannelMessage(channelData, channelName, message)
				}
			}
		}
	}

	return nil
}

func dispatchChannelMessage(c *channelPeers, channelName string, message *channelMessage) {
	jsonValue, err := json.Marshal(messageJSON{
		ChannelName: channelName,
		Message:     *message,
	})
	if err != nil {
		return
	}

	if c.isCommon {
		for _, user := range bootstrap.ConnectionsByUser {
			user.Mutex.Lock()
			for conn := range user.Connections {
				_ = conn.WriteMessage(websocket.TextMessage, jsonValue)
			}
			user.Mutex.Unlock()
		}
	}
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)

	if result {
		channelsList.AddChannel(name, true)
	}

	return commandListChannels(conn, nil)
}
