//go:build darwin || linux

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
	cmd.Args = []string{"test url", "https://github.com/worlpaker/go-syntax/tree/master/examples"}

	err := cmd.Start()
	require.NoError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
		assert.NoError(t, err)
	}()

	var execErr *exec.ExitError
	if err := cmd.Wait(); errors.As(err, &execErr) && execErr.Success() {
		return
	}

	if execErr != nil {
		assert.Equal(t, 1, execErr.ExitCode(), "want exit status 1")
	}
}
