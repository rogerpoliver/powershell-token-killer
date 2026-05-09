package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"ptk/internal/config"
	"ptk/internal/embedded"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install PTK hooks into Claude Code settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		claudeFlag, _ := cmd.Flags().GetBool("claude")
		copilotFlag, _ := cmd.Flags().GetBool("copilot")
		// Default (no flags): install both
		if !claudeFlag && !copilotFlag {
			claudeFlag = true
		}
		if claudeFlag {
			if err := runInit(); err != nil {
				return err
			}
		}
		if copilotFlag {
			if err := runCopilotInit(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("claude", false, "Install Claude Code hooks (default when no flags specified)")
	initCmd.Flags().Bool("copilot", false, "Install GitHub Copilot MCP config and instructions")
}

func runInit() error {
	// 1. Find settings.json
	settingsPath, err := findSettings()
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	// 2. Ensure .claude/hooks exists
	claudeHooks := filepath.Join(filepath.Dir(settingsPath), "hooks")
	if err := os.MkdirAll(claudeHooks, 0755); err != nil {
		return fmt.Errorf("init: create hooks dir: %w", err)
	}

	// 3. Extract hook files from embedded FS
	hookFiles := []string{
		"hooks/caveman-activate.js",
		"hooks/caveman-mode-tracker.js",
		"hooks/caveman-config.js",
		"hooks/caveman-stats.js",
		"hooks/claude/ptk-rewrite.ps1",
	}
	if runtime.GOOS == "windows" {
		hookFiles = append(hookFiles, "hooks/caveman-statusline.ps1")
	} else {
		hookFiles = append(hookFiles, "hooks/caveman-statusline.sh")
	}

	for _, embeddedPath := range hookFiles {
		data, err := fs.ReadFile(embedded.FS, embeddedPath)
		if err != nil {
			fmt.Printf("  warning: could not read embedded %s: %v\n", embeddedPath, err)
			continue
		}
		dst := filepath.Join(claudeHooks, filepath.Base(embeddedPath))
		if err := os.WriteFile(dst, data, 0644); err != nil {
			fmt.Printf("  warning: could not write %s: %v\n", filepath.Base(embeddedPath), err)
		} else {
			fmt.Printf("  copied %s\n", filepath.Base(embeddedPath))
		}
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

	// Merge into existing hooks map (preserve any user-defined hooks)
	existing, _ := settings["hooks"].(map[string]interface{})
	if existing == nil {
		existing = map[string]interface{}{}
	}
	existing["PreToolUse"] = []interface{}{
		map[string]interface{}{
			"matcher": "Bash",
			"hooks": []interface{}{
				map[string]interface{}{"type": "command", "command": rewriteCmd},
			},
		},
	}
	existing["SessionStart"] = []interface{}{
		map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{"type": "command", "command": fmt.Sprintf("node %s", cavemanActivate)},
			},
		},
	}
	existing["UserPromptSubmit"] = []interface{}{
		map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{"type": "command", "command": fmt.Sprintf("node %s", cavemanTracker)},
			},
		},
	}
	settings["hooks"] = existing
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

// runCopilotInit installs GitHub Copilot MCP config and instructions.
// It writes (or merges) .vscode/settings.json with the ptk MCP server entry
// and writes .github/copilot-instructions.md from the embedded FS.
func runCopilotInit() error {
	// 1. Write .vscode/settings.json with MCP server config (merge, don't replace)
	if err := os.MkdirAll(".vscode", 0755); err != nil {
		return fmt.Errorf("copilot init: create .vscode: %w", err)
	}
	vscodePath := ".vscode/settings.json"

	var vsSettings map[string]interface{}
	existing, err := os.ReadFile(vscodePath)
	if err == nil {
		if jsonErr := json.Unmarshal(existing, &vsSettings); jsonErr != nil {
			vsSettings = map[string]interface{}{}
		}
	} else {
		vsSettings = map[string]interface{}{}
	}

	// Merge mcp.servers.ptk into existing settings
	mcpBlock, _ := vsSettings["mcp"].(map[string]interface{})
	if mcpBlock == nil {
		mcpBlock = map[string]interface{}{}
	}
	servers, _ := mcpBlock["servers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
	}
	servers["ptk"] = map[string]interface{}{
		"type":    "stdio",
		"command": "ptk",
		"args":    []interface{}{"serve", "--mcp"},
	}
	mcpBlock["servers"] = servers
	vsSettings["mcp"] = mcpBlock

	out, err := json.MarshalIndent(vsSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("copilot init: marshal settings: %w", err)
	}
	if err := os.WriteFile(vscodePath, out, 0644); err != nil {
		return fmt.Errorf("copilot init: write %s: %w", vscodePath, err)
	}
	fmt.Printf("  wrote %s\n", vscodePath)

	// 2. Write .github/copilot-instructions.md from embedded FS
	if err := os.MkdirAll(".github", 0755); err != nil {
		return fmt.Errorf("copilot init: create .github: %w", err)
	}
	instructionsData, err := fs.ReadFile(embedded.FS, "skills/copilot/copilot-instructions.md")
	if err != nil {
		return fmt.Errorf("copilot init: read embedded instructions: %w", err)
	}
	instructionsPath := ".github/copilot-instructions.md"
	if err := os.WriteFile(instructionsPath, instructionsData, 0644); err != nil {
		return fmt.Errorf("copilot init: write %s: %w", instructionsPath, err)
	}
	fmt.Printf("  wrote %s\n", instructionsPath)

	fmt.Println("\nCopilot MCP server configured. Open VS Code to activate.")
	return nil
}

