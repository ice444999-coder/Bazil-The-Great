"""Simple ping test to verify WebSocket connectivity."""
import asyncio
import websockets
import json


async def test_ping():
    uri = "ws://localhost:8765"
    print("Connecting to", uri)
    
    try:
        async with websockets.connect(uri) as ws:
            print("Connected! Sending ping...")
            await ws.send(json.dumps({"type": "ping"}))
            
            response = await ws.recv()
            print("Received:", response)
            
            msg = json.loads(response)
            if msg['type'] == 'pong':
                print("✅ Ping/Pong successful!")
                return True
            else:
                print("❌ Unexpected response type:", msg['type'])
                return False
                
    except Exception as e:
        print(f"❌ Error: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = asyncio.run(test_ping())
    exit(0 if success else 1)
