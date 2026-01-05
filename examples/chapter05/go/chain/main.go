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

func main() {
	// Load environment variables
	_ = godotenv.Load()

	ctx := context.Background()

	// Initialize ChatModel
	temp := float32(0)
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o-mini",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
		Temperature: &temp,
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create prompt template as a Lambda
	promptLambda := compose.InvokableLambda(func(ctx context.Context, input map[string]any) ([]*schema.Message, error) {
		question := input["input"].(string)
		return []*schema.Message{
			schema.UserMessage(fmt.Sprintf("Answer this question: %s", question)),
		}, nil
	})

	// Create chain: prompt | llm (LCEL-style)
	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.AppendLambda(promptLambda).
		AppendChatModel(llm)

	// Compile the chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("Failed to compile chain: %v", err)
	}

	// Invoke the chain
	result, err := runnable.Invoke(ctx, map[string]any{
		"input": "What is the capital of France?",
	})
	if err != nil {
		log.Fatalf("Failed to invoke chain: %v", err)
	}

	fmt.Println(result.Content)
}
