import psycopg2
import json

conn = psycopg2.connect('host=localhost dbname=ares_db user=ARES password=ARESISWAKING')
cur = conn.cursor()
cur.execute("SELECT task_id, task_type, status, result FROM task_queue WHERE task_id='8a3251a8-7092-4f5f-bcea-b4fa54ce169e'")
result = cur.fetchone()
print(f'\nTask ID: {result[0][:8]}...')
print(f'Type: {result[1]}')
print(f'Status: {result[2]}')
print(f'\nFORGE Result:')
if result[3]:
    result_str = str(result[3])
    print(result_str[:1500] if len(result_str) > 1500 else result_str)
else:
    print('No result')
conn.close()
