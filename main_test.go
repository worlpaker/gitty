package main

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Parallel()
	// Discard output during test.
	defer func(stdout, stderr *os.File) {
		os.Stdout = stdout
		os.Stderr = stderr
	}(os.Stdout, os.Stderr)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)
	os.Stderr = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	exitCode := run()
	assert.Equal(t, 0, exitCode, "want exit status 0")
}
