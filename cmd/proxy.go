package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/tracking"
)

var proxyCmd = &cobra.Command{
	Use:                "proxy <command> [args...]",
	Short:              "Execute command unfiltered but track token usage",
	DisableFlagParsing: true,
	Args:               cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProxy(args)
	},
}

func runProxy(args []string) error {
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr

	out, err := c.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprint(os.Stdout, string(out))
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	fmt.Fprint(os.Stdout, string(out))

	// Track with 0% savings (proxy = no filter)
	cmdName := strings.Join(args, " ")
	tokens := tracking.CountTokens(string(out))
	_ = tracking.Record(cmdName, tokens, tokens)

	return nil
}
