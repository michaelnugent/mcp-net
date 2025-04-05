package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// MCPProxy handles forwarding MCP (Model Context Protocol) requests to an HTTP endpoint
type MCPProxy struct {
	httpEndpoint string
	contentType  string
	httpClient   *http.Client
	mu           sync.Mutex // protects concurrent access to the proxy
}

// NewMCPProxy creates a new MCP proxy with the specified endpoint and content type
func NewMCPProxy(httpEndpoint, contentType string, timeoutSeconds int) *MCPProxy {
	return &MCPProxy{
		httpEndpoint: httpEndpoint,
		contentType:  contentType,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// ProcessRequest forwards a request to the HTTP endpoint and returns the response
func (p *MCPProxy) ProcessRequest(ctx context.Context, request []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.httpEndpoint, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", p.contentType)

	// Send the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func main() {
	// Define command line flags
	endpoint := flag.String("endpoint", "http://localhost:8080", "HTTP endpoint to proxy requests to")
	contentType := flag.String("content-type", "application/json", "Content-Type header for HTTP requests")
	timeout := flag.Int("timeout", 30, "HTTP request timeout in seconds")
	bufferSize := flag.Int("buffer", 64, "Buffer size in KB for reading from stdin")
	flag.Parse()

	// Create a new proxy
	proxy := NewMCPProxy(*endpoint, *contentType, *timeout)

	// Set up a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Fprintf(os.Stderr, "Received signal %v, shutting down...\n", sig)
		cancel()
	}()

	fmt.Fprintf(os.Stderr, "MCP Proxy started. Forwarding requests to %s\n", *endpoint)

	// Process stdin/stdout in the main goroutine
	stdin := os.Stdin
	stdout := os.Stdout

	// Create a buffer for reading from stdin
	buffer := make([]byte, *bufferSize*1024)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "MCP Proxy shutting down\n")
			return
		default:
			// Read from stdin
			n, err := stdin.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				}
				cancel()
				return
			}

			if n > 0 {
				// Process the request
				response, err := proxy.ProcessRequest(ctx, buffer[:n])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error processing request: %v\n", err)
					continue
				}

				// Write the response to stdout
				_, err = stdout.Write(response)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
					cancel()
					return
				}
			}
		}
	}
}
