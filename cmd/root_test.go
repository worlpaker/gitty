package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/worlpaker/gitty/gitty"
	"github.com/worlpaker/gitty/gitty/token"
)

func fakeNewGitty() gitty.Gitty {
	return &mock{}
}

type mock struct{}

func (m *mock) Status(_ context.Context) error {
	return nil
}

func (m *mock) Download(_ context.Context, _ string) error {
	return nil
}

func (m *mock) Auth(_ context.Context) error {
	return nil
}

func TestSubCommands(t *testing.T) {
	t.Parallel()
	c := &cobra.Command{}
	subCommands(c)
	assert.True(t, c.HasSubCommands())
}

func TestCmdSettings(t *testing.T) {
	t.Parallel()
	c := &cobra.Command{}
	cmdSettings(c)
	f := c.FlagErrorFunc()
	err := f(c, errors.New("test error"))
	require.NoError(t, err)
}

func TestExecute(t *testing.T) {
	t.Parallel()
	// Set fake args during tests.
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})
	os.Args = []string{"test flag", "--check"}

	err := Execute(context.Background(), "1.0.0")
	require.NoError(t, err)
}

func TestRunRoot(t *testing.T) {
	t.Parallel()
	// Restore token.
	localToken := token.Get()
	t.Cleanup(func() {
		if localToken != "" {
			err := token.Set(localToken)
			require.NoError(t, err)
		}
	})

	tests := []struct {
		name  string
		flags flags
		args  []string
	}{
		{
			name:  "auth flag",
			flags: flags{auth: true},
		},
		{
			name:  "check flag",
			flags: flags{check: true},
		},
		{
			name:  "set flag",
			flags: flags{set: "test_token"},
		},
		{
			name:  "unset flag",
			flags: flags{unset: true},
		},
		{
			name:  "insufficient arguments",
			flags: flags{},
			args:  []string{},
		},
		{
			name:  "default case",
			flags: flags{},
			args:  []string{"arg1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := &cobra.Command{}
			g := fakeNewGitty()
			runFunc := runRoot(context.Background(), &test.flags, g)
			err := runFunc(c, test.args)
			// Similar issue in the token_test.go.
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
	}
}
