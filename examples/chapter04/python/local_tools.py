from dotenv import load_dotenv
load_dotenv()

from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langchain_core.tools import tool


# 使用简洁的函数定义来定义工具
@tool
def multiply(x: float, y: float) -> float:
   """Multiply 'x' times 'y'. (计算 'x' 乘以 'y')"""
   return x * y

@tool
def exponentiate(x: float, y: float) -> float:
   """Raise 'x' to the 'y'. (计算 'x' 的 'y' 次幂)"""
   return x**y

@tool
def add(x: float, y: float) -> float:
   """Add 'x' and 'y'. (计算 'x' 加 'y')"""
   return x + y

tools = [multiply, exponentiate, add]
# 使用 GPT-4o 初始化 LLM 并绑定工具
llm = ChatOpenAI(model_name="gpt-4o", temperature=0)
llm_with_tools = llm.bind_tools(tools)

query = "What is 393 * 12.25? Also, what is 11 + 49?"
messages = [HumanMessage(query)]
ai_msg = llm_with_tools.invoke(messages)
messages.append(ai_msg)

for tool_call in ai_msg.tool_calls:
   selected_tool = {"add": add, "multiply": multiply, 
       "exponentiate": exponentiate}[tool_call["name"].lower()]
   tool_msg = selected_tool.invoke(tool_call)
  
   print(f"{tool_msg.name} {tool_call['args']} {tool_msg.content}")
   messages.append(tool_msg)

final_response = llm_with_tools.invoke(messages)
print(final_response.content)