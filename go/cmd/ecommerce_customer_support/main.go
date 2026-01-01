package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"agents-go/go/internal/scenarios/ecommerce_customer_support"

	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	evalMode := flag.Bool("eval", false, "Run in evaluation mode (JSON stdin/stdout)")
	flag.Parse()

	if os.Getenv("OPENAI_API_KEY") == "" {
		// Log to stderr in eval mode
		fmt.Fprintf(os.Stderr, "OPENAI_API_KEY is not set\n")
		os.Exit(1)
	}

	ctx := context.Background()
	agent, err := ecommerce_customer_support.NewAgent(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	var initialState *ecommerce_customer_support.AgentState

	if *evalMode {
		// Read from stdin
		inputBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
			os.Exit(1)
		}

		initialState = &ecommerce_customer_support.AgentState{}
		if err := json.Unmarshal(inputBytes, initialState); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal input: %v\nInput: %s\n", err, string(inputBytes))
			os.Exit(1)
		}
	} else {
		// Demo mode with hardcoded inputs
		initialState = &ecommerce_customer_support.AgentState{
			Order: &ecommerce_customer_support.Order{
				OrderID:    "A12345",
				Status:     "Delivered",
				Total:      19.99,
				CustomerID: "CUST001",
			},
			Messages: []*schema.Message{
				schema.UserMessage("My mug arrived broken. Refund?"),
			},
		}
	}

	// Run Agent
	finalState, err := agent.Invoke(ctx, initialState)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}

	if *evalMode {
		// Output JSON to stdout
		outputBytes, err := json.Marshal(finalState)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(outputBytes))
	} else {
		// Print Result Human Readable
		fmt.Println("Final Messages:")
		for _, msg := range finalState.Messages {
			fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
			for _, tc := range msg.ToolCalls {
				fmt.Printf("  (Tool Call: %s args=%s)\n", tc.Function.Name, tc.Function.Arguments)
			}
		}
	}
}
