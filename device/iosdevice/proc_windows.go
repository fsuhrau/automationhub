package iosdevice

import "os/exec"

// xcodebuild only exists on macOS; these keep Windows builds working.

func setProcessGroup(*exec.Cmd) {}

func interruptProcessGroup(cmd *exec.Cmd) { _ = cmd.Process.Kill() }

func killProcessGroup(cmd *exec.Cmd) { _ = cmd.Process.Kill() }
