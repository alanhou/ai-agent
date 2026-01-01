package legal

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
	Matter   *Matter           `json:"matter"`
	Messages []*schema.Message `json:"messages"`
}

type Matter struct {
	MatterID   string `json:"matter_id"`
	ClientID   string `json:"client_id"`
	MatterType string `json:"matter_type"`
	Status     string `json:"status,omitempty"`
}

// -- Tool Args --

type ReviewContractArgs struct {
	ContractType string `json:"contract_type" desc:"Type of contract"`
}

type ResearchCaseLawArgs struct {
	CaseType     string `json:"case_type" desc:"Type of case"`
	Jurisdiction string `json:"jurisdiction" desc:"Jurisdiction"`
}

type ClientIntakeArgs struct {
	ClientName string `json:"client_name" desc:"Client Name"`
	MatterType string `json:"matter_type" desc:"Matter Type"`
}

type AssessComplianceArgs struct {
	Regulations    []string `json:"regulations" desc:"Regulations"`
	ClientIndustry string   `json:"client_industry" desc:"Industry"`
}

type ManageDiscoveryArgs struct {
	DiscoveryType string `json:"discovery_type" desc:"Type of discovery"`
}

type CalculateDamagesArgs struct {
	CaseType string `json:"case_type" desc:"Type of case"`
}

type TrackDeadlinesArgs struct {
	CaseName       string `json:"case_name" desc:"Case Name"`
	DocumentType   string `json:"document_type" desc:"Doc Type"`
	FilingDeadline string `json:"filing_deadline" desc:"Deadline"`
}

type SendLegalResponseArgs struct {
	ClientID string `json:"client_id" desc:"Client ID"`
	Message  string `json:"message" desc:"Message content"`
}

// -- Tool Impls --

func ReviewContract(ctx context.Context, args *ReviewContractArgs) (string, error) {
	fmt.Printf("[TOOL] review_contract(type=%s)\n", args.ContractType)
	return "contract_review_complete", nil
}

func ResearchCaseLaw(ctx context.Context, args *ResearchCaseLawArgs) (string, error) {
	fmt.Printf("[TOOL] research_case_law(type=%s, juris=%s)\n", args.CaseType, args.Jurisdiction)
	return "legal_research_complete", nil
}

func ClientIntake(ctx context.Context, args *ClientIntakeArgs) (string, error) {
	fmt.Printf("[TOOL] client_intake(name=%s, type=%s)\n", args.ClientName, args.MatterType)
	return "client_intake_complete", nil
}

func AssessCompliance(ctx context.Context, args *AssessComplianceArgs) (string, error) {
	fmt.Printf("[TOOL] assess_compliance(ind=%s)\n", args.ClientIndustry)
	return "compliance_assessment_complete", nil
}

func ManageDiscovery(ctx context.Context, args *ManageDiscoveryArgs) (string, error) {
	fmt.Printf("[TOOL] manage_discovery(type=%s)\n", args.DiscoveryType)
	return "discovery_management_initiated", nil
}

func CalculateDamages(ctx context.Context, args *CalculateDamagesArgs) (string, error) {
	fmt.Printf("[TOOL] calculate_damages(type=%s)\n", args.CaseType)
	return "damages_calculation_complete", nil
}

func TrackDeadlines(ctx context.Context, args *TrackDeadlinesArgs) (string, error) {
	fmt.Printf("[TOOL] track_deadlines(case=%s, dead=%s)\n", args.CaseName, args.FilingDeadline)
	return "deadline_tracking_updated", nil
}

func SendLegalResponse(ctx context.Context, args *SendLegalResponseArgs) (string, error) {
	fmt.Printf("[TOOL] send_legal_response -> %s\n", args.Message)
	return "legal_response_sent", nil
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
	arrParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.Array, Desc: desc, Required: false, ElemInfo: &schema.ParameterInfo{Type: schema.String}}
	}

	tools := []*schema.ToolInfo{
		{
			Name: "review_contract", Desc: "Review contracts.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"contract_type": strParamOpt("Contract Type")}),
		},
		{
			Name: "research_case_law", Desc: "Research case law.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"case_type": strParamOpt("Case Type"), "jurisdiction": strParamOpt("Jurisdiction")}),
		},
		{
			Name: "client_intake", Desc: "Process client intake.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"client_name": strParamOpt("Client Name"), "matter_type": strParamOpt("Matter Type")}),
		},
		{
			Name: "assess_compliance", Desc: "Assess compliance.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"regulations": arrParam("Regulations"), "client_industry": strParamOpt("Industry")}),
		},
		{
			Name: "manage_discovery", Desc: "Manage discovery.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"discovery_type": strParamOpt("Type")}),
		},
		{
			Name: "calculate_damages", Desc: "Calculate damages.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"case_type": strParamOpt("Case Type")}),
		},
		{
			Name: "track_deadlines", Desc: "Track deadlines.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"case_name": strParamOpt("Case"), "document_type": strParamOpt("Doc Type"), "filing_deadline": strParamOpt("Deadline")}),
		},
		{
			Name: "send_legal_response", Desc: "Send legal response.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"client_id": strParamOpt("Client ID"), "message": strParam("Message")}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"review_contract": func(ctx context.Context, args interface{}) (string, error) {
			return ReviewContract(ctx, args.(*ReviewContractArgs))
		},
		"research_case_law": func(ctx context.Context, args interface{}) (string, error) {
			return ResearchCaseLaw(ctx, args.(*ResearchCaseLawArgs))
		},
		"client_intake": func(ctx context.Context, args interface{}) (string, error) {
			return ClientIntake(ctx, args.(*ClientIntakeArgs))
		},
		"assess_compliance": func(ctx context.Context, args interface{}) (string, error) {
			return AssessCompliance(ctx, args.(*AssessComplianceArgs))
		},
		"manage_discovery": func(ctx context.Context, args interface{}) (string, error) {
			return ManageDiscovery(ctx, args.(*ManageDiscoveryArgs))
		},
		"calculate_damages": func(ctx context.Context, args interface{}) (string, error) {
			return CalculateDamages(ctx, args.(*CalculateDamagesArgs))
		},
		"track_deadlines": func(ctx context.Context, args interface{}) (string, error) {
			return TrackDeadlines(ctx, args.(*TrackDeadlinesArgs))
		},
		"send_legal_response": func(ctx context.Context, args interface{}) (string, error) {
			return SendLegalResponse(ctx, args.(*SendLegalResponseArgs))
		},
	}

	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		matterJSON, _ := json.Marshal(state.Matter)
		sysPrompt := fmt.Sprintf(
			"You are an experienced legal professional.\n"+
				"Roles: Contract Review, Research, Intake, Compliance, Discovery, Damages, Deadlines.\n"+
				"1) Call appropriate tool.\n2) send_legal_response.\n\n"+
				"MATTER: %s", string(matterJSON))

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
			case "review_contract":
				var a ReviewContractArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "research_case_law":
				var a ResearchCaseLawArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "client_intake":
				var a ClientIntakeArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "assess_compliance":
				var a AssessComplianceArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "manage_discovery":
				var a ManageDiscoveryArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "calculate_damages":
				var a CalculateDamagesArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "track_deadlines":
				var a TrackDeadlinesArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_legal_response":
				var a SendLegalResponseArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			}
			if err != nil {
				resultStr = fmt.Sprintf("Error: %v", err)
			}
			state.Messages = append(state.Messages, &schema.Message{Role: schema.Tool, Content: resultStr, ToolCallID: tc.ID})
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
