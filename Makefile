# Makefile for MCP-Net

# Go parameters
GO := go
GOFLAGS :=
GOFMT := gofmt
BUILD_DIR := build
MCPS_DIR := mcps

# Binary names
MCP_PROXY := mcp-proxy
MCP_SERVER := mcp-server
HELLO_MCP := hello-mcp
CALCULATOR_MCP := calculator-mcp

# Source directories
CMD_DIR := cmd
EXAMPLES_DIR := examples
SERVER_DIR := server

# Main targets
.PHONY: all clean build build-proxy build-server build-examples examples test run-server run-proxy fmt vet tidy install

all: clean build

# Build all binaries
build: build-proxy build-server build-examples

# Build proxy
build-proxy:
	@echo "Building MCP proxy..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(MCP_PROXY) ./$(CMD_DIR)/$(MCP_PROXY)

# Build server
build-server:
	@echo "Building MCP server..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(MCP_SERVER) ./$(CMD_DIR)/$(MCP_SERVER)

# Build example MCPs
build-examples:
	@echo "Building example MCPs..."
	@mkdir -p $(BUILD_DIR)/$(EXAMPLES_DIR)/$(HELLO_MCP)
	@mkdir -p $(BUILD_DIR)/$(EXAMPLES_DIR)/$(CALCULATOR_MCP)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(EXAMPLES_DIR)/$(HELLO_MCP)/$(HELLO_MCP) ./$(EXAMPLES_DIR)/$(HELLO_MCP)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(EXAMPLES_DIR)/$(CALCULATOR_MCP)/$(CALCULATOR_MCP) ./$(EXAMPLES_DIR)/$(CALCULATOR_MCP)

# Shortcut for building and installing examples to mcps directory
examples: build-examples
	@echo "Installing example MCPs to $(MCPS_DIR) directory..."
	@mkdir -p $(MCPS_DIR)
	@cp $(BUILD_DIR)/$(EXAMPLES_DIR)/$(HELLO_MCP)/$(HELLO_MCP) $(MCPS_DIR)/
	@cp $(BUILD_DIR)/$(EXAMPLES_DIR)/$(CALCULATOR_MCP)/$(CALCULATOR_MCP) $(MCPS_DIR)/
	@chmod +x $(MCPS_DIR)/*

# Run server in HTTP mode
run-server: build-server examples
	@echo "Running MCP server in HTTP mode..."
	@$(BUILD_DIR)/$(MCP_SERVER) -http=:8080 -mcp-dir=$(MCPS_DIR)

# Run server in stdio mode
run-server-stdio: build-server examples
	@echo "Running MCP server in stdio mode..."
	@$(BUILD_DIR)/$(MCP_SERVER) -stdio -mcp-dir=$(MCPS_DIR)

# Run proxy to local server
run-proxy: build-proxy
	@echo "Running MCP proxy to localhost:8080..."
	@$(BUILD_DIR)/$(MCP_PROXY) -endpoint=http://localhost:8080

# Test proxy with a sample request to the server (requires server running)
test-proxy: build-proxy
	@echo "Testing MCP proxy with sample request..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | $(BUILD_DIR)/$(MCP_PROXY) -endpoint=http://localhost:8080

# Run a hello world request through proxy to server (requires server running)
test-hello: build-proxy
	@echo "Testing hello tool via proxy..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"hello-mcp.hello","arguments":{"name":"World"}}}' | $(BUILD_DIR)/$(MCP_PROXY) -endpoint=http://localhost:8080

# Run a calculator add request through proxy to server (requires server running)
test-add: build-proxy
	@echo "Testing calculator add tool via proxy..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"calculator-mcp.add","arguments":{"x":10,"y":20}}}' | $(BUILD_DIR)/$(MCP_PROXY) -endpoint=http://localhost:8080

# Clean generated files
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f mcps/$(HELLO_MCP) mcps/$(CALCULATOR_MCP)
	@rm -f $(EXAMPLES_DIR)/$(HELLO_MCP)/$(HELLO_MCP)
	@rm -f $(EXAMPLES_DIR)/$(CALCULATOR_MCP)/$(CALCULATOR_MCP)
	@rm -f $(MCP_PROXY) $(MCP_SERVER)
	@echo "Clean complete"

# Format code
fmt:
	$(GOFMT) -w ./$(CMD_DIR)
	$(GOFMT) -w ./$(SERVER_DIR)
	$(GOFMT) -w ./$(EXAMPLES_DIR)

# Vet code
vet:
	$(GO) vet ./...

# Update and tidy dependencies
tidy:
	$(GO) mod tidy

# Install binaries to $GOPATH/bin
install: build
	@echo "Installing mcp-proxy and mcp-server to GOPATH..."
	@cp $(BUILD_DIR)/$(MCP_PROXY) $(GOPATH)/bin/
	@cp $(BUILD_DIR)/$(MCP_SERVER) $(GOPATH)/bin/
