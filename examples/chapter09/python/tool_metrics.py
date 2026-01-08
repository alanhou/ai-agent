"""
Tool Evaluation Metrics

This module provides metrics for evaluating tool selection and parameter accuracy
in AI agent systems.
"""
from typing import List, Dict


def tool_metrics(pred_tools: List[str], expected_calls: List[dict]) -> Dict[str, float]:
    """
    Calculate tool recall and precision metrics.
    
    Args:
        pred_tools: List of tool names that were predicted/called
        expected_calls: List of expected tool call dicts with 'tool' key
        
    Returns:
        Dict with 'tool_recall' and 'tool_precision' values
    """
    expected_names = [c.get("tool") for c in expected_calls]
    if not expected_names:
        return {"tool_recall": 1.0, "tool_precision": 1.0}
    pred_set = set(pred_tools)
    exp_set = set(expected_names)
    tp = len(exp_set & pred_set)
    recall = tp / len(exp_set)
    precision = tp / len(pred_set) if pred_set else 0.0
    return {"tool_recall": recall, "tool_precision": precision}


def param_accuracy(pred_calls: List[dict], expected_calls: List[dict]) -> float:
    """
    Calculate parameter accuracy for tool calls.
    
    Args:
        pred_calls: List of predicted tool call dicts with 'tool' and 'params' keys
        expected_calls: List of expected tool call dicts with 'tool' and 'params' keys
        
    Returns:
        Accuracy score (0.0 to 1.0)
    """
    if not expected_calls:
        return 1.0
    matched = 0
    for exp in expected_calls:
        for pred in pred_calls:
            if pred.get("tool") == exp.get("tool") \
                and pred.get("params") == exp.get("params"):
                matched += 1
                break
    return matched / len(expected_calls)


if __name__ == "__main__":
    # Example usage
    pred_tools = ["get_weather", "send_email"]
    expected_calls = [
        {"tool": "get_weather", "params": {"city": "Seattle"}},
        {"tool": "send_email", "params": {"to": "user@example.com"}}
    ]
    
    metrics = tool_metrics(pred_tools, expected_calls)
    print(f"Tool Recall: {metrics['tool_recall']}")
    print(f"Tool Precision: {metrics['tool_precision']}")
    
    pred_calls = [
        {"tool": "get_weather", "params": {"city": "Seattle"}},
        {"tool": "send_email", "params": {"to": "user@example.com"}}
    ]
    accuracy = param_accuracy(pred_calls, expected_calls)
    print(f"Parameter Accuracy: {accuracy}")
