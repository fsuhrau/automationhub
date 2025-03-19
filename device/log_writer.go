package device

type LogWriter interface {
	LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string)
	Data(source, data string)
	Log(source, format string, params ...interface{})
	Error(source, format string, params ...interface{})
	TestProtocolId() *uint
	Device() interface{}
	Parent() LogWriter
}
