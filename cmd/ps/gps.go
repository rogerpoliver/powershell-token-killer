package ps

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"ptk/internal/powershell"
	"ptk/internal/tracking"
)

var GpsCmd = &cobra.Command{
	Use:                "gps [name] [flags]",
	Short:              "Filter Get-Process output",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGps(args)
	},
}

type gpsEntry struct {
	Name        string  `json:"Name"`
	Id          int     `json:"Id"`
	CPU         float64 `json:"CPU"`
	WorkingSet  int64   `json:"WorkingSet"`
}

func runGps(args []string) error {
	if !powershell.Available() {
		return fmt.Errorf("ptk gps: PowerShell not found on PATH")
	}

	nameFilter := ""
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			nameFilter = a
		}
	}

	cmdlet := "Get-Process"
	if nameFilter != "" {
		cmdlet += fmt.Sprintf(" -Name %s", nameFilter)
	}
	cmdlet += " | Sort-Object CPU -Descending | Select-Object -First 20" +
		" | Select-Object Name,Id,CPU,WorkingSet"

	raw, err := powershell.InvokePwshJSON(cmdlet)
	if err != nil {
		rawText, _ := powershell.InvokePwshText("Get-Process")
		fmt.Println(rawText)
		return nil
	}

	entries, err := parseGPSJSON(string(raw))
	if err != nil || len(entries) == 0 {
		fmt.Println("(no processes)")
		return nil
	}

	compressed := formatGPS(entries)
	fmt.Print(compressed)

	inTok := tracking.CountTokens(string(raw))
	outTok := tracking.CountTokens(compressed)
	_ = tracking.Record("gps", inTok, outTok)
	return nil
}

func parseGPSJSON(raw string) ([]gpsEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil, nil
	}
	if strings.HasPrefix(raw, "{") {
		var e gpsEntry
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			return nil, err
		}
		return []gpsEntry{e}, nil
	}
	var entries []gpsEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func formatGPS(entries []gpsEntry) string {
	// Sort by CPU descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CPU > entries[j].CPU
	})

	var sb strings.Builder
	for _, e := range entries {
		name := e.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-20s %6d  %6s  %s\n",
			name,
			e.Id,
			fmtCPU(e.CPU),
			humanSize(e.WorkingSet),
		))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d processes\n", len(entries)))
	return sb.String()
}

func fmtCPU(cpu float64) string {
	if cpu < 0 {
		return "0.0s"
	}
	return fmt.Sprintf("%.1fs", cpu)
}
