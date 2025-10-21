# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
import psycopg2

conn = psycopg2.connect('host=localhost dbname=ares_db user=ARES password=ARESISWAKING')
cur = conn.cursor()
cur.execute("UPDATE task_queue SET status='assigned', assigned_to_agent='FORGE' WHERE task_id='8a3251a8-7092-4f5f-bcea-b4fa54ce169e' AND status='failed'")
conn.commit()
print(f'✓ Task reset for FORGE retry with Claude 3.7 Sonnet')
conn.close()
