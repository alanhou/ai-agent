package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// --- Tool Definitions ---

type GetStockPriceArgs struct {
	Ticker string `json:"ticker" jsonschema:"description=The stock ticker symbol (e.g. AAPL)"`
}

func GetStockPrice(ctx context.Context, args *GetStockPriceArgs) (string, error) {
	apiKey := os.Getenv("FINHUB_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("FINHUB_API_KEY not set")
	}

	apiURL := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", args.Ticker, apiKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch stock price: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API Error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	// Finnhub returns 0 for invalid tickers sometimes, or checks 'c'
	price, ok := result["c"].(float64)
	if !ok || price == 0 {
		return "", fmt.Errorf("Ticker %s not found or price is zero.", args.Ticker)
	}

	return fmt.Sprintf("%f", price), nil
}

type QueryWolframAlphaArgs struct {
	Expression string `json:"expression" jsonschema:"description=The mathematical expression or query to evaluate"`
}

func QueryWolframAlpha(ctx context.Context, args *QueryWolframAlphaArgs) (string, error) {
	appID := os.Getenv("WOLFRAM_ALPHA_APP_ID")
	if appID == "" {
		return "", fmt.Errorf("WOLFRAM_ALPHA_APP_ID not set")
	}

	baseURL := "https://api.wolframalpha.com/v1/result"
	params := url.Values{}
	params.Add("i", args.Expression)
	params.Add("appid", appID)

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to query Wolfram Alpha: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Wolfram Alpha API Error: %d - %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

type SendSlackMessageArgs struct {
	Channel string `json:"channel" jsonschema:"description=The Slack channel ID or name where the message will be sent"`
	Message string `json:"message" jsonschema:"description=The content of the message to send"`
}

func SendSlackMessage(ctx context.Context, args *SendSlackMessageArgs) (string, error) {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("SLACK_BOT_TOKEN not set")
	}

	apiURL := "https://slack.com/api/chat.postMessage"
	payload := map[string]string{
		"channel": args.Channel,
		"text":    args.Message,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if ok, _ := result["ok"].(bool); ok {
		return fmt.Sprintf("Message successfully sent to Slack channel '%s'.", args.Channel), nil
	}

	errMsg := "Unknown error"
	if e, ok := result["error"].(string); ok {
		errMsg = e
	}
	return "", fmt.Errorf("Slack API Error: %s", errMsg)
}

// --- Main ---

func main() {
	// 1. Load environment variables
	if err := godotenv.Load("../../../.env"); err != nil {
		// fallback to loading from current directory or system env
		_ = godotenv.Load()
	}

	ctx := context.Background()

	// 2. Initialize ChatModel
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 3. Create Tools
	stockTool, err := utils.InferTool("get_stock_price", "Get stock price via Finnhub REST API.", GetStockPrice)
	if err != nil {
		log.Fatalf("Failed to create stockTool: %v", err)
	}

	wolframTool, err := utils.InferTool("query_wolfram_alpha", "Query Wolfram Alpha to compute expressions or retrieve information.", QueryWolframAlpha)
	if err != nil {
		log.Fatalf("Failed to create wolframTool: %v", err)
	}

	slackTool, err := utils.InferTool("send_slack_message", "Send a message to a specified Slack channel.", SendSlackMessage)
	if err != nil {
		log.Fatalf("Failed to create slackTool: %v", err)
	}

	tools := []tool.InvokableTool{stockTool, wolframTool, slackTool}
	toolMap := make(map[string]tool.InvokableTool)
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))

	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Fatalf("Failed to get tool info: %v", err)
		}
		toolInfos = append(toolInfos, info)
		toolMap[info.Name] = t
	}

	// 4. Bind Tools
	if err := model.BindTools(toolInfos); err != nil {
		log.Fatalf("Failed to bind tools: %v", err)
	}

	// 5. Build Conversation Flow
	// User: "What is the stock price of Apple?"
	messages := []*schema.Message{
		schema.UserMessage("What is the stock price of Apple?"),
	}

	fmt.Println("User:", messages[0].Content)

	// First Run: Get Tool Calls
	resp, err := model.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	// Append AIMessage (with tool calls) to history
	messages = append(messages, resp)

	// Execute Tools
	if len(resp.ToolCalls) > 0 {
		for _, tc := range resp.ToolCalls {
			fmt.Printf("Tool Call: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)

			t, exists := toolMap[tc.Function.Name]
			if !exists {
				log.Printf("Tool %s not found", tc.Function.Name)
				continue
			}

			// Execute
			result, err := t.InvokableRun(ctx, tc.Function.Arguments)
			if err != nil {
				// In a real agent, you might want to return the error to the model
				result = fmt.Sprintf("Error: %v", err)
			}
			fmt.Printf("Tool Result: %s\n", result)

			// Append ToolMessage
			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// Second Run: Get Final Answer
		finalResp, err := model.Generate(ctx, messages)
		if err != nil {
			log.Fatalf("Failed to generate final response: %v", err)
		}
		fmt.Println("AI:", finalResp.Content)
	} else {
		// No tool calls, just print response
		fmt.Println("AI:", resp.Content)
	}
}
