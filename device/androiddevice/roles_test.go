package androiddevice

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestRoles(t *testing.T) {
	want := []string{"Recorder", "Screenshotter", "FeatureExecutor", "NativeScripter", "AppDataClearer", "InputCapable", "LogReader", "LogFollower", "VideoStreamer"}
	if got := testkit.Roles(newTestDevice(&testkit.FakeRunner{})); !reflect.DeepEqual(got, want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
}
