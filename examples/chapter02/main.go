package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// -- 1) Agent State
type AgentState struct {
	Order    map[string]any    `json:"order"`
	Messages []*schema.Message `json:"messages"`
}

// -- 2) Tool Implementation
func cancelOrder(orderID string) string {
	// In a real app, call backend API here
	return fmt.Sprintf("Order %s has been cancelled.", orderID)
}

func main() {
	// Load .env if present
	_ = godotenv.Load()

	ctx := context.Background()
	api_key := os.Getenv("OPENAI_API_KEY")
	base_url := os.Getenv("OPENAI_BASE_URL")

	if api_key == "" {
		fmt.Println("Error: OPENAI_API_KEY is not set")
		return
	}

	// -- 3) Initialize Model
	// We use temperature 0 for deterministic tool usage
	// temp := float32(0.0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  api_key,
		BaseURL: base_url,
		Model:   "gpt-5",
		// Temperature: &temp, // Model has beta limitations, must be 1 (default)
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	// -- 4) Define Tools
	// Schema for cancel_order
	cancelOrderTool := &schema.ToolInfo{
		Name: "cancel_order",
		Desc: "Cancel an order that hasn't shipped.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"order_id": {
				Type:     schema.String,
				Desc:     "The ID of the order to cancel",
				Required: true,
			},
		}),
	}

	// Bind tool to model
	if err := chatModel.BindTools([]*schema.ToolInfo{cancelOrderTool}); err != nil {
		log.Fatalf("Failed to bind tools: %v", err)
	}

	// -- 5) Define Graph Nodes

	// Node: Assistant (Calls LLM)
	assistant := compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		// Prepare System Prompt
		// In Eino, we can prepend a system message or rely on the state having one.
		// Use a dynamic system prompt based on order details if needed.
		orderID := "UNKNOWN"
		if oid, ok := state.Order["order_id"].(string); ok {
			orderID = oid
		}

		systemPrompt := fmt.Sprintf(
			"You are an e-commerce customer service agent.\n"+
				"Order ID: %s\n"+
				"If the customer asks to cancel the order, call cancel_order(order_id).\n"+
				"Then send a simple confirmation.\n"+
				"Otherwise, reply normally.",
			orderID,
		)

		// Construct messages for the LLM: System Prompt + History
		messages := []*schema.Message{schema.SystemMessage(systemPrompt)}
		messages = append(messages, state.Messages...)

		// Generate response
		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return nil, err
		}

		// Update state with new message
		state.Messages = append(state.Messages, resp)
		return state, nil
	})

	// Node: Tools (Executes Tools)
	toolsNode := compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) == 0 {
			return state, nil
		}

		for _, tc := range lastMsg.ToolCalls {
			if tc.Function.Name == "cancel_order" {
				// Parse arguments
				var args struct {
					OrderID string `json:"order_id"`
				}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					return nil, fmt.Errorf("failed to parse args: %v", err)
				}

				// Execute Tool
				result := cancelOrder(args.OrderID)

				// Append Tool Output Message
				state.Messages = append(state.Messages, &schema.Message{
					Role:       schema.Tool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
		}
		return state, nil
	})

	// -- 6) Construct Graph
	graph := compose.NewGraph[*AgentState, *AgentState]()

	_ = graph.AddLambdaNode("assistant", assistant)
	_ = graph.AddLambdaNode("tools", toolsNode)

	_ = graph.AddEdge(compose.START, "assistant")

	// Helper to check if we should go to tools
	shouldCallTool := func(_ context.Context, state *AgentState) (string, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return compose.END, nil
	}

	branch := compose.NewGraphBranch(shouldCallTool, map[string]bool{
		"tools":     true,
		compose.END: true,
	})

	_ = graph.AddBranch("assistant", branch)
	_ = graph.AddEdge("tools", "assistant") // Loop back after tool execution

	// Compile
	runnable, err := graph.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// -- 7) Run it
	initialOrder := map[string]interface{}{
		"order_id": "A12345",
	}
	initialMessages := []*schema.Message{
		schema.UserMessage("Please cancel my order A12345."),
	}

	initialState := &AgentState{
		Order:    initialOrder,
		Messages: initialMessages,
	}

	fmt.Println("Running Eino Agent...")
	finalState, err := runnable.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("Agent run failed: %v", err)
	}

	// Print Results
	for _, msg := range finalState.Messages {
		fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
	}

	// -- 8) Minimal Evaluation Check
	evalOrder := map[string]interface{}{
		"order_id": "B73973",
	}
	evalMessages := []*schema.Message{
		schema.UserMessage(`Please cancel order #B73973. 
        I found a cheaper option elsewhere.`),
	}

	evalState := &AgentState{
		Order:    evalOrder,
		Messages: evalMessages,
	}

	fmt.Println("\nRunning Minimal Evaluation Check...")
	evalResult, err := runnable.Invoke(ctx, evalState)
	if err != nil {
		log.Fatalf("Eval run failed: %v", err)
	}

	// Check 1: Tool called?
	hasToolCall := false
	for _, msg := range evalResult.Messages {
		for _, tc := range msg.ToolCalls {
			if tc.Function.Name == "cancel_order" {
				hasToolCall = true
				break
			}
		}
	}
	if !hasToolCall {
		log.Fatal("Eval Failed: Cancel order tool not called")
	}

	// Check 2: Confirmation message?
	hasConfirmation := false
	for _, msg := range evalResult.Messages {
		contentLower := strings.ToLower(msg.Content)
		if strings.Contains(contentLower, "cancel") || strings.Contains(msg.Content, "取消") {
			hasConfirmation = true
			break
		}
	}
	if !hasConfirmation {
		log.Fatal("Eval Failed: Confirmation message missing")
	}

	fmt.Println("✅ Agent passed minimal evaluation.")
}
