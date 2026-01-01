package financial_services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// -- State --

type AgentState struct {
	Account  *Account          `json:"account"`
	Messages []*schema.Message `json:"messages"`
}

type Account struct {
	AccountID  string  `json:"account_id"`
	CustomerID string  `json:"customer_id"`
	Status     string  `json:"status"`
	Balance    float64 `json:"balance,omitempty"`
}

// -- Tool Args --

type InvestigateTransactionArgs struct {
	AccountID  string `json:"account_id" desc:"Account ID"`
	CustomerID string `json:"customer_id" desc:"Customer ID"`
	AlertType  string `json:"alert_type" desc:"Type of alert"`
}

type FreezeAccountArgs struct {
	AccountID  string `json:"account_id" desc:"Account ID"`
	Reason     string `json:"reason" desc:"Reason for freeze"`
	FreezeType string `json:"freeze_type" desc:"Type of freeze (immediate)"`
}

type ProcessLoanApplicationArgs struct {
	CustomerID string `json:"customer_id" desc:"Customer ID"`
	LoanType   string `json:"loan_type" desc:"Type of loan"`
	LoanAmount string `json:"loan_amount" desc:"Amount requested"`
}

type ResolveDisputeArgs struct {
	AccountID   string `json:"account_id" desc:"Account ID"`
	CustomerID  string `json:"customer_id" desc:"Customer ID"`
	DisputeType string `json:"dispute_type" desc:"Type of dispute"`
}

type RebalancePortfolioArgs struct {
	CustomerID string `json:"customer_id" desc:"Customer ID"`
}

type IncreaseCreditLimitArgs struct {
	AccountID      string `json:"account_id" desc:"Account ID"`
	CurrentLimit   string `json:"current_limit" desc:"Current limit"`
	RequestedLimit string `json:"requested_limit" desc:"Requested limit"`
}

type VerifyDocumentsArgs struct {
	CustomerID string `json:"customer_id" desc:"Customer ID"`
}

type UpdateAccountArgs struct {
	AccountID  string `json:"account_id" desc:"Account ID"`
	CustomerID string `json:"customer_id" desc:"Customer ID"`
}

type ProcessTransactionArgs struct {
	CustomerID      string `json:"customer_id" desc:"Customer ID"`
	TransactionType string `json:"transaction_type" desc:"Type of transaction"`
}

type SendCustomerResponseArgs struct {
	CustomerID string `json:"customer_id" desc:"Customer ID"`
	Message    string `json:"message" desc:"Message content"`
}

// -- Tool Impls --

func InvestigateTransaction(ctx context.Context, args *InvestigateTransactionArgs) (string, error) {
	fmt.Printf("[TOOL] investigate_transaction(acc=%s, type=%s)\n", args.AccountID, args.AlertType)
	return "investigation_initiated", nil
}

func FreezeAccount(ctx context.Context, args *FreezeAccountArgs) (string, error) {
	fmt.Printf("[TOOL] freeze_account(acc=%s, reason=%s)\n", args.AccountID, args.Reason)
	return "account_frozen", nil
}

func ProcessLoanApplication(ctx context.Context, args *ProcessLoanApplicationArgs) (string, error) {
	fmt.Printf("[TOOL] process_loan_application(cust=%s, type=%s)\n", args.CustomerID, args.LoanType)
	return "application_submitted", nil
}

func ResolveDispute(ctx context.Context, args *ResolveDisputeArgs) (string, error) {
	fmt.Printf("[TOOL] resolve_dispute(acc=%s, dispute=%s)\n", args.AccountID, args.DisputeType)
	return "dispute_filed", nil
}

func RebalancePortfolio(ctx context.Context, args *RebalancePortfolioArgs) (string, error) {
	fmt.Printf("[TOOL] rebalance_portfolio(cust=%s)\n", args.CustomerID)
	return "portfolio_updated", nil
}

func IncreaseCreditLimit(ctx context.Context, args *IncreaseCreditLimitArgs) (string, error) {
	fmt.Printf("[TOOL] increase_credit_limit(acc=%s, req=%s)\n", args.AccountID, args.RequestedLimit)
	return "credit_limit_updated", nil
}

func VerifyDocuments(ctx context.Context, args *VerifyDocumentsArgs) (string, error) {
	fmt.Printf("[TOOL] verify_documents(cust=%s)\n", args.CustomerID)
	return "documents_verified", nil
}

func UpdateAccount(ctx context.Context, args *UpdateAccountArgs) (string, error) {
	fmt.Printf("[TOOL] update_account(acc=%s)\n", args.AccountID)
	return "account_updated", nil
}

func ProcessTransaction(ctx context.Context, args *ProcessTransactionArgs) (string, error) {
	fmt.Printf("[TOOL] process_transaction(cust=%s, type=%s)\n", args.CustomerID, args.TransactionType)
	return "transaction_processed", nil
}

func SendCustomerResponse(ctx context.Context, args *SendCustomerResponseArgs) (string, error) {
	fmt.Printf("[TOOL] send_customer_response -> %s\n", args.Message)
	return "message_sent", nil
}

// -- Graph --

func NewAgent(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {
	temp := float32(0.0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init chat model: %v", err)
	}

	strParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: true}
	}
	strParamOpt := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: false}
	}

	tools := []*schema.ToolInfo{
		{
			Name: "investigate_transaction",
			Desc: "Investigate suspicious transactions, fraud alerts, or security concerns.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"account_id":  strParamOpt("Account ID"),
				"customer_id": strParamOpt("Customer ID"),
				"alert_type":  strParamOpt("Type of alert"),
			}),
		},
		{
			Name: "freeze_account",
			Desc: "Freeze account to prevent unauthorized access or transactions.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"account_id":  strParam("Account ID"),
				"reason":      strParam("Reason for freeze"),
				"freeze_type": strParamOpt("Type of freeze"),
			}),
		},
		{
			Name: "process_loan_application",
			Desc: "Process loan applications including personal, business, mortgage, and auto loans.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"customer_id": strParam("Customer ID"),
				"loan_type":   strParam("Type of loan"),
				"loan_amount": strParamOpt("Amount requested"),
			}),
		},
		{
			Name: "resolve_dispute",
			Desc: "Handle disputes including unauthorized charges, fees, and credit report errors.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"account_id":   strParamOpt("Account ID"),
				"customer_id":  strParamOpt("Customer ID"),
				"dispute_type": strParamOpt("Type of dispute"),
			}),
		},
		{
			Name: "rebalance_portfolio",
			Desc: "Manage investment portfolios, retirement planning, and asset allocation.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"customer_id": strParam("Customer ID"),
			}),
		},
		{
			Name: "increase_credit_limit",
			Desc: "Process credit limit increase requests.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"account_id":      strParam("Account ID"),
				"current_limit":   strParam("Current limit"),
				"requested_limit": strParam("Requested limit"),
			}),
		},
		{
			Name: "verify_documents",
			Desc: "Verify customer documents for various banking services.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"customer_id": strParam("Customer ID"),
			}),
		},
		{
			Name: "update_account",
			Desc: "Update account information, add joint holders, close accounts, etc.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"account_id":  strParamOpt("Account ID"),
				"customer_id": strParamOpt("Customer ID"),
			}),
		},
		{
			Name: "process_transaction",
			Desc: "Process various transactions like currency exchange, transfers, etc.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"customer_id":      strParam("Customer ID"),
				"transaction_type": strParam("Type of transaction"),
			}),
		},
		{
			Name: "send_customer_response",
			Desc: "Send a response message to the customer.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"customer_id": strParam("Customer ID"),
				"message":     strParam("Message content"),
			}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"investigate_transaction": func(ctx context.Context, args interface{}) (string, error) {
			return InvestigateTransaction(ctx, args.(*InvestigateTransactionArgs))
		},
		"freeze_account": func(ctx context.Context, args interface{}) (string, error) {
			return FreezeAccount(ctx, args.(*FreezeAccountArgs))
		},
		"process_loan_application": func(ctx context.Context, args interface{}) (string, error) {
			return ProcessLoanApplication(ctx, args.(*ProcessLoanApplicationArgs))
		},
		"resolve_dispute": func(ctx context.Context, args interface{}) (string, error) {
			return ResolveDispute(ctx, args.(*ResolveDisputeArgs))
		},
		"rebalance_portfolio": func(ctx context.Context, args interface{}) (string, error) {
			return RebalancePortfolio(ctx, args.(*RebalancePortfolioArgs))
		},
		"increase_credit_limit": func(ctx context.Context, args interface{}) (string, error) {
			return IncreaseCreditLimit(ctx, args.(*IncreaseCreditLimitArgs))
		},
		"verify_documents": func(ctx context.Context, args interface{}) (string, error) {
			return VerifyDocuments(ctx, args.(*VerifyDocumentsArgs))
		},
		"update_account": func(ctx context.Context, args interface{}) (string, error) {
			return UpdateAccount(ctx, args.(*UpdateAccountArgs))
		},
		"process_transaction": func(ctx context.Context, args interface{}) (string, error) {
			return ProcessTransaction(ctx, args.(*ProcessTransactionArgs))
		},
		"send_customer_response": func(ctx context.Context, args interface{}) (string, error) {
			return SendCustomerResponse(ctx, args.(*SendCustomerResponseArgs))
		},
	}

	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		accountJSON, _ := json.Marshal(state.Account)
		sysPrompt := fmt.Sprintf(
			"You are a professional financial services agent specializing in banking, fraud prevention, loans, and investments.\n"+
				"When you assist customers, you should:\n"+
				"  1) Analyze their request and call the appropriate business tool\n"+
				"  2) Call send_customer_response with a helpful confirmation message\n"+
				"Always prioritize security and compliance with banking regulations.\n\n"+
				"ACCOUNT: %s", string(accountJSON))

		inputMsgs := append([]*schema.Message{schema.SystemMessage(sysPrompt)}, state.Messages...)
		resp, err := chatModel.Generate(ctx, inputMsgs)
		if err != nil {
			return nil, err
		}
		state.Messages = append(state.Messages, resp)
		return state, nil
	}

	toolExecutor := func(ctx context.Context, state *AgentState) (*AgentState, error) {
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
			case "investigate_transaction":
				var a InvestigateTransactionArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "freeze_account":
				var a FreezeAccountArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "process_loan_application":
				var a ProcessLoanApplicationArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "resolve_dispute":
				var a ResolveDisputeArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "rebalance_portfolio":
				var a RebalancePortfolioArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "increase_credit_limit":
				var a IncreaseCreditLimitArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "verify_documents":
				var a VerifyDocumentsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "update_account":
				var a UpdateAccountArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "process_transaction":
				var a ProcessTransactionArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_customer_response":
				var a SendCustomerResponseArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			}

			if err != nil {
				resultStr = fmt.Sprintf("Error: %v", err)
			}
			state.Messages = append(state.Messages, &schema.Message{
				Role:       schema.Tool,
				Content:    resultStr,
				ToolCallID: tc.ID,
			})
		}
		return state, nil
	}

	g := compose.NewGraph[*AgentState, *AgentState]()
	_ = g.AddLambdaNode("assistant", compose.InvokableLambda(assistant))
	_ = g.AddLambdaNode("tools", compose.InvokableLambda(toolExecutor))
	_ = g.AddEdge(compose.START, "assistant")

	_ = g.AddBranch("assistant", compose.NewGraphBranch(func(_ context.Context, state *AgentState) (string, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return compose.END, nil
	}, map[string]bool{"tools": true, compose.END: true}))

	_ = g.AddEdge("tools", "assistant")

	return g.Compile(ctx)
}
