#!/usr/bin/env python3
"""
Math MCP Server using FastMCP framework.
Provides a 'calculate' tool for evaluating arithmetic expressions.
"""
import ast
import operator
from fastmcp import FastMCP

# Create the MCP server
mcp = FastMCP("Math Server")

# ─── Safe Expression Evaluation ────────────────────────────────────────────────
# Restrict nodes to only arithmetic operators (+, -, *, /, **, parentheses).
# This prevents code injection via eval().

ALLOWED_OPERATORS = {
    ast.Add: operator.add,
    ast.Sub: operator.sub,
    ast.Mult: operator.mul,
    ast.Div: operator.truediv,
    ast.Pow: operator.pow,
    ast.USub: operator.neg,
}


def eval_expr(node: ast.AST) -> float:
    if isinstance(node, ast.Constant):  # Python 3.8+
        return node.value
    if isinstance(node, ast.Num):  # Python < 3.8 compatibility
        return node.n
    if isinstance(node, ast.BinOp) and type(node.op) in ALLOWED_OPERATORS:
        left = eval_expr(node.left)
        right = eval_expr(node.right)
        return ALLOWED_OPERATORS[type(node.op)](left, right)
    if isinstance(node, ast.UnaryOp) and type(node.op) in ALLOWED_OPERATORS:
        operand = eval_expr(node.operand)
        return ALLOWED_OPERATORS[type(node.op)](operand)
    raise ValueError(f"Unsupported expression: {ast.dump(node)}")


def compute_math(expression: str) -> float:
    """Parse and evaluate a simple arithmetic expression safely."""
    try:
        expr_ast = ast.parse(expression, mode="eval").body
        return eval_expr(expr_ast)
    except Exception as e:
        raise ValueError(f"Error parsing expression '{expression}': {e}")


@mcp.tool
def calculate(expression: str) -> str:
    """
    Evaluate a simple arithmetic expression safely.
    
    Supports: +, -, *, /, ** (power), and parentheses.
    Examples: "3 + 5", "(10 - 2) * 4", "2 ** 8"
    
    Args:
        expression: A mathematical expression to evaluate (e.g., "3 + 5 * 2")
    
    Returns:
        The result of the calculation as a string.
    """
    # Clean the expression: remove text, keep only math symbols
    cleaned = "".join(ch for ch in expression if ch.isdigit() or ch in "+-*/().^ ")
    cleaned = cleaned.replace("^", "**")  # Allow caret as power operator
    
    if not cleaned.strip():
        return "Error: No valid mathematical expression found."
    
    try:
        result = compute_math(cleaned)
        return str(result)
    except ValueError as e:
        return f"Error: {e}"


if __name__ == "__main__":
    # Run using stdio transport (default) for subprocess communication
    mcp.run()
