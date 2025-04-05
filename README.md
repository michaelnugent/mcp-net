# MCP Tools

This repository contains tools for working with the Model Context Protocol (MCP).

## MCP Proxy

A proxy for the Model Context Protocol that forwards requests to an HTTP endpoint.

### Overview

The MCP proxy forwards data from stdin to a specified HTTP endpoint and returns responses to stdout. It implements a simple proxy that can be used to communicate with remote MCP-compatible servers.

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

### Example MCPs

The `examples` directory contains sample MCPs that can be used for testing:

- `hello-mcp`: A simple MCP that provides a "hello" tool
- `calculator-mcp`: An MCP that provides math operations

To build the examples:

```bash
cd examples/hello-mcp
go build -o hello-mcp
cd ../calculator-mcp
go build -o calculator-mcp
```

Then copy the built executables to your MCP directory:

```bash
mkdir -p mcps
cp examples/hello-mcp/hello-mcp mcps/
cp examples/calculator-mcp/calculator-mcp mcps/
```

## Building

To build the proxy and server:

```bash
go build -o mcp-proxy ./cmd/mcp-proxy
go build -o mcp-server ./cmd/mcp-server
```

## License

MIT 