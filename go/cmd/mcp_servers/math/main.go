package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// evalExpr safely evaluates an arithmetic expression AST node
func evalExpr(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unsupported literal type: %v", n.Kind)

	case *ast.ParenExpr:
		return evalExpr(n.X)

	case *ast.UnaryExpr:
		val, err := evalExpr(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.SUB:
			return -val, nil
		case token.ADD:
			return val, nil
		default:
			return 0, fmt.Errorf("unsupported unary operator: %v", n.Op)
		}

	case *ast.BinaryExpr:
		left, err := evalExpr(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalExpr(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("unsupported binary operator: %v", n.Op)
		}

	default:
		return 0, fmt.Errorf("unsupported expression type: %T", node)
	}
}

// calculate safely evaluates a simple arithmetic expression
func calculate(expression string) (string, error) {
	// Parse the expression
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return "", fmt.Errorf("failed to parse expression: %v", err)
	}

	result, err := evalExpr(expr)
	if err != nil {
		return "", err
	}

	// Format result nicely (remove trailing zeros)
	if result == float64(int64(result)) {
		return fmt.Sprintf("%d", int64(result)), nil
	}
	return fmt.Sprintf("%g", result), nil
}

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Math Server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add the calculate tool
	calcTool := mcp.NewTool("calculate",
		mcp.WithDescription("Evaluate a simple arithmetic expression safely. Supports: +, -, *, /, and parentheses. Examples: '3 + 5', '(10 - 2) * 4'"),
		mcp.WithString("expression",
			mcp.Required(),
			mcp.Description("A mathematical expression to evaluate (e.g., '3 + 5 * 2')"),
		),
	)

	s.AddTool(calcTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		expression, err := request.RequireString("expression")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := calculate(expression)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Calculation error: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	})

	// Start the server using stdio transport
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
