package cmd

import (
	"os"

	"github.com/spf13/cobra"
	pscmds "ptk/cmd/ps"
)

var rootCmd = &cobra.Command{
	Use:     "ptk",
	Short:   "PowerShell Token Killer — compress PS cmdlet output for AI agents",
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(rewriteCmd)
	rootCmd.AddCommand(gainCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(pscmds.GciCmd)
	rootCmd.AddCommand(pscmds.SlsCmd)
	rootCmd.AddCommand(pscmds.GpsCmd)
	rootCmd.AddCommand(pscmds.GsvCmd)
	rootCmd.AddCommand(pscmds.MeasureCmd)
	rootCmd.AddCommand(pscmds.HistoryCmd)
	rootCmd.AddCommand(pscmds.HelpPSCmd)
}
