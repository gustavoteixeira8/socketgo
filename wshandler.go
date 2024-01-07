package socketgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WsHandler struct {
	connMap  map[string]*websocket.Conn
	events   map[string]func(data *Message) (any, error)
	mutex    *sync.RWMutex
	upgrader websocket.Upgrader
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
		connMap:  make(map[string]*websocket.Conn),
		events:   make(map[string]func(data *Message) (any, error)),
		mutex:    &sync.RWMutex{},
		upgrader: wsUpgrader,
	}
}

type Message struct {
	Action   string `json:"action"`
	ClientID string
	Data     any `json:"data"`
}

func (c *WsHandler) InitServer(port string) error {
	router := http.NewServeMux()

	handleFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, err := c.upgrader.Upgrade(w, r, nil)
		if err != nil {
			Response(
				w,
				fmt.Sprintf("error upgrading connection: %v", err),
				http.StatusInternalServerError,
			)
			return
		}

		id := uuid.NewString()

		c.mutex.Lock()
		c.connMap[id] = conn
		c.mutex.Unlock()

		go c.listen(id)
	})

	router.Handle("/ws", handleFunc)
	router.Handle("/ws/", handleFunc)

	return http.ListenAndServe(port, router)
}

func (c *WsHandler) On(event string, fn func(data *Message) (any, error)) {
	c.mutex.Lock()

	_, ok := c.events[event]
	if ok {
		return
	}

	c.events[event] = fn

	c.mutex.Unlock()
}

func (c *WsHandler) NotifyAll(event string) {

	for id, conn := range c.connMap {
		msgInstance := &Message{ClientID: id, Action: event}

		fn, ok := c.events[event]
		if !ok {
			conn.WriteJSON("action not found")
			continue
		}

		msgInstance.ClientID = id

		out, err := fn(msgInstance)
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(out)
	}

}

func (c *WsHandler) listen(id string) {

	conn := c.connMap[id]

	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil || msgType == websocket.CloseMessage {
			logrus.Errorf("error reading message to %s: %v", err, id)
			return
		}

		msgInstance := &Message{}

		err = json.Unmarshal(msg, msgInstance)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("error reading message"))
			continue
		}

		fn, ok := c.events[msgInstance.Action]
		if !ok {
			conn.WriteJSON("action not found")
			continue
		}

		msgInstance.ClientID = id

		out, err := fn(msgInstance)
		if err != nil {
			conn.WriteJSON(err.Error())
			continue
		}

		conn.WriteJSON(out)

	}
}
