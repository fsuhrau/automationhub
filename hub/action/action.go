package action

type ActionType int32

const (
	ActionType_Custom                ActionType = 0
	ActionType_GetSceneGraph         ActionType = 1
	ActionType_ElementIsDisplayed    ActionType = 2
	ActionType_ElementSetValue       ActionType = 3
	ActionType_ElementGetValue       ActionType = 4
	ActionType_Move                  ActionType = 5
	ActionType_TouchDown             ActionType = 6
	ActionType_TouchMove             ActionType = 7
	ActionType_TouchUp               ActionType = 8
	ActionType_DragAndDrop           ActionType = 9
	ActionType_LongTouch             ActionType = 10
	ActionType_ElementTouch          ActionType = 11
	ActionType_GetScreenshot         ActionType = 12
	ActionType_GetTests              ActionType = 13
	ActionType_ExecuteTest           ActionType = 14
	ActionType_ExecutionResult       ActionType = 15
	ActionType_Log                   ActionType = 16
	ActionType_UnityReset            ActionType = 17
	ActionType_Performance           ActionType = 18
	ActionType_NativeScript          ActionType = 19
	ActionType_ExecuteMethodStart    ActionType = 20
	ActionType_ExecuteMethodFinished ActionType = 21
)

type LogType int32

const (
	LogType_DeviceLog      LogType = 0
	LogType_StepLog        LogType = 1
	LogType_StatusLog      LogType = 2
	LogType_CheckpointLog  LogType = 3
	LogType_PerformanceLog LogType = 4
)

type LogLevel int32

const (
	LogLevel_Debug     LogLevel = 0
	LogLevel_Info      LogLevel = 1
	LogLevel_Warning   LogLevel = 2
	LogLevel_Error     LogLevel = 3
	LogLevel_Exception LogLevel = 4
)

type ContentType int32

const (
	ContentType_Flatbuffer ContentType = 0
	ContentType_Json       ContentType = 1
	ContentType_Xml        ContentType = 2
)

type AppType int32

const (
	AppType_Cocos AppType = 0
	AppType_Unity AppType = 1
)

type SetAttr struct {
	Id   string `json:"id"`
	Attr string `json:"attr"`
	Val  string `json:"val"`
}

type GetAttr struct {
	Id   string `json:"id"`
	Attr string `json:"attr"`
}

type MoveOffset struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type MoveElement struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Touch struct {
	Id      string `json:"id"`
	Xoffset int64  `json:"xoffset"`
	Yoffset int64  `json:"yoffset"`
}
type RequestData struct {
	Id          *string      `json:"id,omitempty"`
	Data        *[]byte      `json:"data,omitempty"`
	SetAttr     *SetAttr     `json:"setAttr,omitempty"`
	GetAttr     *GetAttr     `json:"getAttr,omitempty"`
	MoveOffset  *MoveOffset  `json:"moveOffset,omitempty"`
	Touch       *Touch       `json:"touch,omitempty"`
	MoveElement *MoveElement `json:"moveElement,omitempty"`
	Test        *Test        `json:"test,omitempty"`
}

type Request struct {
	ActionID   string      `json:"actionID"`
	ActionType ActionType  `json:"actionType"`
	Payload    RequestData `json:"payload"`
}

type Connect struct {
	CustomerId string  `json:"customerId"`
	AppID      string  `json:"appID"`
	AppType    AppType `json:"appType"`
	DeviceID   string  `json:"deviceID"`
	SessionID  string  `json:"sessionID"`
	Version    string  `json:"version"`
	AppVersion string  `json:"appVersion"`
}

type Screenshot struct {
	Sceengraph  []byte      `json:"sceengraph"`
	Screenshot  []byte      `json:"screenshot"`
	ContentType ContentType `json:"contentType"`
}

type Test struct {
	Assembly   string            `json:"assembly"`
	Class      string            `json:"class"`
	Method     string            `json:"method"`
	Parameter  map[string]string `json:"parameter"`
	Categories []string          `json:"categories"`
}

type Tests struct {
	Tests []Test `json:"tests"`
}

type LogData struct {
	Type    LogType  `json:"type"`
	Level   LogLevel `json:"level"`
	Message string   `json:"message"`
}

type TestDetails struct {
	Test       string   `json:"test"`
	Timeout    int64    `json:"timeout"`
	Categories []string `json:"categories"`
}

type PerformanceData struct {
	Checkpoint  string  `json:"checkpoint"`
	CPU         float64 `json:"CPU"`
	Memory      float64 `json:"memory"`
	FPS         float64 `json:"FPS"`
	VertexCount float64 `json:"VertexCount"`
	Triangles   float64 `json:"Triangles"`
}

type ResponseData struct {
	Visible         *bool            `json:"visible,omitempty"`
	Data            *[]byte          `json:"data,omitempty"`
	Value           *string          `json:"value,omitempty"`
	Screenshot      *Screenshot      `json:"screenshot,omitempty"`
	Connect         *Connect         `json:"connect,omitempty"`
	Tests           *Tests           `json:"tests,omitempty"`
	LogData         *LogData         `json:"logData,omitempty"`
	TestDetails     *TestDetails     `json:"testDetails,omitempty"`
	PerformanceData *PerformanceData `json:"performanceData,omitempty"`
}

type Response struct {
	ActionID   string       `json:"actionID"`
	ActionType ActionType   `json:"actionType"`
	Success    bool         `json:"success"`
	Payload    ResponseData `json:"payload"`
}
