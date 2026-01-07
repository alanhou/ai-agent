from __future__ import annotations
from dotenv import load_dotenv
load_dotenv()

import os
import json
import operator
import sys
from pathlib import Path
from typing import Annotated, Sequence, TypedDict, Optional, List

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

# ========== 工具定义 ==========
@tool
def send_logistics_response(operation_id=None, message=None):
   """向利益相关者发送物流更新、建议或状态报告。"""
   print(f"[TOOL] send_logistics_response → {message}")
   log_to_loki("tool.send_logistics_response", 
               f"operation_id={operation_id}, message={message}")
   return "logistics_response_sent"

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
def evaluate_suppliers(supplier_name: str = None, **kwargs) -> str:
   """评估供应商绩效、进行审计并管理供应商关系。"""
   print(f"[TOOL] evaluate_suppliers(supplier_name={supplier_name}, kwargs={kwargs})")
   log_to_loki("tool.evaluate_suppliers", f"supplier_name={supplier_name}")
   return "supplier_evaluation_complete"

@tool
def optimize_costs(cost_type: str = None, **kwargs) -> str:
   """分析并优化运输、仓储和运营成本。"""
   print(f"[TOOL] optimize_costs(cost_type={cost_type}, kwargs={kwargs})")
   log_to_loki("tool.optimize_costs", f"cost_type={cost_type}")
   return "cost_optimization_initiated"

# 所有可用工具
all_tools = [
    manage_inventory, optimize_warehouse, forecast_demand,
    track_shipments, arrange_shipping, evaluate_suppliers, 
    optimize_costs, send_logistics_response
]

# ========== 初始化 ==========
Traceloop.init(disable_batch=True, app_name="actor_critic_agent")
llm = ChatOpenAI(model="gpt-4o", temperature=0.5, 
    callbacks=[StreamingStdOutCallbackHandler()], verbose=True)

# ========== AgentState 定义 ==========
class AgentState(TypedDict):
   operation: Optional[dict]
   messages: Annotated[Sequence[BaseMessage], operator.add]
   candidates: Optional[List[dict]]  # Actor 生成的候选计划

# ========== Actor 节点：生成候选计划 ==========
def actor_node(state: AgentState):
   history = state["messages"]
   operation = state.get("operation", {})
   operation_json = json.dumps(operation, ensure_ascii=False)
   
   actor_prompt = f'''你是供应链策略专家。基于用户需求，生成 3 个候选供应链计划。
请以 JSON 格式返回，格式如下:
[
  {{"plan": "计划描述", "tools": [{{"tool": "tool_name", "args": {{"key": "value"}}}}]}},
  ...
]

可用工具: manage_inventory, optimize_warehouse, forecast_demand, track_shipments, 
arrange_shipping, evaluate_suppliers, optimize_costs, send_logistics_response

当前运营数据: {operation_json}'''
   
   response = llm.invoke([SystemMessage(content=actor_prompt)] + list(history))
   
   # 解析 JSON 响应
   try:
       content = response.content
       # 提取 JSON 部分（处理可能的 markdown 代码块）
       if "```json" in content:
           content = content.split("```json")[1].split("```")[0]
       elif "```" in content:
           content = content.split("```")[1].split("```")[0]
       candidates = json.loads(content.strip())
   except json.JSONDecodeError:
       candidates = [{"plan": response.content, "tools": []}]
   
   return {"messages": [response], "candidates": candidates}

# ========== Critic 节点：评估并选择/迭代 ==========
def critic_node(state: AgentState):
   candidates = state.get("candidates", [])
   history = state["messages"]
   
   critic_prompt = f'''你是供应链评估专家。评估以下候选计划:

{json.dumps(candidates, ensure_ascii=False, indent=2)}

对每个计划按以下维度打分（1-10分）:
- 可行性 (feasibility)
- 成本效益 (cost_effectiveness)  
- 风险控制 (risk_management)

返回 JSON 格式:
{{
  "evaluations": [
    {{"plan_index": 0, "feasibility": 8, "cost": 7, "risk": 9, "total": 24}},
    ...
  ],
  "best_index": 0,
  "best_score": 8.0,
  "selected": {{"plan": "...", "tools": [...]}},
  "feedback": "改进建议（如果最高分 <= 8）"
}}

如果最高平均分 > 8，选择最佳计划；否则提供改进反馈。'''

   response = llm.invoke([SystemMessage(content=critic_prompt)] + list(history))
   
   try:
       content = response.content
       if "```json" in content:
           content = content.split("```json")[1].split("```")[0]
       elif "```" in content:
           content = content.split("```")[1].split("```")[0]
       eval_result = json.loads(content.strip())
   except json.JSONDecodeError:
       # 解析失败时返回需要重新生成的状态
       return {"messages": [AIMessage(content="regenerate: 无法解析评估结果，请重新生成计划。")]}
   
   best_score = eval_result.get('best_score', 0)
   
   if best_score > 8:
       # 执行获胜计划的工具
       winning_plan = eval_result.get('selected', {})
       messages = [response]
       
       for tool_info in winning_plan.get('tools', []):
           tool_name = tool_info.get('tool')
           tool_args = tool_info.get('args', {})
           tc_id = f"critic_{tool_name}"
           
           try:
               fn = next(t for t in all_tools if t.name == tool_name)
               out = fn.invoke(tool_args)
               messages.append(ToolMessage(content=str(out), tool_call_id=tc_id))
           except StopIteration:
               print(f"Warning: Tool {tool_name} not found")
       
       # 发送最终响应
       send_logistics_response.invoke({
           "operation_id": state.get("operation", {}).get("operation_id"),
           "message": winning_plan.get('plan', '计划已执行')
       })
       
       return {"messages": messages}
   else:
       # 迭代：将反馈添加到历史记录以供 Actor 使用
       feedback = eval_result.get('feedback', '请改进计划的可行性和成本效益')
       return {"messages": [AIMessage(content=f"regenerate: 根据反馈进行改进并重新生成: {feedback}")]}

# ========== 路由函数 ==========
def should_continue(state: AgentState):
   """检查是否需要继续迭代"""
   if not state["messages"]:
       return END
   last_message = state["messages"][-1]
   if hasattr(last_message, 'content') and "regenerate" in last_message.content.lower():
       return "actor"
   return END

# ========== 构建 Actor-Critic 图 ==========
def construct_actor_critic_graph():
   g = StateGraph(AgentState)
   g.add_node("actor", actor_node)
   g.add_node("critic", critic_node)
   
   g.set_entry_point("actor")
   g.add_edge("actor", "critic")
   # 如果未获批准则循环回退（条件边）
   g.add_conditional_edges("critic", should_continue, 
       {"actor": "actor", END: END})
   
   return g.compile()

graph = construct_actor_critic_graph()

# ========== 主函数 ==========
if __name__ == "__main__":
   example = {"operation_id": "OP-12345", "type": "inventory_management", 
              "priority": "high", "location": "Warehouse A"}
   convo = [HumanMessage(content='''We're running critically low on SKU-12345. 
Current stock is 50 units but we have 200 units on backorder. 
What's our reorder strategy?''')]
   
   print("=== Actor-Critic Supply Chain Agent ===\n")
   result = graph.invoke({"operation": example, "messages": convo, "candidates": None})
   
   print("\n=== Final Messages ===")
   for m in result["messages"]:
       print(f"{m.type}: {m.content[:200] if len(m.content) > 200 else m.content}...")