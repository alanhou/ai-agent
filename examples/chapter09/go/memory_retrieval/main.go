// Package main demonstrates memory retrieval evaluation metrics
//
// This module provides metrics for evaluating memory/retrieval systems
// in AI agent applications.
package main

import (
	"fmt"
)

// RetrieveFunc is a function type that takes a query and k, returns k results
type RetrieveFunc func(query string, k int) []string

// EvaluateMemoryRetrieval evaluates a retrieval function against expected results
//
// Given a retrieval function that returns k memory items,
// evaluate performance across multiple queries.
//
// Args:
//
//	retrieveFn: Function that takes (query, k) and returns list of k results
//	queries: List of query strings to test
//	expectedResults: List of expected result lists for each query
//	topK: Number of top results to consider
//
// Returns:
//
//	Map with "retrieval_accuracy@k": proportion of queries where at least
//	one expected item appears in the top k results.
func EvaluateMemoryRetrieval(
	retrieveFn RetrieveFunc,
	queries []string,
	expectedResults [][]string,
	topK int,
) map[string]float64 {
	hits := 0

	for i, query := range queries {
		if i >= len(expectedResults) {
			break
		}
		expect := expectedResults[i]
		results := retrieveFn(query, topK)

		// Create set from expected results
		expectSet := make(map[string]bool)
		for _, e := range expect {
			expectSet[e] = true
		}

		// Check if we retrieved any of the expected items
		for _, r := range results {
			if expectSet[r] {
				hits++
				break
			}
		}
	}

	accuracy := 1.0
	if len(queries) > 0 {
		accuracy = float64(hits) / float64(len(queries))
	}

	return map[string]float64{
		fmt.Sprintf("retrieval_accuracy@%d", topK): accuracy,
	}
}

// mockRetrieveFn is a mock retrieval function for demonstration
func mockRetrieveFn(query string, k int) []string {
	memoryStore := map[string][]string{
		"weather query": {"weather_doc_1", "weather_doc_2", "unrelated"},
		"email query":   {"email_doc_1", "contact_doc", "email_doc_2"},
	}

	results, ok := memoryStore[query]
	if !ok {
		return []string{"unknown"}
	}

	if k > len(results) {
		k = len(results)
	}
	return results[:k]
}

func main() {
	queries := []string{"weather query", "email query"}
	expectedResults := [][]string{
		{"weather_doc_1", "weather_doc_2"}, // Expected for weather query
		{"email_doc_1", "email_doc_2"},     // Expected for email query
	}

	// Test with topK=1
	result := EvaluateMemoryRetrieval(mockRetrieveFn, queries, expectedResults, 1)
	fmt.Printf("Retrieval Accuracy @1: %.1f\n", result["retrieval_accuracy@1"])

	// Test with topK=3
	result = EvaluateMemoryRetrieval(mockRetrieveFn, queries, expectedResults, 3)
	fmt.Printf("Retrieval Accuracy @3: %.1f\n", result["retrieval_accuracy@3"])
}
