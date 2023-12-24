package pubsub

import (
	"go-service-template/config"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
)

func CreateSubscriber(config config.MessageBrokerConfig) (message.Subscriber, error) {
	return nats.NewSubscriber(
		nats.SubscriberConfig{
			URL:              config.URL,
			QueueGroupPrefix: config.QueueGroupPrefix, // Ensures ConsumerGroup functionality
		},
		watermill.NewStdLogger(true, true),
	)
}
