package socketgo

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Client struct {
	id     string
	conn   *websocket.Conn
	events map[string]EventFn
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) Send(data any) error {
	return c.conn.WriteJSON(data)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Read() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *Client) Listen() {
	defer c.Close()

	for {
		msgType, msg, err := c.Read()
		if err != nil || msgType == websocket.CloseMessage {
			logrus.Errorf("error reading message to %s: %v", err, c.id)
			return
		}

		msgInstance := &Message{}

		err = json.Unmarshal(msg, msgInstance)
		if err != nil {
			c.Send(fmt.Sprintf("error reading message %v", err))
			continue
		}

		fn, ok := c.events[msgInstance.Action]
		if !ok {
			c.Send("event not found")
			continue
		}

		msgInstance.ClientID = c.id

		out, err := fn(msgInstance)
		if err != nil {
			c.Send(err.Error())
			continue
		}

		c.Send(out)
	}
}

func (c *Client) Notify(event string, msgInstance *Message) {
	fn, ok := c.events[event]
	if !ok {
		c.Send("action not found")
		return
	}

	out, err := fn(msgInstance)
	if err != nil {
		c.Send(err.Error())
		return
	}

	c.Send(out)
}

func NewClient(conn *websocket.Conn, events map[string]EventFn) *Client {
	return &Client{
		id:     uuid.NewString(),
		conn:   conn,
		events: events,
	}
}
