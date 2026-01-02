import os, json, base64, asyncio, websockets
from fastapi import FastAPI, WebSocket
from dotenv import load_dotenv

load_dotenv()

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
VOICE          = "alloy"                 # GPT-4o 声音
PCM_SR         = 16000                   # 我们将在客户端使用的采样率
PORT           = 5050

app = FastAPI()

@app.websocket("/voice")
async def voice_bridge(ws: WebSocket) -> None:
    """
    1. 浏览器打开 ws://localhost:5050/voice
    2. 浏览器流式传输 base64 编码的 16 位单声道 PCM 数据块: {"audio": "<b64>"}
    3. 我们将数据块转发给 OpenAI Realtime (`input_audio_buffer.append`)
    4. 我们以同样的方式将助手的音频增量中继回浏览器
    5. 我们监听 'speech_started' 事件，如果用户打断则发送截断指令
    """
    await ws.accept()

    # websockets < 13 uses extra_headers, >= 13 uses additional_headers
    import websockets.version
    ws_version = tuple(map(int, websockets.version.version.split('.')[:2]))
    headers = {
        "Authorization": f"Bearer {OPENAI_API_KEY}",
        "OpenAI-Beta": "realtime=v1"
    }
    header_param = "additional_headers" if ws_version >= (13, 0) else "extra_headers"
    
    openai_ws = await websockets.connect(
        "wss://api.openai.com/v1/realtime?" + 
            "model=gpt-4o-realtime-preview-2024-10-01", 
        **{header_param: headers},
        max_size=None, max_queue=None  # 为了演示简单，不做限制
    )

    # 初始化实时会话
    await openai_ws.send(json.dumps({
        "type": "session.update",
        "session": {
            "turn_detection": {"type": "server_vad"},
            "input_audio_format": f"pcm_{PCM_SR}",
            "output_audio_format": f"pcm_{PCM_SR}",
            "voice": VOICE,
            "modalities": ["audio"],
            "instructions": "You are a concise AI assistant."
        }
    }))

    last_assistant_item = None          # 追踪当前助手回复
    latest_pcm_ts       = 0             # 来自客户端的 ms 时间戳
    pending_marks       = []

    async def from_client() -> None:
        """将麦克风 PCM 数据块从浏览器中继 → OpenAI。"""
        nonlocal latest_pcm_ts
        async for msg in ws.iter_text():
            data = json.loads(msg)
            pcm = base64.b64decode(data["audio"])
            latest_pcm_ts += int(len(pcm) / (PCM_SR * 2) * 1000) 
            await openai_ws.send(json.dumps({
                "type": "input_audio_buffer.append",
                "audio": base64.b64encode(pcm).decode("ascii")
            }))

    async def to_client() -> None:
        """中继助手音频 + 处理打断。"""
        nonlocal last_assistant_item, pending_marks
        async for raw in openai_ws:
            msg = json.loads(raw)

            # 助手发言
            if msg["type"] == "response.audio.delta":
                pcm = base64.b64decode(msg["delta"])
                await ws.send_json({"audio": 
                    base64.b64encode(pcm).decode("ascii")})
                last_assistant_item = msg.get("item_id")

            # 用户开始说话 → 取消助手语音
            started = "input_audio_buffer.speech_started"
            if msg["type"] == started and last_assistant_item:
                await openai_ws.send(json.dumps({
                    "type": "conversation.item.truncate",
                    "item_id": last_assistant_item,
                    "content_index": 0,
                    "audio_end_ms": 0   # 立即停止
                }))
                last_assistant_item = None
                pending_marks.clear()

    try:
        await asyncio.gather(from_client(), to_client())
    finally:
        await openai_ws.close()
        await ws.close()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=PORT)