package log

import (
	"go-service-template/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	env, _ := config.GetEnvironment()

	var zapConfig zap.Config
	var err error

	switch env {
	case config.Production:
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.CallerKey = zapcore.OmitKey
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.CallerKey = zapcore.OmitKey
	}

	logger, err = zapConfig.Build()
	if err != nil {
		panic(err)
	}
}

type LoggingParam struct {
	Name  string
	Value any
}

type AppLogger interface {
	Debug(function, cid, msg string, params ...LoggingParam)
	Info(function, cid, msg string, params ...LoggingParam)
	Warn(function, cid, msg string, params ...LoggingParam)
	Error(function, cid, msg string, err error, params ...LoggingParam)
}

type StdLogger struct {
	Object     string
	AppVersion string
}

func GetStdLogger(object string) StdLogger {
	return StdLogger{
		Object:     object,
		AppVersion: config.ServiceConf.AppConfig.Version,
	}
}

func (c StdLogger) Debug(function, cid, msg string, params ...LoggingParam) {
	logger.Debug(msg, c.getFields(function, cid, params)...)
}

func (c StdLogger) Info(function, cid, msg string, params ...LoggingParam) {
	logger.Info(msg, c.getFields(function, cid, params)...)
}

func (c StdLogger) Warn(function, cid, msg string, params ...LoggingParam) {
	logger.Warn(msg, c.getFields(function, cid, params)...)
}

func (c StdLogger) Error(function, cid, msg string, err error, params ...LoggingParam) {
	fields := c.getFields(function, cid, params)
	fields = append(fields, zap.Error(err))

	logger.Error(msg, fields...)
}

func (c StdLogger) getFields(function, cid string, params []LoggingParam) []zap.Field {
	fields := []zap.Field{
		zap.String("function", function),
		zap.String("object", c.Object),
		zap.String("app_version", c.AppVersion),
		zap.String("correlation_id", cid),
	}

	for _, param := range params {
		fields = append(fields, zap.Any(param.Name, param.Value))
	}

	return fields
}

func FlushLogger() {
	if logger != nil {
		_ = logger.Sync()
	}
}
