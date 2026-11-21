package androiddevice

import (
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestDumpsysFlags(t *testing.T) {
	const awake = "adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker"
	const locked = "adb -s S1 shell dumpsys window | grep mDreamingLockscreen"
	tests := []struct {
		name    string
		line    string
		out     string
		err     error
		read    func(*Device) (bool, error)
		want    bool
		wantErr bool
	}{
		{"awake", awake, "  mHoldingDisplaySuspendBlocker=true", nil, (*Device).IsAwake, true, false},
		{"asleep", awake, "  mHoldingDisplaySuspendBlocker=false", nil, (*Device).IsAwake, false, false},
		{"awake read fails", awake, "", errBoom, (*Device).IsAwake, true, true},
		{"awake unreadable", awake, "nothing", nil, (*Device).IsAwake, true, true},
		{"locked", locked, "mDreamingLockscreen=true", nil, (*Device).IsPinLocked, true, false},
		{"unlocked", locked, "mDreamingLockscreen=false", nil, (*Device).IsPinLocked, false, false},
		{"lock read fails", locked, "", errBoom, (*Device).IsPinLocked, true, true},
		{"lock unreadable", locked, "nothing", nil, (*Device).IsPinLocked, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &testkit.FakeRunner{}
			runner.Script(tt.line, []byte(tt.out), tt.err)
			got, err := tt.read(newTestDevice(runner))
			if got != tt.want || (err != nil) != tt.wantErr {
				t.Fatalf("got %v, err %v", got, err)
			}
		})
	}
}

func TestUnlockScreenFailsWhenTheLockStateIsUnknown(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell dumpsys power | grep mHoldingDisplaySuspendBlocker", []byte("mHoldingDisplaySuspendBlocker=true"), nil)
	if err := newTestDevice(runner).unlockScreen(nil); err == nil {
		t.Fatal("expected an error")
	}
}

func TestEnterPinIgnoresAnEmptyPin(t *testing.T) {
	runner := &testkit.FakeRunner{}
	newTestDevice(runner).enterPin("")
	assertCommands(t, runner)
}
