package chatapi

import (
	"chat-demo/go-chat-server/bootstrap"
	"github.com/gorilla/websocket"
)

var commandSetUserName bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	if result {
		bootstrap.PendingConnections.RemoveConn(conn)
		pUser, exists := users.LoadStoreUser(name)
		pUser.AddConn(conn)
		for _, ch := range publicChannels {
			pUser.ConnectChannel(ch)
		}

		if !exists {
			createChannelConnectPeers(newChannelAttributes{
				isSelf:   true,
				isDM:     false,
				name:     "self",
				isPublic: false,
				peers:    []*User{pUser},
			})
		}
		userSocketConnections.Store(conn, pUser)
		userSocketConnections.sendOnlineUsers.Write(nil)
	}
	return commandListChannels(conn, nil)
}

var commandListChannels bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	user, ok := userSocketConnections.Get(conn)
	if ok {
		return user.GetChannels()
	}
	return nil
}

var commandListChannelMessages bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	var channelData, err = decodeChannelAttributes(data)
	var (
		ok, exists bool
		ch         *channel
		user       *User
	)
	if err == nil {
		channelId := channelData.channelId
		user, ok = userSocketConnections.Get(conn)
		if ok {
			if channelData.isDM {
				if len(channelData.peers) != 1 {
					return nil
				}

				ch, exists = allChannelsList.Get(channelId)
				if !exists {
					ch, ok = user.FindOrCreateDMChannel(channelData.peers[0])
				} else {
					ok = true
				}
			} else {
				ch, ok = user.channels[channelId]
			}

			if ok {
				ch.m.RLock()
				defer ch.m.RUnlock()
				if !exists {
					go ch.SendChannelsListToPeers()
				}
				return messagesJSON{
					Messages: &ch.messages,
				}
			}
		}
	}
	return nil
}

var commandStoreUserMessage bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	valueMap, success := data.(map[string]interface{})
	if !success {
		return nil
	}

	channelData, err := decodeChannelAttributes(data)

	if err != nil {
		return nil
	}

	channelId := channelData.channelId

	message, exists := valueMap["message"]
	user, ok := userSocketConnections.Get(conn)
	if exists && ok && len(channelId) > 0 && len(message.(string)) > 0 {
		ch, ok := user.channels[channelId]
		if ok {
			newMessage := ch.AppendMessage(message.(string), user.name)
			go dispatchChannelMessage(ch, &newChannelMessageJSON{
				Message:   newMessage,
				ChannelId: channelId,
			})
		}
	}

	return nil
}

func dispatchChannelMessage(ch *channel, message *newChannelMessageJSON) {
	ch.m.RLock()
	defer ch.m.RUnlock()
	for _, user := range ch.peers {
		user.SendMessage(message)
	}
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	channelData, err := decodeChannelAttributes(data)
	user, ok := userSocketConnections.Get(conn)

	if err == nil && ok {
		ch := createChannelConnectPeers(newChannelAttributes{
			isPublic: channelData.isPublic,
			peers:    []*User{user},
			name:     channelData.channelName,
		})
		go ch.SendChannelsListToPeers()
	}

	return nil
}
