package powershell

import "strings"

// aliasMap maps PowerShell aliases and Unix-imported names to canonical cmdlet names.
// PowerShell is case-insensitive so we lowercase-match.
var aliasMap = map[string]string{
	// Filesystem
	"ls":      "Get-ChildItem",
	"dir":     "Get-ChildItem",
	"gci":     "Get-ChildItem",
	"cat":     "Get-Content",
	"gc":      "Get-Content",
	"type":    "Get-Content",
	"find":    "Get-ChildItem",
	"tree":    "Get-ChildItem",
	// Search
	"grep":    "Select-String",
	"sls":     "Select-String",
	// Processes
	"ps":      "Get-Process",
	"gps":     "Get-Process",
	"kill":    "Stop-Process",
	"spps":    "Stop-Process",
	// Services
	"gsv":     "Get-Service",
	// Measurement
	"measure": "Measure-Object",
	"wc":      "Measure-Object",
	// History
	"history": "Get-History",
	"h":       "Get-History",
	"ghy":     "Get-History",
	// Help
	"man":     "Get-Help",
	"help":    "Get-Help",
	// Location
	"sl":      "Set-Location",
	"cd":      "Set-Location",
	"chdir":   "Set-Location",
}

// Normalize maps a PowerShell alias or Unix-imported command to its canonical cmdlet.
// Preserves arguments unchanged. Returns cmd unchanged if no alias found.
func Normalize(cmd string) string {
	first, rest, hasRest := strings.Cut(cmd, " ")
	canonical, ok := aliasMap[strings.ToLower(first)]
	if !ok {
		return cmd
	}
	if !hasRest || rest == "" {
		return canonical
	}
	return canonical + " " + rest
}
