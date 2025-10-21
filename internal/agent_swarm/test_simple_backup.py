# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
#!/usr/bin/env python3
"""Simple test for create_backup via WebSocket"""

import asyncio
import websockets
import json

async def test_backup():
    uri = "ws://localhost:8765"
    try:
        async with websockets.connect(uri) as websocket:
            print("✅ Connected")
            
            # Send ping first
            await websocket.send(json.dumps({"type": "ping"}))
            response = json.loads(await websocket.recv())
            print(f"Ping response: {response}")
            
            # Test create_backup
            print("\nTesting create_backup...")
            await websocket.send(json.dumps({
                "type": "create_backup",
                "data": {"workspace_path": "C:/ARES_Workspace/ARES_API/internal/agent_swarm"}
            }))
            
            response = await websocket.recv()
            print(f"Raw response: {response}")
            data = json.loads(response)
            print(f"Parsed response: {data}")
            
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()

asyncio.run(test_backup())
