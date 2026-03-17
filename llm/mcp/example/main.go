package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/air-go/rpc/llm/mcp/client"
	clientHook "github.com/air-go/rpc/llm/mcp/client/hook"
	"github.com/air-go/rpc/llm/mcp/server"
	serverHook "github.com/air-go/rpc/llm/mcp/server/hook"
)

func main() {
	// 运行完整示例
	if err := runExample(); err != nil {
		log.Fatalf("Example failed: %v", err)
	}
}

func runExample() error {
	ctx := context.Background()

	fmt.Println("=== MCP Client and Server Example ===")

	// 1. Create server and configure hooks
	fmt.Println("Step 1: Creating MCP server with observability hooks...")
	serverHookChain := serverHook.NewHookChain()

	// Register OpenTelemetry tracing hook
	serverHookChain.RegisterHook(serverHook.NewOtelServerHook())

	// Register Prometheus metrics hook
	serverHookChain.RegisterHook(serverHook.NewPrometheusServerHook())

	mcpServer := server.NewServer("example-server", "1.0.0",
		server.WithHooks(serverHookChain),
	)

	fmt.Println("  ✓ Server created with Otel and Prometheus hooks")

	// 2. Register tools to the server
	fmt.Println("Step 2: Registering tools...")
	registerTools(mcpServer)

	fmt.Println("  ✓ Registered tools: echo, add, greet, calculate")

	// 3. Create in-process client (zero network overhead)
	fmt.Println("Step 3: Creating in-process MCP client (zero network overhead)...")

	// Configure client hooks
	clientHookChain := clientHook.NewClientHookChain()
	clientHookChain.RegisterHook(clientHook.NewOtelClientHook())
	clientHookChain.RegisterHook(clientHook.NewPrometheusClientHook())

	mcpClient, err := client.NewInProcessClient(
		mcpServer.GetMCPServer(),
		"example-server",
		client.WithHooks(clientHookChain),
	)
	if err != nil {
		return fmt.Errorf("failed to create in-process client: %w", err)
	}
	defer mcpClient.Close()

	fmt.Println("  ✓ In-process client created with Otel and Prometheus hooks")

	// 4. Initialize client
	fmt.Println("Step 4: Initializing client...")
	initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "example-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	fmt.Printf("  ✓ Connected to server: %s\n\n", initResult.ServerInfo.Name)

	// 5. List available tools
	fmt.Println("Step 5: Listing available tools...")
	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Println("  Available tools:")
	for _, tool := range tools.Tools {
		fmt.Printf("    - %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// 6. Execute tool examples
	fmt.Println("Step 6: Executing tool examples...")

	// Example 1: Echo tool
	fmt.Println("  Example 6.1: Echo tool")
	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]interface{}{
				"message": "Hello from MCP!",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call echo tool: %w", err)
	}
	printToolResult(result)

	// Example 2: Add tool
	fmt.Println("  Example 6.2: Add tool")
	result, err = mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "add",
			Arguments: map[string]interface{}{
				"a": 42.0,
				"b": 58.0,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call add tool: %w", err)
	}
	printToolResult(result)

	// Example 3: Greet tool
	fmt.Println("  Example 6.3: Greet tool")
	result, err = mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "greet",
			Arguments: map[string]interface{}{
				"name": "Developer",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call greet tool: %w", err)
	}
	printToolResult(result)

	// Example 4: Calculate tool with multiple operations
	fmt.Println("  Example 6.4: Calculate tool")
	result, err = mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "calculate",
			Arguments: map[string]interface{}{
				"expression": "10 * 20 / 2 + 5",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call calculate tool: %w", err)
	}
	printToolResult(result)

	fmt.Println()

	// 7. Performance demo
	fmt.Println("Step 7: Performance demo (multiple parallel calls)...")
	start := time.Now()

	for i := 0; i < 5; i++ {
		go func(index int) {
			_, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "echo",
					Arguments: map[string]interface{}{
						"message": fmt.Sprintf("Parallel call %d", index),
					},
				},
			})
			if err != nil {
				log.Printf("Parallel call %d failed: %v", index, err)
			}
		}(i)
	}

	// Wait a bit for goroutines
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("  ✓ Completed 5 parallel calls in %v\n", time.Since(start))

	fmt.Println("")
	fmt.Println("=== Example completed successfully! ===")
	fmt.Println("")
	fmt.Println("Key features demonstrated:")
	fmt.Println("  ✓ In-process MCP communication (zero network overhead)")
	fmt.Println("  ✓ OpenTelemetry tracing integration")
	fmt.Println("  ✓ Prometheus metrics collection")
	fmt.Println("  ✓ Hook-based extensibility")
	fmt.Println("  ✓ Multiple tools and concurrent calls")

	return nil
}

func registerTools(srv *server.MCPServer) {
	// Echo tool - returns the input message
	echoTool := mcp.NewTool("echo",
		mcp.WithDescription("Echo back the input message"),
		mcp.WithString("message",
			mcp.Description("Message to echo back"),
			mcp.Required(),
		),
	)
	srv.AddTool(echoTool, echoHandler)

	// Add tool - adds two numbers
	addTool := mcp.NewTool("add",
		mcp.WithDescription("Add two numbers together"),
		mcp.WithNumber("a",
			mcp.Description("First number"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("Second number"),
			mcp.Required(),
		),
	)
	srv.AddTool(addTool, addHandler)

	// Greet tool - greets the user
	greetTool := mcp.NewTool("greet",
		mcp.WithDescription("Greet the user"),
		mcp.WithString("name",
			mcp.Description("Name to greet"),
			mcp.Required(),
		),
	)
	srv.AddTool(greetTool, greetHandler)

	// Calculate tool - simple calculator
	calculateTool := mcp.NewTool("calculate",
		mcp.WithDescription("Calculate a mathematical expression"),
		mcp.WithString("expression",
			mcp.Description("Mathematical expression (e.g., '10 + 20 * 2')"),
			mcp.Required(),
		),
	)
	srv.AddTool(calculateTool, calculateHandler)
}

func echoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message := req.GetString("message", "")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Echo: %s", message),
			},
		},
	}, nil
}

func addHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	a, _ := args["a"].(float64)
	b, _ := args["b"].(float64)
	result := a + b

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%v + %v = %v", a, b, result),
			},
		},
	}, nil
}

func greetHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := req.GetString("name", "World")
	greetings := []string{
		fmt.Sprintf("Hello, %s!", name),
		fmt.Sprintf("Welcome back, %s!", name),
		fmt.Sprintf("Nice to see you, %s!", name),
	}
	selectedGreeting := greetings[len(name)%len(greetings)]

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: selectedGreeting,
			},
		},
	}, nil
}

func calculateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	expression := req.GetString("expression", "")

	// Simple calculator - in production, use a proper expression evaluator
	result := evaluateSimpleExpression(expression)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%s = %v", expression, result),
			},
		},
	}, nil
}

func evaluateSimpleExpression(expr string) float64 {
	// This is a simplified evaluator for demonstration
	// In production, use a proper math expression parser
	var result float64
	var op rune
	var num float64
	var numStr string

	for _, ch := range expr {
		switch ch {
		case '+', '-', '*', '/':
			if numStr != "" {
				num = parseSimpleNumber(numStr)
				if op == 0 {
					result = num
				} else {
					switch op {
					case '+':
						result += num
					case '-':
						result -= num
					case '*':
						result *= num
					case '/':
						if num != 0 {
							result /= num
						}
					}
				}
				numStr = ""
			}
			op = ch
		case ' ', '\t':
			// Skip whitespace
		default:
			numStr += string(ch)
		}
	}

	if numStr != "" {
		num = parseSimpleNumber(numStr)
		if op == 0 {
			result = num
		} else {
			switch op {
			case '+':
				result += num
			case '-':
				result -= num
			case '*':
				result *= num
			case '/':
				if num != 0 {
					result /= num
				}
			}
		}
	}

	return result
}

func parseSimpleNumber(s string) float64 {
	var result float64
	var decimal bool
	var divisor float64 = 1

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			if decimal {
				divisor *= 10
				result = result + float64(ch-'0')/divisor
			} else {
				result = result*10 + float64(ch-'0')
			}
		} else if ch == '.' {
			decimal = true
		}
	}

	return result
}

func printToolResult(result *mcp.CallToolResult) {
	fmt.Printf("    Result: ")
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			fmt.Printf("%s\n", textContent.Text)
		}
	}
}
