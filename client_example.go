// nolint
package main

import (
	"context"
	"fmt"
	"go-service-template/config"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
)

func registerSyncTraceProvider(collectorEndpoint string) {
	// Create the Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(collectorEndpoint)))
	if err != nil {
		panic(err)
	}

	env, _ := config.GetEnvironment()

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithSyncer(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("CLIENT"),
			attribute.String("environment", env),
			attribute.Int64("ID", 1),
		)),
	)

	// Register our TracerProvider as the global so any imported instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func runClient() {
	_, _ = config.GetEnvironment()
	appCfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Create sync otel trace provider
	registerSyncTraceProvider(appCfg.OpenTelemetryConfig.CollectorEndpoint)

	// Create http client, it already has OTEL support to propagate spans
	customHTTPClient := customHTTP.CreateCustomHTTPClient(appCfg.HTTPClientConfig)

	tr := otel.Tracer("client-service")
	spanCtx, span := tr.Start(context.Background(), "clientCall", trace.WithSpanKind(trace.SpanKindClient))

	appCtx := monitor.CreateAppContextFromContext(spanCtx, "client-service-tracer", "clientCorrID")

	header := http.Header{}
	header.Set("Correlation-Id", appCtx.GetCorrelationID())

	response, err := customHTTPClient.Do(appCtx, customHTTP.RequestValues{
		URL:       "http://localhost:8080/v1/locations?direction=next&limit=456",
		Method:    http.MethodGet,
		Headers:   header,
		Body:      nil,
		BasicAuth: nil,
	})
	if err != nil {
		panic(err)
	}

	span.End()
	time.Sleep(3 * time.Second)
	fmt.Println(response)
}
