package chatapi

import (
	"chat-demo/go-chat-server/bootstrap"
	"chat-demo/go-chat-server/constants"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var users = usersList{users: make(map[string]*User)}
var userSocketConnections userSocketConnection

var allChannelsList = channels{
	chs: make(map[string]*channel),
}

var publicChannels = make([]*channel, len(constants.PublicChannels), len(constants.PublicChannels))

func Setup() {
	for i, channelName := range constants.PublicChannels {
		publicChannels[i] = createChannelConnectPeers(newChannelAttributes{
			name:     channelName,
			isPublic: true,
		})
	}

	userSocketConnections.sendOnlineUsers = createDebouncedWriter(time.Millisecond*500,
		func(data ...interface{}) {
			userSocketConnections.DispatchToAll(users.GetOnlineUsers())
		})

	bootstrap.AddEndPoints("/ws", &bootstrap.HttpHandler{
		ApiHandlers: map[string]bootstrap.ApiHandler{
			"get": wsHandler,
		},
	})

	bootstrap.AddCommandListener("SET_USERNAME", commandSetUserName)
	bootstrap.AddCommandListener("GET_CHANNELS", commandListChannels)
	bootstrap.AddCommandListener("GET_CHANNEL_MESSAGES", commandListChannelMessages)
	bootstrap.AddCommandListener("POST_MESSAGE", commandStoreUserMessage)
	bootstrap.AddCommandListener("CREATE_CHANNEL", commandCreateChannel)

	bootstrap.MaintenanceRoutines.StartFunc(checkActiveConnections)
}

func wsHandler(r *http.Request) (status int, response *[]byte, e error) {
	var body = []byte("PONG")
	return http.StatusOK, &body, nil
}

func checkActiveConnections(signalChannel <-chan bootstrap.Void, args ...interface{}) {
	var usersListUpdated bool
	timer := time.NewTimer(time.Second * 30)

	networkControlMsg := bootstrap.NetworkMessage{
		IsControl: true,
		ResultCh:  make(chan error),
	}

	for {
		select {
		case <-signalChannel:
			return

		case <-timer.C:
			usersListUpdated = false

			userSocketConnections.m.Lock()
			userSocketConnections.connMap.Range(func(key, value interface{}) bool {
				conn := key.(*websocket.Conn)
				user := value.(*User)
				networkControlMsg.Conn = conn

				bootstrap.NetworkMessagesChannel <- networkControlMsg
				err := <-networkControlMsg.ResultCh

				if err != nil {
					_ = conn.Close()
					userSocketConnections.connMap.Delete(key)
					user.RemoveConn(conn)
					fmt.Printf("Disconnecting %v\n", user.name)
					usersListUpdated = true
				}

				return true
			})
			userSocketConnections.m.Unlock()

			if usersListUpdated {
				userSocketConnections.sendOnlineUsers.Write(nil)
			}
			timer.Reset(time.Second * 30)
		}
	}
}

func decodeChannelAttributes(data interface{}) (attrs clientChannelAttributes, err error) {
	var (
		channelData map[string]interface{}
		s           string
		b           bool
		rawPeers    []interface{}
		peers       []string
	)
	err = errors.New("")

	attrs.peers = make([]string, 0)

	channelData, success := data.(map[string]interface{})

	if !success {
		return
	}

	s, success = channelData["channelName"].(string)
	if !success {
		s = ""
	}
	attrs.channelName = s

	s, success = channelData["channelId"].(string)
	if !success {
		s = ""
	}
	attrs.channelId = s

	b, success = channelData["isPublic"].(bool)
	if !success {
		b = false
	}
	attrs.isPublic = b

	b, success = channelData["isDM"].(bool)
	if !success {
		b = false
	}
	attrs.isDM = b

	rawPeers, success = channelData["peers"].([]interface{})

	if success {
		peers = make([]string, len(rawPeers))
		for i, v := range rawPeers {
			s, success = v.(string)
			if success {
				peers[i] = s
			}
		}
		attrs.peers = peers
	}

	err = nil
	return
}

func createChannelConnectPeers(attrs newChannelAttributes) *channel {
	var ch *channel
	if !attrs.isPublic && len(attrs.peers) == 0 {
		panic("Private channels must have owner")
	}

	if attrs.isDM && !attrs.isSelf {
		if len(attrs.peers) != 2 {
			panic("Invalid DM channel peers")
		}
		peer0, peer1 := attrs.peers[0], attrs.peers[1]
		ch, ok := peer0.FindDMChannel(peer1)

		if ok {
			ch.AddPeer(peer1)
			peer1.ConnectChannel(ch)
			return ch
		}

		ch, ok = peer1.FindDMChannel(peer0)

		if ok {
			ch.AddPeer(peer0)
			peer0.ConnectChannel(ch)
			return ch
		}
	}

	ch = allChannelsList.Add(attrs)

	for _, user := range attrs.peers {
		user.ConnectChannel(ch)
		ch.AddPeer(user)
	}

	if attrs.isPublic {
		for _, user := range users.users {
			user.ConnectChannel(ch)
			ch.AddPeer(user)
		}
	}

	return ch
}

/*func debounceWritePacket(ch <-chan interface{}) {
	var data interface{}

	for {
		select {
		case data = <-ch:
		case <-time.After(time.Second):
			break
		}
	}
}
*/

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	lengthCharset := len(charset)
	buf := make([]byte, length, length)
	size, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	if size != length {
		panic("Invalid size")
	}

	for index, c := range buf {
		buf[index] = charset[int(c)%lengthCharset]
	}
	return string(buf)
}
