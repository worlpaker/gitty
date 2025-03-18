//go:build windows

package script

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()
	fakeKey := "GITTY_TEST_KEY"
	fakeValue := gofakeit.LoremIpsumWord()

	err := Run(fakeKey, fakeValue)
	require.NoError(t, err)

	err = Run(fakeKey, "")
	// The result might not be nil in very rare cases.
	// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) && !execErr.Success() {
			return
		}
	}
	require.NoError(t, err)
}

func TestScript(t *testing.T) {
	t.Parallel()
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
			expected: errors.New("exec: Stdout already set"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.expected == nil {
				t.Cleanup(func() {
					err := del(fakeKey).execute()
					// The result might not be nil in very rare cases.
					// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
					if err != nil {
						var execErr *exec.ExitError
						if errors.As(err, &execErr) && !execErr.Success() {
							return
						}
					}
					require.NoError(t, err)
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
