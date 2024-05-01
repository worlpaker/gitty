package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/worlpaker/gitty/gitty"
	"github.com/worlpaker/gitty/gitty/token"
)

func fakeNewGitty() gitty.Gitty {
	return &mock{}
}

type mock struct{}

func (m *mock) Status(ctx context.Context) error {
	return nil
}
func (m *mock) Download(ctx context.Context, url string) error {
	return nil
}
func (m *mock) Auth(ctx context.Context) error {
	return nil
}

func TestSubCommands(t *testing.T) {
	c := &cobra.Command{}
	subCommands(c)
	assert.True(t, c.HasSubCommands())
}

func TestCmdSettings(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	c := &cobra.Command{}
	cmdSettings(c)
	f := c.FlagErrorFunc()
	err := f(c, fmt.Errorf("test error"))
	assert.Nil(t, err)
}

func TestExecute(t *testing.T) {
	// Set fake args and discard output during tests.
	oldArgs := os.Args
	defer func(stdout *os.File) {
		os.Args = oldArgs
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)
	os.Args = []string{"test flag", "--check"}

	err := Execute(context.Background(), "1.0.0")
	assert.Nil(t, err)
}

func TestRunRoot(t *testing.T) {
	// Restore token and discard output during tests.
	localToken := token.Get()
	defer func(stdout *os.File) {
		if localToken != "" {
			err := token.Set(localToken)
			assert.Nil(t, err)
		}
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

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
			c := &cobra.Command{}
			g := fakeNewGitty()
			runFunc := runRoot(context.Background(), &test.flags, g)
			err := runFunc(c, test.args)
			// Similar issue in the token_test.go.
			// The result might not be nil in very rare cases.
			// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
			if err != nil {
				if e, ok := err.(*exec.ExitError); ok && !e.Success() {
					return
				}
			}
			assert.Nil(t, err)
		})
	}
}
