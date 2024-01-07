package main

import (
	"fmt"
	"socketgo"
)

var msgs = map[string][]string{}

func main() {
	c := socketgo.NewWsHandler()

	c.On("message", func(data *socketgo.Message) (any, error) {
		clientMsgs := msgs[data.ClientID]

		clientMsgs = append(clientMsgs, data.Data.(string))

		msgs[data.ClientID] = clientMsgs

		c.NotifyAll("get-all-messages")

		return nil, nil
	})

	c.On("get-all-messages", func(data *socketgo.Message) (any, error) {
		fmt.Println(msgs)

		return msgs, nil
	})

	c.InitServer(":3000")
}
