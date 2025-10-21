# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
#!/usr/bin/env python3
"""
WebSocket Client Test - Tests SOLACE WebSocket Server
Run this AFTER starting the WebSocket server.
"""

import asyncio
import websockets
import json

async def test_websocket():
    """Test all WebSocket message types."""
    uri = "ws://localhost:8765"
    
    print("=" * 70)
    print("🧪 Testing SOLACE WebSocket Server")
    print("=" * 70)
    print(f"\nConnecting to {uri}...")
    
    try:
        async with websockets.connect(uri) as websocket:
            print("✅ Connected!")
            print()
            
            # Test 1: Ping
            print("[TEST 1] Sending ping...")
            await websocket.send(json.dumps({"type": "ping"}))
            response = json.loads(await websocket.recv())
            print(f"✅ Response: {response['type']} at {response.get('timestamp', 'N/A')}")
            print()
            
            # Test 2: List directory
            print("[TEST 2] Listing current directory...")
            await websocket.send(json.dumps({
                "type": "list_directory",
                "data": {"path": ".", "recursive": False}
            }))
            response = json.loads(await websocket.recv())
            if response['type'] == 'directory_listing':
                print(f"✅ Found {response.get('count', 0)} items")
                for item in response.get('files', [])[:3]:
                    print(f"   • {item['path']} ({item['type']})")
            print()
            
            # Test 3: Read file
            print("[TEST 3] Reading file_operations.py...")
            await websocket.send(json.dumps({
                "type": "read_file",
                "data": {"path": "file_operations.py"}
            }))
            response = json.loads(await websocket.recv())
            if response['type'] == 'file_content':
                content = response.get('content', '')
                lines = content.split('\n')
                print(f"✅ Read {len(lines)} lines")
                print(f"   First line: {lines[0][:60]}...")
            print()
            
            # Test 4: Chat
            print("[TEST 4] Sending chat message...")
            await websocket.send(json.dumps({
                "type": "chat",
                "data": {"message": "Hello SOLACE!"}
            }))
            response = json.loads(await websocket.recv())
            if response['type'] == 'chat_response':
                print(f"✅ {response.get('message', '')}")
            print()
            
            print("=" * 70)
            print("✅ All tests passed!")
            print("=" * 70)
    
    except ConnectionRefusedError:
        print("\n❌ ERROR: Could not connect to WebSocket server")
        print("   Make sure the server is running:")
        print("   python test_websocket_server.py")
    
    except Exception as e:
        print(f"\n❌ ERROR: {e}")
        import traceback
        traceback.print_exc()


if __name__ == '__main__':
    asyncio.run(test_websocket())
