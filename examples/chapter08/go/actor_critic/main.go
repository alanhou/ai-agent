package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// Tool definitions
var tools = []*schema.ToolInfo{
	{Name: "manage_inventory", Desc: "管理库存水平、库存补货、审计和优化策略。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"sku": {Type: "string", Desc: "SKU identifier"}})},
	{Name: "optimize_warehouse", Desc: "优化仓库运营、布局、容量和存储效率。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": {Type: "string", Desc: "Type of operation"}})},
	{Name: "forecast_demand", Desc: "分析需求模式、季节性趋势并创建预测模型。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"season": {Type: "string", Desc: "Season"}})},
	{Name: "track_shipments", Desc: "跟踪货物状态、延误并协调交付物流。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"origin": {Type: "string", Desc: "Origin location"}})},
	{Name: "arrange_shipping", Desc: "安排运输方式、加急交付和多式联运。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"shipping_type": {Type: "string", Desc: "Shipping type"}})},
	{Name: "evaluate_suppliers", Desc: "评估供应商绩效、进行审计并管理供应商关系。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier_name": {Type: "string", Desc: "Supplier name"}})},
	{Name: "optimize_costs", Desc: "分析并优化运输、仓储和运营成本。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"cost_type": {Type: "string", Desc: "Cost type"}})},
	{Name: "send_logistics_response", Desc: "向利益相关者发送物流更新、建议或状态报告。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_id": {Type: "string", Desc: "Operation ID"}, "message": {Type: "string", Desc: "Message"}})},
}

// Candidate plan structure
type CandidatePlan struct {
	Plan  string     `json:"plan"`
	Tools []ToolCall `json:"tools"`
}

type ToolCall struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

// Critic evaluation result
type CriticEvaluation struct {
	Evaluations []struct {
		PlanIndex   int `json:"plan_index"`
		Feasibility int `json:"feasibility"`
		Cost        int `json:"cost"`
		Risk        int `json:"risk"`
		Total       int `json:"total"`
	} `json:"evaluations"`
	BestIndex int            `json:"best_index"`
	BestScore float64        `json:"best_score"`
	Selected  *CandidatePlan `json:"selected"`
	Feedback  string         `json:"feedback"`
}

// Execute tool
func executeTool(name string, args map[string]interface{}) string {
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("[TOOL] %s(%s)\n", name, string(argsJSON))

	results := map[string]string{
		"manage_inventory":        "inventory_management_initiated",
		"optimize_warehouse":      "warehouse_optimization_initiated",
		"forecast_demand":         "demand_forecast_generated",
		"track_shipments":         "shipment_tracking_updated",
		"arrange_shipping":        "shipping_arranged",
		"evaluate_suppliers":      "supplier_evaluation_complete",
		"optimize_costs":          "cost_optimization_initiated",
		"send_logistics_response": "logistics_response_sent",
	}
	if r, ok := results[name]; ok {
		return r
	}
	return "unknown_tool_result"
}

// Extract JSON from response (handles markdown code blocks)
func extractJSON(content string) string {
	if strings.Contains(content, "```json") {
		parts := strings.Split(content, "```json")
		if len(parts) > 1 {
			inner := strings.Split(parts[1], "```")
			if len(inner) > 0 {
				return strings.TrimSpace(inner[0])
			}
		}
	} else if strings.Contains(content, "```") {
		parts := strings.Split(content, "```")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return content
}

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to init chat model: %v", err)
	}

	operation := map[string]string{
		"operation_id": "OP-12345",
		"type":         "inventory_management",
		"priority":     "high",
		"location":     "Warehouse A",
	}
	operationJSON, _ := json.Marshal(operation)

	userMessage := "We're running critically low on SKU-12345. Current stock is 50 units but we have 200 units on backorder. What's our reorder strategy?"

	fmt.Println("=== Actor-Critic Supply Chain Agent (Go) ===")
	fmt.Printf("User: %s\n\n", userMessage)

	maxIterations := 3
	var candidates []CandidatePlan

	for iteration := 0; iteration < maxIterations; iteration++ {
		fmt.Printf("--- Iteration %d ---\n", iteration+1)

		// === ACTOR NODE ===
		fmt.Println("Actor: Generating candidate plans...")

		actorPrompt := fmt.Sprintf(`你是供应链策略专家。基于用户需求，生成 3 个候选供应链计划。
请以 JSON 格式返回，格式如下:
[
  {"plan": "计划描述", "tools": [{"tool": "tool_name", "args": {"key": "value"}}]},
  ...
]

可用工具: manage_inventory, optimize_warehouse, forecast_demand, track_shipments, 
arrange_shipping, evaluate_suppliers, optimize_costs, send_logistics_response

当前运营数据: %s`, string(operationJSON))

		actorMessages := []*schema.Message{
			schema.SystemMessage(actorPrompt),
			schema.UserMessage(userMessage),
		}

		actorResp, err := chatModel.Generate(ctx, actorMessages)
		if err != nil {
			log.Fatalf("Actor error: %v", err)
		}

		// Parse candidates
		jsonContent := extractJSON(actorResp.Content)
		if err := json.Unmarshal([]byte(jsonContent), &candidates); err != nil {
			log.Printf("Warning: Failed to parse actor response: %v", err)
			candidates = []CandidatePlan{{Plan: actorResp.Content, Tools: nil}}
		}

		fmt.Printf("Actor generated %d candidates\n", len(candidates))

		// === CRITIC NODE ===
		fmt.Println("Critic: Evaluating plans...")

		candidatesJSON, _ := json.MarshalIndent(candidates, "", "  ")
		criticPrompt := fmt.Sprintf(`你是供应链评估专家。评估以下候选计划:

%s

对每个计划按以下维度打分（1-10分）:
- 可行性 (feasibility)
- 成本效益 (cost_effectiveness)  
- 风险控制 (risk_management)

返回 JSON 格式:
{
  "evaluations": [
    {"plan_index": 0, "feasibility": 8, "cost": 7, "risk": 9, "total": 24},
    ...
  ],
  "best_index": 0,
  "best_score": 8.0,
  "selected": {"plan": "...", "tools": [...]},
  "feedback": "改进建议（如果最高分 <= 8）"
}

如果最高平均分 > 8，选择最佳计划；否则提供改进反馈。`, string(candidatesJSON))

		criticMessages := []*schema.Message{
			schema.SystemMessage(criticPrompt),
			schema.UserMessage(userMessage),
		}

		criticResp, err := chatModel.Generate(ctx, criticMessages)
		if err != nil {
			log.Fatalf("Critic error: %v", err)
		}

		// Parse evaluation
		var evalResult CriticEvaluation
		jsonContent = extractJSON(criticResp.Content)
		if err := json.Unmarshal([]byte(jsonContent), &evalResult); err != nil {
			log.Printf("Warning: Failed to parse critic response: %v", err)
			continue
		}

		fmt.Printf("Critic best score: %.1f\n", evalResult.BestScore)

		if evalResult.BestScore > 8 && evalResult.Selected != nil {
			// Execute winning plan
			fmt.Printf("\n=== Executing Winning Plan ===\n")
			fmt.Printf("Plan: %s\n\n", evalResult.Selected.Plan)

			for _, tc := range evalResult.Selected.Tools {
				result := executeTool(tc.Tool, tc.Args)
				fmt.Printf("Result: %s\n", result)
			}

			// Send final response
			executeTool("send_logistics_response", map[string]interface{}{
				"operation_id": operation["operation_id"],
				"message":      evalResult.Selected.Plan,
			})

			fmt.Println("\n=== Actor-Critic Complete ===")
			return
		}

		fmt.Printf("Feedback: %s\n", evalResult.Feedback)
		fmt.Println("Continuing to next iteration...")

		// Add feedback to user message for next iteration
		userMessage = fmt.Sprintf("%s\n\nPrevious feedback: %s", userMessage, evalResult.Feedback)
	}

	fmt.Println("\n=== Max iterations reached ===")
}
