package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ps "ptk/cmd/ps"
	"ptk/internal/tracking"
)

type mcpRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type toolDef struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	InputSchema schemaObj `json:"inputSchema"`
}

type schemaObj struct {
	Type       string                `json:"type"`
	Properties map[string]schemaProp `json:"properties,omitempty"`
}

type schemaProp struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// Run starts the MCP stdio server and blocks until stdin closes.
// All non-JSON output must go to stderr (stdout is the JSON-RPC channel).
func Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "[ptk mcp] invalid JSON:", err)
			continue
		}

		// Notifications have no id — don't respond
		if req.ID == nil {
			continue
		}

		result, rpcErr := dispatch(req)
		resp := mcpResponse{Jsonrpc: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("mcp: encode: %w", err)
		}
	}
	return scanner.Err()
}

func dispatch(req mcpRequest) (interface{}, *mcpError) {
	switch req.Method {
	case "initialize":
		return map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo":      map[string]interface{}{"name": "ptk", "version": "0.1.0"},
		}, nil
	case "tools/list":
		return map[string]interface{}{"tools": toolList()}, nil
	case "tools/call":
		return callTool(req.Params)
	default:
		return nil, &mcpError{Code: -32601, Message: "method not found: " + req.Method}
	}
}

func toolList() []toolDef {
	return []toolDef{
		{
			Name:        "gci",
			Description: "Compressed Get-ChildItem — directory listing with noise dirs filtered and human-readable sizes",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"path":     {Type: "string", Description: "Directory path", Default: "."},
					"show_all": {Type: "boolean", Description: "Show hidden and noise directories (node_modules, .git, target)", Default: false},
				},
			},
		},
		{
			Name:        "sls",
			Description: "Compressed Select-String — file search with long lines truncated at 120 chars",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"pattern": {Type: "string", Description: "Search pattern (regex)"},
					"path":    {Type: "string", Description: "File or directory to search"},
				},
			},
		},
		{
			Name:        "gps",
			Description: "Compressed Get-Process — top 20 processes by CPU with human-readable memory",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"name": {Type: "string", Description: "Process name filter (optional)"},
				},
			},
		},
		{
			Name:        "gsv",
			Description: "Compressed Get-Service — service list without verbose table headers",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"name":   {Type: "string", Description: "Service name filter (optional)"},
					"status": {Type: "string", Description: "Filter by status: Running, Stopped (optional)"},
				},
			},
		},
		{
			Name:        "measure",
			Description: "Compressed Measure-Object — drops null properties, returns only requested metrics",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"lines": {Type: "boolean", Description: "Count lines", Default: true},
					"words": {Type: "boolean", Description: "Count words", Default: false},
					"chars": {Type: "boolean", Description: "Count characters", Default: false},
				},
			},
		},
		{
			Name:        "history",
			Description: "Compressed Get-History — PowerShell command history without table header",
			InputSchema: schemaObj{Type: "object"},
		},
		{
			Name:        "phelp",
			Description: "Compressed Get-Help — synopsis section only (~95% token savings vs full help)",
			InputSchema: schemaObj{
				Type: "object",
				Properties: map[string]schemaProp{
					"cmdlet":   {Type: "string", Description: "PowerShell cmdlet name"},
					"examples": {Type: "boolean", Description: "Include up to 2 examples", Default: false},
				},
			},
		},
	}
}

type callParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func callTool(raw json.RawMessage) (interface{}, *mcpError) {
	var p callParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &mcpError{Code: -32602, Message: "invalid params: " + err.Error()}
	}

	var (
		out string
		err error
	)

	switch p.Name {
	case "gci":
		path := stringArg(p.Arguments, "path", ".")
		showAll := boolArg(p.Arguments, "show_all", false)
		out, err = ps.GCIResult(path, showAll)
	case "sls":
		pattern := stringArg(p.Arguments, "pattern", "")
		path := stringArg(p.Arguments, "path", ".")
		out, err = ps.SLSResult(pattern, path, nil)
	case "gps":
		name := stringArg(p.Arguments, "name", "")
		out, err = ps.GPSResult(name)
	case "gsv":
		name := stringArg(p.Arguments, "name", "")
		status := stringArg(p.Arguments, "status", "")
		out, err = ps.GSVResult(name, status)
	case "measure":
		lines := boolArg(p.Arguments, "lines", true)
		words := boolArg(p.Arguments, "words", false)
		chars := boolArg(p.Arguments, "chars", false)
		out, err = ps.MeasureResult(lines, words, chars)
	case "history":
		out, err = ps.HistoryResult()
	case "phelp":
		cmdlet := stringArg(p.Arguments, "cmdlet", "")
		examples := boolArg(p.Arguments, "examples", false)
		out, err = ps.HelpResult(cmdlet, examples)
	default:
		return nil, &mcpError{Code: -32602, Message: "unknown tool: " + p.Name}
	}

	isError := false
	if err != nil {
		out = err.Error()
		isError = true
	}

	_ = tracking.Record(p.Name, tracking.CountTokens(out)*3, tracking.CountTokens(out))

	return toolResult{
		Content: []toolContent{{Type: "text", Text: out}},
		IsError: isError,
	}, nil
}

func stringArg(args map[string]interface{}, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func boolArg(args map[string]interface{}, key string, def bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}
