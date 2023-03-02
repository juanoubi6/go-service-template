# go-service-template

Simple CONTROLLER &rarr; SERVICE &rarr; REPOSITORY service template for a Golang Application

## Features

+ Route handling using [Echo](https://echo.labstack.com/)
+ Swagger support using [Swag](https://github.com/swaggo/swag)
+ Custom HTTP Client that includes retry support
+ [OpenTelemetry](https://opentelemetry.io/docs/instrumentation/go/) support, using [Jaeger](https://www.jaegertracing.io/) as Exporter
    * Logs using [Zap](https://github.com/uber-go/zap)
    * Traces using [Golang OTEL SDK](https://github.com/open-telemetry/opentelemetry-go)
    * Metrics using [Prometheus](https://prometheus.io/docs/guides/go-application/) via [Echo Prometheus Middleware](https://echo.labstack.com/middleware/prometheus/)
    * SQL Query tracing using [OTELSQL Wrapper](https://github.com/nhatthm/otelsql)
    * Propagate spans via Custom HTTP Client using [OTELHTTP](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp)
    * Request tracing using [OTELECHO](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho)
+ DB Migrations using [Golang Migrate](https://github.com/golang-migrate/migrate)
