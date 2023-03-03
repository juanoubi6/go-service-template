package pubsub

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func CreateRouter(
	middleware []message.HandlerMiddleware,
	handlers []EventHandler,
	subscriber message.Subscriber,
) (*message.Router, error) {
	logger := watermill.NewStdLogger(false, false)

	// Create router
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	// Apply middleware
	router.AddMiddleware(middleware...)

	// Register event handlers
	for _, handler := range handlers {
		handlerName, topic := handler.GetData()

		router.AddNoPublisherHandler(
			handlerName,
			topic,
			subscriber,
			handler.Process,
		)
	}

	return router, nil
}
