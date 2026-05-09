package ps

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var HelpPSCmd = &cobra.Command{
	Use:                "phelp <cmdlet> [flags]",
	Short:              "Filter Get-Help output (synopsis only)",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHelp(args)
	},
}

func runHelp(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk help: PowerShell not found on PATH")
	}

	showExamples := false
	var target string
	for _, a := range args {
		lower := strings.ToLower(a)
		switch {
		case lower == "-examples":
			showExamples = true
		case !strings.HasPrefix(a, "-"):
			target = a
		}
	}

	if target == "" {
		fmt.Println("Usage: ptk help <cmdlet>")
		return nil
	}

	cmdlet := fmt.Sprintf("Get-Help %s", target)
	raw, err := powershell.InvokePwshText(cmdlet)
	if err != nil || strings.TrimSpace(raw) == "" {
		fmt.Printf("(no help found for %s)\n", target)
		return nil
	}

	var compressed string
	if showExamples {
		compressed = ExtractSynopsisAndExamples(raw, 2)
	} else {
		compressed = ExtractSynopsis(raw)
	}

	fmt.Print(compressed)

	inTok := tracking.CountTokens(raw)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("help", inTok, outTok)
	return nil
}

// ExtractSynopsis pulls only the SYNOPSIS or first SYNTAX block from Get-Help output.
// Drops DESCRIPTION, PARAMETERS, NOTES, RELATED LINKS — typically 95% token savings.
func ExtractSynopsis(raw string) string {
	lines := strings.Split(raw, "\n")
	var sb strings.Builder
	inSynopsis := false
	inSyntax := false
	syntaxDone := false

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		upper := strings.ToUpper(strings.TrimSpace(line))

		switch {
		case upper == "SYNOPSIS":
			inSynopsis = true
			inSyntax = false
			continue
		case upper == "SYNTAX":
			inSyntax = true
			inSynopsis = false
			continue
		case isSection(upper) && (inSynopsis || inSyntax):
			// Hit next section — stop
			inSynopsis = false
			inSyntax = false
			syntaxDone = true
		}

		if syntaxDone {
			break
		}

		if (inSynopsis || inSyntax) && strings.TrimSpace(line) != "" {
			sb.WriteString(strings.TrimSpace(line))
			sb.WriteByte('\n')
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		// Fallback: first non-empty 3 lines
		count := 0
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if strings.TrimSpace(line) != "" {
				sb.WriteString(line)
				sb.WriteByte('\n')
				count++
				if count >= 3 {
					break
				}
			}
		}
		return sb.String()
	}
	return result + "\n"
}

func ExtractSynopsisAndExamples(raw string, maxExamples int) string {
	synopsis := ExtractSynopsis(raw)

	lines := strings.Split(raw, "\n")
	var examples strings.Builder
	exampleCount := 0
	inExample := false

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		upper := strings.ToUpper(strings.TrimSpace(line))

		if strings.HasPrefix(upper, "EXAMPLE") && strings.Contains(upper, "EXAMPLE") {
			exampleCount++
			if exampleCount > maxExamples {
				break
			}
			inExample = true
			examples.WriteString(fmt.Sprintf("\nExample %d:\n", exampleCount))
			continue
		}
		if inExample && isSection(upper) && !strings.HasPrefix(upper, "EXAMPLE") {
			inExample = false
		}
		if inExample && strings.TrimSpace(line) != "" {
			examples.WriteString("  ")
			examples.WriteString(strings.TrimSpace(line))
			examples.WriteByte('\n')
		}
	}

	return synopsis + examples.String()
}

func isSection(upper string) bool {
	sections := []string{
		"SYNOPSIS", "SYNTAX", "DESCRIPTION", "PARAMETERS",
		"INPUTS", "OUTPUTS", "NOTES", "EXAMPLES", "RELATED LINKS",
		"EXAMPLE", "NAME", "ALIASES",
	}
	for _, s := range sections {
		if upper == s {
			return true
		}
	}
	return false
}

// HelpResult runs Get-Help and returns compressed output.
func HelpResult(target string, examples bool) (string, error) {
	if !powershell.Available() {
		return "", fmt.Errorf("PowerShell not available")
	}
	if target == "" {
		return "Usage: ptk phelp <cmdlet>\n", nil
	}
	raw, err := powershell.InvokePwshText(fmt.Sprintf("Get-Help %s", target))
	if err != nil || strings.TrimSpace(raw) == "" {
		return fmt.Sprintf("(no help found for %s)\n", target), nil
	}
	if examples {
		return ExtractSynopsisAndExamples(raw, 2), nil
	}
	return ExtractSynopsis(raw), nil
}
