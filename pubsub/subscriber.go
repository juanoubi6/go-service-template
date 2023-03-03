package pubsub

import (
	"errors"
	"github.com/Shopify/sarama"
	"go-service-template/config"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func CreateSubscriber(kafkaCfg *sarama.Config, kafkaParams config.KafkaConfig) (message.Subscriber, error) {
	if len(kafkaParams.Brokers) == 0 {
		return nil, errors.New("brokers slice cannot be empty")
	}

	kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	return kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               kafkaParams.Brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: kafkaCfg,
			ConsumerGroup:         kafkaParams.ConsumerGroup,
			OTELEnabled:           true,
		},
		watermill.NewStdLogger(true, true),
	)
}
