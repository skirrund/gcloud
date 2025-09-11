package mq

type Consumer interface {
	OnMessage(message Message) error
}
