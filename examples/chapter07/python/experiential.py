from dotenv import load_dotenv
load_dotenv()
from typing import Annotated
from typing_extensions import TypedDict
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, MessagesState, START
from langchain_core.messages import HumanMessage

# 初始化 LLM
llm = ChatOpenAI(model="gpt-4o")

# 调用 LLM 的函数
def call_model(state: MessagesState):
   response = llm.invoke(state["messages"])
   return {"messages": response}

class InsightAgent:
    def __init__(self):
       self.insights = []
       self.promoted_insights = []
       self.demoted_insights = []
       self.reflections = []

    def generate_insight(self, observation):
       # 使用 LLM 基于观察生成洞察
       messages = [HumanMessage(content=f"Generate an insightful analysis based on the following observation: '{observation}'")]

       # 构建状态图
       builder = StateGraph(MessagesState)
       builder.add_node("generate_insight", call_model)
       builder.add_edge(START, "generate_insight")
       graph = builder.compile()

       # 使用消息调用图
       result = graph.invoke({"messages": messages})
       # 提取生成的洞察
       generated_insight = result["messages"][-1].content
       self.insights.append(generated_insight)
       print(f"Generated: {generated_insight}")
       return generated_insight
    
    def promote_insight(self, insight):
        if insight in self.insights:
            self.insights.remove(insight)
            self.promoted_insights.append(insight)
            print(f"Promoted: {insight}")
        else:
            print(f"Insight '{insight}' not found in insights.")

    def demote_insight(self, insight):
        if insight in self.promoted_insights:
            self.promoted_insights.remove(insight)
            self.demoted_insights.append(insight)
            print(f"Demoted: {insight}")
        else:
            print(f"Insight '{insight}' not found in promoted insights.")

    def edit_insight(self, old_insight, new_insight):
        # 在所有列表中检查
        if old_insight in self.insights:
            index = self.insights.index(old_insight)
            self.insights[index] = new_insight
        elif old_insight in self.promoted_insights:
            index = self.promoted_insights.index(old_insight)
            self.promoted_insights[index] = new_insight
        elif old_insight in self.demoted_insights:
            index = self.demoted_insights.index(old_insight)
            self.demoted_insights[index] = new_insight
        else:
            print(f"Insight '{old_insight}' not found.")
            return
        print(f"Edited: '{old_insight}' to '{new_insight}'")
   
    def show_insights(self):
        print("\nCurrent Insights:")
        print(f"Insights: {self.insights}")
        print(f"Promoted Insights: {self.promoted_insights}")
        print(f"Demoted Insights: {self.demoted_insights}")

    def reflect(self, reflexion_prompt):
        # 构建用于反思的状态图
        builder = StateGraph(MessagesState)
        builder.add_node("reflection", call_model)
        builder.add_edge(START, "reflection")
        graph = builder.compile()
        # 使用反思提示词调用图
        result = graph.invoke(
            {
                "messages": [
                    HumanMessage(
                        content=reflexion_prompt
                    )
                ]
            }
        )
        reflection = result["messages"][-1].content
        self.reflections.append(reflection)
        print(f"Reflection: {reflection}")


agent = InsightAgent()

# 模拟的观察序列以及是否达到 KPI 目标
reports = [
    ("Website traffic rose by 15%, but bounce rate jumped from 40% to 55%.", False),
    ("Email open rates improved to 25%, exceeding our 20% goal.", True),
    ("Cart abandonment increased from 60% to 68%, missing the 50% target.", False),
    ("Average order value climbed 8%, surpassing our 5% uplift target.", True),
    ("New subscription sign-ups dipped by 5%, just below our 10% growth goal.",  False),]

# 1) 在报告期间生成并确定洞察的优先级
for text, hit_target in reports:
    insight = agent.generate_insight(text)
    if hit_target:
        agent.promote_insight(insight)
    else:
        agent.demote_insight(insight)

# 2) 通过人在回路（human-in-the-loop）编辑优化其中一个已提升的洞察
if agent.promoted_insights:
    original = agent.promoted_insights[0]
    agent.edit_insight(original, f'''Refined: {original} Investigate landing-page UX changes to reduce bounce.''')

# 3) 显示智能体的最终洞察状态
agent.show_insights()

# 4) 反思顶级洞察以规划改进
reflection_prompt = (
    "Based on our promoted insights, suggest one high-impact experiment we can run next quarter:"
    f"\n{agent.promoted_insights}")
agent.reflect(reflection_prompt)