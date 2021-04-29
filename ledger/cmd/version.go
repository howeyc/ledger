package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ledger",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ledger v0.2.3")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
