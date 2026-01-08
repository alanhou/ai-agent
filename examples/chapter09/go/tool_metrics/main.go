// Package main demonstrates tool evaluation metrics
//
// This module provides metrics for evaluating tool selection and parameter accuracy
// in AI agent systems.
package main

import (
	"fmt"
	"reflect"
)

// ToolCall represents an expected or predicted tool call
type ToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

// ToolMetricsResult contains tool recall and precision values
type ToolMetricsResult struct {
	ToolRecall    float64 `json:"tool_recall"`
	ToolPrecision float64 `json:"tool_precision"`
}

// ToolMetrics calculates tool recall and precision metrics
//
// Args:
//
//	predTools: List of tool names that were predicted/called
//	expectedCalls: List of expected tool calls with Tool field
//
// Returns:
//
//	ToolMetricsResult with tool_recall and tool_precision values
func ToolMetrics(predTools []string, expectedCalls []ToolCall) ToolMetricsResult {
	// Extract expected tool names
	expectedNames := make([]string, 0, len(expectedCalls))
	for _, c := range expectedCalls {
		expectedNames = append(expectedNames, c.Tool)
	}

	if len(expectedNames) == 0 {
		return ToolMetricsResult{ToolRecall: 1.0, ToolPrecision: 1.0}
	}

	// Create sets
	predSet := make(map[string]bool)
	for _, t := range predTools {
		predSet[t] = true
	}

	expSet := make(map[string]bool)
	for _, t := range expectedNames {
		expSet[t] = true
	}

	// Calculate true positives (intersection)
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
//
// Args:
//
//	predCalls: List of predicted tool calls with Tool and Params
//	expectedCalls: List of expected tool calls with Tool and Params
//
// Returns:
//
//	Accuracy score (0.0 to 1.0)
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

func main() {
	// Example usage
	predTools := []string{"get_weather", "send_email"}
	expectedCalls := []ToolCall{
		{Tool: "get_weather", Params: map[string]interface{}{"city": "Seattle"}},
		{Tool: "send_email", Params: map[string]interface{}{"to": "user@example.com"}},
	}

	metrics := ToolMetrics(predTools, expectedCalls)
	fmt.Printf("Tool Recall: %.1f\n", metrics.ToolRecall)
	fmt.Printf("Tool Precision: %.1f\n", metrics.ToolPrecision)

	predCalls := []ToolCall{
		{Tool: "get_weather", Params: map[string]interface{}{"city": "Seattle"}},
		{Tool: "send_email", Params: map[string]interface{}{"to": "user@example.com"}},
	}
	accuracy := ParamAccuracy(predCalls, expectedCalls)
	fmt.Printf("Parameter Accuracy: %.1f\n", accuracy)
}
