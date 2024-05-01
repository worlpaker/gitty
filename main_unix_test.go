//go:build darwin || linux

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
	cmd.Args = []string{"test url", "https://github.com/worlpaker/go-syntax/tree/master/examples"}

	err := cmd.Start()
	assert.Nil(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
		assert.Nil(t, err)
	}()

	exit := cmd.Wait()
	if e, ok := exit.(*exec.ExitError); ok && !e.Success() {
		return
	}
	assert.Equal(t, 1, exit, "want exit status 1")
}
