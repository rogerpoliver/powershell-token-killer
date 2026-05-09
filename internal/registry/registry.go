package registry

import "strings"

// cmdletToPTK maps canonical PS cmdlet names to ptk subcommands.
// Keys are PascalCase canonical form.
var cmdletToPTK = map[string]string{
	"Get-ChildItem":  "ptk gci",
	"Select-String":  "ptk sls",
	"Get-Process":    "ptk gps",
	"Get-Service":    "ptk gsv",
	"Measure-Object": "ptk measure",
	"Get-Content":    "ptk gc",
	"Get-History":    "ptk history",
	"Get-Help":       "ptk phelp",
	"Get-WinEvent":   "ptk winevent",
}

// Rewrite maps a normalized cmdlet command to its ptk equivalent.
// normalized should already be passed through powershell.Normalize().
// Returns (rewritten, true) or ("", false) if no mapping exists.
func Rewrite(normalized string) (string, bool) {
	first, rest, hasRest := strings.Cut(normalized, " ")
	// Case-insensitive lookup (PS is case-insensitive)
	subcmd := lookupCaseInsensitive(first)
	if subcmd == "" {
		return "", false
	}
	if !hasRest || rest == "" {
		return subcmd, true
	}
	return subcmd + " " + rest, true
}

func lookupCaseInsensitive(cmdlet string) string {
	// Direct lookup first (fast path for already-canonical form)
	if s, ok := cmdletToPTK[cmdlet]; ok {
		return s
	}
	// Slow path: normalize case
	lower := strings.ToLower(cmdlet)
	for k, v := range cmdletToPTK {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	return ""
}
