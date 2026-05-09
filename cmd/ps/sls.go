package ps

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var SlsCmd = &cobra.Command{
	Use:                "sls [pattern] [path] [flags]",
	Short:              "Filter Select-String output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSls(args)
	},
}

func runSls(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk sls: PowerShell not found on PATH")
	}

	psArgs := buildSlsArgs(args)
	cmdlet := fmt.Sprintf("Select-String %s", psArgs)
	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil {
		// No matches is exit 1 in PS — not an error for us
		fmt.Println("(no matches)")
		return nil
	}
	if strings.TrimSpace(raw) == "" {
		fmt.Println("(no matches)")
		return nil
	}

	compressed := filterSLS(raw)
	fmt.Print(compressed)

	inTok := tracking.CountTokens(raw)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("sls", inTok, outTok)
	return nil
}

func buildSlsArgs(args []string) string {
	var psArgs []string
	for _, a := range args {
		lower := strings.ToLower(a)
		switch {
		case lower == "-i" || lower == "-ignorecase":
			psArgs = append(psArgs, "-CaseSensitive:$false")
		case lower == "-v" || lower == "-notmatch":
			psArgs = append(psArgs, "-NotMatch")
		case lower == "-r" || lower == "-recurse":
			psArgs = append(psArgs, "-Recurse")
		case lower == "-l" || lower == "-list":
			psArgs = append(psArgs, "-List")
		case lower == "-f" || lower == "-simplematch":
			psArgs = append(psArgs, "-SimpleMatch")
		case lower == "-n" || lower == "-linenumber":
			// line numbers are default in SLS output
		default:
			psArgs = append(psArgs, a)
		}
	}
	return strings.Join(psArgs, " ")
}

// filterSLS compresses Select-String output.
// Format per line: "file:linenum:content" or "file:linenum:col:content"
func filterSLS(raw string) string {
	lines := strings.Split(strings.TrimRight(raw, "\r\n"), "\n")
	if len(lines) == 0 {
		return "(no matches)\n"
	}

	// Detect single-file mode: all lines have same filename prefix
	singleFile := detectSingleFile(lines)

	const maxLineLen = 120
	var sb strings.Builder
	matchCount := 0

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		matchCount++

		if singleFile {
			// Drop "filename:" prefix
			line = dropFilePrefix(line)
		}

		// Truncate very long lines
		if len(line) > maxLineLen {
			line = line[:maxLineLen] + "…"
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	if matchCount == 0 {
		return "(no matches)\n"
	}

	// Summary when many matches
	if matchCount > 5 {
		sb.WriteString(fmt.Sprintf("\nFound %d matches\n", matchCount))
	}

	return sb.String()
}

func detectSingleFile(lines []string) bool {
	var firstFile string
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			return false
		}
		file := parts[0]
		if firstFile == "" {
			firstFile = file
		} else if file != firstFile {
			return false
		}
	}
	return firstFile != ""
}

func dropFilePrefix(line string) string {
	// Drop "filename:" prefix, keep "linenum:content"
	idx := strings.Index(line, ":")
	if idx == -1 {
		return line
	}
	return line[idx+1:]
}
