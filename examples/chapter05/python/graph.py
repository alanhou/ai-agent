from dotenv import load_dotenv
load_dotenv()
from langgraph.graph import StateGraph, START, END
from langchain_openai import ChatOpenAI
from typing_extensions import TypedDict
from typing import Optional

# Initialize LLM
llm = ChatOpenAI(model_name="gpt-4o-mini", temperature=0)

# Define state schema - simplified to only update what changes
class State(TypedDict, total=False):
    user_message: str
    user_id: str
    issue_type: Optional[str]
    step_result: Optional[str]
    response: Optional[str]
# 1. Node definitions
def categorize_issue(state: dict) -> dict:
    prompt = (
        f"Classify this support request as 'billing' or 'technical'.\n\n"
        f"Message: {state['user_message']}"
    )
    response = llm.invoke(prompt)
    kind = response.content.strip().lower()
    return {"issue_type": kind}

def handle_invoice(state: dict) -> dict:
    # Fetch invoice details...
    return {"step_result": f"Invoice details for {state['user_id']}"}

def handle_refund(state: dict) -> dict:
    # Initiate refund workflow...
    return {"step_result": "Refund process initiated"}

def handle_login(state: dict) -> dict:
    # Troubleshoot login...
    return {"step_result": "Password reset link sent"}

def handle_performance(state: dict) -> dict:
    # Check performance metrics...
    return {"step_result": "Performance metrics analyzed"}

def summarize_response(state: dict) -> dict:
    # Consolidate previous step_result into a user-facing message
    details = state.get("step_result", "")
    response = llm.invoke(f"Write a concise customer reply based on: {details}")
    summary = response.content.strip()
    return {"response": summary}

# 2. Build the graph
graph = StateGraph(State)

# Add all nodes first
graph.add_node("categorize_issue", categorize_issue)
graph.add_node("handle_invoice", handle_invoice)
graph.add_node("handle_refund", handle_refund)
graph.add_node("handle_login", handle_login)
graph.add_node("handle_performance", handle_performance)
graph.add_node("summarize_response", summarize_response)

# Start → categorize_issue
graph.add_edge(START, "categorize_issue")

# categorize_issue → billing or technical
def top_router(state):
    return "billing" if state["issue_type"] == "billing" else "technical"

graph.add_conditional_edges(
    "categorize_issue",
    top_router,
    {"billing": "handle_invoice", "technical": "handle_login"}
)

# Billing sub-branches: invoice vs. refund
def billing_router(state):
    msg = state["user_message"].lower()
    return "invoice" if "invoice" in msg else "refund"

graph.add_conditional_edges(
    "handle_invoice",
    billing_router,
    {"invoice": "handle_invoice", "refund": "handle_refund"}
)

# Technical sub-branches: login vs. performance
def tech_router(state):
    msg = state["user_message"].lower()
    return "login" if "login" in msg else "performance"

graph.add_conditional_edges(
    "handle_login",
    tech_router,
    {"login": "handle_login", "performance": "handle_performance"}
)

# Consolidation: both refund and performance (and invoice/login) lead here
graph.add_edge("handle_refund", "summarize_response")
graph.add_edge("handle_performance", "summarize_response")
# Also cover paths where invoice or login directly go to summary
graph.add_edge("handle_invoice", "summarize_response")
graph.add_edge("handle_login", "summarize_response")

# Final: summary → END
graph.add_edge("summarize_response", END)

# Compile the graph
app = graph.compile()
# 3. Execute the graph
initial_state = {
    "user_message": "Hi, I need help with my invoice and possibly a refund.",
    "user_id": "U1234"
}
result = app.invoke(initial_state)
print(result["response"])