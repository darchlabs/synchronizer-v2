package logger

type Client interface {
	Errorf(format string, args ...interface{})
	WithData(message string, data map[string]interface{})
	Infof(format string, args ...interface{})
}
