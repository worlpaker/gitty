package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCmd(t *testing.T) {
	t.Parallel()
	c := &cobra.Command{Version: "1.0.0"}
	v := versionCmd()
	v.Run(c, nil)
}
