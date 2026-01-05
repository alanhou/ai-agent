package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// State represents the state of the conversation
type State struct {
	Messages []*schema.Message
}

var llm *openai.ChatModel

// callModel invokes the LLM with the current messages
func callModel(ctx context.Context, state *State) (*State, error) {
	resp, err := llm.Generate(ctx, state.Messages)
	if err != nil {
		return state, err
	}
	// Append the assistant's response to the messages
	state.Messages = append(state.Messages, resp)
	return state, nil
}

func main() {
	// Load .env from project root
	_ = godotenv.Load()
	ctx := context.Background()

	// Initialize ChatModel
	var err error
	llm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o-mini", // Using 4o-mini as a proxy for the python example's gpt-5 or standard model
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// ==========================================
	// Part 1: Stateless Interaction (No Memory)
	// ==========================================
	fmt.Println("--- Stateless Interaction ---")

	// Build graph
	builder := compose.NewGraph[*State, *State]()
	builder.AddLambdaNode("call_model", compose.InvokableLambda(callModel))
	builder.AddEdge(compose.START, "call_model")
	builder.AddEdge("call_model", compose.END)

	graph, err := builder.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// 1. "Hi! I'm Bob"
	input1 := &State{
		Messages: []*schema.Message{
			schema.UserMessage("hi! I'm bob"),
		},
	}
	out1, err := graph.Invoke(ctx, input1)
	if err != nil {
		log.Fatalf("Stateless 1 failed: %v", err)
	}
	printLastMessage(out1.Messages)

	// 2. "What's my name?" (Should not know)
	input2 := &State{
		Messages: []*schema.Message{
			schema.UserMessage("what's my name?"),
		},
	}
	out2, err := graph.Invoke(ctx, input2)
	if err != nil {
		log.Fatalf("Stateless 2 failed: %v", err)
	}
	printLastMessage(out2.Messages)

	// ==========================================
	// Part 2: Stateful Interaction (Simulated Memory)
	// ==========================================
	fmt.Println("\n--- Stateful Interaction (Simulated Memory) ---")

	// In a real application, this would be a database.
	// We simulate LangGraph's MemorySaver/checkpointer.
	memoryStore := make(map[string][]*schema.Message)
	threadID := "1"

	// Helper to load/save state
	loadState := func(threadID string) []*schema.Message {
		if msgs, ok := memoryStore[threadID]; ok {
			return msgs
		}
		return []*schema.Message{}
	}
	saveState := func(threadID string, msgs []*schema.Message) {
		memoryStore[threadID] = msgs
	}

	// 1. "Hi! I'm Bob"
	// Load existing (empty)
	currentMsgs := loadState(threadID)
	// Add user message
	currentMsgs = append(currentMsgs, schema.UserMessage("hi! I'm bob"))

	inputState1 := &State{Messages: currentMsgs}
	outState1, err := graph.Invoke(ctx, inputState1)
	if err != nil {
		log.Fatalf("Stateful 1 failed: %v", err)
	}

	// Save updated history (including assistant response)
	saveState(threadID, outState1.Messages)
	printLastMessage(outState1.Messages)

	// 2. "What's my name?" (Should now know "Bob")
	// Load existing (contains "hi I'm bob" conversation)
	currentMsgs2 := loadState(threadID)
	// Add user message
	currentMsgs2 = append(currentMsgs2, schema.UserMessage("what's my name?"))

	inputState2 := &State{Messages: currentMsgs2}
	outState2, err := graph.Invoke(ctx, inputState2)
	if err != nil {
		log.Fatalf("Stateful 2 failed: %v", err)
	}

	// Save updated history
	saveState(threadID, outState2.Messages)
	printLastMessage(outState2.Messages)
}

func printLastMessage(msgs []*schema.Message) {
	if len(msgs) == 0 {
		return
	}
	last := msgs[len(msgs)-1]
	fmt.Printf("[%s]: %s\n", last.Role, last.Content)
}
