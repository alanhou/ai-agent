from dotenv import load_dotenv
load_dotenv()
import json
from openai import OpenAI
from http.server import BaseHTTPRequestHandler, HTTPServer

#  Agent Card (JSON descriptor for discovery)
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


class AgentHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/.well-known/agent.json':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(agent_card).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == '/api':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            rpc_request = json.loads(post_data)
            # 处理 JSON-RPC 请求（A2A 的核心）
            if rpc_request.get('jsonrpc') == '2.0' \
                and rpc_request['method'] == 'summarizeText':
                text = rpc_request['params']['text']
                # 使用 OpenAI API 进行真正的 LLM 摘要
                client = OpenAI()
                try:
                    llm_response = client.chat.completions.create(
                        model="gpt-4o",
                        messages=[
                            {"role": "system", "content": '''You are a helpful assistant that provides concise summaries.'''},
                            {"role": "user", "content": f"""Summarize the following text: {text}"""}
                        ],
                        max_tokens=150,
                        temperature=0.7
                    )
                    summary = llm_response.choices[0].message.content.strip()
                except Exception as e:
                    summary = f"Error in summarization: {str(e)}"  # 错误的后备方案
                
                response = {
                    "jsonrpc": "2.0",
                    "result": {"summary": summary},
                    "id": rpc_request['id']
                }
                # 发送响应
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            else:
                # 错误响应
                error_response = {
                    "jsonrpc": "2.0",
                    "error": {"code": -32601, "message": "Method not found"},
                    "id": rpc_request.get('id')
                }
                self.send_response(400)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(error_response).encode())
        else:
            self.send_response(404)
            self.end_headers()

if __name__ == '__main__':
    server_address = ('', 8000)
    httpd = HTTPServer(server_address, AgentHandler)
    print("Starting A2A agent server on http://localhost:8000")
    httpd.serve_forever()