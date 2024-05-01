package main

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_OSExit(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		run()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRun_OSExit")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Args = []string{"test flag", "foo"}

	exit := cmd.Run()
	if e, ok := exit.(*exec.ExitError); ok && !e.Success() {
		return
	}
	assert.Equal(t, 1, exit, "want exit status 1")
}

func Test_Main(t *testing.T) {
	// Discard output during tests.
	defer func(stdout, stderr *os.File) {
		os.Stdout = stdout
		os.Stderr = stderr
	}(os.Stdout, os.Stderr)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)
	os.Stderr = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	main()
}
