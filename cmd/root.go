package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/worlpaker/gitty/gitty"
	"github.com/worlpaker/gitty/gitty/token"
)

// nArgs represents the maximum number of args.
const nArgs = 1

// subCommands adds sub-commands to the root command.
func subCommands(c *cobra.Command) {
	c.AddCommand(versionCmd())
}

// cmdSettings configures settings for the root command.
func cmdSettings(c *cobra.Command) {
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Println("Error:", err)
		return c.Help()
	})
}

// runRoot prepares and returns a function to execute the root command.
func runRoot(ctx context.Context, f *flags, g gitty.Gitty) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		switch {
		case f.auth:
			return g.Auth(ctx)
		case f.check:
			return g.Status(ctx)
		case f.set != "":
			return token.Set(f.set)
		case f.unset:
			return token.Unset()
		case len(args) < nArgs:
			return cmd.Help()
		default:
			return g.Download(ctx, args[0])
		}
	}
}

// Execute executes the root command.
func Execute(ctx context.Context, version string) error {
	g := gitty.New()
	f := &flags{}
	c := &cobra.Command{
		Use:          "gitty [github url]",
		Short:        "Download GitHub File & Directory",
		RunE:         runRoot(ctx, f, g),
		Args:         cobra.MaximumNArgs(nArgs),
		Version:      version,
		SilenceUsage: true,
	}

	// Configurations for root.
	cmdFlags(c, f)
	cmdSettings(c)
	subCommands(c)

	return c.Execute()
}
