package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Hooks      HooksConfig      `toml:"hooks"`
	PowerShell PowerShellConfig `toml:"powershell"`
	Caveman    CavemanConfig    `toml:"caveman"`
	Tracking   TrackingConfig   `toml:"tracking"`
}

type HooksConfig struct {
	ExcludeCommands []string `toml:"exclude_commands"`
	Allow           []string `toml:"allow"`
	Deny            []string `toml:"deny"`
	Ask             []string `toml:"ask"`
}

type PowerShellConfig struct {
	PreferPwsh bool `toml:"prefer_pwsh"`
	MaxItems   int  `toml:"max_items"`
}

type CavemanConfig struct {
	DefaultMode string `toml:"default_mode"`
}

type TrackingConfig struct {
	Enabled bool `toml:"enabled"`
}

// defaults ships safe allow rules so read-only PS cmdlets auto-allow without
// prompting the user every time (exit 0 from ptk rewrite).
var defaults = Config{
	Hooks: HooksConfig{
		Allow: []string{
			"Get-ChildItem *", "gci *", "ls *", "dir *",
			"Select-String *", "sls *", "grep *",
			"Get-Process", "Get-Process *", "gps", "gps *", "ps", "ps *",
			"Get-Service", "Get-Service *", "gsv", "gsv *",
			"Measure-Object *", "measure *", "wc *",
			"Get-History", "Get-History *", "history *",
			"Get-Help *", "man *",
			"git *", "cargo *", "go *", "npm *", "pnpm *",
			"docker *", "kubectl *",
		},
		Deny: []string{
			"Remove-Item * -Recurse -Force",
			"Format-Volume",
			"Set-ExecutionPolicy Bypass",
			"Set-ExecutionPolicy Unrestricted",
		},
		Ask: []string{
			"Stop-Process *",
			"Remove-Item * -Recurse",
			"Set-ExecutionPolicy *",
			"Stop-Service *",
			"Restart-Service *",
		},
	},
	PowerShell: PowerShellConfig{PreferPwsh: true, MaxItems: 100},
	Caveman:    CavemanConfig{DefaultMode: "full"},
	Tracking:   TrackingConfig{Enabled: true},
}

// Load reads the platform-appropriate config file.
// Returns defaults merged with user overrides. Missing file returns pure defaults.
func Load() Config {
	cfg := defaults
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg
	}
	// User file overrides defaults field by field
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg
	}
	return cfg
}

func configPath() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, "ptk", "config.toml")
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "ptk", "config.toml")
}

// DefaultTOML returns the starter config file content written by ptk init.
func DefaultTOML() string {
	return `[hooks]
# Commands to auto-allow (ptk rewrite exits 0 → Claude Code skips confirmation)
allow = [
  "Get-ChildItem *", "gci *", "ls *", "dir *",
  "Select-String *", "sls *", "grep *",
  "Get-Process", "Get-Process *", "gps *", "ps *",
  "Get-Service", "Get-Service *", "gsv *",
  "Measure-Object *",
  "Get-History *",
  "Get-Help *",
  "git *", "cargo *", "go *", "npm *", "pnpm *",
  "docker *", "kubectl *",
]

# Commands to block entirely
deny = [
  "Remove-Item * -Recurse -Force",
  "Format-Volume",
  "Set-ExecutionPolicy Bypass",
  "Set-ExecutionPolicy Unrestricted",
]

# Commands that require user confirmation
ask = [
  "Stop-Process *",
  "Remove-Item * -Recurse",
  "Set-ExecutionPolicy *",
  "Stop-Service *",
  "Restart-Service *",
]

[powershell]
prefer_pwsh = true
max_items   = 100

[caveman]
default_mode = "full"

[tracking]
enabled = true
`
}
