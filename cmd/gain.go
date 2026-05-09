package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/tracking"
)

var gainCmd = &cobra.Command{
	Use:   "gain",
	Short: "Show token savings analytics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGain()
	},
}

func runGain() error {
	stats, err := tracking.Gain()
	if err != nil {
		return fmt.Errorf("gain: %w", err)
	}
	if len(stats) == 0 {
		fmt.Println("No data yet. Run some ptk commands first.")
		return nil
	}

	// Header
	fmt.Printf("%-20s %6s %8s %8s %8s\n", "Command", "Calls", "In", "Out", "Savings")
	fmt.Println(strings.Repeat("-", 58))

	var totalIn, totalOut int
	for _, s := range stats {
		fmt.Printf("%-20s %6d %8s %8s %7.1f%%\n",
			s.Command,
			s.Calls,
			fmtTokens(s.InputTokens),
			fmtTokens(s.OutputTokens),
			s.SavingsPct,
		)
		totalIn += s.InputTokens
		totalOut += s.OutputTokens
	}

	fmt.Println(strings.Repeat("-", 58))
	var totalPct float64
	if totalIn > 0 {
		totalPct = (1 - float64(totalOut)/float64(totalIn)) * 100
	}
	fmt.Printf("%-20s %6s %8s %8s %7.1f%%\n",
		"TOTAL", "",
		fmtTokens(totalIn),
		fmtTokens(totalOut),
		totalPct,
	)
	return nil
}

func fmtTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
