package action

import "github.com/fsuhrau/automationhub/device/appproto"

// The app protocol types live in device/appproto so the device package
// needn't import hub; these aliases keep the old names working.

type (
	ActionType      = appproto.ActionType
	LogType         = appproto.LogType
	LogLevel        = appproto.LogLevel
	ContentType     = appproto.ContentType
	AppType         = appproto.AppType
	SetAttr         = appproto.SetAttr
	GetAttr         = appproto.GetAttr
	MoveOffset      = appproto.MoveOffset
	MoveElement     = appproto.MoveElement
	Touch           = appproto.Touch
	RequestData     = appproto.RequestData
	Request         = appproto.Request
	Connect         = appproto.Connect
	Screenshot      = appproto.Screenshot
	Test            = appproto.Test
	Tests           = appproto.Tests
	LogData         = appproto.LogData
	TestDetails     = appproto.TestDetails
	PerformanceData = appproto.PerformanceData
	RunKey          = appproto.RunKey
	ResponseData    = appproto.ResponseData
	Response        = appproto.Response
	ActionHandler   = appproto.ActionHandler
)

const (
	ActionType_Custom                = appproto.ActionType_Custom
	ActionType_GetSceneGraph         = appproto.ActionType_GetSceneGraph
	ActionType_ElementIsDisplayed    = appproto.ActionType_ElementIsDisplayed
	ActionType_ElementSetValue       = appproto.ActionType_ElementSetValue
	ActionType_ElementGetValue       = appproto.ActionType_ElementGetValue
	ActionType_Move                  = appproto.ActionType_Move
	ActionType_TouchDown             = appproto.ActionType_TouchDown
	ActionType_TouchMove             = appproto.ActionType_TouchMove
	ActionType_TouchUp               = appproto.ActionType_TouchUp
	ActionType_DragAndDrop           = appproto.ActionType_DragAndDrop
	ActionType_LongTouch             = appproto.ActionType_LongTouch
	ActionType_ElementTouch          = appproto.ActionType_ElementTouch
	ActionType_GetScreenshot         = appproto.ActionType_GetScreenshot
	ActionType_GetTests              = appproto.ActionType_GetTests
	ActionType_ExecuteTest           = appproto.ActionType_ExecuteTest
	ActionType_ExecutionResult       = appproto.ActionType_ExecutionResult
	ActionType_Log                   = appproto.ActionType_Log
	ActionType_UnityReset            = appproto.ActionType_UnityReset
	ActionType_Performance           = appproto.ActionType_Performance
	ActionType_NativeScript          = appproto.ActionType_NativeScript
	ActionType_ExecuteMethodStart    = appproto.ActionType_ExecuteMethodStart
	ActionType_ExecuteMethodFinished = appproto.ActionType_ExecuteMethodFinished
	ActionType_RunKey                = appproto.ActionType_RunKey
	LogType_DeviceLog                = appproto.LogType_DeviceLog
	LogType_StepLog                  = appproto.LogType_StepLog
	LogType_StatusLog                = appproto.LogType_StatusLog
	LogType_CheckpointLog            = appproto.LogType_CheckpointLog
	LogType_PerformanceLog           = appproto.LogType_PerformanceLog
	LogLevel_Debug                   = appproto.LogLevel_Debug
	LogLevel_Info                    = appproto.LogLevel_Info
	LogLevel_Warning                 = appproto.LogLevel_Warning
	LogLevel_Error                   = appproto.LogLevel_Error
	LogLevel_Exception               = appproto.LogLevel_Exception
	ContentType_Flatbuffer           = appproto.ContentType_Flatbuffer
	ContentType_Json                 = appproto.ContentType_Json
	ContentType_Xml                  = appproto.ContentType_Xml
	AppType_Cocos                    = appproto.AppType_Cocos
	AppType_Unity                    = appproto.AppType_Unity
)
