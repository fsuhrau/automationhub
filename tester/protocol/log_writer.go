package protocol

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"math"
	"time"
)

type PerformanceMetric struct {
	Count int
	CPU   float64
	FPS   float64
	MEM   float64
}

type LogWriter struct {
	db                 *gorm.DB
	protocolId         uint
	errs               []error
	startTime          time.Time
	performanceMetrics PerformanceMetric
}

func (w *LogWriter) GetAvgPerformanceMetrics() (cpu, fps, mem float64) {
	cpu = w.performanceMetrics.CPU / float64(w.performanceMetrics.Count)
	fps = w.performanceMetrics.FPS / float64(w.performanceMetrics.Count)
	mem = w.performanceMetrics.MEM / float64(w.performanceMetrics.Count)
	return
}

func (w *LogWriter) LogPerformance(checkpoint string, cpu, fps, mem float64, other string) {
	w.performanceMetrics.Count++
	if math.IsNaN(cpu) {
		cpu = 0
	}
	if math.IsNaN(fps) {
		fps = 0
	}
	if math.IsNaN(mem) {
		mem = 0
	}
	w.performanceMetrics.CPU += cpu
	w.performanceMetrics.FPS += fps
	w.performanceMetrics.MEM += mem
	entry := models.ProtocolPerformanceEntry{
		TestProtocolID: w.protocolId,
		Checkpoint:     checkpoint,
		CPU:            cpu,
		FPS:            fps,
		MEM:            mem,
		Other:          other,
		Runtime:        w.getRuntime(),
	}
	w.db.Create(&entry)
}

func (w *LogWriter) getRuntime() float64 {
	return float64(time.Now().UTC().UnixNano()-w.startTime.UnixNano()) / float64(time.Second)
}

func (w *LogWriter) write(source, level, message, data string) {
	entry := models.ProtocolEntry{
		Timestamp:      time.Now(),
		TestProtocolID: w.protocolId,
		Source:         source,
		Level:          level,
		Message:        message,
		Data:           data,
		Runtime:        w.getRuntime(),
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
		startTime:  time.Now().UTC(),
	}
}
