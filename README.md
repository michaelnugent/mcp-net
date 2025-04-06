# MCP Tools

This repository contains tools for working with the Model Context Protocol (MCP).

## MCP Proxy

A proxy for the Model Context Protocol that forwards requests to an HTTP endpoint.

### Overview

The MCP proxy forwards data from stdin to a specified HTTP endpoint and returns responses to stdout. It implements a simple proxy that can be used to communicate with remote MCP-compatible servers.

### Why golang?

This code runs on the client side.  Many languages require a runtime and correct set of libraries installed to do this.  Go compiles
to a single executable file.  This cuts down on deployment difficulties, user support requests, breaking updates to the runtime
or other issues that are usually taken care of by providing an http endpoint instead of deployment.

### Usage

```bash
./mcp-proxy [options]
```

#### Options

- `-endpoint`: HTTP endpoint to proxy requests to (default: "http://localhost:8080")
- `-content-type`: Content-Type header for HTTP requests (default: "application/json")
- `-timeout`: HTTP request timeout in seconds (default: 30)
- `-buffer`: Buffer size in KB for reading from stdin (default: 64)

### Example

```bash
./mcp-proxy -endpoint="https://api.example.com/mcp" -content-type="application/json"
```

## MCP Server

A server that loads and manages MCP executables from a directory and exposes them to clients via HTTP or stdio.

### Overview

The MCP server loads MCP executables from a directory and serves them to clients. When clients request a list of tools, the server returns all tools from all loaded MCPs, namespaced by the MCP name. The server can run in either HTTP mode or stdio mode.

### Usage

```bash
./mcp-server [options]
```

#### Options

- `-mcp-dir`: Directory containing MCP executables (default: "./mcps")
- `-http`: HTTP server address (default: ":8080")
- `-name`: Name of the MCP server (default: "MCP Server")
- `-version`: Version of the MCP server (default: "1.0.0")
- `-stdio`: Use stdio instead of HTTP (default: false)

### MCP Directory Structure

The server expects a directory containing MCP executables. Each executable must implement the MCP protocol using stdio. The server will:

1. Scan the directory for executable files
2. Run each executable to discover the tools it provides
3. Make these tools available to clients with namespaced names (`mcpname.toolname`)

## Building and Running with Make

This project includes a Makefile that simplifies building and running the components.

### Building

Build all components (proxy, server, and examples):
```bash
make
```

Or build specific components:
```bash
make build-proxy    # Build only the proxy
make build-server   # Build only the server
make build-examples # Build only the example MCPs
```

### Running

Start the MCP server in HTTP mode (after building examples and copying them to the mcps directory):
```bash
make run-server
```

Start the MCP server in stdio mode:
```bash
make run-server-stdio
```

Run the proxy to connect to the local MCP server:
```bash
make run-proxy
```

### Testing

Test the proxy with a basic tools/list request (requires server running):
```bash
make test-proxy
```

Test specific example tools through the proxy (requires server running):
```bash
make test-hello   # Test the hello tool
make test-add     # Test the calculator's add tool
```

### Example MCPs

The repository includes example MCPs that demonstrate how to implement MCP-compatible tools:

- `hello-mcp`: A simple MCP that provides a "hello" tool
- `calculator-mcp`: An MCP that provides math operations

Build and install the examples to the mcps directory:
```bash
make examples
```

### Development Tasks

Format Go code:
```bash
make fmt
```

Run Go vet for code analysis:
```bash
make vet
```

Update and tidy Go dependencies:
```bash
make tidy
```

Clean build artifacts:
```bash
make clean
```

Install binaries to GOPATH/bin:
```bash
make install
```

## License

MIT 