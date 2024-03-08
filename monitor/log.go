package monitor

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go-service-template/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *otelzap.Logger

func NewGlobalLogger() {
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

	zapLogger, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	logger = otelzap.New(zapLogger)
}

type LoggingParam struct {
	Name  string
	Value any
}

type AppLogger interface {
	Debug(function, corrID, msg string, params ...LoggingParam)
	DebugCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam)
	Info(function, corrID, msg string, params ...LoggingParam)
	InfoCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam)
	Warn(function, corrID, msg string, params ...LoggingParam)
	WarnCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam)
	Error(function, corrID, msg string, err error, params ...LoggingParam)
	ErrorCtx(ctx ApplicationContext, function, msg string, err error, params ...LoggingParam)
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

func (c StdLogger) Debug(function, corrID, msg string, params ...LoggingParam) {
	logger.Debug(msg, c.getFields(function, corrID, params)...)
}

func (c StdLogger) DebugCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam) {
	logger.Ctx(ctx).Debug(msg, c.getFields(function, ctx.GetCorrelationID(), params)...)
}

func (c StdLogger) Info(function, corrID, msg string, params ...LoggingParam) {
	logger.Info(msg, c.getFields(function, corrID, params)...)
}

func (c StdLogger) InfoCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam) {
	logger.Ctx(ctx).Info(msg, c.getFields(function, ctx.GetCorrelationID(), params)...)
}

func (c StdLogger) Warn(function, corrID, msg string, params ...LoggingParam) {
	logger.Warn(msg, c.getFields(function, corrID, params)...)
}

func (c StdLogger) WarnCtx(ctx ApplicationContext, function, msg string, params ...LoggingParam) {
	logger.Ctx(ctx).Warn(msg, c.getFields(function, ctx.GetCorrelationID(), params)...)
}

func (c StdLogger) Error(function, corrID, msg string, err error, params ...LoggingParam) {
	fields := c.getFields(function, corrID, params)
	fields = append(fields, zap.Error(err))

	logger.Error(msg, fields...)
}

func (c StdLogger) ErrorCtx(ctx ApplicationContext, function, msg string, err error, params ...LoggingParam) {
	fields := c.getFields(function, ctx.GetCorrelationID(), params)
	fields = append(fields, zap.Error(err))

	logger.Ctx(ctx).Error(msg, fields...)
}

func (c StdLogger) getFields(function, cid string, params []LoggingParam) []zap.Field {
	fields := []zap.Field{
		zap.String(FunctionLogField, function),
		zap.String(ObjectLogField, c.Object),
		zap.String(AppVersionLogField, c.AppVersion),
		zap.String(CorrelationIDField, cid),
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
