package monitor

import (
	"context"
	"go-service-template/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
)

var globalTracerProvider *tracesdk.TracerProvider

func FlushTracerProvider(ctx context.Context) {
	if globalTracerProvider != nil {
		_ = globalTracerProvider.Shutdown(ctx)
	}
}

func RegisterTraceProvider(openTelemetryCfg config.OpenTelemetryConfig, appCfg config.AppConfig) {
	tp, err := createTracerProvider(openTelemetryCfg, appCfg)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	globalTracerProvider = tp

}

// createTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func createTracerProvider(openTelemetryCfg config.OpenTelemetryConfig, appCfg config.AppConfig) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(openTelemetryCfg.CollectorEndpoint)))
	if err != nil {
		return nil, err
	}

	env, _ := config.GetEnvironment()

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(appCfg.Name),
			attribute.String("environment", env),
			attribute.String("version", appCfg.Version),
			attribute.Int64("ID", 1),
		)),
	)

	return tp, nil
}
