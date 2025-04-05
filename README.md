# MCP Proxy

A proxy for the Model Context Protocol (MCP) that forwards requests to an HTTP endpoint.

## Overview

This MCP proxy forwards data from stdin to a specified HTTP endpoint and returns responses to stdout. It implements a simple proxy that can be used to communicate with remote MCP-compatible servers.

## Installation

Clone the repository and build the binary:

```bash
go build -o mcp-proxy
```

## Usage

```bash
./mcp-proxy [options]
```

### Options

- `-endpoint`: HTTP endpoint to proxy requests to (default: "http://localhost:8080")
- `-content-type`: Content-Type header for HTTP requests (default: "application/json")
- `-timeout`: HTTP request timeout in seconds (default: 30)
- `-buffer`: Buffer size in KB for reading from stdin (default: 64)

### Example

```bash
./mcp-proxy -endpoint="https://api.example.com/mcp" -content-type="application/json"
```

## Integration with MCP clients

MCP clients will send requests to stdin and expect responses on stdout. The proxy transparently forwards these requests to the HTTP endpoint and returns the responses to the client.

Example usage with an MCP client:

```bash
mcp-client | ./mcp-proxy -endpoint="https://api.example.com/mcp"
```

## How it works

1. The proxy reads MCP JSON-RPC messages from stdin
2. Each message is forwarded as a POST request to the specified HTTP endpoint
3. The response from the endpoint is written to stdout
4. The process continues until the proxy receives a termination signal (SIGINT or SIGTERM)

## Debugging

The proxy logs errors and startup information to stderr, allowing you to troubleshoot issues while keeping stdout clean for the actual MCP communication.

## License

MIT 