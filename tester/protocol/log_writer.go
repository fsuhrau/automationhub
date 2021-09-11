package protocol

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"
)

type LogWriter struct {
	db         *gorm.DB
	protocolId uint
	errs       []error
}

func (w *LogWriter) write(source, level, message string) {
	entry := models.ProtocolEntry{
		Timestamp:      time.Now(),
		TestProtocolID: w.protocolId,
		Source:         source,
		Level:          level,
		Message:        message,
	}
	w.db.Create(&entry)
}

func (w *LogWriter) Log(source, format string, params ...interface{}) {
	w.write(source, "log", fmt.Sprintf(format, params...))
}

func (w *LogWriter) Error(source, format string, params ...interface{}) {
	msg := fmt.Sprintf(format, params...)
	w.errs = append(w.errs, fmt.Errorf(msg))
	w.write(source,"error", msg)
}

func NewLogWriter(db *gorm.DB, protocolId uint) *LogWriter {
	return &LogWriter{
		db:         db,
		protocolId: protocolId,
	}
}
