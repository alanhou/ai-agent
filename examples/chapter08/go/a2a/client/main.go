package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// --- Types ---

type AgentCard struct {
	Identity     string   `json:"identity"`
	Capabilities []string `json:"capabilities"`
	Endpoint     string   `json:"endpoint"`
	Version      string   `json:"version"`
}

type RPCRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  map[string]string `json:"params"`
	ID      interface{}       `json:"id"`
}

type RPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  map[string]interface{} `json:"result"`
	Error   interface{}            `json:"error"`
	ID      interface{}            `json:"id"`
}

func main() {
	cardURL := "http://localhost:8000/.well-known/agent.json"
	fmt.Printf("Discovering Agent Card at %s...\n", cardURL)

	// 1. Discovery
	resp, err := http.Get(cardURL)
	if err != nil {
		log.Fatalf("Failed to retrieve Agent Card: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Failed to retrieve Agent Card, status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var card AgentCard
	if err := json.Unmarshal(body, &card); err != nil {
		log.Fatalf("Failed to parse Agent Card: %v", err)
	}

	cardBytes, _ := json.MarshalIndent(card, "", "  ")
	fmt.Printf("Discovered Agent Card: %s\n", string(cardBytes))

	// 2. Handshake / Validation
	if card.Version != "1.0" {
		log.Fatal("Incompatible protocol version")
	}
	hasCap := false
	for _, cap := range card.Capabilities {
		if cap == "summarizeText" {
			hasCap = true
			break
		}
	}
	if !hasCap {
		log.Fatal("Required capability 'summarizeText' not supported")
	}
	fmt.Println("Handshake successful: Agent is compatible.")

	// 3. RPC Call
	rpcURL := card.Endpoint
	reqData := RPCRequest{
		JSONRPC: "2.0",
		Method:  "summarizeText",
		Params: map[string]string{
			"text": "This is a long example text that needs summarization. It discusses multiagent systems and communication protocols used in AI agent networks.",
		},
		ID: 123,
	}
	jsonData, _ := json.Marshal(reqData)

	fmt.Printf("Sending RPC Request to %s...\n", rpcURL)
	rpcResp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}
	defer rpcResp.Body.Close()

	respBody, _ := io.ReadAll(rpcResp.Body)
	var result RPCResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Fatalf("Failed to parse RPC response: %v", err)
	}

	resultBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("RPC Response: %s\n", string(resultBytes))
}
