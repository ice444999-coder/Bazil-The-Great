# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
"""
Test all agent connections before starting coordinator
"""
import os
import sys
from dotenv import load_dotenv
from openai import OpenAI
from anthropic import Anthropic
import requests
import psycopg2

load_dotenv()

def test_openai():
    """Test SOLACE connection"""
    print("\n🧪 Testing OpenAI (SOLACE)...")
    try:
        api_key = os.getenv('OPENAI_API_KEY')
        if not api_key:
            print("❌ OPENAI_API_KEY not set")
            return False
        
        client = OpenAI(api_key=api_key)
        response = client.chat.completions.create(
            model='gpt-4-turbo-preview',
            messages=[{'role': 'user', 'content': 'Say "SOLACE online"'}],
            max_tokens=10
        )
        result = response.choices[0].message.content
        print(f"✅ OpenAI connected: {result}")
        return True
    except Exception as e:
        print(f"❌ OpenAI failed: {e}")
        return False

def test_claude():
    """Test FORGE connection"""
    print("\n🧪 Testing Claude (FORGE)...")
    try:
        api_key = os.getenv('CLAUDE_API_KEY')
        if not api_key:
            print("❌ CLAUDE_API_KEY not set")
            return False
        
        client = Anthropic(api_key=api_key)
        message = client.messages.create(
            model='claude-3-5-sonnet-20241022',
            max_tokens=10,
            messages=[{'role': 'user', 'content': 'Say "FORGE online"'}]
        )
        result = message.content[0].text
        print(f"✅ Claude connected: {result}")
        return True
    except Exception as e:
        print(f"❌ Claude failed: {e}")
        return False

def test_deepseek():
    """Test ARCHITECT & SENTINEL connection"""
    print("\n🧪 Testing DeepSeek (ARCHITECT & SENTINEL)...")
    try:
        response = requests.post('http://localhost:11434/api/generate', json={
            'model': 'deepseek-r1:14b',
            'prompt': 'Say "ARCHITECT and SENTINEL online"',
            'stream': False
        }, timeout=30)
        
        if response.status_code == 200:
            result = response.json().get('response', '')
            print(f"✅ DeepSeek connected: {result[:50]}...")
            return True
        else:
            print(f"❌ DeepSeek failed: HTTP {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ DeepSeek failed: {e}")
        print("   Make sure Ollama is running: ollama serve")
        return False

def test_database():
    """Test PostgreSQL connection"""
    print("\n🧪 Testing PostgreSQL...")
    try:
        conn = psycopg2.connect(
            host=os.getenv('DB_HOST', 'localhost'),
            port=os.getenv('DB_PORT', '5432'),
            user=os.getenv('DB_USER'),
            password=os.getenv('DB_PASSWORD'),
            database=os.getenv('DB_NAME')
        )
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM agent_registry")
        count = cursor.fetchone()[0]
        cursor.close()
        conn.close()
        print(f"✅ PostgreSQL connected: {count} agents registered")
        return True
    except Exception as e:
        print(f"❌ PostgreSQL failed: {e}")
        return False

def test_api_endpoints():
    """Test ARES backend API"""
    print("\n🧪 Testing ARES API...")
    try:
        response = requests.get('http://localhost:8080/api/v1/agents', timeout=5)
        if response.status_code == 200:
            agents = response.json()
            print(f"✅ ARES API connected: {len(agents)} agents available")
            return True
        else:
            print(f"❌ ARES API failed: HTTP {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ ARES API failed: {e}")
        print("   Make sure ARES backend is running")
        return False

if __name__ == "__main__":
    print("=" * 60)
    print("ARES AGENT SWARM - CONNECTION TEST")
    print("=" * 60)
    
    results = {
        'OpenAI (SOLACE)': test_openai(),
        'Claude (FORGE)': test_claude(),
        'DeepSeek (ARCHITECT/SENTINEL)': test_deepseek(),
        'PostgreSQL': test_database(),
        'ARES API': test_api_endpoints()
    }
    
    print("\n" + "=" * 60)
    print("TEST RESULTS:")
    print("=" * 60)
    for name, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        print(f"{status} - {name}")
    
    print("=" * 60)
    
    if all(results.values()):
        print("\n🎉 ALL TESTS PASSED - Agent swarm ready to start!")
        print("\nNext step: python internal/agent_swarm/coordinator.py")
        sys.exit(0)
    else:
        print("\n❌ SOME TESTS FAILED - Fix issues before starting coordinator")
        sys.exit(1)
