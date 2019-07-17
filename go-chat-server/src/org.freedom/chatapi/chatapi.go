package chatapi

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"org.freedom/bootstrap"
	"time"
)

//-----------
var users usersList
//-----------

var allChannelsList = ChannelsList{
	channels: make(map[string]*channelPeers),
}

var channelMessages = channelMessagesHistory{
	messages: make(channelsMessagesMap, 0),
}

func Setup() {
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
	//bootstrap.AddCommandListener("LIST_USERS", commandListUsers)
	allChannelsList.AddChannel("general", true)
	allChannelsList.AddChannel("news", true)

	go checkActiveConnections()
}

func wsHandler(r *http.Request) (status int, response *[]byte, e error) {
	var body = []byte("PONG")
	return http.StatusOK, &body, nil
}

func checkActiveConnections() {
	step := 0
	var usersListUpdated bool
	for {
		_ = <-time.After(time.Second * 1)
		step++
		if step > 30 {
			step = 0
			if bootstrap.OsSignal != nil {
				break
			}

			usersListUpdated = false

			bootstrap.ConnectionsByUser.Mutex.RLock()
			for userName, userConns := range bootstrap.ConnectionsByUser.SocketConnections {
				userConns.Mutex.Lock()
				for conn, connData := range userConns.Connections {
					connData.M.Lock()
					err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(time.Second*10))
					if err != nil {
						_ = conn.Close()
						delete(userConns.Connections, conn)
						bootstrap.UserConnections.Delete(conn)
						fmt.Printf("Disconnecting %v\n", userName)
						usersListUpdated = true
					}
					connData.M.Unlock()
				}
				userConns.Mutex.Unlock()
			}
			bootstrap.ConnectionsByUser.Mutex.RUnlock()

			if usersListUpdated {
				go dispatchUsersList()
			}
		}
	}
}

func decodeChannelAttributes(data interface{}) (channelName string, isPrivate bool, err error) {
	var channelData map[string]interface{}

	err = errors.New("")

	channelData, success := data.(map[string]interface{})

	if !success {
		return
	}

	isPrivate, success = channelData["isPrivate"].(bool)
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
