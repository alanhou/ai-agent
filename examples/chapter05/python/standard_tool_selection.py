from dotenv import load_dotenv
load_dotenv()
from langchain_core.tools import tool
import requests
import os

@tool
def get_stock_price(ticker: str) -> float:
    """Get stock price via Finnhub REST API."""
    # Finnhub uses 'c' for the current price in their JSON response
    url = f"https://finnhub.io/api/v1/quote?symbol={ticker}&token={os.environ.get('FINHUB_API_KEY')}"
    
    response = requests.get(url)
    
    if response.status_code == 200:
        data = response.json()
        # Check if price exists (Finnhub returns 0 for invalid tickers sometimes)
        if data.get('c', 0) == 0:
             raise ValueError(f"Ticker {ticker} not found or price is zero.")
        return data["c"]
    else:
        raise ValueError(f"API Error: {response.status_code}")

@tool
def query_wolfram_alpha(expression: str) -> str:
    """
    Query Wolfram Alpha to compute expressions or retrieve information.
    Args: expression (str): The mathematical expression or query to evaluate.
    Returns: str: The result of the computation or the retrieved information.
    """
    api_url = f"https://api.wolframalpha.com/v1/result?i={requests.utils.quote(expression)}&appid={os.environ.get('WOLFRAM_ALPHA_APP_ID')}"
    
    try:
        response = requests.get(api_url)
        if response.status_code == 200:
            return response.text
        else: 
            raise ValueError(f"Wolfram Alpha API Error: {response.status_code} - {response.text}")
    except requests.exceptions.RequestException as e:
        raise ValueError(f"Failed to query Wolfram Alpha: {e}")

@tool
def send_slack_message(channel: str, message: str) -> str:
    """
    Send a message to a specified Slack channel.
    Args:
    channel (str): The Slack channel ID or name where the message will be sent.
    message (str): The content of the message to send.
    Returns:
    str: Confirmation message upon successful sending of the Slack message.
    Raises: ValueError: If the API request fails or returns an error.
    """
    api_url = "https://slack.com/api/chat.postMessage"
    token = os.environ.get("SLACK_BOT_TOKEN")
    headers = { "Authorization": f"Bearer {token}",
                "Content-Type": "application/json" }
    payload = { "channel": channel, "text": message }
    try:
        response = requests.post(api_url, headers=headers, json=payload)
        response_data = response.json()
        if response.status_code == 200 and response_data.get("ok"):
            return f"Message successfully sent to Slack channel '{channel}'."
        else:
            error_msg = response_data.get("error", "Unknown error")
            raise ValueError(f"Slack API Error: {error_msg}")
    except requests.exceptions.RequestException as e:
        raise ValueError(f'''Failed to send message to Slack channel "{channel}": {e}''')

    
if __name__ == "__main__":
    from langchain_core.messages import ToolMessage
    from langchain_openai import ChatOpenAI
    from langchain_core.messages import HumanMessage
    
    # Initialize the LLM with GPT-4o and bind the tools
    llm = ChatOpenAI(model_name="gpt-4o")
    tools = [get_stock_price, send_slack_message, query_wolfram_alpha]
    tool_map = {tool.name: tool for tool in tools}
    llm_with_tools = llm.bind_tools(tools)

    messages = [HumanMessage("What is the stock price of Apple?")]
    ai_msg = llm_with_tools.invoke(messages)
    messages.append(ai_msg)

    for tool_call in ai_msg.tool_calls:
        tool_name = tool_call["name"]
        selected_tool = tool_map[tool_name]
        tool_output = selected_tool.invoke(tool_call)
        messages.append(ToolMessage(tool_output, tool_call_id=tool_call["id"]))

    final_response = llm_with_tools.invoke(messages)
    print(final_response.content)