package monitor

import (
	"context"
	"go-service-template/config"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
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
		attribute.String("deployment.environment", env),
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

func createTracerProvider(openTelemetryCfg config.OpenTelemetryConfig, res *resource.Resource) (trace.TracerProvider, error) {
	if openTelemetryCfg.OtlpEndpoint == "" {
		return nooptrace.NewTracerProvider(), nil
	}

	// Create the OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithHeaders(buildHeaders(openTelemetryCfg.OtlpHeaders)),
		otlptracegrpc.WithEndpointURL(openTelemetryCfg.OtlpEndpoint),
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

func createMeterProvider(openTelemetryCfg config.OpenTelemetryConfig, res *resource.Resource) (metric.MeterProvider, error) {
	if openTelemetryCfg.OtlpEndpoint == "" {
		return noopmetric.NewMeterProvider(), nil
	}

	// Create the metrics exporter
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithHeaders(buildHeaders(openTelemetryCfg.OtlpHeaders)),
		otlpmetricgrpc.WithEndpointURL(openTelemetryCfg.OtlpEndpoint),
	)
	if err != nil {
		return nil, err
	}

	reader := metricsdk.NewPeriodicReader(exporter)

	mp := metricsdk.NewMeterProvider(
		metricsdk.WithReader(reader),
		// Record information about this application in a Resource.
		metricsdk.WithResource(res),
	)

	return mp, nil
}

func buildHeaders(headersStr string) map[string]string {
	if headersStr == "" {
		return nil
	}

	headers := make(map[string]string)
	for _, header := range strings.Split(headersStr, ",") {
		kv := strings.Split(header, "=")
		headers[kv[0]] = kv[1]
	}

	return headers
}
