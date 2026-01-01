package ecommerce_customer_support

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AgentState mimics the Python TypedDict state.
type AgentState struct {
	Order    *Order            `json:"order"`
	Messages []*schema.Message `json:"messages"`
}

type Order struct {
	OrderID    string  `json:"order_id"`
	Status     string  `json:"status"`
	Total      float64 `json:"total"`
	CustomerID string  `json:"customer_id"`
}

// Tool Args Definitions
type SendCustomerMessageArgs struct {
	OrderID string `json:"order_id" desc:"The ID of the order"`
	Text    string `json:"text" desc:"The content of the message to send to the customer"`
}

type IssueRefundArgs struct {
	OrderID string  `json:"order_id" desc:"The ID of the order"`
	Amount  float64 `json:"amount" desc:"The amount to refund"`
}

type CancelOrderArgs struct {
	OrderID string `json:"order_id" desc:"The ID of the order"`
}

type UpdateAddressArgs struct {
	OrderID         string `json:"order_id" desc:"The ID of the order"`
	ShippingAddress string `json:"shipping_address" desc:"The new shipping address"`
}

// -- Tool Implementations --

func SendCustomerMessage(ctx context.Context, args *SendCustomerMessageArgs) (string, error) {
	fmt.Printf("[TOOL] send_customer_message -> %s (Order: %s)\n", args.Text, args.OrderID)
	return "sent", nil
}

func IssueRefund(ctx context.Context, args *IssueRefundArgs) (string, error) {
	fmt.Printf("[TOOL] issue_refund(order_id=%s, amount=%.2f)\n", args.OrderID, args.Amount)
	return "refund_queued", nil
}

func CancelOrder(ctx context.Context, args *CancelOrderArgs) (string, error) {
	fmt.Printf("[TOOL] cancel_order(order_id=%s)\n", args.OrderID)
	return "cancelled", nil
}

func UpdateAddressForOrder(ctx context.Context, args *UpdateAddressArgs) (string, error) {
	fmt.Printf("[TOOL] update_address_for_order(order_id=%s, address=%s)\n", args.OrderID, args.ShippingAddress)
	return "address_updated", nil
}

// NewAgent creates the runnable graph
func NewAgent(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {

	// 1. Model Init
	temp := float32(0.0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init chat model: %v", err)
	}

	// 2. Bind Tools
	// Helper to create ParameterInfo
	strParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: true}
	}
	numParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.Number, Desc: desc, Required: true}
	}

	tools := []*schema.ToolInfo{
		{
			Name: "send_customer_message",
			Desc: "Send a plain response to the customer.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"order_id": strParam("The ID of the order"),
				"text":     strParam("The content of the message to send to the customer"),
			}),
		},
		{
			Name: "issue_refund",
			Desc: "Issue a refund for the given order.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"order_id": strParam("The ID of the order"),
				"amount":   numParam("The amount to refund"),
			}),
		},
		{
			Name: "cancel_order",
			Desc: "Cancel an order that hasn't shipped.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"order_id": strParam("The ID of the order"),
			}),
		},
		{
			Name: "update_address_for_order",
			Desc: "Change the shipping address for a pending order.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"order_id":         strParam("The ID of the order"),
				"shipping_address": strParam("The new shipping address"),
			}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	// Tool Executors
	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"send_customer_message": func(ctx context.Context, args interface{}) (string, error) {
			return SendCustomerMessage(ctx, args.(*SendCustomerMessageArgs))
		},
		"issue_refund": func(ctx context.Context, args interface{}) (string, error) {
			return IssueRefund(ctx, args.(*IssueRefundArgs))
		},
		"cancel_order": func(ctx context.Context, args interface{}) (string, error) {
			return CancelOrder(ctx, args.(*CancelOrderArgs))
		},
		"update_address_for_order": func(ctx context.Context, args interface{}) (string, error) {
			return UpdateAddressForOrder(ctx, args.(*UpdateAddressArgs))
		},
	}

	// 3. Nodes

	// Assistant Node
	assistant := compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		// Prepare System Prompt
		orderJSON, _ := json.Marshal(state.Order)
		sysPrompt := fmt.Sprintf(
			"You are a helpful e-commerce support agent.\n"+
				"When you act, you MUST do exactly TWO steps in order:\n"+
				"  1) call one business tool (issue_refund / cancel_order / modify_order)\n"+
				"  2) call send_customer_message with confirmation text\n"+
				"Then STOP.\n\n"+
				"ORDER: %s", string(orderJSON),
		)

		// Construct full messages with System Prompt at start
		inputMsgs := append([]*schema.Message{schema.SystemMessage(sysPrompt)}, state.Messages...)

		// Call Model
		resp, err := chatModel.Generate(ctx, inputMsgs)
		if err != nil {
			return nil, err
		}

		// Append result to history
		state.Messages = append(state.Messages, resp)
		return state, nil
	})

	// Tool Executor Node
	toolExecutor := compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) == 0 {
			return state, nil
		}

		for _, tc := range lastMsg.ToolCalls {
			handler, ok := toolHandlers[tc.Function.Name]
			if !ok {
				log.Printf("Tool %s not found", tc.Function.Name)
				continue
			}

			var resultStr string
			var err error

			switch tc.Function.Name {
			case "send_customer_message":
				var args SendCustomerMessageArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				resultStr, err = handler(ctx, &args)
			case "issue_refund":
				var args IssueRefundArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				resultStr, err = handler(ctx, &args)
			case "cancel_order":
				var args CancelOrderArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				resultStr, err = handler(ctx, &args)
			case "update_address_for_order":
				var args UpdateAddressArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				resultStr, err = handler(ctx, &args)
			}

			if err != nil {
				resultStr = fmt.Sprintf("Error: %v", err)
			}

			// Append Tool Message
			state.Messages = append(state.Messages, &schema.Message{
				Role:       schema.Tool,
				Content:    resultStr,
				ToolCallID: tc.ID,
			})
		}
		return state, nil
	})

	// 4. Graph Definition
	g := compose.NewGraph[*AgentState, *AgentState]()

	_ = g.AddLambdaNode("assistant", assistant)
	_ = g.AddLambdaNode("tools", toolExecutor)

	_ = g.AddEdge(compose.START, "assistant")

	// Branch from Assistant
	branch := compose.NewGraphBranch(func(_ context.Context, state *AgentState) (string, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return compose.END, nil
	}, map[string]bool{"tools": true, compose.END: true})

	_ = g.AddBranch("assistant", branch)

	// Loop back from Tools to Assistant
	_ = g.AddEdge("tools", "assistant")

	return g.Compile(ctx)
}
