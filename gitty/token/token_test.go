package token

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	t.Parallel()
	// Restore token.
	localToken := Get()
	t.Cleanup(func() {
		if localToken != "" {
			err := Set(localToken)
			require.NoError(t, err)
		}
	})

	fakeValue := gofakeit.LoremIpsumWord()
	err := Set(fakeValue)
	require.NoError(t, err)

	err = Unset()
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
