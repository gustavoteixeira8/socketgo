package socketgo

type EventFn func(data *Message) (any, error)

type Message struct {
	Event    string `json:"Event"`
	ClientID string `json:"clientId"`
	Data     any    `json:"data"`
}
