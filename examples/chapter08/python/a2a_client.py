import requests
import json

# 发现 Agent Card（模拟为直接访问；在生产中，查询注册表）
card_url = 'http://localhost:8000/.well-known/agent.json'
response = requests.get(card_url)
if response.status_code != 200:
    raise ValueError("Failed to retrieve Agent Card")
agent_card = response.json()
print("Discovered Agent Card:", json.dumps(agent_card, indent=2))

agent_card = {
    "identity": "SummarizerAgent",
    "capabilities": ["summarizeText"],
    "schemas": {
        "summarizeText": {
            "input": {"text": "string"},
            "output": {"summary": "string"}
        }
    },
    "endpoint": "http://localhost:8000/api",
    "auth_methods": ["none"],  # 在生产环境中：OAuth2, API keys 等
    "version": "1.0"}


# 握手：检查兼容性
if agent_card['version'] != '1.0':
    raise ValueError("Incompatible protocol version")
if "summarizeText" not in agent_card['capabilities']:
    raise ValueError("Required capability not supported")
print("Handshake successful: Agent is compatible.")

# 发出 JSON-RPC 请求
rpc_url = agent_card['endpoint']
rpc_request = {
    "jsonrpc": "2.0",
    "method": "summarizeText",
    "params": {"text": '''This is a long example text that needs summarization.
         It discusses multiagent systems and communication protocols.'''},
    "id": 123  # 唯一的请求 ID
}
response = requests.post(rpc_url, json=rpc_request)
if response.status_code == 200:
    rpc_response = response.json()
    print("RPC Response:", json.dumps(rpc_response, indent=2))
else:
    print("Error:", response.status_code, response.text)
