package hub

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

func getTypeOfLog(logType action.LogType) string {
	switch logType {
	case action.LogType_DeviceLog:
		return "app"
	case action.LogType_StepLog:
		return "step"
	case action.LogType_StatusLog:
		return "status"
	case action.LogType_CheckpointLog:
		return "checkpoint"
	case action.LogType_PerformanceLog:
		return "performance"
	}
	return ""
}

// logError writes an app log entry to the device log; it reports whether the
// entry was an exception.
func logError(dev device.Device, log *action.LogData) bool {
	switch log.Level {
	case action.LogLevel_Exception:
		dev.Exception(getTypeOfLog(log.Type), log.Message)
		return true
	case action.LogLevel_Error:
		dev.Error(getTypeOfLog(log.Type), log.Message)
	default:
		dev.Log(getTypeOfLog(log.Type), log.Message)
	}
	return false
}
