package socketgo

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsHandler struct {
	httpServer *http.Server
	connMap    map[string]*Client
	events     map[string]EventFn
	mutex      *sync.RWMutex
	upgrader   websocket.Upgrader
}

func NewWsHandler() *WsHandler {
	wsUpgrader := websocket.Upgrader{
		WriteBufferSize: 1024 * 1024 * 2,
		ReadBufferSize:  1024 * 1024 * 2,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WsHandler{
		connMap:  make(map[string]*Client),
		events:   make(map[string]EventFn),
		mutex:    &sync.RWMutex{},
		upgrader: wsUpgrader,
	}
}

func (ws *WsHandler) Serve(endpoint string, port string) error {
	router := http.NewServeMux()

	handleFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			Response(
				w,
				fmt.Sprintf("error upgrading connection: %v", err),
				http.StatusInternalServerError,
			)
			return
		}

		client := NewClient(conn, ws.events)

		ws.mutex.Lock()
		ws.connMap[client.GetID()] = client
		ws.mutex.Unlock()

		go client.Listen()
	})

	router.Handle(endpoint, handleFunc)
	router.Handle(fmt.Sprintf("%s/", endpoint), handleFunc)

	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	ws.httpServer = server

	return server.ListenAndServe()
}

func (ws *WsHandler) On(event string, fn func(data *Message) (any, error)) {
	ws.mutex.Lock()

	_, ok := ws.events[event]
	if ok {
		return
	}

	ws.events[event] = fn

	ws.mutex.Unlock()
}

func (ws *WsHandler) NotifyAll(event string, data *Message) {
	for _, client := range ws.connMap {
		client.Notify(event, data)
	}
}

func (ws *WsHandler) Close() error {
	var err error

	for _, client := range ws.connMap {
		err = client.Close()
		if err != nil {
			return err
		}
	}

	err = ws.httpServer.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return nil
}
