package chatapi

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"org.freedom/bootstrap"
	"org.freedom/constants"
	"time"
)

var users = usersList{users: make(map[string]*User)}
var userSocketConnections userSocketConnection

var allChannelsList = channels{
	chs: make(map[string]*channel),
}

func Setup() {
	for _, channelName := range constants.PublicChannels {
		allChannelsList.Add(true, nil, channelName)
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


				err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(time.Second*10))
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

func decodeChannelAttributes(data interface{}) (channelName string, err error) {
	var channelData map[string]interface{}

	err = errors.New("")

	channelData, success := data.(map[string]interface{})

	if !success {
		return
	}

	channelName, success = channelData["channel"].(string)

	if !success {
		return
	}

	err = nil
	return
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

	for c, index := range buf {
		buf[index] = charset[c%lengthCharset]
	}
	return string(buf)
}
*/
