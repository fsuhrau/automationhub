package unity

import (
	"context"
	"fmt"
	"strings"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/tester/base"
)

// testCase is one test on one device.
type testCase struct {
	dev  base.DeviceMap
	task action.TestStart
}

// outcome is how a test ended.
type outcome struct {
	passed bool
	errs   []string
}

// RunTest runs one test on the device and writes its protocol.
func (s *strategy) RunTest(ctx context.Context, dev base.DeviceMap, task action.TestStart) {
	r := s.runner
	method := base.MethodName(task.Method)
	prot, err := r.ProtocolWriter.NewProtocol(dev.Model, fmt.Sprintf("%s/%s", task.Class, method))
	if err != nil {
		r.LogError("Unable to create LogWriter for %s: %v", dev.Device.DeviceID(), err)
		return
	}
	dev.Device.SetLogWriter(prot.Writer)
	defer func() {
		dev.Device.SetLogWriter(nil)
		prot.Close()
	}()

	r.LogInfo("Run test '%s/%s' on device '%s'", task.Class, method, dev.Device.DeviceID())
	err = NewExecutor(r.Devices, r.ProtocolWriter).Execute(ctx, dev.Device, task, DefaultTestTimeout)
	s.report(ctx, testCase{dev: dev, task: task}, outcome{passed: prot.Writer.HasPassed(), errs: errorLines(err, prot.Errors())})
}

// report logs how the test ended; a test with errors also gets a screenshot.
func (s *strategy) report(ctx context.Context, tc testCase, out outcome) {
	finished := "finished successful"
	if !out.passed {
		finished = "failed"
	}
	if len(out.errs) == 0 {
		s.runner.LogInfo("Test execution %s ", finished)
		return
	}
	s.captureScreenShot(ctx, tc)
	s.runner.LogError("Test execution %s with errors: %v", finished, strings.Join(out.errs, "\n"))
}

// errorLines lists the error of the execution and the errors of the protocol.
func errorLines(err error, protocolErrs []error) []string {
	var lines []string
	if err != nil {
		lines = append(lines, err.Error())
	}
	for _, e := range protocolErrs {
		lines = append(lines, e.Error())
	}
	return lines
}
