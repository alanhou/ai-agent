package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// Tool sets for specialists
var inventoryTools = []*schema.ToolInfo{
	{Name: "manage_inventory", Desc: "管理库存水平、库存补货、审计和优化策略。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"sku": {Type: "string", Desc: "SKU identifier"}})},
	{Name: "optimize_warehouse", Desc: "优化仓库运营、布局、容量和存储效率。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": {Type: "string", Desc: "Type of operation"}})},
	{Name: "forecast_demand", Desc: "分析需求模式、季节性趋势并创建预测模型。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"season": {Type: "string", Desc: "Season"}})},
	{Name: "manage_quality", Desc: "管理质量控制、缺陷跟踪和供应商质量标准。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier": {Type: "string", Desc: "Supplier"}})},
	{Name: "scale_operations", Desc: "针对旺季、容量规划和劳动力管理扩展运营。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"scaling_type": {Type: "string", Desc: "Scaling type"}})},
	{Name: "optimize_costs", Desc: "分析并优化运输、仓储和运营成本。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"cost_type": {Type: "string", Desc: "Cost type"}})},
	{Name: "send_logistics_response", Desc: "向利益相关者发送物流更新、建议或状态报告。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_id": {Type: "string", Desc: "Operation ID"}, "message": {Type: "string", Desc: "Message"}})},
}

var transportationTools = []*schema.ToolInfo{
	{Name: "track_shipments", Desc: "跟踪货物状态、延误并协调交付物流。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"origin": {Type: "string", Desc: "Origin location"}})},
	{Name: "arrange_shipping", Desc: "安排运输方式、加急交付和多式联运。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"shipping_type": {Type: "string", Desc: "Shipping type"}})},
	{Name: "coordinate_operations", Desc: "协调复杂操作，如越库配送、集货和转运。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": {Type: "string", Desc: "Operation type"}})},
	{Name: "manage_special_handling", Desc: "处理危险品、冷链和敏感产品的特殊要求。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"product_type": {Type: "string", Desc: "Product type"}})},
	{Name: "process_returns", Desc: "处理退货、逆向物流和产品处置。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"returned_quantity": {Type: "string", Desc: "Returned quantity"}})},
	{Name: "optimize_delivery", Desc: "优化交付路线、最后一英里物流和可持续性计划。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"delivery_type": {Type: "string", Desc: "Delivery type"}})},
	{Name: "manage_disruption", Desc: "管理供应链中断、应急计划和风险缓解。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"disruption_type": {Type: "string", Desc: "Disruption type"}})},
	{Name: "send_logistics_response", Desc: "向利益相关者发送物流更新、建议或状态报告。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_id": {Type: "string", Desc: "Operation ID"}, "message": {Type: "string", Desc: "Message"}})},
}

var supplierTools = []*schema.ToolInfo{
	{Name: "evaluate_suppliers", Desc: "评估供应商绩效、进行审计并管理供应商关系。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier_name": {Type: "string", Desc: "Supplier name"}})},
	{Name: "handle_compliance", Desc: "管理监管合规、海关、文件和认证。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"compliance_type": {Type: "string", Desc: "Compliance type"}})},
	{Name: "send_logistics_response", Desc: "向利益相关者发送物流更新、建议或状态报告。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_id": {Type: "string", Desc: "Operation ID"}, "message": {Type: "string", Desc: "Message"}})},
}

// executeTool simulates tool execution
func executeTool(name string, args map[string]interface{}) string {
	argsJSON, _ := json.Marshal(args)
	fmt.Printf("[TOOL] %s(%s)\n", name, string(argsJSON))

	results := map[string]string{
		"manage_inventory":        "inventory_management_initiated",
		"track_shipments":         "shipment_tracking_updated",
		"evaluate_suppliers":      "supplier_evaluation_complete",
		"optimize_warehouse":      "warehouse_optimization_initiated",
		"forecast_demand":         "demand_forecast_generated",
		"manage_quality":          "quality_management_initiated",
		"arrange_shipping":        "shipping_arranged",
		"coordinate_operations":   "operations_coordinated",
		"manage_special_handling": "special_handling_managed",
		"handle_compliance":       "compliance_handled",
		"process_returns":         "returns_processed",
		"scale_operations":        "operations_scaled",
		"optimize_costs":          "cost_optimization_initiated",
		"optimize_delivery":       "delivery_optimization_complete",
		"manage_disruption":       "disruption_managed",
		"send_logistics_response": "logistics_response_sent",
	}
	if r, ok := results[name]; ok {
		return r
	}
	return "unknown_tool_result"
}

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	// Initialize chat model
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "gpt-4o",
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("Failed to init chat model: %v", err)
	}

	// Operation context
	operation := map[string]string{
		"operation_id": "OP-12345",
		"type":         "inventory_management",
		"priority":     "high",
		"location":     "Warehouse A",
	}
	operationJSON, _ := json.Marshal(operation)

	userMessage := "We're running critically low on SKU-12345. Current stock is 50 units but we have 200 units on backorder. What's our reorder strategy?"

	fmt.Println("=== Supply Chain Multi-Agent System (Go) ===")
	fmt.Printf("User: %s\n\n", userMessage)

	// === SUPERVISOR NODE ===
	fmt.Println("--- Supervisor ---")
	supervisorPrompt := fmt.Sprintf(`你是一名协调供应链专家团队的监督者。
团队成员：
- inventory: 处理库存水平、预测、质量、仓库优化、扩展和成本。
- transportation: 处理运输跟踪、安排、运营协调、特殊处理、退货、交付优化和中断。
- supplier: 处理供应商评估和合规性。

根据用户查询，选择一名团队成员来处理它。
仅输出所选成员的名称（inventory, transportation, 或 supplier），不要输出其他内容。

当前运营数据: %s`, string(operationJSON))

	supervisorMessages := []*schema.Message{
		schema.SystemMessage(supervisorPrompt),
		schema.UserMessage(userMessage),
	}

	supervisorResp, err := chatModel.Generate(ctx, supervisorMessages)
	if err != nil {
		log.Fatalf("Supervisor error: %v", err)
	}

	selectedAgent := strings.TrimSpace(strings.ToLower(supervisorResp.Content))
	fmt.Printf("Supervisor selected: %s\n\n", selectedAgent)

	// === SPECIALIST NODE ===
	var specialistTools []*schema.ToolInfo
	var specialistPrompt string

	switch selectedAgent {
	case "inventory":
		fmt.Println("--- Inventory Specialist ---")
		specialistTools = inventoryTools
		specialistPrompt = `你是一名库存和仓库管理专家。
在管理时：
  1) 分析库存/仓库挑战
  2) 调用适当的工具
  3) 跟进 send_logistics_response
考虑成本、效率和可扩展性。`
	case "transportation":
		fmt.Println("--- Transportation Specialist ---")
		specialistTools = transportationTools
		specialistPrompt = `你是一名运输和物流专家。
在管理时：
  1) 分析运输/交付挑战
  2) 调用适当的工具
  3) 跟进 send_logistics_response
考虑效率、可持续性和风险缓解。`
	case "supplier":
		fmt.Println("--- Supplier Specialist ---")
		specialistTools = supplierTools
		specialistPrompt = `你是一名供应商关系和合规专家。
在管理时：
  1) 分析供应商/合规性问题
  2) 调用适当的工具
  3) 跟进 send_logistics_response
考虑绩效、法规和关系。`
	default:
		log.Fatalf("Unknown agent: %s", selectedAgent)
	}

	specialistPrompt += fmt.Sprintf("\n\nOPERATION: %s", string(operationJSON))

	specialistMessages := []*schema.Message{
		schema.SystemMessage(specialistPrompt),
		schema.UserMessage(userMessage),
	}

	// First specialist call with tools
	specialistResp, err := chatModel.Generate(ctx, specialistMessages, model.WithTools(specialistTools))
	if err != nil {
		log.Fatalf("Specialist error: %v", err)
	}

	// Handle tool calls
	if len(specialistResp.ToolCalls) > 0 {
		// Add assistant message with tool calls
		specialistMessages = append(specialistMessages, &schema.Message{
			Role:      schema.Assistant,
			Content:   specialistResp.Content,
			ToolCalls: specialistResp.ToolCalls,
		})

		for _, tc := range specialistResp.ToolCalls {
			fmt.Printf("Tool Call: %s\n", tc.Function.Name)

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("Warning: failed to parse tool args: %v", err)
			}

			result := executeTool(tc.Function.Name, args)
			specialistMessages = append(specialistMessages, schema.ToolMessage(result, tc.ID))
		}

		// Final specialist response
		finalResp, err := chatModel.Generate(ctx, specialistMessages)
		if err != nil {
			log.Fatalf("Final specialist error: %v", err)
		}
		fmt.Printf("\nAssistant: %s\n", finalResp.Content)
	} else {
		fmt.Printf("Assistant: %s\n", specialistResp.Content)
	}
}
