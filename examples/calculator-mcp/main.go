package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Calculator MCP",
		"1.0.0",
	)

	// Add calculator tools
	addTool := mcp.NewTool("add",
		mcp.WithDescription("Add two numbers"),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	multiplyTool := mcp.NewTool("multiply",
		mcp.WithDescription("Multiply two numbers"),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	divideTool := mcp.NewTool("divide",
		mcp.WithDescription("Divide two numbers"),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("Numerator"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Denominator"),
		),
	)

	// Add tool handlers
	s.AddTool(addTool, addHandler)
	s.AddTool(multiplyTool, multiplyHandler)
	s.AddTool(divideTool, divideHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func addHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	x, ok := request.Params.Arguments["x"].(float64)
	if !ok {
		return nil, fmt.Errorf("x must be a number")
	}

	y, ok := request.Params.Arguments["y"].(float64)
	if !ok {
		return nil, fmt.Errorf("y must be a number")
	}

	result := x + y
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

func multiplyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	x, ok := request.Params.Arguments["x"].(float64)
	if !ok {
		return nil, fmt.Errorf("x must be a number")
	}

	y, ok := request.Params.Arguments["y"].(float64)
	if !ok {
		return nil, fmt.Errorf("y must be a number")
	}

	result := x * y
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

func divideHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	x, ok := request.Params.Arguments["x"].(float64)
	if !ok {
		return nil, fmt.Errorf("x must be a number")
	}

	y, ok := request.Params.Arguments["y"].(float64)
	if !ok {
		return nil, fmt.Errorf("y must be a number")
	}

	if y == 0 {
		return nil, errors.New("cannot divide by zero")
	}

	result := x / y
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}
