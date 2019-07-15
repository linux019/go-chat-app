package chatapi

import (
	"sync"
	"time"
)

//type channelData struct {
//	IsPublic bool
//	messages []channelMessageJSON
//}
//
//type User struct {
//	name     string
//	channels map[string]*channelData
//}

type ChannelsList struct {
	mutex    sync.RWMutex
	channels map[string]*channelPeers
}

type channelJSON struct {
	IsPublic bool `json:"isPublic"`
}

type channelsJSON struct {
	Channels map[string]channelJSON `json:"channels"`
}

type messagesJSON struct {
	Messages *[]channelMessageJSON `json:"messages"`
}

type messageJSON struct {
	ChannelName string             `json:"channelName"`
	Message     channelMessageJSON `json:"message"`
}

type channelMessageJSON struct {
	Time    int64  `json:"time"`
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

type userJSON struct {
	Online bool `json:"online"`
}

type usersJSON struct {
	Users map[string]userJSON `json:"users"`
}

type channelPeers struct {
	mutex    sync.RWMutex
	IsPublic bool
	peers    []string
}

type channelsMessagesMap map[string][]channelMessageJSON

type channelMessagesHistory struct {
	mutex    sync.RWMutex
	messages channelsMessagesMap
}

func (c *channelMessagesHistory) AppendMessage(channelName, text, sender string) *channelMessageJSON {
	var newMessage = channelMessageJSON{
		Message: text,
		Time:    time.Now().Unix(),
		Sender:  sender,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	channelsMessagesArray, _ := c.messages[channelName]
	if channelsMessagesArray == nil {
		channelsMessagesArray = make([]channelMessageJSON, 0, 1)
	}
	channelsMessagesArray = append(channelsMessagesArray, newMessage)
	channelMessages.messages[channelName] = channelsMessagesArray
	return &newMessage
}

func (cl *ChannelsList) AddChannel(name string, IsPublic bool) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	_, exists := cl.channels[name]
	if !exists {
		cl.channels[name] = &channelPeers{IsPublic: IsPublic}
	}
}
