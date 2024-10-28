package monitor

import (
	"context"
	"go-service-template/config"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	nooplog "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

func FlushMonitorTools(ctx context.Context) {
	if gtp, ok := otel.GetTracerProvider().(*tracesdk.TracerProvider); ok {
		gtp.ForceFlush(ctx)
		gtp.Shutdown(ctx)
	}

	if gmp, ok := otel.GetMeterProvider().(*metricsdk.MeterProvider); ok {
		gmp.ForceFlush(ctx)
		gmp.Shutdown(ctx)
	}

	if glp, ok := logglobal.GetLoggerProvider().(*logsdk.LoggerProvider); ok {
		glp.ForceFlush(ctx)
		glp.Shutdown(ctx)
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
	otel.SetTracerProvider(tp)

	// Create a MeterProvider
	mp, err := createMeterProvider(openTelemetryCfg, toolsResource)
	if err != nil {
		panic(err)
	}
	otel.SetMeterProvider(mp)

	// Create a LogProvider (beta)
	lp, err := createLogProvider(openTelemetryCfg, toolsResource)
	if err != nil {
		panic(err)
	}
	logglobal.SetLoggerProvider(lp)

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

func createLogProvider(openTelemetryCfg config.OpenTelemetryConfig, res *resource.Resource) (log.LoggerProvider, error) {
	if openTelemetryCfg.OtlpEndpoint == "" {
		return nooplog.NewLoggerProvider(), nil
	}

	// Create the OTLP exporter
	exporter, err := otlploggrpc.New(
		context.Background(),
		otlploggrpc.WithHeaders(buildHeaders(openTelemetryCfg.OtlpHeaders)),
		otlploggrpc.WithEndpointURL(openTelemetryCfg.OtlpEndpoint),
	)
	if err != nil {
		return nil, err
	}

	lp := logsdk.NewLoggerProvider(
		// Always be sure to batch in production.
		logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter)),
		// Record information about this application in a Resource.
		logsdk.WithResource(res),
	)

	return lp, nil
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
