package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sendCtrlBreak sends a Ctrl+Break signal to the specified pid.
// Equivalent to "syscall.Kill" in linux.
//
// Example found on: https://github.com/golang/go/blob/master/src/os/signal/signal_windows_test.go
func sendCtrlBreak(t *testing.T, pid int) {
	t.Helper()
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		t.Fatalf("LoadDLL: %v\n", e)
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		t.Fatalf("FindProc: %v\n", e)
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		t.Fatalf("GenerateConsoleCtrlEvent: %v\n", e)
	}
}

func TestMain(t *testing.T) {
	t.Parallel()
	if os.Getenv("BE_CRASHER") == "1" {
		main()
		return
	}

	name := os.Args[0]
	cmd := exec.Command(name, "-test.run=TestMain")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Args = []string{"test url", "https://github.com/worlpaker/go-syntax/tree/master/examples"}

	err := cmd.Start()
	require.NoError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		sendCtrlBreak(t, cmd.Process.Pid)
	}()

	var execErr *exec.ExitError
	if err := cmd.Wait(); !errors.As(err, &execErr) {
		return
	}

	if execErr != nil {
		assert.False(t, execErr.Success(), "want failure exit, most likely 1")
	}
}
