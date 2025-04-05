package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// DefaultRequestTimeout is the default timeout for MCP requests
const DefaultRequestTimeout = 30 * time.Second

// MCPServer is the server that manages MCPs
type MCPServer struct {
	mcpManager *MCPManager
	server     *mcpserver.MCPServer
}

// NewMCPServer creates a new MCP server
func NewMCPServer(mcpDirectory string, name, version string) (*MCPServer, error) {
	// Create the MCP manager
	mcpManager := NewMCPManager(mcpDirectory)
	if err := mcpManager.LoadMCPs(); err != nil {
		return nil, fmt.Errorf("failed to load MCPs: %w", err)
	}

	// Create the MCP server
	server := mcpserver.NewMCPServer(name, version,
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithLogging(),
	)

	// Create the server
	mcpServer := &MCPServer{
		mcpManager: mcpManager,
		server:     server,
	}

	// Register our custom tools
	mcpServer.registerToolsHandler()

	return mcpServer, nil
}

// registerToolsHandler registers custom tools for the server
func (s *MCPServer) registerToolsHandler() {
	// Add a dummy tool to tell clients we're running in server mode
	dummyTool := mcp.NewTool("server_info",
		mcp.WithDescription("Get information about the MCP server"),
	)

	s.server.AddTool(dummyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("Running in MCP server mode"), nil
	})
}

// ServeHTTP serves the MCP over HTTP
func (s *MCPServer) ServeHTTP(addr string) error {
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// Process the request
			response, err := s.ProcessRequest(r.Context(), body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to process request: %v", err), http.StatusInternalServerError)
				return
			}

			// Write the response
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
		}),
	}

	// Start the server
	fmt.Fprintf(os.Stderr, "MCP Server listening on %s\n", addr)
	return server.ListenAndServe()
}

// ServeStdio serves the MCP over standard input/output
func (s *MCPServer) ServeStdio() error {
	// Start the stdio server
	return mcpserver.ServeStdio(s.server)
}

// ProcessRequest processes a raw MCP request
func (s *MCPServer) ProcessRequest(ctx context.Context, rawRequest []byte) ([]byte, error) {
	// Parse the request
	var request struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id"`
		Method  string      `json:"method"`
	}
	if err := json.Unmarshal(rawRequest, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Handle tools/list specially
	if request.Method == "tools/list" {
		return s.handleToolsList(ctx, request.ID)
	}

	// Handle tools/call specially
	if request.Method == "tools/call" {
		return s.handleToolsCall(ctx, request.ID, rawRequest)
	}

	// For other methods, let the server handle it
	// In a real implementation, you would create a function to handle the request directly
	// For now, we'll just return an error since we're not handling these methods yet
	return nil, fmt.Errorf("method not implemented: %s", request.Method)
}

// handleToolsList handles the tools/list method
func (s *MCPServer) handleToolsList(ctx context.Context, id interface{}) ([]byte, error) {
	// Get all tools from all MCPs
	tools := s.mcpManager.GetAllTools()

	// Create the response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}

	// Serialize the response
	return json.Marshal(response)
}

// handleToolsCall handles the tools/call method
func (s *MCPServer) handleToolsCall(ctx context.Context, id interface{}, rawRequest []byte) ([]byte, error) {
	// Parse the request parameters
	var request struct {
		Params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
	}
	if err := json.Unmarshal(rawRequest, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Execute the tool
	result, err := s.mcpManager.ExecuteTool(ctx, request.Params.Name, request.Params.Arguments)
	if err != nil {
		// Create an error response
		errorResponse := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]interface{}{
				"code":    -32000,
				"message": fmt.Sprintf("Failed to execute tool: %v", err),
			},
		}
		return json.Marshal(errorResponse)
	}

	// Create the success response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}

	// Serialize the response
	return json.Marshal(response)
}
