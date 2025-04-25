package models

type ProtocolPerformanceEntry struct {
	Model
	TestProtocolID uint    `json:"testProtocolId"`
	Checkpoint     string  `json:"checkpoint"`
	FPS            float64 `sql:"type:decimal(10,2);" json:"fps"`
	MEM            float64 `sql:"type:decimal(10,2);" json:"mem"`
	CPU            float64 `sql:"type:decimal(10,2);" json:"cpu"`
	Other          string  `json:"other"`
	Runtime        float64 `sql:"type:decimal(10,2);" json:"runtime"`
	VertexCount    float64 `sql:"type:decimal(10,2);" json:"vertexCount"`
	Triangles      float64 `sql:"type:decimal(10,2);" json:"triangles"`
}
