package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// --- Common Tool Definitions ---

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

type TriggerZapierWebhookArgs struct {
	ZapID   string                 `json:"zap_id" jsonschema:"description=The unique identifier for the Zap to be triggered"`
	Payload map[string]interface{} `json:"payload" jsonschema:"description=The data to send to the Zapier webhook"`
}

func TriggerZapierWebhook(ctx context.Context, args *TriggerZapierWebhookArgs) (string, error) {
	zapierWebhookURL := fmt.Sprintf("https://hooks.zapier.com/hooks/catch/%s/", args.ZapID)

	jsonPayload, err := json.Marshal(args.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", zapierWebhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Zapier API Error: %d - %s", resp.StatusCode, string(body))
	}

	return fmt.Sprintf("Zapier webhook '%s' successfully triggered.", args.ZapID), nil
}
