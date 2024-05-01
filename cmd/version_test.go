package cmd

import (
	"os"
	"syscall"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCmd(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	c := &cobra.Command{Version: "1.0.0"}
	v := versionCmd()
	v.Run(c, nil)
}
