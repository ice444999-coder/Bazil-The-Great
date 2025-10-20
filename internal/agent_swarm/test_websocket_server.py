#!/usr/bin/env python3
"""
SOLACE WebSocket Server - Standalone Test
Tests WebSocket functionality without database dependencies.
"""

import asyncio
import websockets
import json
import logging
from datetime import datetime
from pathlib import Path

# Import file operations
import sys
sys.path.append(str(Path(__file__).parent))
from file_operations import read_file, write_file, list_directory

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s'
)
logger = logging.getLogger(__name__)


async def handle_websocket(websocket):
    """Handle WebSocket connections for SOLACE."""
    logger.info(f"[WEBSOCKET] New connection from {websocket.remote_address}")
    
    try:
        async for message in websocket:
            try:
                # Parse incoming JSON message
                logger.debug(f"[WEBSOCKET] Raw message: {message}")
                data = json.loads(message)
                msg_type = data.get('type')
                msg_data = data.get('data', {})
                
                logger.info(f"[WEBSOCKET] Received: {msg_type}")
                
                # Handle different message types
                if msg_type == 'ping':
                    response = {'type': 'pong', 'timestamp': datetime.now().isoformat()}
                
                elif msg_type == 'read_file':
                    file_path = msg_data.get('path')
                    if not file_path:
                        response = {'type': 'error', 'message': 'Missing file path'}
                    else:
                        content = read_file(file_path)
                        response = {'type': 'file_content', 'path': file_path, 'content': content}
                
                elif msg_type == 'write_file':
                    file_path = msg_data.get('path')
                    content = msg_data.get('content')
                    if not file_path or content is None:
                        response = {'type': 'error', 'message': 'Missing path or content'}
                    else:
                        write_file(file_path, content)
                        response = {'type': 'write_success', 'path': file_path}
                
                elif msg_type == 'list_directory':
                    dir_path = msg_data.get('path', '.')
                    recursive = msg_data.get('recursive', True)
                    max_depth = msg_data.get('max_depth', 5)
                    files = list_directory(dir_path, recursive=recursive, max_depth=max_depth)
                    response = {'type': 'directory_listing', 'path': dir_path, 'files': files, 'count': len(files)}
                
                elif msg_type == 'chat':
                    user_message = msg_data.get('message', '')
                    response = {
                        'type': 'chat_response',
                        'message': f'SOLACE WebSocket connected. Received: "{user_message}"',
                        'note': 'OpenAI integration coming soon'
                    }
                
                elif msg_type == 'get_architecture':
                    # Import here to avoid dependency issues if coordinator has database deps
                    try:
                        from coordinator import query_architecture_rules
                        feature_type = msg_data.get('feature_type')
                        rules = query_architecture_rules(feature_type)
                        response = {
                            'type': 'architecture_rules',
                            'rules': rules,
                            'count': len(rules),
                            'filter': feature_type
                        }
                    except Exception as e:
                        response = {'type': 'error', 'message': f'Architecture query failed: {str(e)}'}
                
                elif msg_type == 'create_backup':
                    workspace_path = msg_data.get('workspace_path')
                    if not workspace_path:
                        response = {'type': 'error', 'message': 'Missing workspace_path'}
                    else:
                        from file_operations import create_backup
                        backup_path = create_backup(workspace_path)
                        response = {
                            'type': 'backup_created',
                            'backup_path': backup_path
                        }
                
                elif msg_type == 'restore_backup':
                    backup_path = msg_data.get('backup_path')
                    workspace_path = msg_data.get('workspace_path')
                    if not backup_path or not workspace_path:
                        response = {'type': 'error', 'message': 'Missing backup_path or workspace_path'}
                    else:
                        from file_operations import restore_backup
                        restore_backup(backup_path, workspace_path)
                        response = {
                            'type': 'restore_complete'
                        }
                
                elif msg_type == 'execute_command':
                    command = msg_data.get('command')
                    if not command:
                        response = {'type': 'error', 'message': 'Missing command'}
                    else:
                        cwd = msg_data.get('cwd', '.')
                        import subprocess
                        result = subprocess.run(
                            ["powershell", "-Command", command],
                            cwd=cwd,
                            capture_output=True,
                            text=True,
                            timeout=300
                        )
                        response = {
                            'type': 'command_output',
                            'stdout': result.stdout,
                            'stderr': result.stderr,
                            'exit_code': result.returncode
                        }
                
                else:
                    response = {'type': 'error', 'message': f'Unknown message type: {msg_type}'}
                
                # Send response
                await websocket.send(json.dumps(response))
                logger.info(f"[WEBSOCKET] ✅ Sent response for {msg_type}")
                
            except json.JSONDecodeError as e:
                error_response = {'type': 'error', 'message': f'Invalid JSON: {str(e)}'}
                await websocket.send(json.dumps(error_response))
                logger.error(f"[WEBSOCKET] JSON decode error: {e}")
            
            except FileNotFoundError as e:
                error_response = {'type': 'error', 'message': f'File not found: {str(e)}'}
                await websocket.send(json.dumps(error_response))
                logger.error(f"[WEBSOCKET] File not found: {e}")
            
            except Exception as e:
                error_response = {'type': 'error', 'message': str(e)}
                await websocket.send(json.dumps(error_response))
                logger.error(f"[WEBSOCKET] ❌ Error handling message: {e}")
                import traceback
                logger.error(traceback.format_exc())
    
    except websockets.exceptions.ConnectionClosed:
        logger.info(f"[WEBSOCKET] Connection closed")
    
    except Exception as e:
        logger.error(f"[WEBSOCKET] Connection error: {e}")


async def start_server():
    """Start the WebSocket server."""
    print("=" * 70)
    print("SOLACE WebSocket Server - Standalone Test")
    print("=" * 70)
    print()
    
    async with websockets.serve(handle_websocket, "localhost", 8765):
        print(">> WebSocket server started on ws://localhost:8765")
        print()
        print("Available message types:")
        print("  * ping              - Health check")
        print("  * read_file         - Read file contents")
        print("  * write_file        - Write file contents")
        print("  * list_directory    - List directory")
        print("  * chat              - Chat with SOLACE")
        print("  * create_backup     - Create workspace backup")
        print("  * restore_backup    - Restore from backup")
        print("  * execute_command   - Execute PowerShell commands")
        print()
        print("Press Ctrl+C to stop")
        print("=" * 70)
        
        # Run forever
        await asyncio.Future()


if __name__ == '__main__':
    try:
        asyncio.run(start_server())
    except KeyboardInterrupt:
        print("\n[WEBSOCKET] Server stopped by user")
    except Exception as e:
        print(f"\n[ERROR] {e}")
        import traceback
        traceback.print_exc()
