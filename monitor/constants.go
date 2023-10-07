package monitor

type ContextKeyType string

const (
	CorrelationIDField                     = "correlation_id"
	CorrelationIDContextKey ContextKeyType = "correlation_id"
	AppVersionLogField                     = "app_version"
	ObjectLogField                         = "object"
	FunctionLogField                       = "function"
)
