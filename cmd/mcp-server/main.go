package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mcp-net/mcp-proxy/server"
)

func main() {
	// Define command line flags
	mcpDirectory := flag.String("mcp-dir", "./mcps", "Directory containing MCP executables")
	httpAddr := flag.String("http", ":8080", "HTTP server address")
	name := flag.String("name", "MCP Server", "Name of the MCP server")
	version := flag.String("version", "1.0.0", "Version of the MCP server")
	useStdio := flag.Bool("stdio", false, "Use stdio instead of HTTP")
	flag.Parse()

	// Ensure the MCP directory exists
	if _, err := os.Stat(*mcpDirectory); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "MCP directory does not exist: %s\n", *mcpDirectory)
		if err := os.MkdirAll(*mcpDirectory, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create MCP directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Created MCP directory: %s\n", *mcpDirectory)
	}

	// Get the absolute path of the MCP directory
	absPath, err := filepath.Abs(*mcpDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get absolute path of MCP directory: %v\n", err)
		os.Exit(1)
	}

	// Create the MCP server
	mcpServer, err := server.NewMCPServer(absPath, *name, *version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP server: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Fprintf(os.Stderr, "Received signal %v, shutting down...\n", sig)
		os.Exit(0)
	}()

	// Start the server
	var serverErr error
	if *useStdio {
		fmt.Fprintf(os.Stderr, "Starting MCP server in stdio mode\n")
		serverErr = mcpServer.ServeStdio()
	} else {
		fmt.Fprintf(os.Stderr, "Starting MCP server in HTTP mode on %s\n", *httpAddr)
		serverErr = mcpServer.ServeHTTP(*httpAddr)
	}

	if serverErr != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", serverErr)
		os.Exit(1)
	}
}
