from dotenv import load_dotenv
load_dotenv()

from typing import Annotated, TypedDict, List
from langchain.tools import tool
from langchain_openai import ChatOpenAI
from langchain_core.messages import BaseMessage, SystemMessage, HumanMessage
from langgraph.graph import StateGraph, START, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode, tools_condition

# -- 1) 定义我们唯一的业务工具
@tool
def cancel_order(order_id: str) -> str:
    """取消尚未发货的订单。"""
    # (在这里调用真实的后台 API)
    return f"订单 {order_id} 已被取消。"

tools = [cancel_order]

# -- 2) 定义 State
class AgentState(TypedDict):
    # add_messages 会自动处理消息的追加，而不是覆盖
    messages: Annotated[List[BaseMessage], add_messages]
    order: dict

# -- 3) 定义节点
def call_model(state: AgentState):
    order = state.get("order", {"order_id": "UNKNOWN"})
    
    # 系统提示词告诉模型具体要做什么
    system_prompt = (
        f'''你是一名电商客服代理。
        订单 ID: {order['order_id']}
        如果客户要求取消订单，请调用 cancel_order(order_id)
        然后发送一个简单的确认。
        否则，只需正常回复。'''
    )
    
    # 将系统消息放在消息列表的最前面
    # 注意：我们不在 state 中持久化 SystemMessage，而是每次调用模型时动态添加
    # 这样可以保持 state['messages'] 干净，且允许根据 order 动态改变 prompt
    messages = [SystemMessage(content=system_prompt)] + state["messages"]
    
    # 初始化模型并绑定工具
    # 使用 gpt-4o 或其他支持工具调用的现代模型
    llm = ChatOpenAI(model="gpt-5", temperature=0)
    llm_with_tools = llm.bind_tools(tools)
    
    response = llm_with_tools.invoke(messages)
    
    # 返回的内容会被 add_messages reducer 追加到 state 的 messages 中
    return {"messages": [response]}

# -- 4) 构建图
def construct_graph():
    workflow = StateGraph(AgentState)
    
    # 添加节点
    workflow.add_node("assistant", call_model)
    workflow.add_node("tools", ToolNode(tools))
    
    # 设置入口点
    workflow.add_edge(START, "assistant")
    
    # 添加条件边：
    # 如果 assistant 返回了 tool_calls，tools_condition 会路由到 "tools"
    # 否则路由到 END
    workflow.add_conditional_edges(
        "assistant",
        tools_condition,
    )
    
    # 工具执行完后，返回给 assistant 继续生成回复 (ReAct 循环)
    workflow.add_edge("tools", "assistant")
    
    return workflow.compile()

graph = construct_graph()

if __name__ == "__main__":
    example_order = {"order_id": "A12345"}
    # 注意：这里我们只传入初始的 HumanMessage
    convo = [HumanMessage(content="Please cancel my order A12345.")]
    
    # 运行图
    result = graph.invoke({"order": example_order, "messages": convo})
    
    # 打印结果
    for msg in result["messages"]:
        print(f"{msg.type}: {msg.content}")


    # 最小化评估检查
    example_order = {"order_id": "B73973"}
    convo = [HumanMessage(content='''Please cancel order #B73973. 
        I found a cheaper option elsewhere.''')]

    result = graph.invoke({"order": example_order, "messages": convo})

    # check if tool calls were made
    has_tool_call = False
    for m in result["messages"]:
        if getattr(m, "tool_calls", None):
            for tc in m.tool_calls:
                if tc["name"] == "cancel_order":
                    has_tool_call = True
                    break
    
    assert has_tool_call, "未调用取消订单工具 (Cancel order tool not called)"
        
    assert any("cancel" in m.content.lower() or "取消" in m.content for m in result["messages"]), \
        "缺失确认消息 (Confirmation message missing)"

    print("✅ Agent passed minimal evaluation.")
