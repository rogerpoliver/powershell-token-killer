package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ptk/internal/config"
	"ptk/internal/permissions"
	"ptk/internal/powershell"
	"ptk/internal/registry"
)

var rewriteCmd = &cobra.Command{
	Use:   "rewrite <command>",
	Short: "Rewrite a command to its ptk equivalent (hook protocol)",
	Long: `Translates a raw shell command into its ptk-optimized equivalent.

Exit code protocol (read by ptk-rewrite.ps1):
  0 + stdout  Rewrite found, no deny/ask rule → auto-allow
  1           No ptk equivalent → pass through unchanged
  2           Deny rule matched → let agent handle natively
  3 + stdout  Ask rule matched → rewrite but prompt user`,
	Args:               cobra.ExactArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRewrite(args[0])
	},
}

func runRewrite(rawCmd string) error {
	cfg := config.Load()

	// Check deny/ask BEFORE rewriting — covers both PTK and non-PTK commands.
	verdict := permissions.Check(rawCmd, cfg.Hooks)
	if verdict == permissions.VerdictDeny {
		os.Exit(2)
	}

	// Normalize PS alias → canonical cmdlet → ptk subcommand
	normalized := powershell.Normalize(rawCmd)
	rewritten, ok := registry.Rewrite(normalized)
	if !ok {
		// No PTK equivalent — hook should pass through unchanged.
		os.Exit(1)
	}

	fmt.Print(rewritten)

	switch verdict {
	case permissions.VerdictAllow:
		// exit 0 (normal return)
	case permissions.VerdictAsk, permissions.VerdictDefault:
		os.Exit(3)
	}
	return nil
}
