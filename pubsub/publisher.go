package pubsub

import (
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"go-service-template/config"
	"go-service-template/monitor"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

const MessageKey = "message_key"

func CreatePublisher(kafkaCfg *sarama.Config, kafkaParams config.KafkaConfig) (message.Publisher, error) {
	if len(kafkaParams.Brokers) == 0 {
		return nil, errors.New("brokers slice cannot be empty")
	}

	return kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               kafkaParams.Brokers,
			Marshaler:             kafka.NewWithPartitioningMarshaler(GetMessageKeyFromMessage),
			OverwriteSaramaConfig: kafkaCfg,
			OTELEnabled:           true,
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
