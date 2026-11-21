package device

import "os/exec"

// Windows has no process groups here: a process is stopped alone.

func setProcessGroup(*exec.Cmd) {}

func killProcess(cmd *exec.Cmd) error { return cmd.Process.Kill() }

func interruptProcess(cmd *exec.Cmd) error { return cmd.Process.Kill() }
