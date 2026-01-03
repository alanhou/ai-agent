from typing import TypedDict, Sequence, Any
import sys
from langchain_core.tools import Tool
from langchain_mcp_adapters.client import MultiServerMCPClient


class AgentState(TypedDict):
    messages: Sequence[Any]  # BaseMessage/HumanMessage/... 的列表


# MCP Client configuration with standard MCP servers
mcp_client = MultiServerMCPClient(
    {
        "calculate": {
            "command": sys.executable,  # Use same Python interpreter
            "args": ["python/src/common/mcp/MCP_math_server.py"],
            "transport": "stdio",  # 子进程 -> STDIO JSON-RPC
        },
        "get_weather": {
            # Weather server runs on HTTP transport
            "url": "http://localhost:8000/mcp",
            "transport": "streamable_http",
        },
    })


async def get_mcp_tools() -> list[Tool]:
    return await mcp_client.get_tools()


async def call_mcp_tools(state: AgentState) -> dict[str, Any]:
    messages = state["messages"]
    last_msg = messages[-1].content.lower()

    # 在第一次调用时获取并缓存 MCP 工具
    global MCP_TOOLS
    if "MCP_TOOLS" not in globals():
        MCP_TOOLS = await mcp_client.get_tools()

    # 简单的启发式方法：如果出现任何数字运算符标记，选择 "calculate"
    if any(token in last_msg for token in ["+", "-", "*", "/", "(", ")"]):
        tool_name = "calculate"
    elif "weather" in last_msg:
        tool_name = "get_weather"
    else:
        # 无匹配 -> 直接响应
        return {
            "messages": [
                {
                    "role": "assistant",
                    "content": "Sorry, I can only answer math or weather queries."
                }
            ]
        }

    tool_obj = next((t for t in MCP_TOOLS if t.name == tool_name), None)
    if tool_obj is None:
        return {
            "messages": [
                {"role": "assistant", "content": f"Tool '{tool_name}' not available."}
            ]
        }

    user_input = messages[-1].content
    mcp_result = await tool_obj.ainvoke({"expression": user_input} if tool_name == "calculate" else {"city": user_input})

    return {
        "messages": [
            {"role": "assistant", "content": str(mcp_result)}
        ]
    }


if __name__ == "__main__":
    import asyncio
    from dataclasses import dataclass

    @dataclass
    class MockMessage:
        content: str

    async def test_mcp_tools():
        print("=== Testing MCP Tools ===\n")

        # Show available tools
        tools = await get_mcp_tools()
        print(f"Available tools: {[t.name for t in tools]}\n")

        # Test 1: Math query
        print("Test 1: Math query (3 + 5)")
        state = {"messages": [MockMessage(content="3 + 5")]}
        result = await call_mcp_tools(state)
        print(f"Result: {result}\n")

        # Test 2: Complex math
        print("Test 2: Complex math ((10 - 2) * 4)")
        state = {"messages": [MockMessage(content="(10 - 2) * 4")]}
        result = await call_mcp_tools(state)
        print(f"Result: {result}\n")

        # Test 3: Weather query
        print("Test 3: Weather query (Beijing)")
        state = {"messages": [MockMessage(content="weather Beijing")]}
        result = await call_mcp_tools(state)
        print(f"Result: {result}\n")

        # Test 4: Unrecognized query
        print("Test 4: Unrecognized query")
        state = {"messages": [MockMessage(content="Tell me a joke")]}
        result = await call_mcp_tools(state)
        print(f"Result: {result}\n")

        print("=== Tests Complete ===")

    asyncio.run(test_mcp_tools())