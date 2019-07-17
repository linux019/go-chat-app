package bootstrap

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"org.freedom/constants"
	"os"
	"os/signal"
	"syscall"
)

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

var PendingConnections pendingConnectionsType

var MaintenanceRoutines MaintenanceRoutine

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	fmt.Println("New conn", conn.RemoteAddr().String())

	if PendingConnections.GetConnCount() < constants.MaxHandshakeConnections {
		PendingConnections.AddConnection(conn)
		go readSocket(conn)
	} else {
		_ = conn.Close()
	}
}

var signals = make(chan os.Signal, 1)

func ListenForSignals() {
	OsSignal = <-signals
	log.Println("Terminating")
	MaintenanceRoutines.TerminateAll()
	_ = server.Shutdown(context.Background())
}

func StartHttpServer() {
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	PendingConnections.Init()
	MaintenanceRoutines.StartFunc(PendingConnections.CheckPendingConnections)

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

}

func AddEndPoints(endPoint string, handlers *HttpHandler) {
	mux.Handle(endPoint, handlers)
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
