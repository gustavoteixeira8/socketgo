package socketgo

type EventFn func(data *Message) (any, error)

type Message struct {
	Action   string `json:"action"`
	ClientID string
	Data     any `json:"data"`
}
