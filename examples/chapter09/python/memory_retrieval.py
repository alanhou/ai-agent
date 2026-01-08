"""
Memory Retrieval Evaluation Metrics

This module provides metrics for evaluating memory/retrieval systems
in AI agent applications.
"""
from typing import List, Dict, Any, Callable


def evaluate_memory_retrieval(
    retrieve_fn: Callable[[str, int], List[Any]],
    queries: List[str],
    expected_results: List[List[Any]],
    top_k: int = 1
) -> Dict[str, float]:
    """
    Evaluate a retrieval function against expected results.
    
    Given a retrieval function `retrieve_fn(query, k)` that returns k memory items,
    evaluate performance across multiple queries.
    
    Args:
        retrieve_fn: Function that takes (query, k) and returns list of k results
        queries: List of query strings to test
        expected_results: List of expected result lists for each query
        top_k: Number of top results to consider (default: 1)
        
    Returns:
        Dict with `retrieval_accuracy@k`: proportion of queries where at least
        one expected item appears in the top k results.
    """
    hits = 0
    for query, expect in zip(queries, expected_results):
        results = retrieve_fn(query, top_k)
        # Check if we retrieved any of the expected items
        if set(results) & set(expect):
            hits += 1
    accuracy = hits / len(queries) if queries else 1.0
    return {f"retrieval_accuracy@{top_k}": accuracy}


if __name__ == "__main__":
    # Example usage with a mock retrieval function
    def mock_retrieve_fn(query: str, k: int) -> List[str]:
        """Mock retrieval function for demonstration."""
        memory_store = {
            "weather query": ["weather_doc_1", "weather_doc_2", "unrelated"],
            "email query": ["email_doc_1", "contact_doc", "email_doc_2"],
        }
        return memory_store.get(query, ["unknown"])[:k]
    
    queries = ["weather query", "email query"]
    expected_results = [
        ["weather_doc_1", "weather_doc_2"],  # Expected for weather query
        ["email_doc_1", "email_doc_2"]        # Expected for email query
    ]
    
    # Test with top_k=1
    result = evaluate_memory_retrieval(mock_retrieve_fn, queries, expected_results, top_k=1)
    print(f"Retrieval Accuracy @1: {result['retrieval_accuracy@1']}")
    
    # Test with top_k=3
    result = evaluate_memory_retrieval(mock_retrieve_fn, queries, expected_results, top_k=3)
    print(f"Retrieval Accuracy @3: {result['retrieval_accuracy@3']}")
