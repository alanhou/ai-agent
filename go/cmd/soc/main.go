package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"agents-go/go/internal/scenarios/soc"

	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	evalMode := flag.Bool("eval", false, "Run in evaluation mode (JSON stdin/stdout)")
	flag.Parse()

	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Fprintf(os.Stderr, "OPENAI_API_KEY is not set\n")
		os.Exit(1)
	}

	ctx := context.Background()
	agent, err := soc.NewAgent(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	var initialState *soc.AgentState

	if *evalMode {
		inputBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
			os.Exit(1)
		}

		initialState = &soc.AgentState{}
		if err := json.Unmarshal(inputBytes, initialState); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal input: %v\nInput: %s\n", err, string(inputBytes))
			os.Exit(1)
		}
	} else {
		// Demo mode
		initialState = &soc.AgentState{
			Incident: &soc.Incident{
				IncidentID: "INC-888",
				Severity:   "high",
				Type:       "Malware",
			},
			Messages: []*schema.Message{
				schema.UserMessage("Analyze this file hash: abc123def456"),
			},
		}
	}

	finalState, err := agent.Invoke(ctx, initialState)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}

	if *evalMode {
		outputBytes, err := json.Marshal(finalState)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(outputBytes))
	} else {
		fmt.Println("Final Messages:")
		for _, msg := range finalState.Messages {
			fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
			for _, tc := range msg.ToolCalls {
				fmt.Printf("  (Tool Call: %s args=%s)\n", tc.Function.Name, tc.Function.Arguments)
			}
		}
	}
}
