package device

type LogWriter interface {
	Data(source, data string)
	Log(source, format string, params ...interface{})
	Error(source, format string, params ...interface{})
}
