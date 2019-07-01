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
		bootstrap.ConnectionsByUser.AddUserConn(conn, name)
		bootstrap.UserConnections.StoreConnection(conn, name)
	}
	return commandListChannels(conn, nil)
}

var commandListChannels bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	channelsList.mutex.RLock()
	defer channelsList.mutex.RUnlock()

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

	channelMessages.mutex.RLock()
	defer channelMessages.mutex.RUnlock()

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
			channelsList.mutex.RLock()
			defer channelsList.mutex.RUnlock()
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
		bootstrap.ConnectionsByUser.WriteMessageToAll(&jsonValue)
	}
}

func dispatchPublicChannels() {
	channels := make(map[string]channelJSON)

	channelsList.mutex.RLock()
	defer channelsList.mutex.RUnlock()

	for name, attributes := range channelsList.channels {
		channels[name] = channelJSON{
			IsCommon: attributes.isCommon,
		}
	}

	jsonValue, err := json.Marshal(channelsJSON{
		Channels: channels,
	})

	if err != nil {
		return
	}

	bootstrap.ConnectionsByUser.WriteMessageToAll(&jsonValue)
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)

	if result {
		channelsList.AddChannel(name, true)
	}

	go dispatchPublicChannels()

	return nil
}

var commandListUsers bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {

	users := usersJSON{
		Users: make(map[string]userJSON),
	}

	userConns := bootstrap.ConnectionsByUser.GetConnectedUsersStatus()

	for name, connCount := range userConns {
		users.Users[name] = userJSON{
			Online: connCount > 0,
		}
	}

	return users
}
