// Package main demonstrates agent evaluation metrics
//
// This module provides end-to-end evaluation metrics for AI agents,
// including phrase recall, task success, and single instance evaluation.
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ToolCall represents an expected or predicted tool call
type ToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExpectedState represents the expected final state for evaluation
type ExpectedState struct {
	ToolCalls           []ToolCall `json:"tool_calls"`
	CustomerMsgContains []string   `json:"customer_msg_contains"`
}

// Expected wraps the expected final state
type Expected struct {
	FinalState ExpectedState `json:"final_state"`
}

// TestInstance represents a single test case
type TestInstance struct {
	Order        interface{} `json:"order"`
	Conversation []Message   `json:"conversation"`
	Expected     Expected    `json:"expected"`
}

// ToolMetricsResult contains tool recall and precision values
type ToolMetricsResult struct {
	ToolRecall    float64 `json:"tool_recall"`
	ToolPrecision float64 `json:"tool_precision"`
}

// EvaluationResult contains all evaluation metrics
type EvaluationResult struct {
	PhraseRecall  float64 `json:"phrase_recall"`
	ToolRecall    float64 `json:"tool_recall"`
	ToolPrecision float64 `json:"tool_precision"`
	ParamAccuracy float64 `json:"param_accuracy"`
	TaskSuccess   float64 `json:"task_success"`
}

// PhraseRecall calculates the recall of expected phrases in the response
//
// Args:
//
//	response: The agent's response text
//	expectedPhrases: List of phrases that should appear in the response
//
// Returns:
//
//	Recall score (0.0 to 1.0)
func PhraseRecall(response string, expectedPhrases []string) float64 {
	if len(expectedPhrases) == 0 {
		return 1.0
	}

	responseLower := strings.ToLower(response)
	found := 0
	for _, phrase := range expectedPhrases {
		if strings.Contains(responseLower, strings.ToLower(phrase)) {
			found++
		}
	}

	return float64(found) / float64(len(expectedPhrases))
}

// TaskSuccess determines if the task was successfully completed
//
// Args:
//
//	finalReply: The agent's final response
//	predTools: List of tools that were called
//	expected: Expected final state with tool_calls and customer_msg_contains
//
// Returns:
//
//	1.0 if task succeeded, 0.0 otherwise
func TaskSuccess(finalReply string, predTools []string, expected ExpectedState) float64 {
	// Create set of expected tools
	expectedTools := make(map[string]bool)
	for _, c := range expected.ToolCalls {
		expectedTools[c.Tool] = true
	}

	// Check if all expected tools were called
	if len(expectedTools) > 0 {
		predSet := make(map[string]bool)
		for _, t := range predTools {
			predSet[t] = true
		}

		for t := range expectedTools {
			if !predSet[t] {
				return 0.0
			}
		}
	}

	// Check if all expected phrases are in the final reply
	if len(expected.CustomerMsgContains) > 0 {
		phraseScore := PhraseRecall(finalReply, expected.CustomerMsgContains)
		if phraseScore < 1.0 {
			return 0.0
		}
	}

	return 1.0
}

// ToolMetrics calculates tool recall and precision metrics
func ToolMetrics(predTools []string, expectedCalls []ToolCall) ToolMetricsResult {
	expectedNames := make([]string, 0, len(expectedCalls))
	for _, c := range expectedCalls {
		expectedNames = append(expectedNames, c.Tool)
	}

	if len(expectedNames) == 0 {
		return ToolMetricsResult{ToolRecall: 1.0, ToolPrecision: 1.0}
	}

	predSet := make(map[string]bool)
	for _, t := range predTools {
		predSet[t] = true
	}

	expSet := make(map[string]bool)
	for _, t := range expectedNames {
		expSet[t] = true
	}

	tp := 0
	for t := range expSet {
		if predSet[t] {
			tp++
		}
	}

	recall := float64(tp) / float64(len(expSet))
	precision := 0.0
	if len(predSet) > 0 {
		precision = float64(tp) / float64(len(predSet))
	}

	return ToolMetricsResult{ToolRecall: recall, ToolPrecision: precision}
}

// ParamAccuracy calculates parameter accuracy for tool calls
func ParamAccuracy(predCalls []ToolCall, expectedCalls []ToolCall) float64 {
	if len(expectedCalls) == 0 {
		return 1.0
	}

	matched := 0
	for _, exp := range expectedCalls {
		for _, pred := range predCalls {
			if pred.Tool == exp.Tool && reflect.DeepEqual(pred.Params, exp.Params) {
				matched++
				break
			}
		}
	}

	return float64(matched) / float64(len(expectedCalls))
}

// EvaluateSingleInstance evaluates a single test instance
//
// This is a simplified version that demonstrates the structure.
// In a real implementation, this would invoke an actual agent graph.
//
// Args:
//
//	raw: JSON string containing the test instance
//
// Returns:
//
//	EvaluationResult or nil if evaluation failed
func EvaluateSingleInstance(raw string) *EvaluationResult {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var instance TestInstance
	if err := json.Unmarshal([]byte(raw), &instance); err != nil {
		fmt.Printf("[SKIPPED] example failed with error: %v\n", err)
		return nil
	}

	// In a real implementation, you would:
	// 1. Convert messages to the agent's message format
	// 2. Invoke the agent graph
	// 3. Extract the final reply and tool calls from the result

	// For demonstration, we'll simulate with empty results
	finalReply := ""
	predTools := []string{}
	predCalls := []ToolCall{}

	expected := instance.Expected.FinalState
	tm := ToolMetrics(predTools, expected.ToolCalls)

	return &EvaluationResult{
		PhraseRecall:  PhraseRecall(finalReply, expected.CustomerMsgContains),
		ToolRecall:    tm.ToolRecall,
		ToolPrecision: tm.ToolPrecision,
		ParamAccuracy: ParamAccuracy(predCalls, expected.ToolCalls),
		TaskSuccess:   TaskSuccess(finalReply, predTools, expected),
	}
}

func main() {
	fmt.Println("Agent Metrics Module")
	fmt.Println("----------------------------------------")

	// Test PhraseRecall
	response := "Your order has been shipped and will arrive tomorrow."
	expectedPhrases := []string{"shipped", "arrive"}
	recall := PhraseRecall(response, expectedPhrases)
	fmt.Printf("Phrase Recall: %.1f\n", recall)

	// Test TaskSuccess
	expectedState := ExpectedState{
		ToolCalls:           []ToolCall{{Tool: "ship_order"}},
		CustomerMsgContains: []string{"shipped"},
	}
	success := TaskSuccess(response, []string{"ship_order"}, expectedState)
	fmt.Printf("Task Success: %.1f\n", success)

	// Test Message conversion (just show structure)
	msg := Message{Role: "user", Content: "Hello!"}
	fmt.Printf("Message Type: %s, Content: %s\n", msg.Role, msg.Content)
}
