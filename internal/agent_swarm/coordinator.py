#!/usr/bin/env python3
"""
ARES Agent Swarm Coordinator
Watches task_queue and executes tasks using specialized AI agents.

Agents:
- SOLACE (OpenAI GPT-4): Strategy, coordination, decision-making
- FORGE (Claude): UI building, coding, frontend work
- ARCHITECT (DeepSeek-R1): Planning, architecture, design patterns
- SENTINEL (DeepSeek-R1): Debugging, testing, validation

Usage:
    python coordinator.py [--interval SECONDS] [--debug]
"""

import os
import sys
import time
import json
import logging
import argparse
from datetime import datetime
from typing import Optional, Dict, Any, List

import psycopg2
import psycopg2.extras
from openai import OpenAI
from anthropic import Anthropic
from dotenv import load_dotenv
import requests
import asyncio
import websockets
import signal
from file_operations import read_file, write_file, list_directory

# Load environment variables from .env
load_dotenv()

# Validate required environment variables
required_keys = ['OPENAI_API_KEY', 'CLAUDE_API_KEY', 'DB_HOST', 'DB_USER', 'DB_PASSWORD', 'DB_NAME']
missing_keys = [key for key in required_keys if not os.getenv(key)]

if missing_keys:
    print(f"[ERROR] ERROR: Missing environment variables: {', '.join(missing_keys)}")
    print("Please add them to .env file")
    sys.exit(1)

print("All required environment variables loaded")

# Configure comprehensive logging with rotation
from logging.handlers import RotatingFileHandler

os.makedirs("logs", exist_ok=True)

logger = logging.getLogger("solace_coordinator")
logger.setLevel(logging.INFO)

# File handler with rotation (10MB per file, keep 5 backups)
file_handler = RotatingFileHandler(
    "logs/solace_coordinator.log", 
    maxBytes=10*1024*1024,  # 10MB
    backupCount=5,
    encoding='utf-8'
)
file_handler.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(message)s'))
logger.addHandler(file_handler)

# Console handler
console_handler = logging.StreamHandler()
console_handler.setFormatter(logging.Formatter('%(levelname)s: %(message)s'))
logger.addHandler(console_handler)

# Set console output to UTF-8 on Windows
if sys.platform == 'win32':
    import codecs
    sys.stdout = codecs.getwriter('utf-8')(sys.stdout.buffer, 'strict')
    sys.stderr = codecs.getwriter('utf-8')(sys.stderr.buffer, 'strict')

logger.info("=" * 70)
logger.info("SOLACE Coordinator Starting...")
logger.info("=" * 70)


def signal_handler(sig, frame):
    """Handle shutdown signals gracefully."""
    logger.info("=" * 70)
    logger.info("Shutdown signal received (Ctrl+C)")
    logger.info("Cleaning up processes...")
    try:
        cleanup_orphaned_powershells()
        logger.info("Cleanup complete. Exiting.")
    except Exception as e:
        logger.error(f"Error during cleanup: {e}")
    logger.info("=" * 70)
    sys.exit(0)


# Register signal handler for graceful shutdown
signal.signal(signal.SIGINT, signal_handler)
logger.info("Signal handler registered (Ctrl+C for graceful shutdown)")


class AgentCoordinator:
    """Coordinates task execution across multiple AI agents."""
    
    def __init__(self, db_config: Dict[str, str]):
        """Initialize coordinator with database connection."""
        self.db_config = db_config
        self.conn = None
        self.openai_client = None
        self.anthropic_client = None
        self.ollama_base_url = "http://localhost:11434"
        
        # Initialize API clients
        self._init_api_clients()
        
    def _init_api_clients(self):
        """Initialize API clients for different LLM providers."""
        # OpenAI (for SOLACE) - from env
        openai_key = os.getenv('OPENAI_API_KEY')
        if openai_key:
            self.openai_client = OpenAI(api_key=openai_key)
            logger.info("[OK] OpenAI client initialized (SOLACE)")
        else:
            logger.warning("[WARN] OPENAI_API_KEY not set - SOLACE unavailable")
            
        # Anthropic Claude (for FORGE) - from env
        claude_key = os.getenv('CLAUDE_API_KEY')
        if claude_key:
            self.anthropic_client = Anthropic(api_key=claude_key)
            logger.info("[OK] Claude client initialized (FORGE)")
        else:
            logger.warning("[WARN] CLAUDE_API_KEY not set - FORGE unavailable")
            
        # Ollama (for ARCHITECT, SENTINEL) - check if running
        try:
            import requests
            resp = requests.get(f"{self.ollama_base_url}/api/tags", timeout=2)
            if resp.status_code == 200:
                logger.info("[OK] Ollama available (ARCHITECT, SENTINEL)")
            else:
                logger.warning("[WARN] Ollama not responding - ARCHITECT/SENTINEL unavailable")
        except Exception as e:
            logger.warning(f"[WARN] Ollama not available: {e}")
    
    def connect_db(self):
        """Connect to PostgreSQL database."""
        try:
            self.conn = psycopg2.connect(**self.db_config)
            self.conn.autocommit = False
            logger.info("[OK] Connected to PostgreSQL (ares_db)")
        except Exception as e:
            logger.error(f"[ERROR] Database connection failed: {e}")
            raise
    
    def close_db(self):
        """Close database connection."""
        if self.conn:
            self.conn.close()
            logger.info("Database connection closed")
    
    def get_pending_tasks(self) -> List[Dict[str, Any]]:
        """Fetch tasks with status='assigned' that need execution."""
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                cur.execute("""
                    SELECT 
                        task_id, task_type, priority, status, 
                        created_by, assigned_to_agent, file_paths,
                        depends_on_task_ids, description, context,
                        created_at, assigned_at, deadline
                    FROM task_queue
                    WHERE status = 'assigned'
                    ORDER BY priority DESC, created_at ASC
                    LIMIT 10
                """)
                tasks = cur.fetchall()
                return [dict(task) for task in tasks]
        except Exception as e:
            logger.error(f"Error fetching tasks: {e}")
            return []
    
    def update_task_status(self, task_id: str, status: str, result: Optional[Dict] = None, error: Optional[str] = None):
        """Update task status in database."""
        try:
            with self.conn.cursor() as cur:
                if status == 'in_progress':
                    cur.execute("""
                        UPDATE task_queue 
                        SET status = %s, started_at = NOW()
                        WHERE task_id = %s
                    """, (status, task_id))
                elif status == 'completed':
                    cur.execute("""
                        UPDATE task_queue 
                        SET status = %s, completed_at = NOW(), result = %s
                        WHERE task_id = %s
                    """, (status, json.dumps(result) if result else '{}', task_id))
                elif status == 'failed':
                    cur.execute("""
                        UPDATE task_queue 
                        SET status = %s, completed_at = NOW(), error_log = %s, retry_count = retry_count + 1
                        WHERE task_id = %s
                    """, (status, error, task_id))
                
                self.conn.commit()
                logger.info(f"Task {task_id[:8]}... -> {status}")
        except Exception as e:
            self.conn.rollback()
            logger.error(f"Error updating task: {e}")
    
    def update_agent_status(self, agent_name: str, status: str, task_id: Optional[str] = None):
        """Update agent status (idle/busy)."""
        try:
            with self.conn.cursor() as cur:
                cur.execute("""
                    UPDATE agent_registry
                    SET status = %s, current_task_id = %s, last_active_at = NOW()
                    WHERE agent_name = %s
                """, (status, task_id, agent_name))
                self.conn.commit()
        except Exception as e:
            self.conn.rollback()
            logger.error(f"Error updating agent status: {e}")
    
    def log_task_history(self, agent_name: str, task_id: str, task_type: str, 
                        success: bool, duration_ms: int, error_msg: Optional[str] = None,
                        tokens_used: Optional[int] = None):
        """Log task execution to history table."""
        try:
            with self.conn.cursor() as cur:
                cur.execute("""
                    INSERT INTO agent_task_history 
                    (agent_name, task_id, task_type, success, duration_ms, error_message, cost_tokens)
                    VALUES (%s, %s, %s, %s, %s, %s, %s)
                """, (agent_name, task_id, task_type, success, duration_ms, error_msg, tokens_used))
                self.conn.commit()
        except Exception as e:
            self.conn.rollback()
            logger.error(f"Error logging history: {e}")
    
    def query_architecture_rules(self, feature_type: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        Query architecture_rules table directly from database.
        
        Args:
            feature_type: Optional filter (e.g., 'agent_api_endpoint', 'trading_api_endpoint')
        
        Returns:
            List of dicts with: id, feature_type, backend_pattern, frontend_pattern, 
                                integration_points, created_at, updated_at
        """
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                if feature_type:
                    cur.execute("""
                        SELECT 
                            id, feature_type, backend_pattern, frontend_pattern,
                            integration_points, created_at, updated_at
                        FROM architecture_rules
                        WHERE feature_type = %s
                        ORDER BY feature_type
                    """, (feature_type,))
                else:
                    cur.execute("""
                        SELECT 
                            id, feature_type, backend_pattern, frontend_pattern,
                            integration_points, created_at, updated_at
                        FROM architecture_rules
                        ORDER BY feature_type
                    """)
                
                rules = cur.fetchall()
                return [dict(rule) for rule in rules]
        
        except Exception as e:
            logger.error(f"Error querying architecture rules: {e}")
            return []
    
    def execute_with_solace(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute task using OpenAI (SOLACE)."""
        if not self.openai_client:
            raise Exception("OpenAI client not initialized")
        
        prompt = f"""You are SOLACE, the strategic coordinator of the ARES agent swarm.

Task Type: {task['task_type']}
Description: {task['description']}
Priority: {task['priority']}/10
Files: {task.get('file_paths', [])}

Your role is to:
1. Analyze the task
2. Decide if you should handle it OR delegate to:
   - FORGE (UI building, React, HTML, CSS)
   - ARCHITECT (Planning, architecture, design)
   - SENTINEL (Testing, debugging, validation)
3. If delegating, explain why
4. Provide strategic guidance

Respond in JSON format:
{{
    "decision": "handle|delegate",
    "delegate_to": "FORGE|ARCHITECT|SENTINEL|null",
    "reasoning": "explanation",
    "guidance": "strategic advice",
    "estimated_complexity": "low|medium|high"
}}"""
        
        # Define available tools for SOLACE
        tools = [
            {
                "type": "function",
                "function": {
                    "name": "read_file",
                    "description": "Read contents of a file from the repository",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "file_path": {
                                "type": "string",
                                "description": "Relative path to file (e.g., internal/api/controllers/agent_controller.go)"
                            }
                        },
                        "required": ["file_path"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "list_directory",
                    "description": "List all files in a directory",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "dir_path": {
                                "type": "string",
                                "description": "Relative directory path (e.g., internal/api/controllers)"
                            }
                        },
                        "required": ["dir_path"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "get_architecture_rules",
                    "description": "Get all architecture patterns that define where different feature types should be implemented",
                    "parameters": {
                        "type": "object",
                        "properties": {},
                        "required": []
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "search_architecture_rules",
                    "description": "Search architecture rules by keyword",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "keyword": {
                                "type": "string",
                                "description": "Keyword to search for (e.g., 'agent', 'trading', 'health')"
                            }
                        },
                        "required": ["keyword"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "create_user_request",
                    "description": "Create a new user request for orchestration (breaks down into GitHub instructions)",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "request_text": {
                                "type": "string",
                                "description": "User's request (e.g., 'integrate agent dashboard into main UI')"
                            },
                            "request_type": {
                                "type": "string",
                                "description": "Type of request: feature_integration, bug_fix, refactor, new_feature"
                            }
                        },
                        "required": ["request_text"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "analyze_request",
                    "description": "Analyze a user request to determine which architecture rules apply and which files will be affected",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "request_id": {
                                "type": "integer",
                                "description": "ID of the user request to analyze"
                            }
                        },
                        "required": ["request_id"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "generate_instructions",
                    "description": "Generate atomic GitHub instructions for a user request",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "request_id": {
                                "type": "integer",
                                "description": "ID of the analyzed user request"
                            }
                        },
                        "required": ["request_id"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "get_next_instruction",
                    "description": "Get the next pending GitHub instruction to execute",
                    "parameters": {
                        "type": "object",
                        "properties": {},
                        "required": []
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "scan_repository",
                    "description": "Scan the entire repository and cache all file information",
                    "parameters": {
                        "type": "object",
                        "properties": {},
                        "required": []
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "write_file",
                    "description": "Write or modify a file in the repository",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "file_path": {
                                "type": "string",
                                "description": "Relative path to file to write"
                            },
                            "content": {
                                "type": "string",
                                "description": "Complete file content to write"
                            }
                        },
                        "required": ["file_path", "content"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "create_backup",
                    "description": "Create timestamped backup of entire ARES workspace before making changes",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "reason": {
                                "type": "string",
                                "description": "Reason for backup (e.g., 'before integrating agent dashboard')"
                            }
                        },
                        "required": ["reason"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "execute_command",
                    "description": "Execute a shell command (build, test, run) and return output",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "command": {
                                "type": "string",
                                "description": "Command to execute (e.g., 'go build -o ares_api.exe .\\cmd\\main.go')"
                            },
                            "working_directory": {
                                "type": "string",
                                "description": "Directory to run command in (default: C:\\ARES_Workspace\\ARES_API)"
                            }
                        },
                        "required": ["command"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "restore_from_backup",
                    "description": "Restore entire workspace from a timestamped backup",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "backup_timestamp": {
                                "type": "string",
                                "description": "Timestamp of backup to restore (e.g., '2025-10-17_153045')"
                            }
                        },
                        "required": ["backup_timestamp"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "verify_system_running",
                    "description": "Check if ARES system is running and responding correctly",
                    "parameters": {
                        "type": "object",
                        "properties": {},
                        "required": []
                    }
                }
            }
        ]
        
        response = self.openai_client.chat.completions.create(
            model="gpt-4",
            messages=[
                {"role": "system", "content": "You are SOLACE, strategic AI coordinator with repository orchestration capabilities."},
                {"role": "user", "content": prompt}
            ],
            tools=tools,
            tool_choice="auto",
            temperature=0.7,
            max_tokens=1000
        )
        
        # Handle function calls from OpenAI
        if response.choices[0].message.tool_calls:
            for tool_call in response.choices[0].message.tool_calls:
                function_name = tool_call.function.name
                function_args = json.loads(tool_call.function.arguments)
                
                # Execute the function by calling ARES API
                result = self._execute_tool(function_name, function_args)
                
                # Log the tool execution
                print(f"🔧 SOLACE executed tool: {function_name}")
                print(f"   Args: {function_args}")
                print(f"   Result: {result}")
        
        result_text = response.choices[0].message.content
        tokens_used = response.usage.total_tokens
        
        # Try to parse JSON response
        try:
            result = json.loads(result_text)
        except:
            result = {
                "decision": "handle",
                "reasoning": result_text,
                "guidance": "Processed by SOLACE"
            }
        
        result['tokens_used'] = tokens_used
        return result
    
    def _execute_tool(self, function_name: str, args: dict) -> dict:
        """Execute a SOLACE orchestration tool by calling ARES API"""
        base_url = "http://localhost:8080/api/v1/solace-orch"
        
        # Map function names to API endpoints
        endpoint_map = {
            "read_file": f"{base_url}/file/read",
            "list_directory": f"{base_url}/file/list",
            "get_architecture_rules": f"{base_url}/architecture/rules",
            "search_architecture_rules": f"{base_url}/architecture/search",
            "create_user_request": f"{base_url}/orchestrate/request",
            "analyze_request": f"{base_url}/orchestrate/analyze/{args.get('request_id', '')}",
            "generate_instructions": f"{base_url}/orchestrate/generate/{args.get('request_id', '')}",
            "get_next_instruction": f"{base_url}/orchestrate/next",
            "scan_repository": f"{base_url}/orchestrate/scan",
            "write_file": f"{base_url}/file/write",
            "create_backup": f"{base_url}/backup/create",
            "execute_command": f"{base_url}/command/execute",
            "restore_from_backup": f"{base_url}/backup/restore",
            "verify_system_running": f"{base_url}/system/verify"
        }
        
        endpoint = endpoint_map.get(function_name)
        if not endpoint:
            return {"error": f"Unknown function: {function_name}"}
        
        try:
            # Determine HTTP method
            if function_name in ["get_architecture_rules", "get_next_instruction", "verify_system_running"]:
                response = requests.get(endpoint)
            else:
                # Remove request_id from args if it's in the URL
                payload = {k: v for k, v in args.items() if k != 'request_id'}
                response = requests.post(endpoint, json=payload)
            
            if response.status_code == 200:
                return response.json()
            else:
                return {"error": f"API call failed: {response.status_code}", "details": response.text}
        
        except Exception as e:
            return {"error": f"Tool execution failed: {str(e)}"}
    
    def execute_with_forge(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute task using Claude (FORGE)."""
        if not self.anthropic_client:
            raise Exception("Anthropic client not initialized")
        
        prompt = f"""You are FORGE, the UI builder of the ARES agent swarm.

Task: {task['description']}
Type: {task['task_type']}
Files: {task.get('file_paths', [])}

Build the requested UI component or code. Follow these guidelines:
- Use modern, clean design
- Match the existing ARES purple theme (#8B5CF6)
- Ensure responsiveness
- Include error handling
- Add helpful comments

Provide your implementation and explain what you built."""
        
        message = self.anthropic_client.messages.create(
            model="claude-3-7-sonnet-20250219",  # Claude 3.5 deprecated, using 3.7
            max_tokens=4000,
            messages=[
                {"role": "user", "content": prompt}
            ]
        )
        
        result_text = message.content[0].text
        tokens_used = message.usage.input_tokens + message.usage.output_tokens
        
        return {
            "implementation": result_text,
            "agent": "FORGE",
            "tokens_used": tokens_used
        }
    
    def execute_with_architect(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute task using Ollama DeepSeek-R1 (ARCHITECT)."""
        import requests
        
        prompt = f"""You are ARCHITECT, the planning and design expert of the ARES agent swarm.

Task: {task['description']}
Type: {task['task_type']}

Create a detailed architectural plan. Include:
- System design overview
- Component breakdown
- Data flow
- Design patterns to use
- Potential challenges
- Testing strategy

Be thorough and strategic."""
        
        resp = requests.post(
            f"{self.ollama_base_url}/api/generate",
            json={
                "model": "deepseek-r1:14b",
                "prompt": prompt,
                "stream": False
            },
            timeout=120
        )
        
        if resp.status_code != 200:
            raise Exception(f"Ollama error: {resp.status_code}")
        
        result = resp.json()
        return {
            "plan": result.get('response', ''),
            "agent": "ARCHITECT"
        }
    
    def execute_with_sentinel(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute task using Ollama DeepSeek-R1 (SENTINEL)."""
        import requests
        
        prompt = f"""You are SENTINEL, the testing and debugging expert of the ARES agent swarm.

Task: {task['description']}
Type: {task['task_type']}
Files: {task.get('file_paths', [])}

Your mission:
- Identify potential bugs
- Write test cases
- Validate logic
- Check edge cases
- Suggest improvements

Provide a comprehensive quality assessment."""
        
        resp = requests.post(
            f"{self.ollama_base_url}/api/generate",
            json={
                "model": "deepseek-r1:14b",
                "prompt": prompt,
                "stream": False
            },
            timeout=120
        )
        
        if resp.status_code != 200:
            raise Exception(f"Ollama error: {resp.status_code}")
        
        result = resp.json()
        return {
            "analysis": result.get('response', ''),
            "agent": "SENTINEL"
        }
    
    def execute_task(self, task: Dict[str, Any]):
        """Execute a task using the appropriate agent."""
        task_id = task['task_id']
        agent_name = task['assigned_to_agent']
        task_type = task['task_type']
        
        logger.info(f"[EXEC] Executing task {task_id[:8]}... ({task_type}) with {agent_name}")
        
        # Update status to in_progress
        self.update_task_status(task_id, 'in_progress')
        self.update_agent_status(agent_name, 'busy', task_id)
        
        start_time = time.time()
        result = None
        error = None
        tokens_used = 0
        
        try:
            # Route to appropriate agent
            if agent_name == 'SOLACE':
                result = self.execute_with_solace(task)
                tokens_used = result.get('tokens_used', 0)
            elif agent_name == 'FORGE':
                result = self.execute_with_forge(task)
                tokens_used = result.get('tokens_used', 0)
            elif agent_name == 'ARCHITECT':
                result = self.execute_with_architect(task)
            elif agent_name == 'SENTINEL':
                result = self.execute_with_sentinel(task)
            else:
                raise Exception(f"Unknown agent: {agent_name}")
            
            duration_ms = int((time.time() - start_time) * 1000)
            
            # If SOLACE delegates, create new task
            if agent_name == 'SOLACE' and result.get('decision') == 'delegate':
                delegate_to = result.get('delegate_to')
                if delegate_to:
                    logger.info(f"[DELEGATE] SOLACE delegating to {delegate_to}: {result.get('reasoning')}")
                    # Create delegated task (simplified - in production, call API)
                    with self.conn.cursor() as cur:
                        cur.execute("""
                            INSERT INTO task_queue 
                            (task_type, priority, status, created_by, assigned_to_agent, description, context)
                            VALUES (%s, %s, 'assigned', 'SOLACE', %s, %s, %s)
                        """, (task_type, task['priority'], delegate_to, task['description'], 
                              json.dumps({'delegated_from': task_id, 'guidance': result.get('guidance')})))
                        self.conn.commit()
            
            # Mark task complete
            self.update_task_status(task_id, 'completed', result)
            self.update_agent_status(agent_name, 'idle', None)
            self.log_task_history(agent_name, task_id, task_type, True, duration_ms, None, tokens_used)
            
            logger.info(f"[OK] Task completed by {agent_name} in {duration_ms}ms")
            
        except Exception as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error = str(e)
            logger.error(f"[ERROR] Task failed: {error}")
            
            self.update_task_status(task_id, 'failed', None, error)
            self.update_agent_status(agent_name, 'idle', None)
            self.log_task_history(agent_name, task_id, task_type, False, duration_ms, error, tokens_used)
    
    def run(self, interval: int = 10):
        """Main coordinator loop."""
        logger.info(f"[AGENT] Agent Coordinator starting (check interval: {interval}s)")
        logger.info("Agents: SOLACE (OpenAI), FORGE (Claude), ARCHITECT (Ollama), SENTINEL (Ollama)")
        
        self.connect_db()
        
        try:
            while True:
                # Fetch pending tasks
                tasks = self.get_pending_tasks()
                
                if tasks:
                    logger.info(f"[STATS] Found {len(tasks)} pending task(s)")
                    for task in tasks:
                        self.execute_task(task)
                else:
                    logger.debug("No pending tasks")
                
                # Wait before next check
                time.sleep(interval)
                
        except KeyboardInterrupt:
            logger.info("Coordinator stopped by user")
        except Exception as e:
            logger.error(f"Coordinator error: {e}")
        finally:
            self.close_db()


# ============================================================================
# WEBSOCKET SERVER FOR SOLACE REAL-TIME COMMUNICATION
# ============================================================================

def get_db_connection():
    """Get PostgreSQL connection to ARES database."""
    return psycopg2.connect(
        host=os.getenv('DB_HOST', 'localhost'),
        port=int(os.getenv('DB_PORT', 5432)),
        database=os.getenv('DB_NAME', 'ares_db'),
        user=os.getenv('DB_USER', 'ARES'),
        password=os.getenv('DB_PASSWORD', 'ARESISWAKING')
    )


def query_architecture_rules(feature_type: Optional[str] = None) -> List[Dict[str, Any]]:
    """
    Query architecture_rules table directly from database (standalone function for WebSocket).
    
    Args:
        feature_type: Optional filter (e.g., 'agent_api_endpoint', 'trading_api_endpoint')
    
    Returns:
        List of dicts with: id, feature_type, backend_pattern, frontend_pattern, 
                            integration_points, created_at, updated_at
    """
    conn = None
    try:
        conn = get_db_connection()
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            if feature_type:
                cur.execute("""
                    SELECT 
                        id, feature_type, backend_pattern, frontend_pattern,
                        integration_points, created_at, updated_at
                    FROM architecture_rules
                    WHERE feature_type = %s
                    ORDER BY feature_type
                """, (feature_type,))
            else:
                cur.execute("""
                    SELECT 
                        id, feature_type, backend_pattern, frontend_pattern,
                        integration_points, created_at, updated_at
                    FROM architecture_rules
                    ORDER BY feature_type
                """)
            
            rules = cur.fetchall()
            # Convert datetime objects to ISO strings for JSON serialization
            for rule in rules:
                if 'created_at' in rule and rule['created_at']:
                    rule['created_at'] = rule['created_at'].isoformat()
                if 'updated_at' in rule and rule['updated_at']:
                    rule['updated_at'] = rule['updated_at'].isoformat()
            
            return [dict(rule) for rule in rules]
    
    except Exception as e:
        logger.error(f"Error querying architecture rules: {e}")
        return []
    
    finally:
        if conn:
            conn.close()


def get_openai_tools() -> List[Dict[str, Any]]:
    """
    Define function tools for OpenAI function calling.
    
    Returns:
        List of tool definitions for OpenAI API
    """
    return [
        {
            "type": "function",
            "function": {
                "name": "read_file",
                "description": "Read contents of a file from the repository",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the file to read (relative or absolute)"
                        }
                    },
                    "required": ["path"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "write_file",
                "description": "Write content to a file in the repository",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the file to write"
                        },
                        "content": {
                            "type": "string",
                            "description": "Content to write to the file"
                        }
                    },
                    "required": ["path", "content"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "list_directory",
                "description": "List contents of a directory",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Path to the directory to list"
                        },
                        "recursive": {
                            "type": "boolean",
                            "description": "Whether to list recursively (default: true)"
                        }
                    },
                    "required": ["path"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "query_architecture",
                "description": "Query architecture rules from the database",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "feature_type": {
                            "type": "string",
                            "description": "Optional feature type filter (e.g., 'agent_api_endpoint')"
                        }
                    },
                    "required": []
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "create_backup",
                "description": "Create timestamped backup of workspace before making changes. Always call this before modifying files.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "workspace_path": {
                            "type": "string",
                            "description": "Path to workspace directory to backup"
                        }
                    },
                    "required": ["workspace_path"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "restore_backup",
                "description": "Restore workspace from a previous backup. Use when changes need to be rolled back.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "backup_path": {
                            "type": "string",
                            "description": "Path to backup directory"
                        },
                        "workspace_path": {
                            "type": "string",
                            "description": "Path to workspace to restore to"
                        }
                    },
                    "required": ["backup_path", "workspace_path"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "execute_command",
                "description": "Execute PowerShell command and return output. Use for building, testing, or running system commands.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "command": {
                            "type": "string",
                            "description": "PowerShell command to execute"
                        },
                        "cwd": {
                            "type": "string",
                            "description": "Working directory for command execution"
                        }
                    },
                    "required": ["command"]
                }
            }
        }
    ]


async def handle_chat_message(message: str, websocket) -> str:
    """
    Process chat message through OpenAI with function calling.
    
    Args:
        message: User's chat message
        websocket: WebSocket connection for status updates
    
    Returns:
        Final response from OpenAI
    """
    logger.info(f"[CHAT] Received message: {message[:100]}..." if len(message) > 100 else f"[CHAT] Received message: {message}")
    
    # Get OpenAI API key from environment
    openai_api_key = os.getenv('OPENAI_API_KEY')
    if not openai_api_key:
        logger.error("[CHAT] OPENAI_API_KEY not set")
        return "Error: OPENAI_API_KEY not set in environment variables"
    
    # Initialize OpenAI client
    client = OpenAI(api_key=openai_api_key)
    logger.debug("[CHAT] OpenAI client initialized")
    
    # Prepare messages
    messages = [
        {
            "role": "system",
            "content": "You are SOLACE, an AI assistant with direct file system access. You help David build the ARES system. You can read/write files, list directories, query architecture rules, create/restore backups, and execute PowerShell commands."
        },
        {
            "role": "user",
            "content": message
        }
    ]
    
    # Get tools
    tools = get_openai_tools()
    
    try:
        # Call OpenAI with function calling
        max_iterations = 5
        iteration = 0
        
        logger.info(f"[CHAT] Starting function calling loop (max {max_iterations} iterations)")
        
        while iteration < max_iterations:
            iteration += 1
            logger.debug(f"[CHAT] Iteration {iteration}/{max_iterations}")
            
            response = client.chat.completions.create(
                model="gpt-4-turbo-preview",
                messages=messages,
                tools=tools,
                tool_choice="auto"
            )
            
            response_message = response.choices[0].message
            
            # Check if OpenAI wants to call functions
            if response_message.tool_calls:
                logger.info(f"[CHAT] OpenAI requesting {len(response_message.tool_calls)} function call(s)")
                # Add assistant's response to messages
                messages.append(response_message)
                
                # Execute each function call
                for tool_call in response_message.tool_calls:
                    function_name = tool_call.function.name
                    function_args = json.loads(tool_call.function.arguments)
                    
                    # Send status update
                    status_msg = json.dumps({
                        "type": "status",
                        "message": f"Executing {function_name}..."
                    })
                    await websocket.send(status_msg)
                    
                    logger.info(f"[CHAT] Executing {function_name} with args: {function_args}")
                    
                    # Execute the function
                    try:
                        if function_name == "read_file":
                            result = read_file(function_args['path'])
                            function_response = f"File content:\n{result}"
                        
                        elif function_name == "write_file":
                            write_file(function_args['path'], function_args['content'])
                            function_response = f"Successfully wrote to {function_args['path']}"
                        
                        elif function_name == "list_directory":
                            recursive = function_args.get('recursive', True)
                            files = list_directory(function_args['path'], recursive=recursive)
                            function_response = json.dumps(files, indent=2)
                        
                        elif function_name == "query_architecture":
                            feature_type = function_args.get('feature_type')
                            rules = query_architecture_rules(feature_type)
                            function_response = json.dumps(rules, indent=2)
                        
                        elif function_name == "create_backup":
                            workspace_path = function_args.get("workspace_path")
                            from file_operations import create_backup
                            backup_path = create_backup(workspace_path)
                            function_response = json.dumps({"backup_path": backup_path, "success": True}, indent=2)
                        
                        elif function_name == "restore_backup":
                            backup_path = function_args.get("backup_path")
                            workspace_path = function_args.get("workspace_path")
                            from file_operations import restore_backup
                            restore_backup(backup_path, workspace_path)
                            function_response = json.dumps({"success": True, "message": "Backup restored successfully"}, indent=2)
                        
                        elif function_name == "execute_command":
                            command = function_args.get("command")
                            cwd = function_args.get("cwd", ".")
                            import subprocess
                            result = subprocess.run(
                                ["powershell", "-Command", command],
                                cwd=cwd,
                                capture_output=True,
                                text=True,
                                timeout=300
                            )
                            function_response = json.dumps({
                                "stdout": result.stdout,
                                "stderr": result.stderr,
                                "exit_code": result.returncode
                            }, indent=2)
                        
                        else:
                            function_response = f"Unknown function: {function_name}"
                    
                    except Exception as e:
                        function_response = f"Error executing {function_name}: {str(e)}"
                        logger.error(f"[CHAT] Function execution error: {e}", exc_info=True)
                    
                    logger.info(f"[CHAT] Function {function_name} executed successfully")
                    logger.debug(f"[CHAT] Function result: {function_response[:200]}..." if len(function_response) > 200 else f"[CHAT] Function result: {function_response}")
                    
                    # Add function response to messages
                    messages.append({
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": function_response
                    })
                
                # Continue loop to get final response from OpenAI
                continue
            
            else:
                # No more function calls, return final response
                logger.info("[CHAT] OpenAI completed, returning final response")
                return response_message.content
        
        logger.warning(f"[CHAT] Maximum iterations ({max_iterations}) reached")
        return "Maximum iterations reached. Please try a simpler request."
    
    except Exception as e:
        logger.error(f"[CHAT] OpenAI error: {e}")
        return f"Error: {str(e)}"


async def handle_websocket(websocket):
    """
    Handle WebSocket connections for SOLACE real-time communication.
    
    Supports message types:
    - ping: Health check
    - read_file: Read file contents
    - write_file: Write file contents
    - list_directory: List directory contents
    - chat: Chat with SOLACE (OpenAI function calling integration)
    - get_architecture: Query architecture rules from database
    - create_backup: Create timestamped workspace backup
    - restore_backup: Restore workspace from backup
    - execute_command: Execute PowerShell commands
    """
    logger.info(f"[WEBSOCKET] New connection from {websocket.remote_address}")
    
    try:
        async for message in websocket:
            try:
                # Parse incoming JSON message
                data = json.loads(message)
                msg_type = data.get('type')
                msg_data = data.get('data', {})
                
                logger.info(f"[WEBSOCKET] Received message type: {msg_type}")
                logger.debug(f"[WEBSOCKET] Message data: {msg_data}")
                
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
                    response = {'type': 'directory_listing', 'path': dir_path, 'files': files}
                
                elif msg_type == 'chat':
                    user_message = msg_data.get('message', '')
                    logger.info(f"[WEBSOCKET] Processing chat message: {user_message[:50]}...")
                    
                    # Use OpenAI function calling to process the message
                    ai_response = await handle_chat_message(user_message, websocket)
                    
                    response = {
                        'type': 'chat_response',
                        'message': ai_response
                    }
                
                elif msg_type == 'get_architecture':
                    feature_type = msg_data.get('feature_type')
                    rules = query_architecture_rules(feature_type)
                    response = {
                        'type': 'architecture_rules',
                        'rules': rules,
                        'count': len(rules),
                        'filter': feature_type
                    }
                
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
                    logger.warning(f"[WEBSOCKET] Unknown message type: {msg_type}")
                    response = {'type': 'error', 'message': f'Unknown message type: {msg_type}'}
                
                # Send response
                await websocket.send(json.dumps(response))
                logger.info(f"[WEBSOCKET] ✓ Sent response for {msg_type}")
                
            except json.JSONDecodeError as e:
                logger.error(f"[WEBSOCKET] JSON decode error: {e}")
                error_response = {'type': 'error', 'message': f'Invalid JSON: {str(e)}'}
                await websocket.send(json.dumps(error_response))
            
            except FileNotFoundError as e:
                logger.error(f"[WEBSOCKET] File not found: {e}")
                error_response = {'type': 'error', 'message': f'File not found: {str(e)}'}
                await websocket.send(json.dumps(error_response))
            
            except Exception as e:
                logger.error(f"[WEBSOCKET] Error handling message: {e}", exc_info=True)
                error_response = {'type': 'error', 'message': str(e)}
                await websocket.send(json.dumps(error_response))
    
    except websockets.exceptions.ConnectionClosed:
        logger.info(f"[WEBSOCKET] Connection closed: {websocket.remote_address}")
    
    except Exception as e:
        logger.error(f"[WEBSOCKET] Connection error: {e}")


def cleanup_existing_processes():
    """Kill existing coordinator processes to prevent port conflicts."""
    import subprocess
    logger.info("Checking for existing coordinator processes...")
    try:
        result = subprocess.run(
            ["powershell", "-Command", 
             "Get-Process python -ErrorAction SilentlyContinue | Where-Object {$_.CommandLine -like '*coordinator*'} | Stop-Process -Force"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            logger.info("✓ Cleaned up existing coordinator processes")
        else:
            logger.debug("No existing coordinator processes found")
    except Exception as e:
        logger.warning(f"Could not cleanup processes: {e}")


def cleanup_orphaned_powershells():
    """Kill orphaned PowerShell processes older than 5 minutes."""
    import subprocess
    logger.info("Checking for orphaned PowerShell processes...")
    try:
        result = subprocess.run(
            ["powershell", "-Command", 
             "$limit = (Get-Date).AddMinutes(-5); Get-Process powershell -ErrorAction SilentlyContinue | Where-Object {$_.StartTime -lt $limit} | Stop-Process -Force"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            logger.info("✓ Cleaned up orphaned PowerShell processes")
        else:
            logger.debug("No orphaned PowerShell processes found")
    except Exception as e:
        logger.warning(f"Could not cleanup PowerShell processes: {e}")


async def start_websocket_server():
    """Start the WebSocket server for SOLACE."""
    # Cleanup before starting
    cleanup_existing_processes()
    cleanup_orphaned_powershells()
    
    logger.info("=" * 60)
    logger.info("Starting SOLACE WebSocket Server")
    logger.info("=" * 60)
    
    async with websockets.serve(handle_websocket, "localhost", 8765):
        logger.info("✅ WebSocket server started on ws://localhost:8765")
        logger.info("   Available message types: ping, read_file, write_file, list_directory, chat")
        logger.info("   Press Ctrl+C to stop")
        
        # Run forever
        await asyncio.Future()


def main():
    """Entry point."""
    # Validate environment on startup (only when run directly, not when imported)
    try:
        import validate_env
        validate_env.validate_environment()
    except SystemExit:
        # Re-raise exit from validation failure
        raise
    except ImportError:
        logger.warning("validate_env.py not found - skipping environment validation")
    except Exception as e:
        logger.warning(f"Environment validation error: {e}")
    
    parser = argparse.ArgumentParser(description='ARES Agent Swarm Coordinator')
    parser.add_argument('--interval', type=int, default=10, help='Task check interval in seconds')
    parser.add_argument('--debug', action='store_true', help='Enable debug logging')
    parser.add_argument('--websocket', action='store_true', help='Start WebSocket server instead of coordinator')
    args = parser.parse_args()
    
    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)
    
    # WebSocket mode
    if args.websocket:
        try:
            asyncio.run(start_websocket_server())
        except KeyboardInterrupt:
            logger.info("[WEBSOCKET] Server stopped by user")
        except Exception as e:
            logger.error(f"[WEBSOCKET] Server error: {e}")
            import traceback
            traceback.print_exc()
            sys.exit(1)
        return
    
    # Coordinator mode (default)
    # Database configuration
    db_config = {
        'host': os.getenv('DB_HOST', 'localhost'),
        'port': int(os.getenv('DB_PORT', 5432)),
        'database': os.getenv('DB_NAME', 'ares_db'),
        'user': os.getenv('DB_USER', 'ARES'),
        'password': os.getenv('DB_PASSWORD', 'ARESISWAKING')
    }
    
    # Create and run coordinator
    try:
        coordinator = AgentCoordinator(db_config)
        coordinator.run(interval=args.interval)
    except Exception as e:
        logger.error(f"[ERROR] Coordinator crashed: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()

