package main

import (
	"socketgo"

	"github.com/gustavoteixeira8/localdb"
)

type ChatMessages struct {
	*localdb.Base
	Message  string
	ClientID string
}

func main() {
	c := socketgo.NewWsHandler()

	cfg := &localdb.DBManagerConfig{
		Path:        "./chat-messages",
		StorageType: localdb.StorageTypeJSON,
	}

	r := localdb.New[*ChatMessages](cfg)

	err := r.Start()
	if err != nil {
		panic(err)
	}

	err = r.Migrate(&ChatMessages{})
	if err != nil {
		panic(err)
	}

	c.On("message", func(data *socketgo.Message) (any, error) {

		r.Add(&ChatMessages{
			Base:     localdb.NewBase(),
			Message:  data.Data.(string),
			ClientID: data.ClientID,
		})

		c.NotifyAll("get-all-messages", data)

		return nil, nil
	})

	c.On("get-all-messages", func(data *socketgo.Message) (any, error) {

		msgs, err := r.Find(func(model *ChatMessages) *localdb.FindResponse[*ChatMessages] {
			return &localdb.FindResponse[*ChatMessages]{
				Query: true,
			}
		})
		if err != nil {
			return nil, err
		}

		return msgs, nil
	})

	c.Serve("/ws", ":3000")
}
