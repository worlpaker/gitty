//go:build windows

package script

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	fakeKey := "GITTY_TEST_KEY"
	fakeValue := gofakeit.LoremIpsumWord()

	err := Run(fakeKey, fakeValue)
	assert.Nil(t, err)

	err = Run(fakeKey, "")
	// The result might not be nil in very rare cases.
	// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
	}
	assert.Nil(t, err)
}

func TestScript(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	fakeKey := "GITTY_TEST_KEY"
	fakeValue := gofakeit.LoremIpsumWord()

	tests := []struct {
		name     string
		script   *Script
		expected error
	}{
		{
			name:     "success",
			script:   save(fakeKey, fakeValue),
			expected: nil,
		},
		{
			name:     "error",
			script:   save(fakeKey, fakeValue),
			expected: fmt.Errorf("exec: Stdout already set"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expected == nil {
				t.Cleanup(func() {
					err := delete(fakeKey).execute()
					// The result might not be nil in very rare cases.
					// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
					if err != nil {
						if e, ok := err.(*exec.ExitError); ok && !e.Success() {
							return
						}
					}
					assert.Nil(t, err)
				})
			} else {
				var buf bytes.Buffer
				test.script.cmd.Stdout = &buf
			}
			err := test.script.execute()
			assert.Equal(t, test.expected, err)
		})
	}
}
