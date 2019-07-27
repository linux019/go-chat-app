package main

import (
	"org.freedom/go-chat-server/bootstrap"
	"org.freedom/go-chat-server/chatapi"
)

func main() {
	bootstrap.StartHttpServer()
	chatapi.Setup()
	bootstrap.ListenForSignals()
}
