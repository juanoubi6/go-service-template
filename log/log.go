package log

import (
	"fmt"
	"go-service-template/config"
)

type AppLogger interface {
	Debug(function string, cid string, msg string, objs ...interface{})
	Info(function string, cid string, msg string, objs ...interface{})
	Warn(function string, cid string, msg string, objs ...interface{})
	Error(function string, cid string, err error, objs ...interface{})
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

func (c StdLogger) Debug(function string, cid string, msg string, objs ...interface{}) {
	//TODO: implement with ZAP
	fmt.Sprintf("to implement")
}

func (c StdLogger) Info(function string, cid string, msg string, objs ...interface{}) {
	//TODO: implement with ZAP
	fmt.Sprintf("to implement")
}

func (c StdLogger) Warn(function string, cid string, msg string, objs ...interface{}) {
	//TODO: implement with ZAP
	fmt.Sprintf("to implement")
}

func (c StdLogger) Error(function string, cid string, err error, objs ...interface{}) {
	//TODO: implement with ZAP
	fmt.Sprintf("to implement")
}
