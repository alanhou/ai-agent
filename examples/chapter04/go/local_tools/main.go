package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// MathArgs - parameter schema inferred from struct tags (like LangChain's @tool docstring)
type MathArgs struct {
	X float64 `json:"x" jsonschema:"description=First number"`
	Y float64 `json:"y" jsonschema:"description=Second number"`
}

// Tool functions with signature: func(ctx, args) (result, error)
func multiply(ctx context.Context, args *MathArgs) (string, error) {
	return fmt.Sprintf("%v", args.X*args.Y), nil
}

func exponentiate(ctx context.Context, args *MathArgs) (string, error) {
	return fmt.Sprintf("%v", math.Pow(args.X, args.Y)), nil
}

func add(ctx context.Context, args *MathArgs) (string, error) {
	return fmt.Sprintf("%v", args.X+args.Y), nil
}

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	model, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"), BaseURL: os.Getenv("OPENAI_BASE_URL"), Model: "gpt-4o",
	})

	// InferTool: Go's equivalent of LangChain's @tool decorator
	// Infers schema from struct tags, just like docstrings in Python
	multiplyTool, _ := utils.InferTool("multiply", "Multiply x * y (计算 x 乘以 y)", multiply)
	expTool, _ := utils.InferTool("exponentiate", "Raise x to power y (计算 x 的 y 次幂)", exponentiate)
	addTool, _ := utils.InferTool("add", "Add x + y (计算 x 加 y)", add)

	tools := []tool.InvokableTool{multiplyTool, expTool, addTool}
	toolInfos := make([]*schema.ToolInfo, len(tools))
	toolMap := make(map[string]tool.InvokableTool)
	for i, t := range tools {
		info, _ := t.Info(ctx)
		toolInfos[i] = info
		toolMap[info.Name] = t
	}
	_ = model.BindTools(toolInfos)

	// First call - get tool calls
	msgs := []*schema.Message{schema.UserMessage("What is 393 * 12.25? Also, what is 11 + 49?")}
	resp, err := model.Generate(ctx, msgs)
	if err != nil {
		log.Fatal(err)
	}
	msgs = append(msgs, resp)

	// Execute each tool call using InvokableTool
	for _, tc := range resp.ToolCalls {
		result, _ := toolMap[tc.Function.Name].InvokableRun(ctx, tc.Function.Arguments)
		fmt.Printf("%s(%s) = %s\n", tc.Function.Name, tc.Function.Arguments, result)
		msgs = append(msgs, &schema.Message{Role: schema.Tool, Content: result, ToolCallID: tc.ID})
	}

	// Final call with tool results
	final, _ := model.Generate(ctx, msgs)
	fmt.Println(final.Content)
}
