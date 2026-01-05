package main

import (
	"agents-go/examples/chapter05/go/common"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// ToolGroup represents a category of tools
type ToolGroup struct {
	Name        string
	Description string
	Tools       []tool.InvokableTool
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	ctx := context.Background()

	// Initialize ChatModel
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o-mini",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create Tools
	wolframTool, err := utils.InferTool("query_wolfram_alpha",
		"Query Wolfram Alpha to compute expressions or retrieve information.",
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

	// Define tool groups
	toolGroups := []ToolGroup{
		{
			Name:        "Computation",
			Description: "Tools related to mathematical computations and data analysis.",
			Tools:       []tool.InvokableTool{wolframTool},
		},
		{
			Name:        "Automation",
			Description: "Tools that automate workflows and integrate different services.",
			Tools:       []tool.InvokableTool{zapierTool},
		},
		{
			Name:        "Communication",
			Description: "Tools that facilitate communication and messaging.",
			Tools:       []tool.InvokableTool{slackTool},
		},
	}

	// User query
	userQuery := "Solve this equation: 2x + 3 = 7"

	// Step 1: Select the most relevant tool group using LLM
	selectedGroup, err := selectGroupLLM(ctx, model, toolGroups, userQuery)
	if err != nil {
		log.Fatalf("Failed to select group: %v", err)
	}
	fmt.Printf("Selected Tool Group: %s\n", selectedGroup.Name)

	// Step 2: Select the most relevant tool within the group using LLM
	selectedTool, err := selectToolLLM(ctx, model, selectedGroup, userQuery)
	if err != nil {
		log.Fatalf("Failed to select tool: %v", err)
	}

	toolInfo, err := selectedTool.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get tool info: %v", err)
	}
	fmt.Printf("Selected Tool: %s\n", toolInfo.Name)

	// Step 3: Determine parameters and invoke the tool
	args, err := determineParameters(userQuery, toolInfo.Name)
	if err != nil {
		log.Fatalf("Failed to determine parameters: %v", err)
	}

	result, err := selectedTool.InvokableRun(ctx, args)
	if err != nil {
		log.Fatalf("Error invoking tool '%s': %v", toolInfo.Name, err)
	}

	fmt.Printf("Tool '%s' Result: %s\n", toolInfo.Name, result)
}

func selectGroupLLM(ctx context.Context, model *openai.ChatModel, groups []ToolGroup, query string) (*ToolGroup, error) {
	// Build options string
	var options []string
	for _, g := range groups {
		options = append(options, g.Name)
	}

	prompt := fmt.Sprintf("Select the most appropriate tool group for the following query: '%s'.\nOptions are: %s.",
		query, strings.Join(options, ", "))

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := model.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	groupName := strings.TrimSpace(resp.Content)

	// Find matching group
	for i := range groups {
		if strings.EqualFold(groups[i].Name, groupName) {
			return &groups[i], nil
		}
	}

	return nil, fmt.Errorf("group not found: %s", groupName)
}

func selectToolLLM(ctx context.Context, model *openai.ChatModel, group *ToolGroup, query string) (tool.InvokableTool, error) {
	// Build tool names
	var toolNames []string
	for _, t := range group.Tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolNames = append(toolNames, info.Name)
	}

	prompt := fmt.Sprintf("Based on the query: '%s', select the most appropriate tool from the group '%s'.\nAvailable tools: %s",
		query, group.Name, strings.Join(toolNames, ", "))

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := model.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	toolName := strings.TrimSpace(resp.Content)

	// Find matching tool
	for _, t := range group.Tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(toolName), strings.ToLower(info.Name)) {
			return t, nil
		}
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

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
