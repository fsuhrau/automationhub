package protocol_test

import (
	"math"
	"testing"

	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
)

func allNaN() models.TestProtocol {
	n := math.NaN()
	return models.TestProtocol{
		AvgFPS: n, AvgMEM: n, AvgCPU: n, AvgVertexCount: n, AvgTriangles: n,
		HistAvgFPS: n, HistAvgMEM: n, HistAvgCPU: n, HistAvgVertexCount: n, HistAvgTriangles: n,
	}
}

// metricsOf lists the five averages and the five history averages.
func metricsOf(p *models.TestProtocol) []float64 {
	return []float64{
		p.AvgFPS, p.AvgMEM, p.AvgCPU, p.AvgVertexCount, p.AvgTriangles,
		p.HistAvgFPS, p.HistAvgMEM, p.HistAvgCPU, p.HistAvgVertexCount, p.HistAvgTriangles,
	}
}

func measured(fps, mem, cpu, vertices, triangles float64) models.TestProtocol {
	return models.TestProtocol{AvgFPS: fps, AvgMEM: mem, AvgCPU: cpu, AvgVertexCount: vertices, AvgTriangles: triangles}
}

func TestCleanZeroesEveryNaNMetric(t *testing.T) {
	p := allNaN()

	protocol.Clean(&p)

	for i, v := range metricsOf(&p) {
		if v != 0 {
			t.Errorf("metric %d is %v, want 0", i, v)
		}
	}
}

func TestCleanKeepsMeasuredValues(t *testing.T) {
	p := measured(60, 512, 30, 1000, 2000)
	p.HistAvgFPS = 55

	protocol.Clean(&p)

	if p.AvgFPS != 60 || p.AvgTriangles != 2000 || p.HistAvgFPS != 55 {
		t.Fatalf("clean changed measured values: %+v", p)
	}
}

func TestAveragesOfCoversEveryMetric(t *testing.T) {
	p := measured(1, 2, 3, 4, 5)

	got := protocol.AveragesOf(&p)

	want := protocol.Averages{FPS: 1, MEM: 2, CPU: 3, VertexCount: 4, Triangles: 5}
	if got != want {
		t.Fatalf("averages %+v, want %+v", got, want)
	}
}

func TestAveragesOfCleansTheProtocol(t *testing.T) {
	p := allNaN()

	got := protocol.AveragesOf(&p)

	if got != (protocol.Averages{}) || p.AvgFPS != 0 {
		t.Fatalf("averages %+v, protocol %+v", got, p)
	}
}

func TestApplyHistoryAveragesEveryMetricOverTheEarlierProtocols(t *testing.T) {
	current := measured(100, 100, 100, 100, 100)
	history := []models.TestProtocol{measured(10, 20, 30, 40, 50), measured(30, 40, 50, 60, 70), allNaN()}

	protocol.ApplyHistory(&current, history)

	want := []float64{100, 100, 100, 100, 100, 40.0 / 3, 20, 80.0 / 3, 100.0 / 3, 40}
	for i, v := range metricsOf(&current) {
		if math.Abs(v-want[i]) > 1e-9 {
			t.Errorf("metric %d = %v, want %v", i, v, want[i])
		}
	}
	if len(current.TestProtocolHistory) != 3 {
		t.Fatalf("history %d protocols, want 3", len(current.TestProtocolHistory))
	}
}

func TestApplyHistoryWithoutEarlierProtocolsLeavesTheAveragesAtZero(t *testing.T) {
	current := measured(100, 100, 100, 100, 100)

	protocol.ApplyHistory(&current, nil)

	if current.HistAvgFPS != 0 || current.HistAvgTriangles != 0 || len(current.TestProtocolHistory) != 0 {
		t.Fatalf("protocol %+v", current)
	}
}
