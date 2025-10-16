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

# Setup logging with UTF-8 encoding for Windows
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[
        logging.FileHandler('agent_coordinator.log', encoding='utf-8'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# Set console output to UTF-8 on Windows
if sys.platform == 'win32':
    import codecs
    sys.stdout = codecs.getwriter('utf-8')(sys.stdout.buffer, 'strict')
    sys.stderr = codecs.getwriter('utf-8')(sys.stderr.buffer, 'strict')


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
        
        response = self.openai_client.chat.completions.create(
            model="gpt-4",
            messages=[
                {"role": "system", "content": "You are SOLACE, strategic AI coordinator."},
                {"role": "user", "content": prompt}
            ],
            temperature=0.7,
            max_tokens=1000
        )
        
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


def main():
    """Entry point."""
    parser = argparse.ArgumentParser(description='ARES Agent Swarm Coordinator')
    parser.add_argument('--interval', type=int, default=10, help='Task check interval in seconds')
    parser.add_argument('--debug', action='store_true', help='Enable debug logging')
    args = parser.parse_args()
    
    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)
    
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

