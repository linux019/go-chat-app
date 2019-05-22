package bootstrap

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type ApiHandler = func(r *http.Request) (status int, response *[]byte, e error)

type HttpHandler struct {
	ApiHandlers map[string]ApiHandler
}

var signals = make(chan os.Signal, 1)
var mux = new(http.ServeMux)

var server = http.Server{
	Addr:    ":4488",
	Handler: mux,
}

//
//var upgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//}

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	worker := h.ApiHandlers[strings.ToLower(r.Method)]

	if worker != nil {
		status, response, err := worker(r)
		w.WriteHeader(status)
		if status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		if response != nil {
			_, _ = w.Write(*response)
		}

		if err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Invalid Endpoint"))
	}
}

func startServer(server *http.Server) {
	log.Fatal(server.ListenAndServe())
}

func ListenForSignals() {
	<-signals
	log.Println("Terminating")
	_ = server.Shutdown(context.Background())
}

func Setup() {
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go startServer(&server)
}

func AddEndPoints(endPoint string, handlers *HttpHandler) {
	mux.Handle(endPoint, handlers)
}
