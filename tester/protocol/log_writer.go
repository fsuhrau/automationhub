package protocol

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"
)

type LogWriter struct {
	db         *gorm.DB
	protocolId uint
	errs       []error
	startTime time.Time
}

func (w *LogWriter) LogPerformance(checkpoint string, cpu, fps, mem float32, other string) {
	entry := models.ProtocolPerformanceEntry{
		TestProtocolID: w.protocolId,
		Checkpoint:     checkpoint,
		FPS:            fps,
		MEM:            mem,
		CPU:            cpu,
		Other:          other,
		Runtime: w.getRuntime(),
	}
	w.db.Create(&entry)
}

func (w *LogWriter) getRuntime() int64 {
	return time.Now().UTC().Unix() - w.startTime.Unix()
}

func (w *LogWriter) write(source, level, message, data string) {
	entry := models.ProtocolEntry{
		Timestamp:      time.Now(),
		TestProtocolID: w.protocolId,
		Source:         source,
		Level:          level,
		Message:        message,
		Data:           data,
		Runtime: w.getRuntime(),
	}
	w.db.Create(&entry)

	events.NewTestProtocolLog.Trigger(events.NewTestProtocolLogPayload{
		TestProtocolID: w.protocolId,
		Entry:          entry,
	})
}

func (w *LogWriter) Data(source, data string) {
	w.write(source, "log", "", data)
}

func (w *LogWriter) Log(source, format string, params ...interface{}) {
	w.write(source, "log", fmt.Sprintf(format, params...), "")
}

func (w *LogWriter) Error(source, format string, params ...interface{}) {
	msg := fmt.Sprintf(format, params...)
	w.errs = append(w.errs, fmt.Errorf(msg))
	w.write(source, "error", msg, "")
}

func NewLogWriter(db *gorm.DB, protocolId uint) *LogWriter {
	return &LogWriter{
		db:         db,
		protocolId: protocolId,
		startTime: time.Now().UTC(),
	}
}
