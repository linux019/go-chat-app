package main

import (
	"chat-demo/go-chat-server/bootstrap"
	"chat-demo/go-chat-server/chatapi"
)

func main() {
	bootstrap.StartHttpServer()
	chatapi.Setup()
	bootstrap.ListenForSignals()
}
