package it_help_desk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// -- State Definitions --

type AgentState struct {
	Ticket   *Ticket           `json:"ticket"`
	Messages []*schema.Message `json:"messages"`
}

type Ticket struct {
	TicketID string `json:"ticket_id"`
	UserID   string `json:"user_id"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
	Category string `json:"category,omitempty"`
}

// -- Tool Definitions --

type ProvisionUserAccessArgs struct {
	UserID string `json:"user_id" desc:"The ID of the user"`
	Action string `json:"action" desc:"The action to perform (e.g. grant_access)"`
}

func ProvisionUserAccess(ctx context.Context, args *ProvisionUserAccessArgs) (string, error) {
	fmt.Printf("[TOOL] provision_user_access(user_id=%s, action=%s)\n", args.UserID, args.Action)
	return "user_access_updated", nil
}

type TroubleshootNetworkArgs struct {
	Issue    string `json:"issue" desc:"The network issue description"`
	Location string `json:"location" desc:"The location of the issue"`
}

func TroubleshootNetwork(ctx context.Context, args *TroubleshootNetworkArgs) (string, error) {
	fmt.Printf("[TOOL] troubleshoot_network(issue=%s, location=%s)\n", args.Issue, args.Location)
	return "network_issue_diagnosed", nil
}

type DiagnoseSystemIssueArgs struct {
	System  string `json:"system" desc:"The system identifier"`
	Issue   string `json:"issue" desc:"The issue description"`
	Service string `json:"service" desc:"The specific service involved"`
}

func DiagnoseSystemIssue(ctx context.Context, args *DiagnoseSystemIssueArgs) (string, error) {
	fmt.Printf("[TOOL] diagnose_system_issue(system=%s, issue=%s, service=%s)\n", args.System, args.Issue, args.Service)
	return "system_diagnosis_complete", nil
}

type DeploySoftwareArgs struct {
	Software string `json:"software" desc:"Name of the software"`
	Action   string `json:"action" desc:"Action like install/update"`
}

func DeploySoftware(ctx context.Context, args *DeploySoftwareArgs) (string, error) {
	fmt.Printf("[TOOL] deploy_software(software=%s, action=%s)\n", args.Software, args.Action)
	return "software_deployment_initiated", nil
}

type ContainSecurityIncidentArgs struct {
	IncidentType   string `json:"incident_type" desc:"Type of incident (malware, etc)"`
	AffectedSystem string `json:"affected_system" desc:"System affected"`
}

func ContainSecurityIncident(ctx context.Context, args *ContainSecurityIncidentArgs) (string, error) {
	fmt.Printf("[TOOL] contain_security_incident(type=%s, system=%s)\n", args.IncidentType, args.AffectedSystem)
	return "security_incident_contained", nil
}

type TroubleshootHardwareArgs struct {
	Device   string `json:"device" desc:"Device identifier"`
	Location string `json:"location" desc:"Location"`
	Issue    string `json:"issue" desc:"Issue description"`
}

func TroubleshootHardware(ctx context.Context, args *TroubleshootHardwareArgs) (string, error) {
	fmt.Printf("[TOOL] troubleshoot_hardware(device=%s, issue=%s)\n", args.Device, args.Issue)
	return "hardware_troubleshooting_initiated", nil
}

type AssignRolesArgs struct {
	UserID  string `json:"user_id" desc:"User ID"`
	NewRole string `json:"new_role" desc:"Role to assign"`
}

func AssignRoles(ctx context.Context, args *AssignRolesArgs) (string, error) {
	fmt.Printf("[TOOL] assign_roles(user=%s, role=%s)\n", args.UserID, args.NewRole)
	return "role_assignment_complete", nil
}

type EscalateIncidentArgs struct {
	IncidentID string `json:"incident_id" desc:"Incident ID"`
	EscalateTo string `json:"escalate_to" desc:"Team/Person to escalate to"`
}

func EscalateIncident(ctx context.Context, args *EscalateIncidentArgs) (string, error) {
	fmt.Printf("[TOOL] escalate_incident(id=%s, to=%s)\n", args.IncidentID, args.EscalateTo)
	return "incident_escalated", nil
}

type ApplyPatchesArgs struct {
	TargetSystems string `json:"target_systems" desc:"Systems to patch"`
	PatchType     string `json:"patch_type" desc:"Type of patch"`
}

func ApplyPatches(ctx context.Context, args *ApplyPatchesArgs) (string, error) {
	fmt.Printf("[TOOL] apply_patches(target=%s, type=%s)\n", args.TargetSystems, args.PatchType)
	return "patch_deployment_scheduled", nil
}

type SendUserResponseArgs struct {
	UserID  string `json:"user_id" desc:"User ID"`
	Message string `json:"message" desc:"Message content"`
}

func SendUserResponse(ctx context.Context, args *SendUserResponseArgs) (string, error) {
	fmt.Printf("[TOOL] send_user_response -> %s\n", args.Message)
	return "response_sent", nil
}

// -- Agent Construction --

func NewAgent(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {
	temp := float32(0.0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init chat model: %v", err)
	}

	// Define Tools using schema.NewParamsOneOfByParams
	strParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: true}
	}
	// Optional parameter helper
	strParamOpt := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: false}
	}

	tools := []*schema.ToolInfo{
		{
			Name: "provision_user_access",
			Desc: "Manage user access including account creation, password resets, permissions, and account termination.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"user_id": strParam("The ID of the user"),
				"action":  strParamOpt("The action to perform (e.g. grant_access)"),
			}),
		},
		{
			Name: "troubleshoot_network",
			Desc: "Diagnose and resolve network connectivity issues including WiFi, VPN, internet, and firewall problems.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"issue":    strParam("The network issue description"),
				"location": strParamOpt("The location of the issue"),
			}),
		},
		{
			Name: "diagnose_system_issue",
			Desc: "Diagnose server, database, application, and system performance issues.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"system":  strParam("The system identifier"),
				"issue":   strParam("The issue description"),
				"service": strParamOpt("The specific service involved"),
			}),
		},
		{
			Name: "deploy_software",
			Desc: "Handle software installation, updates, license management, and deployment.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"software": strParam("Name of the software"),
				"action":   strParamOpt("Action like install/update"),
			}),
		},
		{
			Name: "contain_security_incident",
			Desc: "Respond to security incidents including malware, ransomware, phishing, and breaches.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"incident_type":   strParam("Type of incident (malware, etc)"),
				"affected_system": strParam("System affected"),
			}),
		},
		{
			Name: "troubleshoot_hardware",
			Desc: "Diagnose and resolve hardware issues with printers, projectors, computers, and peripherals.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"device":   strParam("Device identifier"),
				"location": strParamOpt("Location"),
				"issue":    strParam("Issue description"),
			}),
		},
		{
			Name: "assign_roles",
			Desc: "Manage user roles, permissions, and security policy enforcement.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"user_id":  strParam("User ID"),
				"new_role": strParam("Role to assign"),
			}),
		},
		{
			Name: "escalate_incident",
			Desc: "Escalate complex issues to higher-level support teams or specialists.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"incident_id": strParam("Incident ID"),
				"escalate_to": strParam("Team/Person to escalate to"),
			}),
		},
		{
			Name: "apply_patches",
			Desc: "Apply system patches, updates, and security fixes to infrastructure.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"target_systems": strParam("Systems to patch"),
				"patch_type":     strParamOpt("Type of patch"),
			}),
		},
		{
			Name: "send_user_response",
			Desc: "Send a response or status update to the user.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"user_id": strParamOpt("User ID"),
				"message": strParam("Message content"),
			}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	// Handlers map
	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"provision_user_access": func(ctx context.Context, args interface{}) (string, error) {
			return ProvisionUserAccess(ctx, args.(*ProvisionUserAccessArgs))
		},
		"troubleshoot_network": func(ctx context.Context, args interface{}) (string, error) {
			return TroubleshootNetwork(ctx, args.(*TroubleshootNetworkArgs))
		},
		"diagnose_system_issue": func(ctx context.Context, args interface{}) (string, error) {
			return DiagnoseSystemIssue(ctx, args.(*DiagnoseSystemIssueArgs))
		},
		"deploy_software": func(ctx context.Context, args interface{}) (string, error) {
			return DeploySoftware(ctx, args.(*DeploySoftwareArgs))
		},
		"contain_security_incident": func(ctx context.Context, args interface{}) (string, error) {
			return ContainSecurityIncident(ctx, args.(*ContainSecurityIncidentArgs))
		},
		"troubleshoot_hardware": func(ctx context.Context, args interface{}) (string, error) {
			return TroubleshootHardware(ctx, args.(*TroubleshootHardwareArgs))
		},
		"assign_roles": func(ctx context.Context, args interface{}) (string, error) {
			return AssignRoles(ctx, args.(*AssignRolesArgs))
		},
		"escalate_incident": func(ctx context.Context, args interface{}) (string, error) {
			return EscalateIncident(ctx, args.(*EscalateIncidentArgs))
		},
		"apply_patches": func(ctx context.Context, args interface{}) (string, error) {
			return ApplyPatches(ctx, args.(*ApplyPatchesArgs))
		},
		"send_user_response": func(ctx context.Context, args interface{}) (string, error) {
			return SendUserResponse(ctx, args.(*SendUserResponseArgs))
		},
	}

	// -- Nodes --
	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		ticketJSON, _ := json.Marshal(state.Ticket)
		sysPrompt := fmt.Sprintf(
			"You are an experienced IT Help Desk technician and system administrator.\n"+
				"Your expertise covers: User access, Network, System diagnostics, Software, Security, Hardware, Infrastructure.\n"+
				"When helping users:\n"+
				"  1) Analyze the technical issue and call the appropriate diagnostic/resolution tool\n"+
				"  2) Follow up with send_user_response to explain what actions were taken\n"+
				"  3) Escalate complex issues when they exceed your support level\n\n"+
				"TICKET: %s", string(ticketJSON))

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

			// Dispatch unmarshalling
			switch tc.Function.Name {
			case "provision_user_access":
				var a ProvisionUserAccessArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "troubleshoot_network":
				var a TroubleshootNetworkArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "diagnose_system_issue":
				var a DiagnoseSystemIssueArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "deploy_software":
				var a DeploySoftwareArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "contain_security_incident":
				var a ContainSecurityIncidentArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "troubleshoot_hardware":
				var a TroubleshootHardwareArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "assign_roles":
				var a AssignRolesArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "escalate_incident":
				var a EscalateIncidentArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "apply_patches":
				var a ApplyPatchesArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_user_response":
				var a SendUserResponseArgs
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
