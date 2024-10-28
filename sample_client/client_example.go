// nolint
package main

import (
	"context"
	"go-service-template/config"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func registerSyncTraceProvider(collectorEndpoint string) {
	// Create the OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(collectorEndpoint),
	)
	if err != nil {
		panic(err)
	}

	env, _ := config.GetEnvironment()

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithSyncer(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("mock-client"),
			attribute.String("environment", env),
		)),
	)

	// Register our TracerProvider as the global so any imported instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func main() {
	_, _ = config.GetEnvironment()
	appCfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	appCtx := monitor.CreateAppContextFromContext(context.Background(), "CLIENT-CORRELATION-ID")

	// Create sync otel trace provider
	registerSyncTraceProvider(appCfg.OpenTelemetryConfig.OtlpEndpoint)
	monitor.NewGlobalLogger()

	// Create http client, it already has OTEL support to propagate spans
	customHTTPClient := customHTTP.CreateCustomHTTPClient(appCfg.HTTPClientConfig)

	// Create new tracer and root span
	tr := otel.Tracer("clienTracer")
	spanCtx, span := tr.Start(
		appCtx,
		"clientCall",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String(monitor.CorrelationIDField, appCtx.GetCorrelationID())),
	)

	// Add a baggage to the context with the correlation ID
	bagMember, err := baggage.NewMember(monitor.CorrelationIDField, appCtx.GetCorrelationID())
	if err != nil {
		panic(err)
	}

	bag, err := baggage.New(bagMember)
	if err != nil {
		panic(err)
	}

	spanCtx = baggage.ContextWithBaggage(spanCtx, bag)

	// Create a new app context
	appCtx = &monitor.AppContext{Context: spanCtx}

	header := http.Header{}
	header.Set("Correlation-Id", appCtx.GetCorrelationID())

	// Send first request
	_, err = customHTTPClient.Do(appCtx, customHTTP.RequestValues{
		URL:       "http://localhost:8080/v1/locations?direction=next&limit=456",
		Method:    http.MethodGet,
		Headers:   header,
		Body:      nil,
		BasicAuth: nil,
	})
	if err != nil {
		println("Failed 1st request")
		return
	}

	// Send second request after the 1st one
	_, err = customHTTPClient.Do(appCtx, customHTTP.RequestValues{
		URL:       "http://localhost:8080/v1/location-mock",
		Method:    http.MethodPost,
		Headers:   header,
		Body:      nil,
		BasicAuth: nil,
	})
	if err != nil {
		println("Failed 2nd request")
		return
	}

	span.End()
	time.Sleep(3 * time.Second)
}
