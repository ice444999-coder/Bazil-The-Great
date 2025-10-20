#!/usr/bin/env python3
"""
Test OpenAI Chat Integration
Tests SOLACE's OpenAI function calling via WebSocket.
"""

import asyncio
import websockets
import json
import sys

async def test_chat():
    """Test OpenAI chat with function calling."""
    uri = "ws://localhost:8765"
    
    print("=" * 70)
    print("🤖 Testing SOLACE OpenAI Chat Integration")
    print("=" * 70)
    print(f"\nConnecting to {uri}...")
    
    try:
        async with websockets.connect(uri) as websocket:
            print("✅ Connected to WebSocket server")
            print()
            
            # Test chat with function calling
            test_message = "List files in the current directory"
            print(f"📤 Sending chat message: '{test_message}'")
            
            await websocket.send(json.dumps({
                "type": "chat",
                "data": {"message": test_message}
            }))
            
            print("\n📥 Waiting for responses...\n")
            
            # Receive responses (status updates and final response)
            while True:
                response = await websocket.recv()
                data = json.loads(response)
                
                msg_type = data.get('type')
                
                if msg_type == 'status':
                    # Status update during function execution
                    print(f"⏳ Status: {data.get('message', '')}")
                
                elif msg_type == 'chat_response':
                    # Final response from SOLACE
                    print(f"\n🤖 SOLACE Response:")
                    print("-" * 70)
                    print(data.get('message', ''))
                    print("-" * 70)
                    break
                
                elif msg_type == 'error':
                    print(f"\n❌ Error: {data.get('message', '')}")
                    break
                
                else:
                    print(f"\n❓ Unknown response type: {msg_type}")
                    print(f"   Data: {data}")
            
            print("\n" + "=" * 70)
            print("✅ Test complete!")
            print("=" * 70)
    
    except ConnectionRefusedError:
        print("\n❌ ERROR: Could not connect to WebSocket server")
        print("   Make sure the server is running:")
        print("   python test_websocket_server.py")
        print("   OR")
        print("   python coordinator.py --websocket")
    
    except Exception as e:
        print(f"\n❌ ERROR: {e}")
        import traceback
        traceback.print_exc()


if __name__ == '__main__':
    print("\nNote: This test requires:")
    print("  1. WebSocket server running (port 8765)")
    print("  2. OPENAI_API_KEY environment variable set")
    print()
    
    asyncio.run(test_chat())
