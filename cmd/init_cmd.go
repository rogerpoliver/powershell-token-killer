package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"ptk/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install PTK hooks into Claude Code settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func runInit() error {
	// 1. Find settings.json
	settingsPath, err := findSettings()
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	// 2. Find PTK binary location (hooks relative to binary or cwd)
	hooksDir, err := findHooksDir()
	if err != nil {
		return fmt.Errorf("init: hooks dir: %w", err)
	}

	// 3. Ensure .claude/hooks exists
	claudeHooks := filepath.Join(filepath.Dir(settingsPath), "hooks")
	if err := os.MkdirAll(claudeHooks, 0755); err != nil {
		return fmt.Errorf("init: create hooks dir: %w", err)
	}

	// 4. Copy hook files
	hookFiles := []string{
		"caveman-activate.js",
		"caveman-mode-tracker.js",
		"caveman-config.js",
		"caveman-stats.js",
	}
	if runtime.GOOS == "windows" {
		hookFiles = append(hookFiles, "caveman-statusline.ps1")
	} else {
		hookFiles = append(hookFiles, "caveman-statusline.sh")
	}

	for _, f := range hookFiles {
		src := filepath.Join(hooksDir, f)
		dst := filepath.Join(claudeHooks, f)
		if err := copyFile(src, dst); err != nil {
			fmt.Printf("  warning: could not copy %s: %v\n", f, err)
		} else {
			fmt.Printf("  copied %s\n", f)
		}
	}

	// Copy ptk-rewrite hook
	var rewriteHookSrc, rewriteHookDst string
	if runtime.GOOS == "windows" {
		rewriteHookSrc = filepath.Join(hooksDir, "claude", "ptk-rewrite.ps1")
		rewriteHookDst = filepath.Join(claudeHooks, "ptk-rewrite.ps1")
	} else {
		rewriteHookSrc = filepath.Join(hooksDir, "claude", "ptk-rewrite.ps1")
		rewriteHookDst = filepath.Join(claudeHooks, "ptk-rewrite.ps1")
	}
	if err := copyFile(rewriteHookSrc, rewriteHookDst); err != nil {
		fmt.Printf("  warning: could not copy ptk-rewrite.ps1: %v\n", err)
	} else {
		fmt.Println("  copied ptk-rewrite.ps1")
	}

	// 5. Patch settings.json
	if err := patchSettings(settingsPath, claudeHooks); err != nil {
		fmt.Printf("  warning: could not patch settings.json: %v\n", err)
	} else {
		fmt.Println("  patched settings.json")
	}

	// 6. Create config.toml
	if err := createConfig(); err != nil {
		fmt.Printf("  warning: could not create config.toml: %v\n", err)
	} else {
		fmt.Println("  created ~/.config/ptk/config.toml")
	}

	fmt.Println("\nPTK installed. Restart Claude Code to activate.")
	return nil
}

func findSettings() (string, error) {
	// Try project-local .claude/settings.json first
	if _, err := os.Stat(".claude/settings.json"); err == nil {
		abs, _ := filepath.Abs(".claude/settings.json")
		return abs, nil
	}
	// Fall back to global ~/.claude/settings.json
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	global := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(global), 0755); err != nil {
		return "", err
	}
	// Create empty settings if not exists
	if _, err := os.Stat(global); os.IsNotExist(err) {
		if err := os.WriteFile(global, []byte("{}"), 0644); err != nil {
			return "", err
		}
	}
	return global, nil
}

func findHooksDir() (string, error) {
	// Look for hooks/ relative to cwd (development) or executable
	candidates := []string{"hooks"}
	exe, _ := os.Executable()
	if exe != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "hooks"))
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}
	return "", fmt.Errorf("hooks directory not found (looked in: %v)", candidates)
}

func patchSettings(settingsPath, claudeHooks string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		settings = map[string]interface{}{}
	}

	// Build hook commands
	ptkRewrite := filepath.Join(claudeHooks, "ptk-rewrite.ps1")
	cavemanActivate := filepath.Join(claudeHooks, "caveman-activate.js")
	cavemanTracker := filepath.Join(claudeHooks, "caveman-mode-tracker.js")

	var statusLineCmd string
	if runtime.GOOS == "windows" {
		statusLineCmd = fmt.Sprintf("pwsh -NoProfile -File %s", filepath.Join(claudeHooks, "caveman-statusline.ps1"))
	} else {
		statusLineCmd = fmt.Sprintf("bash %s", filepath.Join(claudeHooks, "caveman-statusline.sh"))
	}

	var rewriteCmd string
	if runtime.GOOS == "windows" {
		rewriteCmd = fmt.Sprintf("pwsh -NoProfile -File %s", ptkRewrite)
	} else {
		rewriteCmd = fmt.Sprintf("pwsh -NoProfile -File %s", ptkRewrite)
	}

	settings["hooks"] = map[string]interface{}{
		"PreToolUse": []interface{}{
			map[string]interface{}{
				"matcher": "Bash",
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": rewriteCmd},
				},
			},
		},
		"SessionStart": []interface{}{
			map[string]interface{}{
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": fmt.Sprintf("node %s", cavemanActivate)},
				},
			},
		},
		"UserPromptSubmit": []interface{}{
			map[string]interface{}{
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": fmt.Sprintf("node %s", cavemanTracker)},
				},
			},
		},
	}
	settings["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": statusLineCmd,
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

func createConfig() error {
	var configDir string
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(base, "ptk")
	} else {
		base := os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(base, "ptk")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configFile); err == nil {
		return nil // already exists
	}
	return os.WriteFile(configFile, []byte(config.DefaultTOML()), 0644)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
