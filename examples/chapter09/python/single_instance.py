"""
Agent Evaluation Metrics

This module provides end-to-end evaluation metrics for AI agents,
including phrase recall, task success, and single instance evaluation.
"""
import json
from typing import List, Dict, Optional, Any

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage, BaseMessage

try:
    from .tool_metrics import tool_metrics, param_accuracy
except ImportError:
    from tool_metrics import tool_metrics, param_accuracy


def to_lc_message(turn: dict) -> BaseMessage:
    """
    Convert a conversation turn dict to a LangChain message.
    
    Args:
        turn: Dict with 'role' and 'content' keys
        
    Returns:
        LangChain message object (HumanMessage, AIMessage, or SystemMessage)
    """
    role = turn.get("role", "user")
    content = turn.get("content", "")
    
    if role == "user":
        return HumanMessage(content=content)
    elif role == "assistant":
        return AIMessage(content=content)
    elif role == "system":
        return SystemMessage(content=content)
    else:
        # Default to HumanMessage for unknown roles
        return HumanMessage(content=content)


def phrase_recall(response: str, expected_phrases: List[str]) -> float:
    """
    Calculate the recall of expected phrases in the response.
    
    Args:
        response: The agent's response text
        expected_phrases: List of phrases that should appear in the response
        
    Returns:
        Recall score (0.0 to 1.0)
    """
    if not expected_phrases:
        return 1.0
    
    response_lower = response.lower()
    found = sum(1 for phrase in expected_phrases if phrase.lower() in response_lower)
    return found / len(expected_phrases)


def task_success(
    final_reply: str,
    pred_tools: List[str],
    expected: dict
) -> float:
    """
    Determine if the task was successfully completed.
    
    Args:
        final_reply: The agent's final response
        pred_tools: List of tools that were called
        expected: Expected final state dict with 'tool_calls' and 'customer_msg_contains'
        
    Returns:
        1.0 if task succeeded, 0.0 otherwise
    """
    # Check if all expected tools were called
    expected_tools = {c.get("tool") for c in expected.get("tool_calls", [])}
    if expected_tools and not expected_tools.issubset(set(pred_tools)):
        return 0.0
    
    # Check if all expected phrases are in the final reply
    expected_phrases = expected.get("customer_msg_contains", [])
    if expected_phrases:
        phrase_score = phrase_recall(final_reply, expected_phrases)
        if phrase_score < 1.0:
            return 0.0
    
    return 1.0


def evaluate_single_instance(raw: str, graph: Any) -> Optional[Dict[str, float]]:
    """
    Evaluate a single test instance against an agent graph.
    
    Args:
        raw: JSON string containing the test instance with:
            - 'order': Order context
            - 'conversation': List of conversation turns
            - 'expected': Expected final state
        graph: LangGraph graph object with invoke() method
        
    Returns:
        Dict of metrics or None if evaluation failed
    """
    if not raw.strip():
        return None
    try:
        ex = json.loads(raw)
        order = ex["order"]
        messages = [to_lc_message(t) for t in ex["conversation"]]
        expected = ex["expected"]["final_state"]

        result = graph.invoke({"order": order, "messages": messages})

        # Extract the assistant's final message
        final_reply = ""
        for msg in reversed(result["messages"]):
            if isinstance(msg, AIMessage) \
                and not msg.additional_kwargs.get("tool_calls"):
                final_reply = msg.content or ""
                break

        # Collect predicted tool names and parameters
        pred_tools, pred_calls = [], []
        for m in result["messages"]:
            if isinstance(m, AIMessage):
                for tc in m.additional_kwargs.get("tool_calls", []):
                    name = tc.get("function", {}).get("name") or tc.get("name")
                    args = json.loads(tc["function"]["arguments"]) \
                        if "function" in tc else tc.get("args", {})
                    pred_tools.append(name)
                    pred_calls.append({"tool": name, "params": args})

        # Calculate and return metrics
        tm = tool_metrics(pred_tools, expected.get("tool_calls", []))
        return {
            "phrase_recall": phrase_recall(final_reply, expected.get("customer_msg_contains", [])),
            "tool_recall": tm["tool_recall"],
            "tool_precision": tm["tool_precision"],
            "param_accuracy": param_accuracy(pred_calls, expected.get("tool_calls", [])),
            "task_success": task_success(final_reply, pred_tools, expected),
        }
    except Exception as e:
        print(f"[SKIPPED] example failed with error: {e!r}")
        return None


if __name__ == "__main__":
    # Example usage (without actual graph)
    print("Agent Metrics Module")
    print("-" * 40)
    
    # Test phrase_recall
    response = "Your order has been shipped and will arrive tomorrow."
    expected = ["shipped", "arrive"]
    recall = phrase_recall(response, expected)
    print(f"Phrase Recall: {recall}")
    
    # Test task_success
    expected_state = {
        "tool_calls": [{"tool": "ship_order"}],
        "customer_msg_contains": ["shipped"]
    }
    success = task_success(response, ["ship_order"], expected_state)
    print(f"Task Success: {success}")
    
    # Test to_lc_message
    turn = {"role": "user", "content": "Hello!"}
    msg = to_lc_message(turn)
    print(f"Message Type: {type(msg).__name__}, Content: {msg.content}")
