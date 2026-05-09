package ps

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var HistoryCmd = &cobra.Command{
	Use:                "history [count]",
	Short:              "Filter Get-History output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory(args)
	},
}

func runHistory(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk history: PowerShell not found on PATH")
	}

	count := "50"
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			count = a
		}
	}

	cmdlet := fmt.Sprintf("Get-History -Count %s", count)
	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil || strings.TrimSpace(raw) == "" {
		fmt.Println("(no history)")
		return nil
	}

	compressed := FilterHistory(raw)
	fmt.Print(compressed)

	inTok := tracking.CountTokens(raw)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("history", inTok, outTok)
	return nil
}

// FilterHistory strips the PS table headers and formats as "id  command".
//
// PS format:
//   Id CommandLine
//   -- -----------
//    1 Get-ChildItem
//    2 cargo build --release
func FilterHistory(raw string) string {
	lines := strings.Split(strings.TrimRight(raw, "\r\n"), "\n")
	var sb strings.Builder
	inData := false

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)

		// Detect and skip header row
		if strings.Contains(trimmed, "Id") && strings.Contains(trimmed, "CommandLine") {
			inData = true
			continue
		}
		// Skip separator line
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		if !inData || trimmed == "" {
			continue
		}

		// Format: leading spaces + id + spaces + command
		// Normalize to "id  command"
		fields := strings.SplitN(trimmed, " ", 2)
		if len(fields) < 2 {
			sb.WriteString(trimmed)
			sb.WriteByte('\n')
			continue
		}
		id := strings.TrimSpace(fields[0])
		cmd := strings.TrimSpace(fields[1])
		sb.WriteString(id)
		sb.WriteString("  ")
		sb.WriteString(cmd)
		sb.WriteByte('\n')
	}

	if sb.Len() == 0 {
		return "(no history)\n"
	}
	return sb.String()
}

// HistoryResult runs Get-History and returns compressed output.
func HistoryResult() (string, error) {
	if !powershell.Available() {
		return "", fmt.Errorf("PowerShell not available")
	}
	raw, err := powershell.InvokePwshText("Get-History")
	if err != nil || strings.TrimSpace(raw) == "" {
		return "(no history)\n", nil
	}
	return FilterHistory(raw), nil
}
