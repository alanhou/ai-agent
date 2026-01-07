package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// --- Types ---

type AgentCard struct {
	Identity     string                 `json:"identity"`
	Capabilities []string               `json:"capabilities"`
	Schemas      map[string]interface{} `json:"schemas"`
	Endpoint     string                 `json:"endpoint"`
	AuthMethods  []string               `json:"auth_methods"`
	Version      string                 `json:"version"`
}

type RPCRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  map[string]string `json:"params"`
	ID      interface{}       `json:"id"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- Global Config ---

var agentCard = AgentCard{
	Identity:     "SummarizerAgent",
	Capabilities: []string{"summarizeText"},
	Schemas: map[string]interface{}{
		"summarizeText": map[string]interface{}{
			"input":  map[string]string{"text": "string"},
			"output": map[string]string{"summary": "string"},
		},
	},
	Endpoint:    "http://localhost:8000/api",
	AuthMethods: []string{"none"},
	Version:     "1.0",
}

// --- Handlers ---

func agentHandler(w http.ResponseWriter, r *http.Request) {
	// CORS basics for demo
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "GET" && r.URL.Path == "/.well-known/agent.json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agentCard)
		return
	}

	if r.Method == "POST" && r.URL.Path == "/api" {
		handleRPC(w, r)
		return
	}

	http.NotFound(w, r)
}

func handleRPC(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.JSONRPC == "2.0" && req.Method == "summarizeText" {
		text := req.Params["text"]

		// Call LLM
		summary := simpleSummarize(text)

		resp := RPCResponse{
			JSONRPC: "2.0",
			Result:  map[string]string{"summary": summary},
			ID:      req.ID,
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		resp := RPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: req.ID,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func simpleSummarize(text string) string {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		BaseURL:     os.Getenv("OPENAI_BASE_URL"),
		Temperature: nil, // default
	})
	if err != nil {
		return fmt.Sprintf("Error initializing LLM: %v", err)
	}

	msgs := []*schema.Message{
		schema.SystemMessage("You are a helpful assistant that provides concise summaries."),
		schema.UserMessage(fmt.Sprintf("Summarize the following text: %s", text)),
	}

	resp, err := chatModel.Generate(ctx, msgs)
	if err != nil {
		return fmt.Sprintf("Error generating summary: %v", err)
	}
	return resp.Content
}

func main() {
	_ = godotenv.Load()

	http.HandleFunc("/", agentHandler)

	port := "8000"
	fmt.Printf("Starting Go A2A agent server on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
