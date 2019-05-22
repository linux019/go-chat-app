package main

import (
	"org.freedom/bootstrap"
	"org.freedom/chatapi"
)

func main() {
	bootstrap.Setup()
	chatapi.Setup()
	bootstrap.ListenForSignals()
}
