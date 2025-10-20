#!/usr/bin/env python3
"""
Test Backup, Restore, and Command Execution via WebSocket
Tests the 3 new WebSocket message types added in Instruction #6.
"""

import asyncio
import websockets
import json
import sys

async def test_all():
    """Test all 3 new WebSocket message types."""
    uri = "ws://localhost:8765"
    
    print("=" * 70)
    print(" Testing Backup, Restore, and Command Execution")
    print("=" * 70)
    print(f"\nConnecting to {uri}...")
    
    try:
        async with websockets.connect(uri) as websocket:
            print(" Connected to WebSocket server")
            print()
            
            # Test 1: Create backup
            print("TEST 1: Create Backup")
            print("-" * 70)
            await websocket.send(json.dumps({
                "type": "create_backup",
                "data": {"workspace_path": "C:/ARES_Workspace/ARES_API/internal/agent_swarm"}
            }))
            response = json.loads(await websocket.recv())
            
            if response.get('type') == 'backup_created':
                backup_path = response.get('backup_path')
                print(f" Backup created successfully!")
                print(f"   Backup path: {backup_path}")
            else:
                print(f" Error: {response}")
                return
            
            print()
            
            # Test 2: Execute command (count Python files)
            print("TEST 2: Execute Command (Count Python files)")
            print("-" * 70)
            await websocket.send(json.dumps({
                "type": "execute_command",
                "data": {
                    "command": "Get-ChildItem -Filter *.py | Measure-Object | Select-Object -ExpandProperty Count",
                    "cwd": "C:/ARES_Workspace/ARES_API/internal/agent_swarm"
                }
            }))
            response = json.loads(await websocket.recv())
            
            if response.get('type') == 'command_output':
                stdout = response.get('stdout', '').strip()
                stderr = response.get('stderr', '').strip()
                exit_code = response.get('exit_code', -1)
                
                print(f" Command executed successfully!")
                print(f"   Python files found: {stdout}")
                print(f"   Exit code: {exit_code}")
                
                if stderr:
                    print(f"   Stderr: {stderr}")
            else:
                print(f" Error: {response}")
                return
            
            print()
            
            # Test 3: Execute another command (list .py files)
            print("TEST 3: Execute Command (List Python files)")
            print("-" * 70)
            await websocket.send(json.dumps({
                "type": "execute_command",
                "data": {
                    "command": "Get-ChildItem -Filter *.py | Select-Object -ExpandProperty Name",
                    "cwd": "C:/ARES_Workspace/ARES_API/internal/agent_swarm"
                }
            }))
            response = json.loads(await websocket.recv())
            
            if response.get('type') == 'command_output':
                stdout = response.get('stdout', '').strip()
                exit_code = response.get('exit_code', -1)
                
                print(f" Command executed successfully!")
                print(f"   Files found:")
                for line in stdout.split('\n')[:10]:  # Show first 10 files
                    if line.strip():
                        print(f"     - {line.strip()}")
                print(f"   Exit code: {exit_code}")
            else:
                print(f" Error: {response}")
                return
            
            print()
            
            # Test 4: Restore backup (COMMENTED OUT - use with caution!)
            print("TEST 4: Restore Backup (SKIPPED - enable manually if needed)")
            print("-" * 70)
            print("  Restore is commented out for safety.")
            print("   To test restore, uncomment the code below and re-run.")
            print(f"   Backup available at: {backup_path}")
            
            # UNCOMMENT TO TEST RESTORE:
            # print(f"   Restoring from: {backup_path}")
            # await websocket.send(json.dumps({
            #     "type": "restore_backup",
            #     "data": {
            #         "backup_path": backup_path,
            #         "workspace_path": "C:/ARES_Workspace/ARES_API/internal/agent_swarm"
            #     }
            # }))
            # response = json.loads(await websocket.recv())
            # 
            # if response.get('type') == 'restore_complete':
            #     print(f" Restore completed successfully!")
            # else:
            #     print(f" Error: {response}")
            
            print()
            print("=" * 70)
            print(" All tests completed!")
            print("=" * 70)
    
    except ConnectionRefusedError:
        print("\n ERROR: Could not connect to WebSocket server")
        print("   Make sure the server is running:")
        print("   python test_websocket_server.py")
        print("   OR")
        print("   python coordinator.py --websocket")
        sys.exit(1)
    
    except Exception as e:
        print(f"\n ERROR: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    print("\n  Note: This test requires:")
    print("  1. WebSocket server running (port 8765)")
    print("  2. file_operations.py with create_backup/restore_backup functions")
    print("  3. PowerShell available on the system")
    print()
    
    asyncio.run(test_all())

