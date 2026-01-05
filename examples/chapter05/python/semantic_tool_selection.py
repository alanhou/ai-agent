from dotenv import load_dotenv
load_dotenv()
import os
import requests
import logging
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_core.messages import HumanMessage, AIMessage, ToolMessage
from langchain_community.vectorstores import FAISS
import faiss
import numpy as np

# Initialize OpenAI embeddings
embeddings = OpenAIEmbeddings()

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
def trigger_zapier_webhook(zap_id: str, payload: dict) -> str: 
    """ Trigger a Zapier webhook to execute a predefined Zap. 
    Args: 
    zap_id (str): The unique identifier for the Zap to be triggered. 
    payload (dict): The data to send to the Zapier webhook. 
    Returns: 
    str: Confirmation message upon successful triggering of the Zap. 
    Raises: ValueError: If the API request fails or returns an error. 
    """ 

    zapier_webhook_url = f"https://hooks.zapier.com/hooks/catch/{zap_id}/" 
    try: 
        response = requests.post(zapier_webhook_url, json=payload) 
        if response.status_code == 200: 
            return f"Zapier webhook '{zap_id}' successfully triggered." 
        else: 
            raise ValueError(f'''Zapier API Error: {response.status_code} - {response.text}''') 
    except requests.exceptions.RequestException as e: 
        raise ValueError(f"Failed to trigger Zapier webhook '{zap_id}': {e}") 

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



# Tool descriptions
tool_descriptions = {
       "query_wolfram_alpha": '''Use Wolfram Alpha to compute mathematical expressions or retrieve information.''',
       "trigger_zapier_webhook": '''Trigger a Zapier webhook to execute predefined automated workflows.''',
       "send_slack_message": '''Send messages to specific Slack channels to communicate with team members.'''
}

# Create embeddings for each tool description
tool_embeddings = []
tool_names = []

for tool_name, description in tool_descriptions.items():
   embedding = embeddings.embed_query(description)
   tool_embeddings.append(embedding)
   tool_names.append(tool_name)
# Initialize FAISS vector store
dimension = len(tool_embeddings[0])  
index = faiss.IndexFlatL2(dimension)

# Normalize embeddings for cosine similarity
faiss.normalize_L2(np.array(tool_embeddings).astype('float32'))

llm = ChatOpenAI(model_name="gpt-4o")
# Convert list to FAISS-compatible format
tool_embeddings_np = np.array(tool_embeddings).astype('float32')
index.add(tool_embeddings_np)

# Map index to tool functions
index_to_tool = {
   0: query_wolfram_alpha,
   1: trigger_zapier_webhook,
   2: send_slack_message
}

def select_tool(query: str, top_k: int = 1) -> list:
   """
   Select the most relevant tool(s) based on the user's query using 
   vector-based retrieval.
  
   Args:
       query (str): The user's input query.
       top_k (int): Number of top tools to retrieve.
      
   Returns:
       list: List of selected tool functions.
   """
#    query_embedding = embeddings.embed_query(query).astype('float32')
   query_embedding = np.array(embeddings.embed_query(query)).astype('float32')
   faiss.normalize_L2(query_embedding.reshape(1, -1))
   D, I = index.search(query_embedding.reshape(1, -1), top_k)
   selected_tools = [index_to_tool[idx] for idx in I[0] if idx in index_to_tool]
   return selected_tools

def determine_parameters(query: str, tool_name: str) -> dict:
   """
   Extract parameters for the tool based on the user's query.

   Args:
       query (str): The user's input query.
       tool_name (str): The selected tool name.

   Returns:
       dict: Parameters for the tool.
   """
   parameters = {}
   if tool_name == "query_wolfram_alpha":
       parameters["expression"] = query
   elif tool_name == "trigger_zapier_webhook":
       parameters["zap_id"] = "123456"
       parameters["payload"] = {"data": query}
   elif tool_name == "send_slack_message":
       parameters["channel"] = "#general"
       parameters["message"] = query

   return parameters

# Example user query
user_query = "Solve this equation: 2x + 3 = 7"

# Select the top tool
selected_tools = select_tool(user_query, top_k=1)
selected_tool = selected_tools[0] if selected_tools else None

if selected_tool:
    tool_name = selected_tool.name
    # Determine the parameters based on the query and the selected tool
    args = determine_parameters(user_query, tool_name)

    # Invoke the selected tool
    try:
       tool_result = selected_tool.invoke(args)
       print(f"Tool '{tool_name}' Result: {tool_result}")
    except ValueError as e:
       print(f"Error invoking tool '{tool_name}': {e}")
else:
   print("No tool was selected.")