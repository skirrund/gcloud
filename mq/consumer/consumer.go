package consumer

type Message struct {
	Value           string
	Payload         []byte
	RedeliveryCount uint32
}

type Consumer interface {
	OnMessage(message Message) error
}
