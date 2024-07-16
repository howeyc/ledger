package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ledger",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("ledger %s\n", version)
		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Print(bi)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
