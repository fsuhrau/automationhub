package protocol

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"
)

type LogWriter struct {
	db         *gorm.DB
	source     string
	protocolId uint
	errs       []error
}

func (w *LogWriter) write(level, message string) {
	entry := models.ProtocolEntry{
		Timestamp:      time.Now(),
		TestProtocolID: w.protocolId,
		Source:         w.source,
		Level:          level,
		Message:        message,
	}
	w.db.Create(&entry)
}

func (w *LogWriter) Log(format string, params ...interface{}) {
	w.write("log", fmt.Sprintf(format, params...))
}

func (w *LogWriter) Error(format string, params ...interface{}) {
	msg := fmt.Sprintf(format, params...)
	w.errs = append(w.errs, fmt.Errorf(msg))
	w.write("error", msg)
}

func NewLogWriter(db *gorm.DB, source string, protocolId uint) *LogWriter {
	return &LogWriter{
		db:         db,
		source:     source,
		protocolId: protocolId,
	}
}
