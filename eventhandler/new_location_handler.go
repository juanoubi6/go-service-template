// nolint
package eventhandler

import (
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"go-service-template/domain"
	"go-service-template/monitor"
)

type NewLocationEventHandler struct {
	logger monitor.AppLogger
}

func CreateNewLocationHandler() *NewLocationEventHandler {
	return &NewLocationEventHandler{
		logger: monitor.GetStdLogger("LocationConsumer"),
	}
}

func (c *NewLocationEventHandler) GetData() (name string, topic string) {
	return "NewLocationEventHandler", domain.LocationsNewTopic
}

func (c *NewLocationEventHandler) Process(msg *message.Message) error {
	fnName := "NewLocationEventHandler.Process"
	var appCtx monitor.ApplicationContext

	appCtx = monitor.CreateAppContextFromContext(msg.Context(), msg.Metadata.Get(monitor.CorrelationIDField))

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	var newLocation domain.Location
	if err := json.Unmarshal(msg.Payload, &newLocation); err != nil {
		c.logger.ErrorCtx(appCtx, fnName, "failed to unmarshal location", err)
		return err
	}

	c.logger.InfoCtx(appCtx, fnName, fmt.Sprintf("Received new location: %v", newLocation))

	return nil
}
