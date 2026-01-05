package main

import (
	"agents-go/examples/chapter05/go/common"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/joho/godotenv"
)

// ToolDescription holds tool metadata for semantic search
type ToolDescription struct {
	Name        string
	Description string
	Tool        tool.InvokableTool
	Embedding   []float64
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	ctx := context.Background()

	// Initialize embeddings model
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   "text-embedding-ada-002",
	})
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	// Create tools
	wolframTool, err := utils.InferTool("query_wolfram_alpha",
		"Use Wolfram Alpha to compute mathematical expressions or retrieve information.",
		common.QueryWolframAlpha)
	if err != nil {
		log.Fatalf("Failed to create wolframTool: %v", err)
	}

	zapierTool, err := utils.InferTool("trigger_zapier_webhook",
		"Trigger a Zapier webhook to execute predefined automated workflows.",
		common.TriggerZapierWebhook)
	if err != nil {
		log.Fatalf("Failed to create zapierTool: %v", err)
	}

	slackTool, err := utils.InferTool("send_slack_message",
		"Send messages to specific Slack channels to communicate with team members.",
		common.SendSlackMessage)
	if err != nil {
		log.Fatalf("Failed to create slackTool: %v", err)
	}

	// Tool descriptions for semantic search
	toolDescriptions := []ToolDescription{
		{
			Name:        "query_wolfram_alpha",
			Description: "Use Wolfram Alpha to compute mathematical expressions or retrieve information.",
			Tool:        wolframTool,
		},
		{
			Name:        "trigger_zapier_webhook",
			Description: "Trigger a Zapier webhook to execute predefined automated workflows.",
			Tool:        zapierTool,
		},
		{
			Name:        "send_slack_message",
			Description: "Send messages to specific Slack channels to communicate with team members.",
			Tool:        slackTool,
		},
	}

	// Create embeddings for each tool description
	var descriptions []string
	for _, td := range toolDescriptions {
		descriptions = append(descriptions, td.Description)
	}

	embeddings, err := embedder.EmbedStrings(ctx, descriptions)
	if err != nil {
		log.Fatalf("Failed to create embeddings: %v", err)
	}

	for i := range toolDescriptions {
		toolDescriptions[i].Embedding = embeddings[i]
	}

	// User query
	userQuery := "Solve this equation: 2x + 3 = 7"

	// Select the top tool using semantic similarity
	selectedTool, err := selectTool(ctx, embedder, toolDescriptions, userQuery)
	if err != nil {
		log.Fatalf("Failed to select tool: %v", err)
	}

	if selectedTool == nil {
		fmt.Println("No tool was selected.")
		return
	}

	fmt.Printf("Selected Tool: %s\n", selectedTool.Name)

	// Determine parameters based on the query and selected tool
	args, err := determineParameters(userQuery, selectedTool.Name)
	if err != nil {
		log.Fatalf("Failed to determine parameters: %v", err)
	}

	// Invoke the selected tool
	result, err := selectedTool.Tool.InvokableRun(ctx, args)
	if err != nil {
		log.Fatalf("Error invoking tool '%s': %v", selectedTool.Name, err)
	}

	fmt.Printf("Tool '%s' Result: %s\n", selectedTool.Name, result)
}

// selectTool selects the most relevant tool based on semantic similarity
func selectTool(ctx context.Context, embedder *openai.Embedder, tools []ToolDescription, query string) (*ToolDescription, error) {
	// Create embedding for the query
	queryEmbeddings, err := embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %v", err)
	}
	queryEmbedding := queryEmbeddings[0]

	// Calculate cosine similarity with each tool
	var bestTool *ToolDescription
	bestScore := -1.0

	for i := range tools {
		score := cosineSimilarity(queryEmbedding, tools[i].Embedding)
		if score > bestScore {
			bestScore = score
			bestTool = &tools[i]
		}
	}

	return bestTool, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// determineParameters extracts parameters based on the query and tool
func determineParameters(query, toolName string) (string, error) {
	var params interface{}

	switch toolName {
	case "query_wolfram_alpha":
		params = common.QueryWolframAlphaArgs{Expression: query}
	case "trigger_zapier_webhook":
		params = common.TriggerZapierWebhookArgs{
			ZapID:   "123456",
			Payload: map[string]interface{}{"data": query},
		}
	case "send_slack_message":
		params = common.SendSlackMessageArgs{
			Channel: "#general",
			Message: query,
		}
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}


