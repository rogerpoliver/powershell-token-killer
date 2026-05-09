package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"ptk/internal/mcp"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start PTK as an MCP server (GitHub Copilot, Cursor, and other MCP clients)",
	Long: `Start PTK as a Model Context Protocol (MCP) server over stdio.

Exposes compressed PowerShell cmdlets as MCP tools:
  gci, sls, gps, gsv, measure, history, phelp

Configure in VS Code (.vscode/settings.json):
  {
    "mcp": {
      "servers": {
        "ptk": {
          "type": "stdio",
          "command": "ptk",
          "args": ["serve", "--mcp"]
        }
      }
    }
  }

Or run: ptk init --copilot`,
	RunE: func(cmd *cobra.Command, args []string) error {
		isMCP, _ := cmd.Flags().GetBool("mcp")
		if !isMCP {
			return fmt.Errorf("use --mcp flag: ptk serve --mcp")
		}
		return mcp.Run()
	},
}

func init() {
	serveCmd.Flags().Bool("mcp", false, "Serve MCP protocol over stdio")
	rootCmd.AddCommand(serveCmd)
}
