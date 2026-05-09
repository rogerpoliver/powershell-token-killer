package ps

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

// noiseDirs are always hidden unless -a/--all is passed.
var noiseDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"target":       true,
	"bin":          true,
	"obj":          true,
	".next":        true,
	"dist":         true,
	"__pycache__":  true,
	".mypy_cache":  true,
	".pytest_cache": true,
	"vendor":       true,
}

var GciCmd = &cobra.Command{
	Use:                "gci [path] [flags]",
	Short:              "Filter Get-ChildItem output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGci(args)
	},
}

type gciEntry struct {
	Name       string `json:"Name"`
	Length     int64  `json:"Length"`
	Attributes string `json:"Attributes"`
}

func runGci(args []string) error {
	path, flags, showAll, recurse := parseGciArgs(args)

	if !powershell.Available() {
		return fmt.Errorf("ptk gci: PowerShell not found on PATH")
	}

	// Build Get-ChildItem invocation
	psArgs := fmt.Sprintf("-Path %s", quotePath(path))
	if recurse {
		psArgs += " -Recurse"
	}
	if showAll {
		psArgs += " -Force"
	}
	for _, f := range flags {
		psArgs += " " + f
	}

	cmdlet := fmt.Sprintf("Get-ChildItem %s", psArgs)
	raw, err := powershell.InvokePwshJSON(cmdlet + " | Select-Object Name,Length,Attributes")
	if err != nil {
		// Fallback: run without filter, print raw
		rawText, _ := powershell.InvokePwshText(cmdlet)
		fmt.Println(rawText)
		return nil
	}

	rawStr := string(raw)
	entries, err := parseGCIJSON(rawStr)
	if err != nil || len(entries) == 0 {
		// Fallback to text output
		rawText, _ := powershell.InvokePwshText(cmdlet)
		if rawText == "" {
			fmt.Println("(empty)")
			return nil
		}
		fmt.Println(rawText)
		return nil
	}

	compressed := formatGCI(entries, showAll)
	fmt.Print(compressed)

	// Track savings
	inTok := tracking.CountTokens(rawStr)
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("gci", inTok, outTok)

	return nil
}

func parseGciArgs(args []string) (path string, flags []string, showAll, recurse bool) {
	path = "."
	for _, a := range args {
		lower := strings.ToLower(a)
		switch {
		case lower == "-a" || lower == "--all" || lower == "-force":
			showAll = true
		case lower == "-r" || lower == "-recurse" || lower == "--recurse":
			recurse = true
		case lower == "-tree":
			recurse = true
		case strings.HasPrefix(a, "-la") || strings.HasPrefix(a, "-al"):
			showAll = true
		case strings.HasPrefix(a, "-"):
			// pass other flags through
			flags = append(flags, a)
		default:
			path = a
		}
	}
	return
}

func parseGCIJSON(raw string) ([]gciEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil, nil
	}
	// PS returns single object (not array) when only 1 item
	if strings.HasPrefix(raw, "{") {
		var entry gciEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			return nil, err
		}
		return []gciEntry{entry}, nil
	}
	var entries []gciEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func isDir(e gciEntry) bool {
	return strings.Contains(e.Attributes, "Directory")
}

func formatGCI(entries []gciEntry, showAll bool) string {
	var dirs, files []gciEntry
	extCount := map[string]int{}

	for _, e := range entries {
		if !showAll && noiseDirs[e.Name] {
			continue
		}
		if isDir(e) {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
			ext := filepath.Ext(e.Name)
			if ext == "" {
				ext = "(no ext)"
			}
			extCount[ext]++
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		return "(empty)\n"
	}

	var sb strings.Builder

	for _, d := range dirs {
		sb.WriteString(d.Name)
		sb.WriteString("/\n")
	}
	for _, f := range files {
		sb.WriteString(f.Name)
		sb.WriteString("  ")
		sb.WriteString(humanSize(f.Length))
		sb.WriteString("\n")
	}

	// Summary line
	if len(dirs)+len(files) > 0 {
		sb.WriteString(fmt.Sprintf("\nSummary: %d files, %d dirs", len(files), len(dirs)))
		if len(extCount) > 0 {
			sb.WriteString(" (")
			exts := topExts(extCount, 5)
			sb.WriteString(strings.Join(exts, ", "))
			if len(extCount) > 5 {
				sb.WriteString(fmt.Sprintf(", +%d more", len(extCount)-5))
			}
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func quotePath(path string) string {
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`
	}
	return path
}

func topExts(extCount map[string]int, n int) []string {
	type kv struct{ k string; v int }
	var sorted []kv
	for k, v := range extCount {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	var result []string
	for i, kv := range sorted {
		if i >= n {
			break
		}
		result = append(result, fmt.Sprintf("%s ×%d", kv.k, kv.v))
	}
	return result
}

// GCIResult runs Get-ChildItem and returns compressed output for the given path.
func GCIResult(path string, showAll bool) (string, error) {
	if !powershell.Available() {
		return "", fmt.Errorf("PowerShell not available")
	}
	cmdlet := fmt.Sprintf("Get-ChildItem -Path %s", quotePath(path))
	if showAll {
		cmdlet += " -Force"
	}
	raw, err := powershell.InvokePwshJSON(cmdlet + " | Select-Object Name,Length,Attributes")
	if err != nil {
		rawText, _ := powershell.InvokePwshText("Get-ChildItem -Path " + quotePath(path))
		return rawText, nil
	}
	entries, err := parseGCIJSON(string(raw))
	if err != nil || len(entries) == 0 {
		rawText, _ := powershell.InvokePwshText("Get-ChildItem -Path " + quotePath(path))
		if rawText == "" {
			return "(empty)\n", nil
		}
		return rawText, nil
	}
	return formatGCI(entries, showAll), nil
}

