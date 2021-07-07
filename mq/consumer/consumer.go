package consumer

type Message struct {
	Value           string
	RedeliveryCount uint32
}

type Consumer interface {
	OnMessage(message Message) error
}
