package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd creates a command to print the version number of gitty.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of gitty",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Println("gitty version", cmd.Root().Version)
		},
	}
}
