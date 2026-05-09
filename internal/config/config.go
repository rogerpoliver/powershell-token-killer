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

var defaults = Config{
	Hooks: HooksConfig{
		Deny: []string{
			"Remove-Item * -Recurse -Force",
			"Format-Volume",
			"Set-ExecutionPolicy Bypass",
		},
		Ask: []string{
			"Stop-Process *",
			"Remove-Item * -Recurse",
			"Set-ExecutionPolicy *",
		},
	},
	PowerShell: PowerShellConfig{PreferPwsh: true, MaxItems: 100},
	Caveman:    CavemanConfig{DefaultMode: "full"},
	Tracking:   TrackingConfig{Enabled: true},
}

// Load reads ~/.config/ptk/config.toml (or %APPDATA%\ptk\config.toml on Windows).
// Returns defaults if file not found.
func Load() Config {
	path := configPath()
	cfg := defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg
	}
	return cfg
}

func configPath() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
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

// DefaultTOML returns the default config file content.
func DefaultTOML() string {
	return `[hooks]
exclude_commands = []
deny = [
  "Remove-Item * -Recurse -Force",
  "Format-Volume",
  "Set-ExecutionPolicy Bypass",
]
ask = [
  "Stop-Process *",
  "Remove-Item * -Recurse",
  "Set-ExecutionPolicy *",
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
