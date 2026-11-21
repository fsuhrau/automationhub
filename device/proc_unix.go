//go:build !windows

package device

import (
	"os"
	"os/exec"
	"syscall"
)

// setProcessGroup gives the command its own process group.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func leadsGroup(cmd *exec.Cmd) bool {
	return cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid
}

// killProcess kills the process, with its group when it leads one.
func killProcess(cmd *exec.Cmd) error {
	if leadsGroup(cmd) {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return cmd.Process.Kill()
}

// interruptProcess sends SIGINT to the process, or to its group when it leads
// one.
func interruptProcess(cmd *exec.Cmd) error {
	if leadsGroup(cmd) {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}
	return cmd.Process.Signal(os.Interrupt)
}
