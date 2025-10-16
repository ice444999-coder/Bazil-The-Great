"""
End-to-end test: Create task, watch agents collaborate, verify result
"""
import requests
import time
import json

BASE_URL = 'http://localhost:8080/api/v1/agents'

def create_test_task():
    """Create a simple test task"""
    print("\n📋 Creating test task for SOLACE...")
    
    response = requests.post(f'{BASE_URL}/tasks', json={
        'task_type': 'test_collaboration',
        'description': '''Test agent collaboration:
        1. ARCHITECT: Plan a simple "Hello World" component
        2. FORGE: Build the component
        3. SENTINEL: Verify it works
        4. SOLACE: Approve and report back''',
        'priority': 10,
        'context': {'test': True}
    })
    
    if response.status_code != 200:
        print(f"❌ Failed to create task: {response.text}")
        return None
    
    task = response.json()
    task_id = task.get('task_id') or task.get('id')
    print(f"✅ Task created: {task_id}")
    return task_id

def watch_task(task_id, max_wait=300):
    """Watch task progress"""
    print(f"\n👀 Watching task {task_id}...")
    print("   (This may take 2-5 minutes as agents collaborate)\n")
    
    start_time = time.time()
    last_status = None
    
    while time.time() - start_time < max_wait:
        response = requests.get(f'{BASE_URL}/tasks/{task_id}')
        if response.status_code != 200:
            print(f"❌ Failed to get task status")
            return False
        
        task = response.json()
        status = task.get('status', 'unknown')
        
        if status != last_status:
            elapsed = int(time.time() - start_time)
            print(f"[{elapsed}s] Status: {status}")
            
            if 'assigned_to_agent' in task and task['assigned_to_agent']:
                print(f"      Assigned to: {task['assigned_to_agent']}")
            
            last_status = status
        
        if status == 'completed':
            print(f"\n✅ Task completed!")
            if 'result' in task and task['result']:
                print(f"\nResult:")
                print(json.dumps(task['result'], indent=2))
            return True
        
        if status == 'failed':
            print(f"\n❌ Task failed!")
            if 'error_log' in task:
                print(f"Error: {task['error_log']}")
            return False
        
        time.sleep(5)
    
    print(f"\n⏰ Timeout after {max_wait}s")
    return False

def check_agent_status():
    """Check all agents are registered"""
    print("\n🤖 Checking agent status...")
    response = requests.get(BASE_URL)
    
    if response.status_code != 200:
        print("❌ Failed to get agents")
        return False
    
    agents = response.json()
    expected = ['SOLACE', 'FORGE', 'ARCHITECT', 'SENTINEL']
    
    for agent_name in expected:
        agent = next((a for a in agents if a.get('agent_name') == agent_name), None)
        if agent:
            status = agent.get('status', 'unknown')
            print(f"   {agent_name}: {status}")
        else:
            print(f"   {agent_name}: ❌ NOT FOUND")
            return False
    
    return True

if __name__ == "__main__":
    print("=" * 60)
    print("ARES AGENT SWARM - END-TO-END TEST")
    print("=" * 60)
    
    # Check agents registered
    if not check_agent_status():
        print("\n❌ Agents not properly registered")
        exit(1)
    
    # Create test task
    task_id = create_test_task()
    if not task_id:
        exit(1)
    
    # Watch execution
    success = watch_task(task_id)
    
    print("\n" + "=" * 60)
    if success:
        print("🎉 END-TO-END TEST PASSED")
        print("=" * 60)
        print("\nAgent swarm is fully operational!")
        print("\nNext: Create real tasks via API or agent dashboard")
        exit(0)
    else:
        print("❌ END-TO-END TEST FAILED")
        print("=" * 60)
        print("\nCheck coordinator logs for details")
        exit(1)
