package bootstrap

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"org.freedom/constants"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type ApiHandler = func(r *http.Request) (status int, response *[]byte, e error)

type HttpHandler struct {
	ApiHandlers map[string]ApiHandler
}

var signals = make(chan os.Signal, 1)
var OsSignal os.Signal = nil
var mux = new(http.ServeMux)

var server = http.Server{
	Addr:    constants.ServerAddress,
	Handler: mux,
}

var webSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)

	//if r.Header.Get("Origin") != "http://"+r.Host {
	//	http.Error(w, "Origin not allowed", http.StatusForbidden)
	//	return
	//}

	if err != nil {
		panic(err)
	}

	newSocketId := atomic.AddUint64(&ConnectionPool.connectionCounter, 1)
	ConnectionPool.connections.Store(newSocketId, conn)
	go ReadSocket(conn)
}

//worker := h.ApiHandlers[strings.ToLower(r.Method)]
//
//var response *[]byte = nil
//
//if worker != nil {
//	wsMessageType, wsBuffer, err := conn.ReadMessage()
//	if err != nil {
//		log.Println("read", wsMessageType, wsBuffer, err)
//	}
//
//	timeText, _ := time.Now().MarshalText()
//	response = &timeText
//	//_, response, err := worker(r)
//	//w.WriteHeader(status)
//	//if status == http.StatusOK {
//	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
//	//} else {
//	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//	//}
//
//	if err != nil {
//		log.Println(err)
//	}
//
//	if response != nil && err == nil {
//		//_, _ = w.Write(*response)
//		log.Println("writing data")
//		err = conn.WriteMessage(websocket.TextMessage, *response)
//	} else {
//		log.Println("none")
//	}
//
//} else {
//	w.WriteHeader(http.StatusNotFound)
//	_, _ = w.Write([]byte("Invalid Endpoint"))
//}

func ListenForSignals() {
	OsSignal = <-signals
	log.Println("Terminating")
	_ = server.Shutdown(context.Background())
}

func StartHttpServer() {
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	go CheckKeepAliveSockets()
}

func AddEndPoints(endPoint string, handlers *HttpHandler) {
	mux.Handle(endPoint, handlers)
}
