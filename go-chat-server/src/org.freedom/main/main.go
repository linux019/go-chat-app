package main

import (
	"org.freedom/bootstrap"
	"org.freedom/chatapi"
)

func main() {
	bootstrap.StartHttpServer()
	chatapi.Setup()
	bootstrap.ListenForSignals()
}
