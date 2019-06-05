#!/bin/sh
export GOPATH=$(pwd)/go-chat-server
export GOBIN=$GOPATH/bin
#go env
go build org.freedom/main
