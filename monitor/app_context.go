package monitor

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type KeyValuePair struct {
	Key   string
	Value string
}

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

	// Search if context has a baggage which contains a correlationID. If not, set the provided one
	baggageCorrelationID := baggage.FromContext(ctx).Member(CorrelationIDField).Value()
	if baggageCorrelationID == "" {
		ctx = addBaggageToContext(ctx, KeyValuePair{CorrelationIDField, correlationID})
	}

	return &AppContext{Context: ctx}
}

func CreateAppContextFromRequest(request *http.Request, correlationID string) *AppContext {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	requestCtx := request.Context()

	// Search if context has a baggage which contains a correlationID. If not, set the provided one
	baggageCorrelationID := baggage.FromContext(requestCtx).Member(CorrelationIDField).Value()
	if baggageCorrelationID == "" {
		requestCtx = addBaggageToContext(requestCtx, KeyValuePair{CorrelationIDField, correlationID})
	}

	return &AppContext{Context: requestCtx}
}

func CreateMockAppContext(operationName string) *AppContext {
	return &AppContext{
		Context: context.WithValue(context.Background(), CorrelationIDContextKey, operationName),
	}
}

func (appCtx *AppContext) GetCorrelationID() string {
	return baggage.FromContext(appCtx).Member(CorrelationIDField).Value()
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

func addBaggageToContext(ctx context.Context, kvPairs ...KeyValuePair) context.Context {
	// Create member list
	members := []baggage.Member{}

	// Transform all key-value pairs into baggage members
	for _, kvPair := range kvPairs {
		member, _ := baggage.NewMember(kvPair.Key, kvPair.Value)
		members = append(members, member)
	}

	// Create the baggage with the members
	bag, err := baggage.New(members...)
	if err != nil {
		panic(err)
	}

	return baggage.ContextWithBaggage(ctx, bag)
}
