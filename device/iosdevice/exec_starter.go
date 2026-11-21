package iosdevice

import (
	"fmt"
	"io"
	"os/exec"
)

// execStarter is the ProcessStarter on top of os/exec. The command runs in its
// own process group, so stopping it also ends the children xcodebuild starts.
type execStarter struct{}

func (execStarter) StartStreaming(out io.Writer, name string, args ...string) (RunnerProcess, error) {
	cmd := exec.Command(name, args...)
	setProcessGroup(cmd)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	return &execProcess{cmd: cmd}, nil
}

type execProcess struct{ cmd *exec.Cmd }

func (p *execProcess) Wait() error      { return p.cmd.Wait() }
func (p *execProcess) Interrupt() error { interruptProcessGroup(p.cmd); return nil }
func (p *execProcess) Kill() error      { killProcessGroup(p.cmd); return nil }
