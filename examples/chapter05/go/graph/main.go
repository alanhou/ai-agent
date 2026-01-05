package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// State represents the workflow state
type State struct {
	UserMessage string
	UserID      string
	IssueType   string
	StepResult  string
	Response    string
}

var llm *openai.ChatModel

func main() {
	// Load environment variables
	_ = godotenv.Load()

	ctx := context.Background()

	temp := float32(0)

	// Initialize LLM
	var err error
	llm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o-mini",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
		Temperature: &temp,
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Build the graph using eino's native Graph
	graph := compose.NewGraph[*State, *State]()

	// Add all nodes as Lambda functions
	graph.AddLambdaNode("categorize_issue", compose.InvokableLambda(categorizeIssue))
	graph.AddLambdaNode("handle_invoice", compose.InvokableLambda(handleInvoice))
	graph.AddLambdaNode("handle_refund", compose.InvokableLambda(handleRefund))
	graph.AddLambdaNode("handle_login", compose.InvokableLambda(handleLogin))
	graph.AddLambdaNode("handle_performance", compose.InvokableLambda(handlePerformance))
	graph.AddLambdaNode("summarize_response", compose.InvokableLambda(summarizeResponse))

	// Add edge from START to categorize_issue
	graph.AddEdge(compose.START, "categorize_issue")

	// Add conditional branch from categorize_issue
	topBranch := compose.NewGraphBranch(topRouter, map[string]bool{
		"handle_invoice": true,
		"handle_login":   true,
	})
	graph.AddBranch("categorize_issue", topBranch)

	// Add conditional branch from handle_invoice
	billingBranch := compose.NewGraphBranch(billingRouter, map[string]bool{
		"summarize_response": true,
		"handle_refund":      true,
	})
	graph.AddBranch("handle_invoice", billingBranch)

	// Add conditional branch from handle_login
	techBranch := compose.NewGraphBranch(techRouter, map[string]bool{
		"summarize_response": true,
		"handle_performance": true,
	})
	graph.AddBranch("handle_login", techBranch)

	// Add direct edges to summarize_response
	graph.AddEdge("handle_refund", "summarize_response")
	graph.AddEdge("handle_performance", "summarize_response")

	// Add edge from summarize_response to END
	graph.AddEdge("summarize_response", compose.END)

	// Compile the graph
	runnable, err := graph.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// Execute the graph
	initialState := &State{
		UserMessage: "Hi, I need help with my invoice and possibly a refund.",
		UserID:      "U1234",
	}

	result, err := runnable.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("Failed to invoke graph: %v", err)
	}

	fmt.Println(result.Response)
}

// Node functions
func categorizeIssue(ctx context.Context, state *State) (*State, error) {
	prompt := fmt.Sprintf("Classify this support request as 'billing' or 'technical'.\n\nMessage: %s", state.UserMessage)
	messages := []*schema.Message{schema.UserMessage(prompt)}

	resp, err := llm.Generate(ctx, messages)
	if err != nil {
		return state, err
	}

	state.IssueType = strings.ToLower(strings.TrimSpace(resp.Content))
	return state, nil
}

func handleInvoice(ctx context.Context, state *State) (*State, error) {
	state.StepResult = fmt.Sprintf("Invoice details for %s", state.UserID)
	return state, nil
}

func handleRefund(ctx context.Context, state *State) (*State, error) {
	state.StepResult = "Refund process initiated"
	return state, nil
}

func handleLogin(ctx context.Context, state *State) (*State, error) {
	state.StepResult = "Password reset link sent"
	return state, nil
}

func handlePerformance(ctx context.Context, state *State) (*State, error) {
	state.StepResult = "Performance metrics analyzed"
	return state, nil
}

func summarizeResponse(ctx context.Context, state *State) (*State, error) {
	details := state.StepResult
	prompt := fmt.Sprintf("Write a concise customer reply based on: %s", details)
	messages := []*schema.Message{schema.UserMessage(prompt)}

	resp, err := llm.Generate(ctx, messages)
	if err != nil {
		return state, err
	}

	state.Response = strings.TrimSpace(resp.Content)
	return state, nil
}

// Router functions (GraphBranchCondition type)
func topRouter(ctx context.Context, state *State) (string, error) {
	if state.IssueType == "billing" {
		return "handle_invoice", nil
	}
	return "handle_login", nil
}

func billingRouter(ctx context.Context, state *State) (string, error) {
	msg := strings.ToLower(state.UserMessage)
	if strings.Contains(msg, "invoice") {
		return "summarize_response", nil
	}
	return "handle_refund", nil
}

func techRouter(ctx context.Context, state *State) (string, error) {
	msg := strings.ToLower(state.UserMessage)
	if strings.Contains(msg, "login") {
		return "summarize_response", nil
	}
	return "handle_performance", nil
}


