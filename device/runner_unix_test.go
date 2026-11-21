//go:build !windows

package device_test

import (
	"context"
	"fmt"
	"syscall"
	"testing"

	"github.com/fsuhrau/automationhub/device"
)

func TestExecRunnerStartStreamingInterruptEndsTheProcessGroup(t *testing.T) {
	var out syncBuffer
	p, err := device.ExecRunner{}.StartStreaming(context.Background(), &out, "sh", "-c", "sh -c 'echo $$; exec sleep 60'; true")
	if err != nil {
		t.Fatal(err)
	}
	var childPID int
	waitUntil(t, "the child pid", func() bool {
		_, scanErr := fmt.Sscan(out.String(), &childPID)
		return scanErr == nil
	})
	interrupter, ok := p.(device.Interrupter)
	if !ok {
		t.Fatal("the process is not a device.Interrupter")
	}
	if err := interrupter.Interrupt(); err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err == nil {
		t.Error("an interrupted process should fail Wait")
	}
	waitUntil(t, "the child to end", func() bool { return syscall.Kill(childPID, 0) != nil })
}
