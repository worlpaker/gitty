package cmd

import (
	"github.com/spf13/cobra"
)

// flags represents the flags for the root command.
type flags struct {
	set   string
	auth  bool
	check bool
	unset bool
}

// cmdFlags configures command flags for the root command.
func cmdFlags(c *cobra.Command, f *flags) {
	c.Flags().StringVarP(&f.set, "set", "s", "", "set github token into os environment variable (e.g., gitty -s=your_github_token)")
	c.Flags().BoolVarP(&f.auth, "auth", "a", false, "print authenticated username")
	c.Flags().BoolVarP(&f.check, "check", "c", false, "check client status and remaining rate limit")
	c.Flags().BoolVarP(&f.unset, "unset", "u", false, "unset github token from os environment variable")
}
