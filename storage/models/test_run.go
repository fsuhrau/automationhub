package models

type TestRun struct {
	Model
	TestID       uint                  `json:"testId"`
	Test         *Test                 `json:"test"`
	AppBinaryID  uint                  `json:"appBinaryId"`
	AppBinary    *AppBinary            `json:"appBinary"`
	StartURL     string                `json:"startUrl"`
	SessionID    string                `json:"sessionId"`
	Parameter    string                `json:"parameter"`
	Protocols    []TestProtocol        `json:"protocols"`
	Log          []TestRunLogEntry     `json:"log"`
	DeviceStatus []TestRunDeviceStatus `json:"deviceStatus"`
}
