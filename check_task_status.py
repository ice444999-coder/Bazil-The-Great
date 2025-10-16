#!/usr/bin/env python3
import sys
sys.path.insert(0, r'C:\ARES_Workspace\ARES_API\Lib\site-packages')
import psycopg2

conn = psycopg2.connect(
    host='localhost',
    database='ares_db',
    user='ARES',
    password='ARESISWAKING'
)

cur = conn.cursor()

print("\n=== TASK QUEUE STATUS ===\n")

# Count by status
cur.execute("SELECT status, COUNT(*) FROM task_queue GROUP BY status ORDER BY status")
for row in cur.fetchall():
    print(f"{row[0]:15} {row[1]:3} tasks")

# Recent completions
print("\n=== RECENTLY COMPLETED (Last 10 min) ===\n")
cur.execute("""
    SELECT task_id, assigned_to_agent, 
           EXTRACT(EPOCH FROM (completed_at - started_at)) as duration_secs
    FROM task_queue
    WHERE status = 'completed' 
    AND completed_at > NOW() - INTERVAL '10 minutes'
    ORDER BY completed_at DESC
    LIMIT 5
""")
for row in cur.fetchall():
    print(f"{row[0][:8]}... by {row[1]:10} in {row[2]:.1f}s")

# Pending/assigned
print("\n=== WAITING FOR EXECUTION ===\n")
cur.execute("""
    SELECT task_id, status, assigned_to_agent
    FROM task_queue
    WHERE status IN ('pending', 'assigned')
    ORDER BY created_at DESC
    LIMIT 5
""")
for row in cur.fetchall():
    print(f"{row[0][:8]}... {row[1]:12} {row[2] or 'unassigned'}")

conn.close()
