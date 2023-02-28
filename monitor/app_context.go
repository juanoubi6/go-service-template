package monitor

import (
	"context"
	"github.com/google/uuid"
)

type ApplicationContext interface {
	context.Context
	GetCorrelationID() string
}

type AppContext struct {
	correlationID string
	context.Context
}

func CreateAppContext(parent context.Context, correlationID string) *AppContext {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	appCtx := &AppContext{
		correlationID: correlationID,
		Context:       parent,
	}

	return appCtx
}

func (appCtx AppContext) GetCorrelationID() string {
	return appCtx.correlationID
}
