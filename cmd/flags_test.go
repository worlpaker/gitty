package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCmdFlags(t *testing.T) {
	t.Parallel()
	c := &cobra.Command{}
	f := &flags{}

	cmdFlags(c, f)

	_, err := c.Flags().GetString("set")
	require.NoError(t, err)
	_, err = c.Flags().GetBool("auth")
	require.NoError(t, err)
	_, err = c.Flags().GetBool("check")
	require.NoError(t, err)
	_, err = c.Flags().GetBool("unset")
	require.NoError(t, err)
}
