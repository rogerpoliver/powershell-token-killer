package ps

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var GsvCmd = &cobra.Command{
	Use:                "gsv [name] [flags]",
	Short:              "Filter Get-Service output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGsv(args)
	},
}

func runGsv(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk gsv: PowerShell not found on PATH")
	}

	nameFilter := ""
	statusFilter := ""
	for i, a := range args {
		lower := strings.ToLower(a)
		switch {
		case lower == "-name" && i+1 < len(args):
			nameFilter = args[i+1]
		case lower == "-status" && i+1 < len(args):
			statusFilter = args[i+1]
		case !strings.HasPrefix(a, "-"):
			nameFilter = a
		}
	}

	cmdlet := "Get-Service"
	if nameFilter != "" {
		cmdlet += fmt.Sprintf(" -Name %s", nameFilter)
	}
	if statusFilter != "" {
		cmdlet += fmt.Sprintf(" -Status %s", statusFilter)
	}

	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil {
		fmt.Println("(no services found)")
		return nil
	}

	compressed := filterGSV(raw)
	fmt.Print(compressed)

	inTok := tracking.CountTokens(raw)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("gsv", inTok, outTok)
	return nil
}

// filterGSV parses tabular Get-Service output and compacts it.
func filterGSV(raw string) string {
	lines := strings.Split(strings.TrimRight(raw, "\r\n"), "\n")
	var sb strings.Builder
	inData := false

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		// Skip header and separator lines
		if strings.Contains(line, "Status") && strings.Contains(line, "Name") {
			inData = true
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "---") || strings.TrimSpace(line) == "" {
			continue
		}
		if !inData {
			continue
		}

		// Parse: Status  Name  DisplayName
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		status := fields[0]
		name := fields[1]
		display := ""
		if len(fields) > 2 {
			display = strings.Join(fields[2:], " ")
		}

		// Truncate display name
		if len(display) > 40 {
			display = display[:37] + "..."
		}

		if display != "" {
			sb.WriteString(fmt.Sprintf("%-10s  %-25s  %s\n", status, name, display))
		} else {
			sb.WriteString(fmt.Sprintf("%-10s  %s\n", status, name))
		}
	}

	if sb.Len() == 0 {
		return "(no services)\n"
	}
	return sb.String()
}
