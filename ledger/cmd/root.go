package cmd

import (
	"os"

	cc "github.com/ivanpirog/coloredcobra"
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
	cc.Init(&cc.Config{
		RootCmd:         rootCmd,
		Headings:        cc.Magenta + cc.Bold + cc.Underline,
		Commands:        cc.Red + cc.Bold,
		Aliases:         cc.Bold + cc.Italic,
		CmdShortDescr:   cc.White,
		Example:         cc.Italic,
		ExecName:        cc.Bold,
		Flags:           cc.Yellow + cc.Bold,
		FlagsDescr:      cc.White,
		FlagsDataType:   cc.Italic + cc.Blue,
		NoExtraNewlines: true,
	})
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
