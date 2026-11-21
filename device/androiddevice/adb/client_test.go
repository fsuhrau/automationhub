package adb

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestClientPrefixesTheSerial(t *testing.T) {
	boom := errors.New("boom")
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell ls", []byte("out"), nil)
	runner.Script("adb -s S1 shell fail", nil, boom)
	c := Client{Runner: runner, Serial: "S1"}

	out, err := c.Run(context.Background(), "shell", "ls")
	if err != nil || string(out) != "out" {
		t.Fatalf("Run = %q, %v", out, err)
	}
	if _, err := c.Run(context.Background(), "shell", "fail"); !errors.Is(err, boom) {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := c.Start(context.Background(), "shell", "ls"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	want := []string{"adb -s S1 shell ls", "adb -s S1 shell fail", "adb -s S1 shell ls"}
	if got := runner.Commands(); !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %v, want %v", got, want)
	}
}

func TestClientStartStreamingWritesTheOutput(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell run-server", []byte("server log\n"), nil)
	var out bytes.Buffer
	c := Client{Runner: runner, Serial: "S1"}

	p, err := c.StartStreaming(context.Background(), &out, "shell", "run-server")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Wait(); err != nil || out.String() != "server log\n" {
		t.Fatalf("Wait = %v, output %q", err, out.String())
	}
	if _, err := c.StartStreaming(context.Background(), &out, "shell", "unscripted"); err == nil {
		t.Fatal("an unscripted command must fail")
	}
}
