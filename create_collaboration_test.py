"""
Create comprehensive agent swarm collaboration test task.
This tests SOLACE → FORGE → ARCHITECT → SENTINEL working together.
"""
import sys
sys.path.insert(0, 'C:\\ARES_Workspace\\ARES_API\\Lib\\site-packages')

import psycopg2
import json
from datetime import datetime

def create_collaboration_test():
    conn = psycopg2.connect('host=localhost dbname=ares_db user=ARES password=ARESISWAKING')
    cur = conn.cursor()
    
    # Create a complex task that requires all agents
    task_description = """AGENT SWARM COLLABORATION TEST - Multi-Agent Workflow

OBJECTIVE: Test all 4 agents working together on a realistic trading dashboard feature.

WORKFLOW:
1. SOLACE (Director) - Analyze requirements and delegate tasks
2. ARCHITECT - Design the system architecture and data flow
3. FORGE - Implement the UI components
4. SENTINEL - Test and validate the implementation

FEATURE TO BUILD:
"Real-time Trading Performance Dashboard Card"

REQUIREMENTS:
- Display: Total P&L, Win Rate, Best Trade, Worst Trade
- Real-time updates every 5 seconds
- Color coding: Green for profit, Red for loss
- Responsive design (mobile + desktop)
- Data fetched from /api/trades/stats endpoint

SUCCESS CRITERIA:
✓ SOLACE delegates to appropriate agents
✓ ARCHITECT creates detailed design doc
✓ FORGE implements working component
✓ SENTINEL validates and reports test results
✓ All agents log their reasoning and decisions

COLLABORATION POINTS:
- SOLACE should recognize this needs architecture + implementation + testing
- ARCHITECT should specify API contract and component structure
- FORGE should reference ARCHITECT's design
- SENTINEL should validate against ARCHITECT's specs

This tests:
- Multi-agent delegation
- Knowledge sharing between agents
- Sequential workflow execution
- Decision logging to PostgreSQL
- Real agent collaboration (not just individual tasks)
"""

    cur.execute("""
        INSERT INTO task_queue (
            task_id, task_type, description, priority, status, 
            context, file_paths, created_at, assigned_to_agent
        ) VALUES (
            gen_random_uuid(), 
            'feature_implementation',
            %s,
            10,
            'assigned',
            %s,
            %s,
            %s,
            'SOLACE'
        ) RETURNING task_id
    """, (
        task_description,
        json.dumps({
            "test_type": "multi_agent_collaboration",
            "requires_agents": ["SOLACE", "ARCHITECT", "FORGE", "SENTINEL"],
            "expected_flow": "SOLACE→ARCHITECT→FORGE→SENTINEL",
            "feature": "TradingPerformanceCard",
            "complexity": "high",
            "test_collaboration": True
        }),
        json.dumps(["frontend/src/components/TradingPerformanceCard.tsx"]),
        datetime.now()
    ))
    
    task_id = cur.fetchone()[0]
    conn.commit()
    
    print(f"\n{'='*70}")
    print(f"✓ AGENT SWARM COLLABORATION TEST CREATED")
    print(f"{'='*70}")
    print(f"Task ID: {task_id}")
    print(f"Assigned to: SOLACE (will delegate)")
    print(f"Expected Flow: SOLACE → ARCHITECT → FORGE → SENTINEL")
    print(f"\nThis will test:")
    print(f"  1. SOLACE's delegation logic")
    print(f"  2. ARCHITECT's design capabilities")
    print(f"  3. FORGE's implementation skills")
    print(f"  4. SENTINEL's testing abilities")
    print(f"  5. Multi-agent knowledge sharing")
    print(f"\nWatch the coordinator log to see agents collaborate!")
    print(f"{'='*70}\n")
    
    conn.close()
    return task_id

if __name__ == '__main__':
    create_collaboration_test()
