package ps

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var MeasureCmd = &cobra.Command{
	Use:                "measure [flags]",
	Short:              "Filter Measure-Object output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMeasure(args)
	},
}

func runMeasure(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk measure: PowerShell not found on PATH")
	}

	psArgs, requested := buildMeasureArgs(args)
	cmdlet := fmt.Sprintf("Measure-Object %s", psArgs)

	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil {
		fmt.Println("(measure failed)")
		return nil
	}

	compressed := FilterMeasure(raw, requested)
	fmt.Print(compressed)

	inTok := tracking.CountTokens(raw)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("measure", inTok, outTok)
	return nil
}

func buildMeasureArgs(args []string) (psArgs string, requested map[string]string) {
	requested = map[string]string{}
	var parts []string
	for _, a := range args {
		lower := strings.ToLower(a)
		switch {
		case lower == "-l" || lower == "-line":
			parts = append(parts, "-Line")
			requested["Count"] = "Lines"
		case lower == "-w" || lower == "-word":
			parts = append(parts, "-Word")
			requested["Words"] = "Words"
		case lower == "-c" || lower == "-character":
			parts = append(parts, "-Character")
			requested["Characters"] = "Chars"
		default:
			parts = append(parts, a)
		}
	}
	// Default to -Line if nothing specified
	if len(requested) == 0 {
		parts = append(parts, "-Line")
		requested["Count"] = "Lines"
	}
	return strings.Join(parts, " "), requested
}

// FilterMeasure parses Measure-Object output, dropping null properties.
// Input format:
//   Count    : 42
//   Average  :
//   Sum      :
//   ...
func FilterMeasure(raw string, requested map[string]string) string {
	lines := strings.Split(strings.TrimRight(raw, "\r\n"), "\n")
	var parts []string

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if val == "" {
			continue // null property — skip
		}

		// Map to human-friendly label
		label := key
		if friendly, ok := requested[key]; ok {
			label = friendly
		}
		parts = append(parts, fmt.Sprintf("%s: %s", label, val))
	}

	if len(parts) == 0 {
		return "(no output)\n"
	}
	return strings.Join(parts, "  ") + "\n"
}

// MeasureResult runs Measure-Object and returns compressed output.
func MeasureResult(lines, words, chars bool) (string, error) {
	if !powershell.Available() {
		return "", fmt.Errorf("PowerShell not available")
	}
	var flags []string
	if lines {
		flags = append(flags, "-l")
	}
	if words {
		flags = append(flags, "-w")
	}
	if chars {
		flags = append(flags, "-c")
	}
	psArgs, requested := buildMeasureArgs(flags)
	cmdlet := fmt.Sprintf("Measure-Object %s", psArgs)
	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil {
		return "(measure failed)\n", nil
	}
	return FilterMeasure(raw, requested), nil
}
