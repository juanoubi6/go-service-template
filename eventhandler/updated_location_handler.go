// nolint
package eventhandler

import (
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"go-service-template/domain"
	"go-service-template/monitor"
)

type UpdatedLocationEventHandler struct {
	logger monitor.AppLogger
}

func CreateUpdatedLocationHandler() *UpdatedLocationEventHandler {
	return &UpdatedLocationEventHandler{
		logger: monitor.GetStdLogger("LocationConsumer"),
	}
}

func (c *UpdatedLocationEventHandler) GetData() (name string, topic string) {
	return "UpdatedLocationEventHandler", domain.LocationsUpdatedTopic
}

func (c *UpdatedLocationEventHandler) Process(msg *message.Message) error {
	fnName := "UpdatedLocationEventHandler.Process"
	ctx := monitor.CreateAppContextFromContext(msg.Context(), msg.Metadata.Get(monitor.CorrelationIDField))

	var newLocation domain.Location
	if err := json.Unmarshal(msg.Payload, &newLocation); err != nil {
		c.logger.ErrorCtx(ctx, fnName, "failed to unmarshal location", err)
		return err
	}

	c.logger.InfoCtx(ctx, fnName, fmt.Sprintf("Received updated location: %v", newLocation))

	return nil
}
