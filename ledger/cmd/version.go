package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "user"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ledger",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ledger %s\n", version)
		fmt.Printf("- build/commit: %s\n", commit)
		fmt.Printf("- build/date: %s\n", date)
		fmt.Printf("- build/user: %s\n", builtBy)
		fmt.Printf("- os/type: %s\n", runtime.GOOS)
		fmt.Printf("- os/arch: %s\n", runtime.GOARCH)
		fmt.Printf("- go/version: %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
