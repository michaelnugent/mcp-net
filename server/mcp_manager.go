package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ToolInfo represents information about a tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// MCPInfo stores information about an MCP executable
type MCPInfo struct {
	Name      string
	Path      string
	ToolInfos []ToolInfo
}

// MCPManager manages a collection of MCP executables
type MCPManager struct {
	mcpMap       map[string]*MCPInfo
	mcpDirectory string
	mutex        sync.RWMutex
}

// NewMCPManager creates a new MCP manager
func NewMCPManager(mcpDirectory string) *MCPManager {
	return &MCPManager{
		mcpMap:       make(map[string]*MCPInfo),
		mcpDirectory: mcpDirectory,
	}
}

// LoadMCPs loads all MCPs from the configured directory
func (m *MCPManager) LoadMCPs() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clear existing MCPs
	m.mcpMap = make(map[string]*MCPInfo)

	// Walk through the MCP directory
	return filepath.WalkDir(m.mcpDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip non-executable files
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&0111 == 0 {
			return nil
		}

		// Get the base name without extension
		name := filepath.Base(path)
		ext := filepath.Ext(name)
		if ext != "" {
			name = name[:len(name)-len(ext)]
		}

		// Create MCP info
		mcpInfo := &MCPInfo{
			Name: name,
			Path: path,
		}

		// Try to get tool info
		toolInfos, err := m.getToolInfos(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get tool info for %s: %v\n", path, err)
		} else {
			mcpInfo.ToolInfos = toolInfos
		}

		// Store MCP info
		m.mcpMap[name] = mcpInfo
		fmt.Fprintf(os.Stderr, "Loaded MCP: %s from %s with %d tools\n", name, path, len(mcpInfo.ToolInfos))

		return nil
	})
}

// getToolInfos queries an MCP executable for its tool information
func (m *MCPManager) getToolInfos(mcpPath string) ([]ToolInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a temporary client to get the tool info
	cmd := exec.CommandContext(ctx, mcpPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP: %w", err)
	}

	// Create a simple JSON-RPC client
	// First, initialize the MCP
	initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocol_version":"2024-11-05"}}`
	_, err = stdin.Write([]byte(initMsg + "\n"))
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to send initialize message: %w", err)
	}

	// Read the initialize response (we don't need to parse it)
	buffer := make([]byte, 4096)
	_, err = stdout.Read(buffer)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to read initialize response: %w", err)
	}

	// Now, send the tools/list request
	listMsg := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	_, err = stdin.Write([]byte(listMsg + "\n"))
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to send tools/list message: %w", err)
	}

	// Read the tools/list response
	n, err := stdout.Read(buffer)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to read tools/list response: %w", err)
	}

	// Kill the process
	cmd.Process.Kill()

	// Parse the response to get the tool info
	response := buffer[:n]

	// Parse the JSON-RPC response
	var resp struct {
		Result struct {
			Tools []ToolInfo `json:"tools"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse tools/list response: %w", err)
	}

	return resp.Result.Tools, nil
}

// GetAllTools returns all tools from all MCPs
func (m *MCPManager) GetAllTools() []ToolInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var allTools []ToolInfo
	for mcpName, mcpInfo := range m.mcpMap {
		for _, tool := range mcpInfo.ToolInfos {
			// Create a copy of the tool with the name prefixed by the MCP name
			toolCopy := tool
			toolCopy.Name = fmt.Sprintf("%s.%s", mcpName, tool.Name)
			allTools = append(allTools, toolCopy)
		}
	}
	return allTools
}

// GetMCPForTool returns the MCP info for a given tool name
func (m *MCPManager) GetMCPForTool(toolName string) (*MCPInfo, string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	parts := strings.SplitN(toolName, ".", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid tool name format, expected 'mcp.tool': %s", toolName)
	}

	mcpName := parts[0]
	localToolName := parts[1]

	mcpInfo, ok := m.mcpMap[mcpName]
	if !ok {
		return nil, "", fmt.Errorf("MCP not found: %s", mcpName)
	}

	return mcpInfo, localToolName, nil
}

// ExecuteTool executes a tool on the appropriate MCP
func (m *MCPManager) ExecuteTool(ctx context.Context, toolName string, parameters map[string]interface{}) (interface{}, error) {
	mcpInfo, localToolName, err := m.GetMCPForTool(toolName)
	if err != nil {
		return nil, err
	}

	// Create a command to execute the MCP
	cmd := exec.CommandContext(ctx, mcpInfo.Path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP: %w", err)
	}

	// Ensure the command is killed when done
	defer cmd.Process.Kill()

	// Initialize the MCP
	initMsg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocol_version":"2024-11-05"}}`
	_, err = stdin.Write([]byte(initMsg + "\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to send initialize message: %w", err)
	}

	// Read the initialize response
	buffer := make([]byte, 4096)
	_, err = stdout.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read initialize response: %w", err)
	}

	// Build the tool call request
	callParams := map[string]interface{}{
		"name":      localToolName,
		"arguments": parameters,
	}

	callRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params":  callParams,
	}

	callJSON, err := json.Marshal(callRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tools/call request: %w", err)
	}

	_, err = stdin.Write(append(callJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("failed to send tools/call message: %w", err)
	}

	// Read the response
	n, err := stdout.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read tools/call response: %w", err)
	}

	// Parse the JSON-RPC response
	var resp struct {
		Result interface{} `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(buffer[:n], &resp); err != nil {
		return nil, fmt.Errorf("failed to parse tools/call response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP tool error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}

	return resp.Result, nil
}
