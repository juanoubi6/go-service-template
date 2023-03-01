package monitor

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type ApplicationContext interface {
	context.Context
	GetCorrelationID() string
	GetTracer() trace.Tracer
	GetRootSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span)
	StartSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span)
}

type AppContext struct {
	rootSpan      trace.Span
	tracer        trace.Tracer
	correlationID string
	context.Context
}

func CreateAppContextFromRequest(request *http.Request, correlationID string) *AppContext {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	tracer := otel.Tracer(request.RequestURI, trace.WithInstrumentationAttributes(
		attribute.String(CorrelationIDField, correlationID),
	))

	ctx, rootSpan := tracer.Start(request.Context(), request.RequestURI,
		trace.WithAttributes(attribute.String(CorrelationIDField, correlationID)),
		trace.WithSpanKind(trace.SpanKindServer),
	)

	appCtx := &AppContext{
		tracer:        tracer,
		rootSpan:      rootSpan,
		correlationID: correlationID,
		Context:       ctx,
	}

	return appCtx
}

func (appCtx *AppContext) GetCorrelationID() string {
	return appCtx.correlationID
}

func (appCtx *AppContext) GetTracer() trace.Tracer {
	if appCtx.tracer == nil {
		return trace.NewNoopTracerProvider().Tracer("no-name-tracer")
	}

	return appCtx.tracer
}

// GetRootSpan returns the appCtx itself as the context and the root span created before
func (appCtx *AppContext) GetRootSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span) {
	opts = append(opts,
		trace.WithAttributes(attribute.String(CorrelationIDField, appCtx.correlationID)),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithNewRoot(),
	)

	newCtx, rootSpan := appCtx.tracer.Start(appCtx, name, opts...)

	return appCtx.clone(newCtx), rootSpan
}

// StartSpan is a wrapper around tracer.Start() that returns an ApplicationContext object instead of a plain context
func (appCtx *AppContext) StartSpan(name string, opts ...trace.SpanStartOption) (ApplicationContext, trace.Span) {
	newCtx, span := appCtx.tracer.Start(appCtx, name, opts...)

	return appCtx.clone(newCtx), span
}

func CreateMockAppContext(operationName string) *AppContext {
	appCtx := &AppContext{
		tracer:        trace.NewNoopTracerProvider().Tracer(operationName),
		correlationID: operationName,
		Context:       context.Background(),
	}

	return appCtx
}

func (appCtx *AppContext) clone(newCtx context.Context) *AppContext {
	return &AppContext{
		rootSpan:      appCtx.rootSpan,
		tracer:        appCtx.tracer,
		correlationID: appCtx.correlationID,
		Context:       newCtx,
	}
}
