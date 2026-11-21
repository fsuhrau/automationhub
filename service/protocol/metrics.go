package protocol

import (
	"math"

	"github.com/fsuhrau/automationhub/storage/models"
)

// HistoryLimit is how many earlier runs the averages and histories of a run
// look at.
const HistoryLimit = 20

// Averages are the performance averages of a protocol, as the baselines
// list them.
type Averages struct {
	FPS         float64 `json:"avgFps"`
	MEM         float64 `json:"avgMem"`
	CPU         float64 `json:"avgCpu"`
	VertexCount float64 `json:"avgVertexCount"`
	Triangles   float64 `json:"avgTriangles"`
}

// metric says where one performance metric lives: its average on a
// protocol, the average over the earlier protocols, and in Averages.
type metric struct {
	avg  func(*models.TestProtocol) *float64
	hist func(*models.TestProtocol) *float64
	of   func(*Averages) *float64
}

// metrics is the one list of performance metrics; cleaning, history
// averages and baselines all run over it.
func metrics() []metric {
	return []metric{
		{
			avg:  func(p *models.TestProtocol) *float64 { return &p.AvgFPS },
			hist: func(p *models.TestProtocol) *float64 { return &p.HistAvgFPS },
			of:   func(a *Averages) *float64 { return &a.FPS },
		},
		{
			avg:  func(p *models.TestProtocol) *float64 { return &p.AvgMEM },
			hist: func(p *models.TestProtocol) *float64 { return &p.HistAvgMEM },
			of:   func(a *Averages) *float64 { return &a.MEM },
		},
		{
			avg:  func(p *models.TestProtocol) *float64 { return &p.AvgCPU },
			hist: func(p *models.TestProtocol) *float64 { return &p.HistAvgCPU },
			of:   func(a *Averages) *float64 { return &a.CPU },
		},
		{
			avg:  func(p *models.TestProtocol) *float64 { return &p.AvgVertexCount },
			hist: func(p *models.TestProtocol) *float64 { return &p.HistAvgVertexCount },
			of:   func(a *Averages) *float64 { return &a.VertexCount },
		},
		{
			avg:  func(p *models.TestProtocol) *float64 { return &p.AvgTriangles },
			hist: func(p *models.TestProtocol) *float64 { return &p.HistAvgTriangles },
			of:   func(a *Averages) *float64 { return &a.Triangles },
		},
	}
}

// Clean makes a protocol serializable: NaN averages, which JSON can't
// carry, become 0.
func Clean(p *models.TestProtocol) {
	for _, m := range metrics() {
		zeroNaN(m.avg(p))
		zeroNaN(m.hist(p))
	}
}

func zeroNaN(v *float64) {
	if math.IsNaN(*v) {
		*v = 0
	}
}

// AveragesOf cleans the protocol and returns its performance averages.
func AveragesOf(p *models.TestProtocol) Averages {
	Clean(p)
	var a Averages
	for _, m := range metrics() {
		*m.of(&a) = *m.avg(p)
	}
	return a
}

// ApplyHistory sets the averages of the earlier protocols on current and
// attaches them as its history.
func ApplyHistory(current *models.TestProtocol, history []models.TestProtocol) {
	for i := range history {
		Clean(&history[i])
		for _, m := range metrics() {
			*m.hist(current) += *m.avg(&history[i])
		}
	}
	if n := float64(len(history)); n > 0 {
		for _, m := range metrics() {
			*m.hist(current) /= n
		}
	}
	current.TestProtocolHistory = history
}
