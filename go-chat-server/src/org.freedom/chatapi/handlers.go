package chatapi

import (
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
)

var commandSetUserName bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	if result {
		bootstrap.ConnectionsByUser.AddUserConn(conn, name)
		bootstrap.UserConnections.StoreConnection(conn, name)
		go dispatchUsersList()
	}
	return commandListChannels(conn, nil)
}

var commandListChannels bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	allChannelsList.mutex.RLock()
	defer allChannelsList.mutex.RUnlock()

	channels := make(map[string]channelJSON)
	for name, attributes := range allChannelsList.channels {
		channels[name] = channelJSON{
			IsPublic: attributes.IsPublic,
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

		if exists && len(channelName) > 0 && len(message.(string)) > 0 {
			allChannelsList.mutex.RLock()
			defer allChannelsList.mutex.RUnlock()
			channelData, exists := allChannelsList.channels[channelName]

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

func dispatchChannelMessage(c *channelPeers, channelName string, message *channelMessageJSON) {
	if c.IsPublic {
		bootstrap.ConnectionsByUser.WriteMessageToAll(&messageJSON{
			ChannelName: channelName,
			Message:     *message,
		})
	} else {
		panic("No private channels")
	}
}

func dispatchPublicChannels() {
	channels := make(map[string]channelJSON)

	allChannelsList.mutex.RLock()
	defer allChannelsList.mutex.RUnlock()

	for name, attributes := range allChannelsList.channels {
		channels[name] = channelJSON{
			IsPublic: attributes.IsPublic,
		}
	}

	bootstrap.ConnectionsByUser.WriteMessageToAll(&channelsJSON{
		Channels: channels,
	})
}

func dispatchUsersList() {
	users := usersJSON{
		Users: make(map[string]userJSON),
	}

	userConns := bootstrap.ConnectionsByUser.GetConnectedUsersStatus()

	for name, connCount := range userConns {
		users.Users[name] = userJSON{
			Online: connCount > 0,
		}
	}

	bootstrap.ConnectionsByUser.WriteMessageToAll(&users)
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)

	if result {

		allChannelsList.AddChannel(name, true)
	}

	go dispatchPublicChannels()
	return nil
}

//var commandListUsers bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
//
//	users := usersJSON{
//		Users: make(map[string]userJSON),
//	}
//
//	userConns := bootstrap.ConnectionsByUser.GetConnectedUsersStatus()
//
//	for name, connCount := range userConns {
//		users.Users[name] = userJSON{
//			Online: connCount > 0,
//		}
//	}
//
//	return users
//}
