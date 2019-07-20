package chatapi

import (
	"github.com/gorilla/websocket"
	"org.freedom/bootstrap"
)

var commandSetUserName bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	if result {
		bootstrap.PendingConnections.RemoveConn(conn)
		pUser := users.LoadStoreUser(name)
		pUser.AddConn(conn)
		pUser.AddPublicChannels()
		pUser.AddPrivateChannels("self")
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
	var channelName, err = decodeChannelAttributes(data)

	if err == nil {
		user, ok := userSocketConnections.Get(conn)
		if ok {
			ch, ok := user.channels[channelName]
			if ok {
				ch.m.RLock()
				defer ch.m.RUnlock()
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

	var channelName, err = decodeChannelAttributes(data)

	if err != nil {
		return nil
	}

	message, exists := valueMap["message"]
	user, ok := userSocketConnections.Get(conn)
	if exists && ok && len(channelName) > 0 && len(message.(string)) > 0 {
		ch, ok := user.channels[channelName]
		if ok {
			newMessage := ch.AppendMessage(message.(string), user.name)
			go dispatchChannelMessage(ch, newMessage)
		}
	}

	return nil
}

func dispatchChannelMessage(ch *channel, message *channelMessageJSON) {
	ch.m.RLock()
	defer ch.m.RUnlock()
	for _, user := range ch.peers {
		user.SendMessage(message)
	}
}

var commandCreateChannel bootstrap.CommandListener = func(conn *websocket.Conn, data interface{}) interface{} {
	name, result := data.(string)
	user, ok := userSocketConnections.Get(conn)
	if result && ok {
		ch := allChannelsList.Add(true, user, name)
		go ch.SendPeersChannelList()
	}

	return nil
}
