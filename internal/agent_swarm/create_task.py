#!/usr/bin/env python3
"""
Task Creator - Properly handles SQL with Python
No more SQL syntax errors in PowerShell!
"""
import sys
import os

# Add package path
sys.path.insert(0, r'C:\ARES_Workspace\ARES_API\Lib\site-packages')

import psycopg2
import json

def create_task(task_type, description, priority, context, file_paths=None):
    """Create task with proper SQL parameter binding"""
    
    conn = psycopg2.connect(
        host='localhost',
        database='ares_db',
        user='ARES',
        password='ARESISWAKING'
    )
    
    try:
        with conn.cursor() as cur:
            # Use parameterized query - NO SQL INJECTION, NO SYNTAX ERRORS
            cur.execute("""
                INSERT INTO task_queue (task_type, description, priority, context, file_paths)
                VALUES (%s, %s, %s, %s, %s)
                RETURNING task_id, created_at
            """, (
                task_type,
                description,
                priority,
                json.dumps(context) if isinstance(context, dict) else context,
                json.dumps(file_paths) if file_paths else None
            ))
            
            task_id, created_at = cur.fetchone()
            conn.commit()
            
            print(f"✓ Task created: {task_id}")
            print(f"  Type: {task_type}")
            print(f"  Priority: {priority}")
            print(f"  Created: {created_at}")
            return task_id
            
    except Exception as e:
        conn.rollback()
        print(f"✗ Error: {e}", file=sys.stderr)
        sys.exit(1)
    finally:
        conn.close()

if __name__ == "__main__":
    # Code Quality IMPLEMENTATION Task - ACTUAL FIXES
    create_task(
        task_type="code_implementation",
        description="""IMPLEMENTATION: Actually FIX all 290 warnings NOW. Not planning - EXECUTE.

FILES TO FIX (DO NOT JUST PLAN - ACTUALLY MODIFY FILES):

1. frontend/src/components/AdvancedOrderForm.tsx
   - Remove ALL 40+ inline style={{}} attributes
   - Create AdvancedOrderForm.module.css
   - Move ALL styles to CSS module
   - Import and use className= instead of style=
   - TEST: Component looks identical visually

2. frontend/src/components/OpenPositionsTable.tsx
   - Remove ALL 5+ inline styles
   - Create OpenPositionsTable.module.css
   - Move table styles to CSS module
   - TEST: Table renders identically

3. web/code-ide.html
   - Remove ALL 7+ inline style="" attributes
   - Move to <style> block in <head>
   - Fix field-sizing (use height: auto; resize: vertical;)
   - TEST: Page works identically

EXECUTION REQUIREMENTS:
- FORGE must MODIFY the actual files (not create plans)
- Use file_registry table to track changes
- Run ESLint after each file: npm run lint
- VERIFY: 0 warnings, 0 errors before marking complete
- If ANY warnings remain, task FAILS

SUCCESS = get_errors() returns ZERO warnings
FAILURE = ANY warnings still exist

This is IMPLEMENTATION not PLANNING. Write actual code to files NOW.""",
        priority=10,
        context={
            "action": "IMPLEMENT_NOW",
            "verify_zero_warnings": True,
            "auto_test": True,
            "strict_mode": True,
            "must_modify_files": True
        },
        file_paths=[
            "frontend/src/components/AdvancedOrderForm.tsx",
            "frontend/src/components/OpenPositionsTable.tsx", 
            "web/code-ide.html"
        ]
    )
