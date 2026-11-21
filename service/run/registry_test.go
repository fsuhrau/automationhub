package run_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

type stubRunner struct{}

func (*stubRunner) Initialize(models.Test, map[string]string) error { return nil }
func (*stubRunner) Run([]models.Device, *models.AppBinary, string) (*models.TestRun, error) {
	return nil, nil
}
func (*stubRunner) Cancel(string) error   { return nil }
func (*stubRunner) LockedDevices() []uint { return nil }

func TestRegistryKeepsTheRunnersByRun(t *testing.T) {
	registry := run.NewRegistry()
	runner := &stubRunner{}

	registry.Add(7, runner)
	got, ok := registry.Get(7)
	if !ok || got != runner {
		t.Fatalf("got %v, %v", got, ok)
	}
	if _, ok := registry.Get(8); ok {
		t.Fatal("another run has a runner")
	}
	registry.Remove(7)
	if _, ok := registry.Get(7); ok {
		t.Fatal("the removed runner is still there")
	}
}
