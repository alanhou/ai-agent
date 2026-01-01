package supply_chain

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
	Operation *Operation        `json:"operation"`
	Messages  []*schema.Message `json:"messages"`
}

type Operation struct {
	OperationID string `json:"operation_id"`
	Type        string `json:"type"`
	Priority    string `json:"priority,omitempty"`
	Status      string `json:"status,omitempty"`
	Location    string `json:"location,omitempty"`
}

// -- Tool Args --

type ManageInventoryArgs struct {
	SKU string `json:"sku" desc:"SKU"`
}

type TrackShipmentsArgs struct {
	Origin string `json:"origin" desc:"Origin"`
}

type EvaluateSuppliersArgs struct {
	SupplierName string `json:"supplier_name" desc:"Supplier"`
}

type OptimizeWarehouseArgs struct {
	OperationType string `json:"operation_type" desc:"Op Type"`
}

type ForecastDemandArgs struct {
	Season string `json:"season" desc:"Season"`
}

type ManageQualityArgs struct {
	Supplier string `json:"supplier" desc:"Supplier"`
}

type ArrangeShippingArgs struct {
	ShippingType string `json:"shipping_type" desc:"Ship Type"`
}

type CoordinateOperationsArgs struct {
	OperationType string `json:"operation_type" desc:"Op Type"`
}

type ManageSpecialHandlingArgs struct {
	ProductType string `json:"product_type" desc:"Product Type"`
}

type HandleComplianceArgs struct {
	ComplianceType string `json:"compliance_type" desc:"Compliance Type"`
}

type ProcessReturnsArgs struct {
	ReturnedQuantity string `json:"returned_quantity" desc:"Qty"`
}

type ScaleOperationsArgs struct {
	ScalingType string `json:"scaling_type" desc:"Scaling Type"`
}

type OptimizeCostsArgs struct {
	CostType string `json:"cost_type" desc:"Cost Type"`
}

type OptimizeDeliveryArgs struct {
	DeliveryType string `json:"delivery_type" desc:"Delivery Type"`
}

type ManageDisruptionArgs struct {
	DisruptionType string `json:"disruption_type" desc:"Disruption Type"`
}

type SendLogisticsResponseArgs struct {
	OperationID string `json:"operation_id" desc:"Op ID"`
	Message     string `json:"message" desc:"Message"`
}

// -- Tool Impls --

func ManageInventory(ctx context.Context, args *ManageInventoryArgs) (string, error) {
	fmt.Printf("[TOOL] manage_inventory(sku=%s)\n", args.SKU)
	return "inventory_management_initiated", nil
}

func TrackShipments(ctx context.Context, args *TrackShipmentsArgs) (string, error) {
	fmt.Printf("[TOOL] track_shipments(origin=%s)\n", args.Origin)
	return "shipment_tracking_updated", nil
}

func EvaluateSuppliers(ctx context.Context, args *EvaluateSuppliersArgs) (string, error) {
	fmt.Printf("[TOOL] evaluate_suppliers(name=%s)\n", args.SupplierName)
	return "supplier_evaluation_complete", nil
}

func OptimizeWarehouse(ctx context.Context, args *OptimizeWarehouseArgs) (string, error) {
	fmt.Printf("[TOOL] optimize_warehouse(type=%s)\n", args.OperationType)
	return "warehouse_optimization_initiated", nil
}

func ForecastDemand(ctx context.Context, args *ForecastDemandArgs) (string, error) {
	fmt.Printf("[TOOL] forecast_demand(season=%s)\n", args.Season)
	return "demand_forecast_generated", nil
}

func ManageQuality(ctx context.Context, args *ManageQualityArgs) (string, error) {
	fmt.Printf("[TOOL] manage_quality(supp=%s)\n", args.Supplier)
	return "quality_management_initiated", nil
}

func ArrangeShipping(ctx context.Context, args *ArrangeShippingArgs) (string, error) {
	fmt.Printf("[TOOL] arrange_shipping(type=%s)\n", args.ShippingType)
	return "shipping_arranged", nil
}

func CoordinateOperations(ctx context.Context, args *CoordinateOperationsArgs) (string, error) {
	fmt.Printf("[TOOL] coordinate_operations(type=%s)\n", args.OperationType)
	return "operations_coordinated", nil
}

func ManageSpecialHandling(ctx context.Context, args *ManageSpecialHandlingArgs) (string, error) {
	fmt.Printf("[TOOL] manage_special_handling(prod=%s)\n", args.ProductType)
	return "special_handling_managed", nil
}

func HandleCompliance(ctx context.Context, args *HandleComplianceArgs) (string, error) {
	fmt.Printf("[TOOL] handle_compliance(type=%s)\n", args.ComplianceType)
	return "compliance_handled", nil
}

func ProcessReturns(ctx context.Context, args *ProcessReturnsArgs) (string, error) {
	fmt.Printf("[TOOL] process_returns(qty=%s)\n", args.ReturnedQuantity)
	return "returns_processed", nil
}

func ScaleOperations(ctx context.Context, args *ScaleOperationsArgs) (string, error) {
	fmt.Printf("[TOOL] scale_operations(type=%s)\n", args.ScalingType)
	return "operations_scaled", nil
}

func OptimizeCosts(ctx context.Context, args *OptimizeCostsArgs) (string, error) {
	fmt.Printf("[TOOL] optimize_costs(type=%s)\n", args.CostType)
	return "cost_optimization_initiated", nil
}

func OptimizeDelivery(ctx context.Context, args *OptimizeDeliveryArgs) (string, error) {
	fmt.Printf("[TOOL] optimize_delivery(type=%s)\n", args.DeliveryType)
	return "delivery_optimization_complete", nil
}

func ManageDisruption(ctx context.Context, args *ManageDisruptionArgs) (string, error) {
	fmt.Printf("[TOOL] manage_disruption(type=%s)\n", args.DisruptionType)
	return "disruption_managed", nil
}

func SendLogisticsResponse(ctx context.Context, args *SendLogisticsResponseArgs) (string, error) {
	fmt.Printf("[TOOL] send_logistics_response -> %s\n", args.Message)
	return "logistics_response_sent", nil
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
		{Name: "manage_inventory", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"sku": strParamOpt("SKU")})},
		{Name: "track_shipments", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"origin": strParamOpt("Origin")})},
		{Name: "evaluate_suppliers", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier_name": strParamOpt("Supplier")})},
		{Name: "optimize_warehouse", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": strParamOpt("Op Type")})},
		{Name: "forecast_demand", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"season": strParamOpt("Season")})},
		{Name: "manage_quality", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"supplier": strParamOpt("Supplier")})},
		{Name: "arrange_shipping", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"shipping_type": strParamOpt("Ship Type")})},
		{Name: "coordinate_operations", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_type": strParamOpt("Op Type")})},
		{Name: "manage_special_handling", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"product_type": strParamOpt("Prod Type")})},
		{Name: "handle_compliance", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"compliance_type": strParamOpt("Comp Type")})},
		{Name: "process_returns", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"returned_quantity": strParamOpt("Qty")})},
		{Name: "scale_operations", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"scaling_type": strParamOpt("Scale Type")})},
		{Name: "optimize_costs", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"cost_type": strParamOpt("Cost Type")})},
		{Name: "optimize_delivery", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"delivery_type": strParamOpt("Delivery Type")})},
		{Name: "manage_disruption", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"disruption_type": strParamOpt("Disruption Type")})},
		{Name: "send_logistics_response", ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{"operation_id": strParamOpt("Op ID"), "message": strParam("Msg")})},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"manage_inventory": func(ctx context.Context, args interface{}) (string, error) {
			return ManageInventory(ctx, args.(*ManageInventoryArgs))
		},
		"track_shipments": func(ctx context.Context, args interface{}) (string, error) {
			return TrackShipments(ctx, args.(*TrackShipmentsArgs))
		},
		"evaluate_suppliers": func(ctx context.Context, args interface{}) (string, error) {
			return EvaluateSuppliers(ctx, args.(*EvaluateSuppliersArgs))
		},
		"optimize_warehouse": func(ctx context.Context, args interface{}) (string, error) {
			return OptimizeWarehouse(ctx, args.(*OptimizeWarehouseArgs))
		},
		"forecast_demand": func(ctx context.Context, args interface{}) (string, error) {
			return ForecastDemand(ctx, args.(*ForecastDemandArgs))
		},
		"manage_quality": func(ctx context.Context, args interface{}) (string, error) {
			return ManageQuality(ctx, args.(*ManageQualityArgs))
		},
		"arrange_shipping": func(ctx context.Context, args interface{}) (string, error) {
			return ArrangeShipping(ctx, args.(*ArrangeShippingArgs))
		},
		"coordinate_operations": func(ctx context.Context, args interface{}) (string, error) {
			return CoordinateOperations(ctx, args.(*CoordinateOperationsArgs))
		},
		"manage_special_handling": func(ctx context.Context, args interface{}) (string, error) {
			return ManageSpecialHandling(ctx, args.(*ManageSpecialHandlingArgs))
		},
		"handle_compliance": func(ctx context.Context, args interface{}) (string, error) {
			return HandleCompliance(ctx, args.(*HandleComplianceArgs))
		},
		"process_returns": func(ctx context.Context, args interface{}) (string, error) {
			return ProcessReturns(ctx, args.(*ProcessReturnsArgs))
		},
		"scale_operations": func(ctx context.Context, args interface{}) (string, error) {
			return ScaleOperations(ctx, args.(*ScaleOperationsArgs))
		},
		"optimize_costs": func(ctx context.Context, args interface{}) (string, error) {
			return OptimizeCosts(ctx, args.(*OptimizeCostsArgs))
		},
		"optimize_delivery": func(ctx context.Context, args interface{}) (string, error) {
			return OptimizeDelivery(ctx, args.(*OptimizeDeliveryArgs))
		},
		"manage_disruption": func(ctx context.Context, args interface{}) (string, error) {
			return ManageDisruption(ctx, args.(*ManageDisruptionArgs))
		},
		"send_logistics_response": func(ctx context.Context, args interface{}) (string, error) {
			return SendLogisticsResponse(ctx, args.(*SendLogisticsResponseArgs))
		},
	}

	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		opJSON, _ := json.Marshal(state.Operation)
		sysPrompt := fmt.Sprintf(
			"You are a Logistics Expert.\n"+
				"Roles: Inventory, Shipping, Warehouse, Suppliers, Forecast, Quality, Costs, Delivery, Risk.\n"+
				"1) Use tools.\n2) send_logistics_response.\n\n"+
				"OPERATION: %s", string(opJSON))

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
			case "manage_inventory":
				var a ManageInventoryArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "track_shipments":
				var a TrackShipmentsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "evaluate_suppliers":
				var a EvaluateSuppliersArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "optimize_warehouse":
				var a OptimizeWarehouseArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "forecast_demand":
				var a ForecastDemandArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "manage_quality":
				var a ManageQualityArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "arrange_shipping":
				var a ArrangeShippingArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "coordinate_operations":
				var a CoordinateOperationsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "manage_special_handling":
				var a ManageSpecialHandlingArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "handle_compliance":
				var a HandleComplianceArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "process_returns":
				var a ProcessReturnsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "scale_operations":
				var a ScaleOperationsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "optimize_costs":
				var a OptimizeCostsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "optimize_delivery":
				var a OptimizeDeliveryArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "manage_disruption":
				var a ManageDisruptionArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_logistics_response":
				var a SendLogisticsResponseArgs
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
