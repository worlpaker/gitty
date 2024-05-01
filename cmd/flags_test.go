package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCmdFlags(t *testing.T) {
	c := &cobra.Command{}
	f := &flags{}

	cmdFlags(c, f)

	_, err := c.Flags().GetString("set")
	assert.Nil(t, err)
	_, err = c.Flags().GetBool("auth")
	assert.Nil(t, err)
	_, err = c.Flags().GetBool("check")
	assert.Nil(t, err)
	_, err = c.Flags().GetBool("unset")
	assert.Nil(t, err)
}
