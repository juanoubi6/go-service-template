package pubsub

import "github.com/ThreeDotsLabs/watermill/message"

type EventHandler interface {
	Process(msg *message.Message) error
	GetData() (name, topic string)
}
