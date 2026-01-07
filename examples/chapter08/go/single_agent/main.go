package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// Tool definitions for OpenAI function calling
var tools = []*schema.ToolInfo{
	{Name: "manage_inventory", Desc: "管理库存水平、库存补货、审计和优化策略。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"sku": {Type: "string", Desc: "SKU identifier"}})},
	{Name: "track_shipments", Desc: "跟踪货物状态、延误并协调交付物流。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"origin": {Type: "string", Desc: "Origin location"}})},
	{Name: "evaluate_suppliers", Desc: "评估供应商绩效、进行审计并管理供应商关系。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier_name": {Type: "string", Desc: "Supplier name"}})},
	{Name: "optimize_warehouse", Desc: "优化仓库运营、布局、容量和存储效率。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": {Type: "string", Desc: "Type of operation"}})},
	{Name: "forecast_demand", Desc: "分析需求模式、季节性趋势并创建预测模型。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"season": {Type: "string", Desc: "Season"}})},
	{Name: "manage_quality", Desc: "管理质量控制、缺陷跟踪和供应商质量标准。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier": {Type: "string", Desc: "Supplier"}})},
	{Name: "arrange_shipping", Desc: "安排运输方式、加急交付和多式联运。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"shipping_type": {Type: "string", Desc: "Shipping type"}})},
	{Name: "coordinate_operations", Desc: "协调复杂操作，如越库配送、集货和转运。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": {Type: "string", Desc: "Operation type"}})},
	{Name: "manage_special_handling", Desc: "处理危险品、冷链和敏感产品的特殊要求。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"product_type": {Type: "string", Desc: "Product type"}})},
	{Name: "handle_compliance", Desc: "管理监管合规、海关、文件和认证。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"compliance_type": {Type: "string", Desc: "Compliance type"}})},
	{Name: "process_returns", Desc: "处理退货、逆向物流和产品处置。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"returned_quantity": {Type: "string", Desc: "Returned quantity"}})},
	{Name: "scale_operations", Desc: "针对旺季、容量规划和劳动力管理扩展运营。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"scaling_type": {Type: "string", Desc: "Scaling type"}})},
	{Name: "optimize_costs", Desc: "分析并优化运输、仓储和运营成本。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"cost_type": {Type: "string", Desc: "Cost type"}})},
	{Name: "optimize_delivery", Desc: "优化交付路线、最后一英里物流和可持续性计划。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"delivery_type": {Type: "string", Desc: "Delivery type"}})},
	{Name: "manage_disruption", Desc: "管理供应链中断、应急计划和风险缓解。", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"disruption_type": {Type: "string", Desc: "Disruption type"}})},
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

	// Initialize chat model with tools
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

	systemPrompt := fmt.Sprintf(`你是一位经验丰富的供应链与物流专业人士。
你的专业知识涵盖：
- 库存管理和需求预测
- 运输和航运优化
- 供应商关系管理和评估
- 仓库运营和容量规划
- 质量控制和合规管理
- 成本优化和运营效率
- 风险管理和中断响应
- 可持续发展和绿色物流倡议

在管理供应链运营时：
  1) 分析物流挑战或机遇
  2) 调用适当的供应链管理工具
  3) 跟进 send_logistics_response 以提供建议
  4) 考虑成本、效率、质量和可持续性影响
  5) 优先考虑客户满意度和业务连续性

始终在成本与质量和风险缓解之间取得平衡。
当前运营数据: %s`, string(operationJSON))

	userMessage := "We're running critically low on SKU-12345. Current stock is 50 units but we have 200 units on backorder. What's our reorder strategy?"

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}

	fmt.Println("=== Supply Chain Logistics Agent (Go) ===")
	fmt.Printf("User: %s\n\n", userMessage)

	// First LLM call with tools
	resp, err := chatModel.Generate(ctx, messages, model.WithTools(tools))
	if err != nil {
		log.Fatalf("Error generating response: %v", err)
	}

	// Handle tool calls
	if len(resp.ToolCalls) > 0 {
		// Add assistant message with tool calls
		messages = append(messages, &schema.Message{
			Role:      schema.Assistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			fmt.Printf("Tool Call: %s\n", tc.Function.Name)

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("Warning: failed to parse tool args: %v", err)
			}

			result := executeTool(tc.Function.Name, args)
			messages = append(messages, schema.ToolMessage(result, tc.ID))
		}

		// Second LLM call for final response
		finalResp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Fatalf("Error generating final response: %v", err)
		}
		fmt.Printf("\nAssistant: %s\n", finalResp.Content)
	} else {
		fmt.Printf("Assistant: %s\n", resp.Content)
	}
}
