#!/bin/sh
set -e
export GOPATH=$(pwd)/go-chat-server
export GOBIN=$GOPATH/bin
#go env
go build -o chat-server org.freedom/main
strip -s chat-server
#go build -o chat-server -compiler gccgo -gccgoflags "-march=native -O3" org.freedom/main
