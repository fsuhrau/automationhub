package generic

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

func (d *Device) SetLogWriter(writer device.LogWriter) {
	d.writer = writer
}

func (d *Device) GetLogWriter() device.LogWriter {
	return d.writer
}

func (d *Device) SetTestResult(success models.TestResultState) {

}

func (d *Device) TestProtocolId() *uint {
	if d.writer != nil {
		return d.writer.TestProtocolId()
	}
	return nil
}

//nolint:revive // argument-limit: signature is fixed by device.TestLogger
func (d *Device) LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string) {
	if d.writer != nil {
		d.writer.LogPerformance(checkpoint, cpu, fps, mem, vertexCount, triangles, other)
	}
}

// SetKey stores a key the app reported for the running test. Outside a test
// there is no protocol, so the key is only logged.
func (d *Device) SetKey(key, value string) {
	if d.writer != nil {
		d.writer.SetKey(key, value)
	} else {
		logrus.Infof("run key %s = %s outside a test, dropped", key, value)
	}
}

func (d *Device) Data(source, path string) {
	if d.writer != nil {
		d.writer.Data(source, path)
	}
}

func (d *Device) Log(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Log(source, format, params...)
	} else {
		logrus.Infof(format, params...)
	}
}

func (d *Device) Passed(result bool) {
	if d.writer != nil {
		d.writer.Passed(result)
	} else {
		logrus.Infof("Test: %v", result)
	}
}

func (d *Device) Error(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Error(source, format, params...)
	} else {
		logrus.Errorf(format, params...)
	}
}

func (d *Device) Exception(source, format string, params ...interface{}) {
	if d.writer != nil {
		d.writer.Error(source, format, params...)
	} else {
		logrus.Errorf(format, params...)
	}
}
