package cmd

import "github.com/spf13/cobra"

var GenCommand = &cobra.Command{
	Use:     "gen",
	Short:   "gen",
	Version: "v0.1.1",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
