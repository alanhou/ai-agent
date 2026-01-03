package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Testing Go MCP Tools ===\n")

	// Test 1: Connect to Math server (stdio - spawn subprocess)
	fmt.Println("Test 1: Math query (3 + 5)")
	mathResult, err := testMathServer(ctx, "3 + 5")
	if err != nil {
		log.Printf("Math test failed: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n\n", mathResult)
	}

	// Test 2: Complex math
	fmt.Println("Test 2: Complex math ((10 - 2) * 4)")
	mathResult2, err := testMathServer(ctx, "(10 - 2) * 4")
	if err != nil {
		log.Printf("Math test 2 failed: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n\n", mathResult2)
	}

	// Test 3: Connect to Weather server (streamable HTTP)
	fmt.Println("Test 3: Weather query (Beijing)")
	fmt.Println("Note: Make sure the weather server is running on port 8001")
	weatherResult, err := testWeatherServer(ctx, "Beijing")
	if err != nil {
		log.Printf("Weather test failed: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n\n", weatherResult)
	}

	// Test 4: Routing logic
	fmt.Println("Test 4: Query routing test")
	testRouting()

	fmt.Println("=== Tests Complete ===")
}

// testMathServer connects to the math MCP server via stdio and calls the calculate tool
func testMathServer(ctx context.Context, expression string) (string, error) {
	// Find the math server command
	cmd := findServerCommand("./go/cmd/mcp_servers/math")

	// Create stdio client that spawns the math server as a subprocess
	c, err := client.NewStdioMCPClient(cmd[0], nil, cmd[1:]...)
	if err != nil {
		return "", fmt.Errorf("failed to create stdio client: %w", err)
	}
	defer c.Close()

	// Initialize the client
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "go-mcp-test-client",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		return "", fmt.Errorf("failed to initialize: %w", err)
	}

	// Call the calculate tool
	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = "calculate"
	callReq.Params.Arguments = map[string]interface{}{
		"expression": expression,
	}

	result, err := c.CallTool(ctx, callReq)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %w", err)
	}

	// Extract text from result
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in result")
}

// testWeatherServer connects to the weather MCP server via streamable HTTP
func testWeatherServer(ctx context.Context, city string) (string, error) {
	// Create streamable HTTP client
	c, err := client.NewStreamableHttpClient("http://localhost:8001/mcp")
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}
	defer c.Close()

	// Start the client with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := c.Start(ctxWithTimeout); err != nil {
		return "", fmt.Errorf("failed to start HTTP client: %w", err)
	}

	// Initialize the client
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "go-mcp-test-client",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		return "", fmt.Errorf("failed to initialize: %w", err)
	}

	// Call the get_weather tool
	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = "get_weather"
	callReq.Params.Arguments = map[string]interface{}{
		"city": city,
	}

	result, err := c.CallTool(ctx, callReq)
	if err != nil {
		return "", fmt.Errorf("failed to call tool: %w", err)
	}

	// Extract text from result
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in result")
}

// findServerCommand returns the command to run a server
func findServerCommand(serverPath string) []string {
	// Try to find pre-built binary first (based on path)
	binaryName := ""
	if strings.Contains(serverPath, "math") {
		binaryName = "./bin/mcp_math_server"
	} else if strings.Contains(serverPath, "weather") {
		binaryName = "./bin/mcp_weather_server"
	}

	if binaryName != "" {
		if _, err := os.Stat(binaryName); err == nil {
			return []string{binaryName}
		}
	}

	// Fall back to go run
	goPath, err := exec.LookPath("go")
	if err != nil {
		log.Fatal("go command not found")
	}
	return []string{goPath, "run", serverPath}
}

// routeQuery determines which tool to use based on the query content
func routeQuery(query string) string {
	queryLower := strings.ToLower(query)

	// Check for math operators
	mathTokens := []string{"+", "-", "*", "/", "(", ")"}
	for _, token := range mathTokens {
		if strings.Contains(queryLower, token) {
			return "calculate"
		}
	}

	// Check for weather keywords
	if strings.Contains(queryLower, "weather") {
		return "get_weather"
	}

	return "unknown"
}

// testRouting tests the query routing logic
func testRouting() {
	testCases := []struct {
		query    string
		expected string
	}{
		{"3 + 5", "calculate"},
		{"(10 - 2) * 4", "calculate"},
		{"What's the weather in Beijing?", "get_weather"},
		{"Tell me a joke", "unknown"},
	}

	for _, tc := range testCases {
		result := routeQuery(tc.query)
		status := "✓"
		if result != tc.expected {
			status = "✗"
		}
		fmt.Printf("  %s Query: %q -> %s (expected: %s)\n", status, tc.query, result, tc.expected)
	}
	fmt.Println()
}
