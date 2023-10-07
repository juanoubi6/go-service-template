package monitor

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type ApplicationContext interface {
	context.Context
	GetCorrelationID() string
	StartSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span)
}

type AppContext struct {
	context.Context
}

func CreateAppContextFromContext(ctx context.Context, correlationID string) *AppContext {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	newCtx := context.WithValue(ctx, CorrelationIDContextKey, correlationID)

	return &AppContext{Context: newCtx}
}

func CreateAppContextFromRequest(request *http.Request, correlationID string) *AppContext {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	existingContext := request.Context()
	newCtx := context.WithValue(existingContext, CorrelationIDContextKey, correlationID)

	return &AppContext{Context: newCtx}
}

func CreateMockAppContext(operationName string) *AppContext {
	return &AppContext{
		Context: context.WithValue(context.Background(), CorrelationIDContextKey, operationName),
	}
}

func (appCtx *AppContext) GetCorrelationID() string {
	val := appCtx.Value(CorrelationIDContextKey)
	if val == nil {
		return ""
	}

	return val.(string)
}

// StartSpan is a wrapper around tracer.Start() that returns an ApplicationContext object instead of a plain context
func (appCtx *AppContext) StartSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span) {
	opts = append(opts,
		trace.WithAttributes(attribute.String(CorrelationIDField, appCtx.GetCorrelationID())),
		trace.WithSpanKind(trace.SpanKindServer),
	)

	newCtx, span := GetGlobalTracer().Start(appCtx, name, opts...)

	return &AppContext{Context: newCtx}, span
}
