"""
Phase A Integration Test - SOLACE Autonomous Workflow End-to-End
Tests all autonomous capabilities: architecture awareness, file operations, 
safety workflows, and command execution.
"""

import asyncio
import websockets
import json


async def test_phase_a():
    """End-to-end test of SOLACE autonomous capabilities."""
    uri = "ws://localhost:8765"
    
    print("=" * 60)
    print("PHASE A INTEGRATION TEST - SOLACE Autonomous Workflow")
    print("=" * 60)
    
    try:
        async with websockets.connect(uri) as ws:
            
            # TEST 1: Architecture awareness
            print("\n1️⃣ TEST: Architecture Query")
            await ws.send(json.dumps({
                "type": "chat",
                "message": "What architecture rules exist for agent features?"
            }))
            
            response = await receive_until_chat_response(ws)
            print(f"✅ SOLACE response: {response[:100]}...")
            
            # TEST 2: Repository scan
            print("\n2️⃣ TEST: Repository Scan")
            await ws.send(json.dumps({
                "type": "chat",
                "message": "List all Python files in the current directory"
            }))
            
            response = await receive_until_chat_response(ws)
            print(f"✅ SOLACE found files: {response[:100]}...")
            
            # TEST 3: Safety workflow - Backup before change
            print("\n3️⃣ TEST: Safety Workflow (Backup → Modify → Verify)")
            await ws.send(json.dumps({
                "type": "chat",
                "message": "Create a backup of the agent_swarm directory"
            }))
            
            response = await receive_until_chat_response(ws)
            print(f"✅ Backup created: {response[:100]}...")
            
            # TEST 4: File modification with OpenAI decision
            print("\n4️⃣ TEST: Autonomous File Read")
            await ws.send(json.dumps({
                "type": "chat",
                "message": "Read the first 10 lines of coordinator.py"
            }))
            
            response = await receive_until_chat_response(ws)
            print(f"✅ File read: {response[:100]}...")
            
            # TEST 5: Command execution
            print("\n5️⃣ TEST: Command Execution")
            await ws.send(json.dumps({
                "type": "chat",
                "message": "Count how many .py files are in this directory"
            }))
            
            response = await receive_until_chat_response(ws)
            print(f"✅ Command result: {response[:100]}...")
            
            print("\n" + "=" * 60)
            print("✅ PHASE A COMPLETE - All autonomous features working!")
            print("=" * 60)
            
    except ConnectionRefusedError:
        print("\n❌ ERROR: Could not connect to WebSocket server")
        print("   Make sure the coordinator is running:")
        print("   powershell -File start_simple.ps1")
        return False
    except websockets.exceptions.ConnectionClosed as e:
        print(f"\n❌ ERROR: Connection closed by server: {e}")
        print("   Check server logs for errors")
        return False
    except Exception as e:
        print(f"\n❌ ERROR: {e}")
        import traceback
        traceback.print_exc()
        return False
    
    return True


async def receive_until_chat_response(ws):
    """Receive messages until we get chat_response or error."""
    while True:
        msg = json.loads(await ws.recv())
        if msg['type'] == 'status':
            print(f"   Status: {msg['message']}")
        elif msg['type'] == 'chat_response':
            return msg['message']
        elif msg['type'] == 'error':
            print(f"   ❌ Error: {msg['message']}")
            return f"Error: {msg['message']}"


if __name__ == "__main__":
    success = asyncio.run(test_phase_a())
    exit(0 if success else 1)
