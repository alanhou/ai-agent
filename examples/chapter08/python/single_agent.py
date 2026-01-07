from __future__ import annotations
from dotenv import load_dotenv
load_dotenv()
"""supply_chain_logistics_agent.py
一个用于供应链和物流管理的 LangGraph 工作流智能体，
处理库存管理、运输运营、供应商关系和仓库优化。
"""
import os
import json
import operator
import builtins
from typing import Annotated, Sequence, TypedDict, Optional
import sys
from pathlib import Path
# Add project root to path for package imports
sys.path.insert(0, str(Path.cwd()))

from langchain_openai.chat_models import ChatOpenAI
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage
from langchain_core.messages.tool import ToolMessage
from langchain_core.callbacks import StreamingStdOutCallbackHandler
from langchain.tools import tool
from langgraph.graph import StateGraph, END
from traceloop.sdk import Traceloop
from python.src.common.observability.loki_logger import log_to_loki

# os.environ["OTEL_EXPORTER_OTLP_ENDPOINT"] = "http://localhost:4317"
# os.environ["OTEL_EXPORTER_OTLP_INSECURE"] = "true"

@tool
def manage_inventory(sku: str = None, **kwargs) -> str:
   """管理库存水平、库存补货、审计和优化策略。"""
   print(f"[TOOL] manage_inventory(sku={sku}, kwargs={kwargs})")
   log_to_loki("tool.manage_inventory", f"sku={sku}")
   return "inventory_management_initiated"

@tool
def track_shipments(origin: str = None, **kwargs) -> str:
   """跟踪货物状态、延误并协调交付物流。"""
   print(f"[TOOL] track_shipments(origin={origin}, kwargs={kwargs})")
   log_to_loki("tool.track_shipments", f"origin={origin}")
   return "shipment_tracking_updated"

@tool
def evaluate_suppliers(supplier_name: str = None, **kwargs) -> str:
   """评估供应商绩效、进行审计并管理供应商关系。"""
   print(f"[TOOL] evaluate_suppliers(supplier_name={supplier_name},  kwargs={kwargs})")
   log_to_loki("tool.evaluate_suppliers", f"supplier_name={supplier_name}")
   return "supplier_evaluation_complete"

@tool
def optimize_warehouse(operation_type: str = None, **kwargs) -> str:
   """优化仓库运营、布局、容量和存储效率。"""
   print(f"[TOOL] optimize_warehouse(operation_type={operation_type}, kwargs={kwargs})")
   log_to_loki("tool.optimize_warehouse", f"operation_type={operation_type}")
   return "warehouse_optimization_initiated"

@tool
def forecast_demand(season: str = None, **kwargs) -> str:
   """分析需求模式、季节性趋势并创建预测模型。"""
   print(f"[TOOL] forecast_demand(season={season}, kwargs={kwargs})")
   log_to_loki("tool.forecast_demand", f"season={season}")
   return "demand_forecast_generated"

@tool
def manage_quality(supplier: str = None, **kwargs) -> str:
   """管理质量控制、缺陷跟踪和供应商质量标准。"""
   print(f"[TOOL] manage_quality(supplier={supplier}, kwargs={kwargs})")
   log_to_loki("tool.manage_quality", f"supplier={supplier}")
   return "quality_management_initiated"

@tool
def arrange_shipping(shipping_type: str = None, **kwargs) -> str:
   """安排运输方式、加急交付和多式联运。"""
   print(f"[TOOL] arrange_shipping(shipping_type={shipping_type}, kwargs={kwargs})")
   log_to_loki("tool.arrange_shipping", f"shipping_type={shipping_type}")
   return "shipping_arranged"

@tool
def coordinate_operations(operation_type: str = None, **kwargs) -> str:
   """协调复杂操作，如越库配送、集货和转运。"""
   print(f"[TOOL] coordinate_operations(operation_type={operation_type}, kwargs={kwargs})")
   log_to_loki("tool.coordinate_operations", f"operation_type={operation_type}")
   return "operations_coordinated"

@tool
def manage_special_handling(product_type: str = None, **kwargs) -> str:
   """处理危险品、冷链和敏感产品的特殊要求。"""
   print(f"[TOOL] manage_special_handling(product_type={product_type}, kwargs={kwargs})")
   log_to_loki("tool.manage_special_handling", f"product_type={product_type}")
   return "special_handling_managed"

@tool
def handle_compliance(compliance_type: str = None, **kwargs) -> str:
   """管理监管合规、海关、文件和认证。"""
   print(f"[TOOL] handle_compliance(compliance_type={compliance_type}, kwargs={kwargs})")
   log_to_loki("tool.handle_compliance", f"compliance_type={compliance_type}")
   return "compliance_handled"

@tool
def process_returns(returned_quantity: str = None, **kwargs) -> str:
   """处理退货、逆向物流和产品处置。"""
   print(f"[TOOL] process_returns(returned_quantity={returned_quantity}, kwargs={kwargs})")
   log_to_loki("tool.process_returns", f"returned_quantity={returned_quantity}")
   return "returns_processed"

@tool
def scale_operations(scaling_type: str = None, **kwargs) -> str:
   """针对旺季、容量规划和劳动力管理扩展运营。"""
   print(f"[TOOL] scale_operations(scaling_type={scaling_type}, kwargs={kwargs})")
   log_to_loki("tool.scale_operations", f"scaling_type={scaling_type}")
   return "operations_scaled"

@tool
def optimize_costs(cost_type: str = None, **kwargs) -> str:
   """分析并优化运输、仓储和运营成本。"""
   print(f"[TOOL] optimize_costs(cost_type={cost_type}, kwargs={kwargs})")
   log_to_loki("tool.optimize_costs", f"cost_type={cost_type}")
   return "cost_optimization_initiated"

@tool
def optimize_delivery(delivery_type: str = None, **kwargs) -> str:
   """优化交付路线、最后一英里物流和可持续性计划。"""
   print(f"[TOOL] optimize_delivery(delivery_type={delivery_type}, kwargs={kwargs})")
   log_to_loki("tool.optimize_delivery", f"delivery_type={delivery_type}")
   return "delivery_optimization_complete"

@tool
def manage_disruption(disruption_type: str = None, **kwargs) -> str:
   """管理供应链中断、应急计划和风险缓解。"""
   print(f"[TOOL] manage_disruption(disruption_type={disruption_type}, kwargs={kwargs})")
   log_to_loki("tool.manage_disruption", f"disruption_type={disruption_type}")
   return "disruption_managed"

@tool
def send_logistics_response(operation_id: str = None, message: str = None):
   """向利益相关者发送物流更新、建议或状态报告。"""
   print(f"[TOOL] send_logistics_response → {message}")
   log_to_loki("tool.send_logistics_response", f"operation_id={operation_id},  message={message}")
   return "logistics_response_sent"

TOOLS = [
   manage_inventory, track_shipments, evaluate_suppliers, optimize_warehouse,
   forecast_demand, manage_quality, arrange_shipping, coordinate_operations,
   manage_special_handling, handle_compliance, process_returns, scale_operations,
   optimize_costs, optimize_delivery, manage_disruption, send_logistics_response]

Traceloop.init(disable_batch=True, app_name="supply_chain_logistics_agent")
llm = ChatOpenAI(model="gpt-4o", temperature=0.0, 
                callbacks=[StreamingStdOutCallbackHandler()], 
                verbose=True).bind_tools(TOOLS)

class AgentState(TypedDict):
   operation: Optional[dict]  # 供应链运营信息
   messages: Annotated[Sequence[BaseMessage], operator.add]

def call_model(state: AgentState):
   history = state["messages"]
  
   # 优雅地处理缺失或不完整的运营数据
   operation = state.get("operation", {})
   if not operation:
       operation = {"operation_id": "UNKNOWN", "type": "general", 
           "priority": "medium", "status": "active"}
  
   operation_json = json.dumps(operation, ensure_ascii=False)
   system_prompt = (
       "你是一位经验丰富的供应链与物流专业人士。\n"
       "你的专业知识涵盖：\n"
       "- 库存管理和需求预测\n"
       "- 运输和航运优化\n"
       "- 供应商关系管理和评估\n"
       "- 仓库运营和容量规划\n"
       "- 质量控制和合规管理\n"
       "- 成本优化和运营效率\n"
       "- 风险管理和中断响应\n"
       "- 可持续发展和绿色物流倡议\n"
       "\n"
       "在管理供应链运营时：\n"
       "  1) 分析物流挑战或机遇\n"
       "  2) 调用适当的供应链管理工具\n"
       "  3) 跟进 send_logistics_response 以提供建议\n"
       "  4) 考虑成本、效率、质量和可持续性影响\n"
       "  5) 优先考虑客户满意度和业务连续性\n"
       "\n"
       "始终在成本与质量和风险缓解之间取得平衡。\n"
       f"当前运营数据: {operation_json}"
   )

   full = [SystemMessage(content=system_prompt)] + history

   first: ToolMessage | BaseMessage = llm.invoke(full)
   messages = [first]

   if getattr(first, "tool_calls", None):
       for tc in first.tool_calls:
           print(first)
           print(tc['name'])
           fn = next(t for t in TOOLS if t.name == tc['name'])
           out = fn.invoke(tc["args"])
           messages.append(ToolMessage(content=str(out), tool_call_id=tc["id"]))

       second = llm.invoke(full + messages)
       messages.append(second)

   return {"messages": messages}

def construct_graph():
   g = StateGraph(AgentState)
   g.add_node("assistant", call_model)
   g.set_entry_point("assistant")
   return g.compile()

graph = construct_graph()

if __name__ == "__main__":
   example = {"operation_id": "OP-12345", "type": "inventory_management", 
              "priority": "high", "location": "Warehouse A"}
   convo = [HumanMessage(content="We're running critically low on SKU-12345. Current stock is 50 units but we have 200 units on backorder. What's our reorder strategy?")]
   result = graph.invoke({"operation": example, "messages": convo})
   for m in result["messages"]:
       print(f"{m.type}: {m.content}")