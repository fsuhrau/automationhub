package events

var TestCreated = NewBus[TestCreatedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type TestCreatedPayload struct {
	TestID uint        `json:"testId"`
	Test   interface{} `json:"test"`
}
