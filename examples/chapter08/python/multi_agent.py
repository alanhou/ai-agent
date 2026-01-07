from __future__ import annotations
from dotenv import load_dotenv
load_dotenv()

import os
import json
import operator
import sys
from pathlib import Path
from typing import Annotated, Sequence, TypedDict, Optional

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

os.environ["OTEL_EXPORTER_OTLP_ENDPOINT"] = "http://localhost:4317"
os.environ["OTEL_EXPORTER_OTLP_INSECURE"] = "true"

# 所有专家共享的工具
@tool
def send_logistics_response(operation_id = None, message = None):
   """向利益相关者发送物流更新、建议或状态报告。"""
   print(f"[TOOL] send_logistics_response → {message}")
   log_to_loki("tool.send_logistics_response", 
               f"operation_id={operation_id}, message={message}")
   return "logistics_response_sent"

# 库存与仓库专家工具
@tool
def manage_inventory(sku: str = None, **kwargs) -> str:
   """管理库存水平、库存补货、审计和优化策略。"""
   print(f"[TOOL] manage_inventory(sku={sku}, kwargs={kwargs})")
   log_to_loki("tool.manage_inventory", f"sku={sku}")
   return "inventory_management_initiated"

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

INVENTORY_TOOLS = [manage_inventory, optimize_warehouse, forecast_demand,
manage_quality, scale_operations, optimize_costs, send_logistics_response]

# 运输与物流专家工具
@tool
def track_shipments(origin: str = None, **kwargs) -> str:
   """跟踪货物状态、延误并协调交付物流。"""
   print(f"[TOOL] track_shipments(origin={origin}, kwargs={kwargs})")
   log_to_loki("tool.track_shipments", f"origin={origin}")
   return "shipment_tracking_updated"

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
def process_returns(returned_quantity: str = None, **kwargs) -> str:
   """处理退货、逆向物流和产品处置。"""
   print(f"[TOOL] process_returns(returned_quantity={returned_quantity}, kwargs={kwargs})")
   log_to_loki("tool.process_returns", f"returned_quantity={returned_quantity}")
   return "returns_processed"

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

TRANSPORTATION_TOOLS = [track_shipments, arrange_shipping, coordinate_operations, 
    manage_special_handling, process_returns, optimize_delivery, 
    manage_disruption, send_logistics_response]

# 供应商与合规专家工具
@tool
def evaluate_suppliers(supplier_name: str = None, **kwargs) -> str:
   """评估供应商绩效、进行审计并管理供应商关系。"""
   print(f"[TOOL] evaluate_suppliers(supplier_name={supplier_name}, kwargs={kwargs})")
   log_to_loki("tool.evaluate_suppliers", f"supplier_name={supplier_name}")
   return "supplier_evaluation_complete"

@tool
def handle_compliance(compliance_type: str = None, **kwargs) -> str:
   """管理监管合规、海关、文件和认证。"""
   print(f"[TOOL] handle_compliance(compliance_type={compliance_type}, kwargs={kwargs})")
   log_to_loki("tool.handle_compliance", f"compliance_type={compliance_type}")
   return "compliance_handled"

SUPPLIER_TOOLS = [evaluate_suppliers, handle_compliance, send_logistics_response]

Traceloop.init(disable_batch=True, app_name="supply_chain_logistics_agent")
llm = ChatOpenAI(model="gpt-4o", temperature=0.0, 
    callbacks=[StreamingStdOutCallbackHandler()], verbose=True)

# 将工具绑定到专门的 LLM
inventory_llm = llm.bind_tools(INVENTORY_TOOLS)
transportation_llm = llm.bind_tools(TRANSPORTATION_TOOLS)
supplier_llm = llm.bind_tools(SUPPLIER_TOOLS)

class AgentState(TypedDict):
   operation: Optional[dict]  # 供应链运营信息
   messages: Annotated[Sequence[BaseMessage], operator.add]

# 监督者（管理者）节点：路由到适当的专家
def supervisor_node(state: AgentState):
   history = state["messages"]
   operation = state.get("operation", {})
   operation_json = json.dumps(operation, ensure_ascii=False)
  
   supervisor_prompt = (
       "你是一名协调供应链专家团队的监督者。\n"
       "团队成员：\n"
       "- inventory: 处理库存水平、预测、\n"
       "质量、仓库优化、扩展和成本。\n"
       "- transportation: 处理运输跟踪、\n"
       "安排、运营协调、\n"
       "特殊处理、退货、交付优化和中断。\n"
       "- supplier: 处理供应商评估和合规性。\n"
       "\n"
       "根据用户查询，选择一名团队成员来处理它。\n"
       "仅输出所选成员的名称\n"
       "（inventory, transportation, 或 supplier），不要输出其他内容。\n\n"
       f"当前运营数据: {operation_json}"
   )

   full = [SystemMessage(content=supervisor_prompt)] + history
   response = llm.invoke(full)
   return {"messages": [response]}

# 专家节点模板
def specialist_node(state: AgentState, specialist_llm, system_prompt: str):
   history = state["messages"]
   operation = state.get("operation", {})
   if not operation:
       operation = {"operation_id": "UNKNOWN", "type": "general",   
           "priority": "medium", "status": "active"}
   operation_json = json.dumps(operation, ensure_ascii=False)
   full_prompt = system_prompt + f"\n\nOPERATION: {operation_json}"
  
   full = [SystemMessage(content=full_prompt)] + history

   first: ToolMessage | BaseMessage = specialist_llm.invoke(full)
   messages = [first]

   if getattr(first, "tool_calls", None):
       for tc in first.tool_calls:
           print(first)
           print(tc['name'])
           # 查找工具（假设工具名称在所有工具中是唯一的）
           all_tools = INVENTORY_TOOLS + TRANSPORTATION_TOOLS + SUPPLIER_TOOLS
           fn = next(t for t in all_tools if t.name == tc['name'])
           out = fn.invoke(tc["args"])
           messages.append(ToolMessage(content=str(out), tool_call_id=tc["id"]))

       second = specialist_llm.invoke(full + messages)
       messages.append(second)

   return {"messages": messages}

# 库存专家节点
def inventory_node(state: AgentState):
   inventory_prompt = (
       "你是一名库存和仓库管理专家。\n"
       "在管理时：\n"
       "  1) 分析库存/仓库挑战\n"
       "  2) 调用适当的工具\n"
       "  3) 跟进 send_logistics_response\n"
       "考虑成本、效率和可扩展性。"
   )
   return specialist_node(state, inventory_llm, inventory_prompt)

# 运输专家节点
def transportation_node(state: AgentState):
   transportation_prompt = (
       "你是一名运输和物流专家。\n"
       "在管理时：\n"
       "  1) 分析运输/交付挑战\n"
       "  2) 调用适当的工具\n"
       "  3) 跟进 send_logistics_response\n"
       "考虑效率、可持续性和风险缓解。"
   )
   return specialist_node(state, transportation_llm, transportation_prompt)

# 供应商专家节点
def supplier_node(state: AgentState):
   supplier_prompt = (
       "你是一名供应商关系和合规专家。\n"
       "在管理时：\n"
       "  1) 分析供应商/合规性问题\n"
       "  2) 调用适当的工具\n"
       "  3) 跟进 send_logistics_response\n"
       "考虑绩效、法规和关系。"
   )
   return specialist_node(state, supplier_llm, supplier_prompt)

# 用于条件边的路由函数
def route_to_specialist(state: AgentState):
   last_message = state["messages"][-1]
   agent_name = last_message.content.strip().lower()
   if agent_name == "inventory":
       return "inventory"
   elif agent_name == "transportation":
       return "transportation"
   elif agent_name == "supplier":
       return "supplier"
   else:
       # 如果没有匹配则回退
       return END

def construct_graph():
   g = StateGraph(AgentState)
   g.add_node("supervisor", supervisor_node)
   g.add_node("inventory", inventory_node)
   g.add_node("transportation", transportation_node)
   g.add_node("supplier", supplier_node)
  
   g.set_entry_point("supervisor")
   g.add_conditional_edges("supervisor", route_to_specialist,
    {"inventory": "inventory", "transportation": 
    "transportation", "supplier": "supplier"})
  
   g.add_edge("inventory", END)
   g.add_edge("transportation", END)
   g.add_edge("supplier", END)
  
   return g.compile()

graph = construct_graph()

if __name__ == "__main__":
   example = {"operation_id": "OP-12345", "type": "inventory_management", 
              "priority": "high", "location": "Warehouse A"}
   convo = [HumanMessage(content='''We're running critically low on SKU-12345. Current stock is 50 units but we have 200 units on backorder. What's our reorder strategy?''')]
   result = graph.invoke({"operation": example, "messages": convo})
   for m in result["messages"]:
       print(f"{m.type}: {m.content}")