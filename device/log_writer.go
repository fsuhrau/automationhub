package device

type LogWriter interface {
	Log(format string, params ...interface{})
	Error(format string, params ...interface{})
}
