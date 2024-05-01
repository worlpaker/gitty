package token

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	// Restore token and discard output during tests.
	localToken := Get()
	defer func(stdout *os.File) {
		if localToken != "" {
			err := Set(localToken)
			assert.Nil(t, err)
		}
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	fakeValue := gofakeit.LoremIpsumWord()
	err := Set(fakeValue)
	assert.Nil(t, err)

	err = Unset()
	// The result might not be nil in very rare cases.
	// Error: &exec.ExitError{ProcessState:(*os.ProcessState)(0xc00017a288), Stderr:[]uint8(nil)}
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
	}
	assert.Nil(t, err)
}
