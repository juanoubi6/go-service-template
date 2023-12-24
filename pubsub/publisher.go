package pubsub

import (
	"encoding/json"
	"go-service-template/config"
	"go-service-template/monitor"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
)

const MessageKey = "message_key"

func CreatePublisher(config config.MessageBrokerConfig) (message.Publisher, error) {
	return nats.NewPublisher(
		nats.PublisherConfig{
			URL: config.URL,
			JetStream: nats.JetStreamConfig{
				TrackMsgId: true,  // Ensures ExactlyOnceDelivery
				AckAsync:   false, // Ensures ExactlyOnceDelivery at the cost of latency
			},
		},
		watermill.NewStdLogger(true, true),
	)

}

func CreateJSONMessage(ctx monitor.ApplicationContext, key string, payload any) (*message.Message, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(watermill.NewUUID(), bytes)
	msg.SetContext(ctx)
	msg.Metadata.Set(MessageKey, key)
	msg.Metadata.Set(monitor.CorrelationIDField, ctx.GetCorrelationID())

	return msg, nil
}

func GetMessageKeyFromMessage(_ string, msg *message.Message) (string, error) {
	return msg.Metadata.Get(MessageKey), nil
}
