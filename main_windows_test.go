package main

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// sendCtrlBreak simulates sending a Ctrl+Break signal to the specified process ID (pid).
// This function is used to trigger a cancellation signal in the test environment.
// Equivalent to `syscall.Kill` in linux.
func sendCtrlBreak(t *testing.T, pid int) {
	// Example found on: https://github.com/golang/go/blob/master/src/os/signal/signal_windows_test.go
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

func TestRun_Cancel(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		run()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRun_Cancel")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Args = []string{"test url", "https://github.com/worlpaker/go-syntax/tree/master/examples"}

	err := cmd.Start()
	assert.Nil(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		sendCtrlBreak(t, cmd.Process.Pid)
	}()

	exit := cmd.Wait()
	if e, ok := exit.(*exec.ExitError); ok && !e.Success() {
		return
	}
	assert.Equal(t, 1, exit, "want exit status 1")
}
