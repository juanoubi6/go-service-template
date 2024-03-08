package monitor

import (
	"context"
	"go-service-template/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

var (
	globalTracerProvider trace.TracerProvider
	globalTracer         trace.Tracer
	globalMeterProvider  metric.MeterProvider
	globalMeter          metric.Meter
)

func GetGlobalTracer() trace.Tracer {
	if globalTracer != nil {
		return globalTracer
	}

	return otel.GetTracerProvider().Tracer("default-tracer")
}

func GetGlobalMeter() metric.Meter {
	if globalMeter != nil {
		return globalMeter
	}

	return otel.GetMeterProvider().Meter("default-meter")
}

func FlushMonitorTools(ctx context.Context) {
	if globalTracerProvider != nil {
		if gtp, ok := globalTracerProvider.(*tracesdk.TracerProvider); ok {
			gtp.ForceFlush(ctx)
			gtp.Shutdown(ctx)
		}
	}

	if globalMeterProvider != nil {
		if gmp, ok := globalMeterProvider.(*metricsdk.MeterProvider); ok {
			gmp.ForceFlush(ctx)
			gmp.Shutdown(ctx)
		}
	}
}

func RegisterMonitoringTools(openTelemetryCfg config.OpenTelemetryConfig, appCfg config.AppConfig) {
	env, _ := config.GetEnvironment()

	toolsResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(appCfg.Name),
		attribute.String("environment", env),
		attribute.String("version", appCfg.Version),
	)

	// Create a TracerProvider
	tp, err := createTracerProvider(openTelemetryCfg, toolsResource)
	if err != nil {
		panic(err)
	}
	globalTracerProvider = tp
	globalTracer = tp.Tracer(appCfg.Name)
	otel.SetTracerProvider(globalTracerProvider)

	// Create a MeterProvider
	mp, err := createMeterProvider(openTelemetryCfg, toolsResource)
	if err != nil {
		panic(err)
	}
	globalMeterProvider = mp
	globalMeter = mp.Meter(appCfg.Name)
	otel.SetMeterProvider(globalMeterProvider)

	// TODO: Create a LogProvider once OpenTelemetry develops the SDK

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Create global logger
	NewGlobalLogger()
}

// createTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the OTLP exporter that will send spans to the provided endpoint. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func createTracerProvider(openTelemetryCfg config.OpenTelemetryConfig, res *resource.Resource) (trace.TracerProvider, error) {
	if openTelemetryCfg.TracesCollectorEndpoint == "" {
		return nooptrace.NewTracerProvider(), nil
	}

	// Create the OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(openTelemetryCfg.TracesCollectorEndpoint),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		// Record information about this application in a Resource.
		tracesdk.WithResource(res),
	)

	return tp, nil
}

// createMeterProvider returns an OpenTelemetry MeterProvider configured to use
// the OTLP exporter that will send spans to the provided endpoint. The returned
// MeterProvider will also use a Resource configured with all the information
// about the application.
func createMeterProvider(openTelemetryCfg config.OpenTelemetryConfig, res *resource.Resource) (metric.MeterProvider, error) {
	if openTelemetryCfg.MetricsCollectorEndpoint == "" {
		return noopmetric.NewMeterProvider(), nil
	}

	// Create the prometheus metrics exporter
	meterReader, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	mp := metricsdk.NewMeterProvider(
		// Always be sure to batch in production.
		metricsdk.WithReader(meterReader),
		// Record information about this application in a Resource.
		metricsdk.WithResource(res),
	)

	return mp, nil
}
