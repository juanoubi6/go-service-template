package pubsub

import (
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"go-service-template/config"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func CreatePublisher(kafkaCfg *sarama.Config, kafkaParams config.KafkaConfig) (message.Publisher, error) {
	if len(kafkaParams.Brokers) == 0 {
		return nil, errors.New("brokers slice cannot be empty")
	}

	return kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               kafkaParams.Brokers,
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: kafkaCfg,
			OTELEnabled:           true,
		},
		watermill.NewStdLogger(true, true),
	)
}

func CreateJSONMessage(payload any) (*message.Message, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return message.NewMessage(watermill.NewUUID(), bytes), nil
}
