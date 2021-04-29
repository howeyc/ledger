package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ledger",
	Short: "Plain text accounting",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var ledgerFilePath string

func init() {
	cobra.OnInitialize(initConfig)

	ledgerFilePath = os.Getenv("LEDGER_FILE")

	rootCmd.PersistentFlags().StringVarP(&ledgerFilePath, "file", "f", ledgerFilePath, "ledger file (default is $LEDGER_FILE)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
