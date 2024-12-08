package device

type LogWriter interface {
	LogPerformance(checkpoint string, cpu, fps, mem float64, other string)
	Data(source, data string)
	Log(source, format string, params ...interface{})
	Error(source, format string, params ...interface{})
}
