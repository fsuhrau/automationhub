package hub

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
)

func getTypeOfLog(logType appproto.LogType) string {
	switch logType {
	case appproto.LogType_DeviceLog:
		return "app"
	case appproto.LogType_StepLog:
		return "step"
	case appproto.LogType_StatusLog:
		return "status"
	case appproto.LogType_CheckpointLog:
		return "checkpoint"
	case appproto.LogType_PerformanceLog:
		return "performance"
	}
	return ""
}

// logError writes an app log entry to the device log; it reports whether the
// entry was an exception.
func logError(dev device.Device, log *appproto.LogData) bool {
	switch log.Level {
	case appproto.LogLevel_Exception:
		dev.Exception(getTypeOfLog(log.Type), log.Message)
		return true
	case appproto.LogLevel_Error:
		dev.Error(getTypeOfLog(log.Type), log.Message)
	default:
		dev.Log(getTypeOfLog(log.Type), log.Message)
	}
	return false
}
