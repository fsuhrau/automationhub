package device

type LogWriter interface {
	Log(source, format string, params ...interface{})
	Error(source, format string, params ...interface{})
}
