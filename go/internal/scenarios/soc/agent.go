package soc

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
	Incident *Incident         `json:"incident"`
	Messages []*schema.Message `json:"messages"`
}

type Incident struct {
	IncidentID string `json:"incident_id"`
	Severity   string `json:"severity"`
	Type       string `json:"type,omitempty"`
	Status     string `json:"status,omitempty"`
	Analyst    string `json:"analyst,omitempty"`
}

// -- Tool Args --

type LookupThreatIntelArgs struct {
	Indicator string `json:"indicator" desc:"IP/Hash/URL"`
	Type      string `json:"type" desc:"Type of indicator"`
}

type QueryLogsArgs struct {
	Query    string `json:"query" desc:"Search query"`
	LogIndex string `json:"log_index" desc:"Index to search"`
}

type TriageIncidentArgs struct {
	IncidentID string `json:"incident_id" desc:"Incident ID"`
	Decision   string `json:"decision" desc:"Decision"`
	Reason     string `json:"reason" desc:"Reason"`
}

type IsolateHostArgs struct {
	HostID string `json:"host_id" desc:"Host ID"`
	Reason string `json:"reason" desc:"Reason"`
}

type SendAnalystResponseArgs struct {
	IncidentID string `json:"incident_id" desc:"Incident ID"`
	Message    string `json:"message" desc:"Message content"`
}

// -- Tool Impls --

func LookupThreatIntel(ctx context.Context, args *LookupThreatIntelArgs) (string, error) {
	fmt.Printf("[TOOL] lookup_threat_intel(ind=%s, type=%s)\n", args.Indicator, args.Type)
	return "threat_intel_retrieved", nil
}

func QueryLogs(ctx context.Context, args *QueryLogsArgs) (string, error) {
	fmt.Printf("[TOOL] query_logs(q=%s, idx=%s)\n", args.Query, args.LogIndex)
	return "log_query_executed", nil
}

func TriageIncident(ctx context.Context, args *TriageIncidentArgs) (string, error) {
	fmt.Printf("[TOOL] triage_incident(id=%s, decision=%s)\n", args.IncidentID, args.Decision)
	return "incident_triaged", nil
}

func IsolateHost(ctx context.Context, args *IsolateHostArgs) (string, error) {
	fmt.Printf("[TOOL] isolate_host(host=%s)\n", args.HostID)
	return "host_isolated", nil
}

func SendAnalystResponse(ctx context.Context, args *SendAnalystResponseArgs) (string, error) {
	fmt.Printf("[TOOL] send_analyst_response -> %s\n", args.Message)
	return "analyst_response_sent", nil
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
			Name: "lookup_threat_intel", Desc: "Threat Intel Lookup.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"indicator": strParam("Indicator"), "type": strParam("Type")}),
		},
		{
			Name: "query_logs", Desc: "Query security logs.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"query": strParam("Query"), "log_index": strParam("Log Index")}),
		},
		{
			Name: "triage_incident", Desc: "Triage incident.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"incident_id": strParam("ID"), "decision": strParam("Decision"), "reason": strParam("Reason")}),
		},
		{
			Name: "isolate_host", Desc: "Isolate compromised host.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"host_id": strParam("Host ID"), "reason": strParam("Reason")}),
		},
		{
			Name: "send_analyst_response", Desc: "Send analyst response.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"incident_id": strParamOpt("ID"), "message": strParam("Message")}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"lookup_threat_intel": func(ctx context.Context, args interface{}) (string, error) {
			return LookupThreatIntel(ctx, args.(*LookupThreatIntelArgs))
		},
		"query_logs": func(ctx context.Context, args interface{}) (string, error) {
			return QueryLogs(ctx, args.(*QueryLogsArgs))
		},
		"triage_incident": func(ctx context.Context, args interface{}) (string, error) {
			return TriageIncident(ctx, args.(*TriageIncidentArgs))
		},
		"isolate_host": func(ctx context.Context, args interface{}) (string, error) {
			return IsolateHost(ctx, args.(*IsolateHostArgs))
		},
		"send_analyst_response": func(ctx context.Context, args interface{}) (string, error) {
			return SendAnalystResponse(ctx, args.(*SendAnalystResponseArgs))
		},
	}

	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		incJSON, _ := json.Marshal(state.Incident)
		sysPrompt := fmt.Sprintf(
			"You are a SOC analyst.\n"+
				"Roles: Threat Intel, Log Analysis, Triage, Host Isolation.\n"+
				"1) Use tools.\n2) send_analyst_response.\n\n"+
				"INCIDENT: %s", string(incJSON))

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
			case "lookup_threat_intel":
				var a LookupThreatIntelArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "query_logs":
				var a QueryLogsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "triage_incident":
				var a TriageIncidentArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "isolate_host":
				var a IsolateHostArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_analyst_response":
				var a SendAnalystResponseArgs
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
