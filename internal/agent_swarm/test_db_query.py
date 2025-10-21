# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
#!/usr/bin/env python3
"""
Test Architecture Rules Database Query
Tests direct SQL query to architecture_rules table.
"""

import sys
import os
from pathlib import Path

# Add agent_swarm directory to path
sys.path.insert(0, str(Path(__file__).parent))

# Load environment variables
from dotenv import load_dotenv
load_dotenv()

# Import the query function
from coordinator import query_architecture_rules

def main():
    print("=" * 70)
    print("🗄️  Testing Architecture Rules Database Query")
    print("=" * 70)
    print()
    
    # Test 1: Get all rules
    print("[TEST 1] Querying all architecture rules...")
    try:
        rules = query_architecture_rules()
        print(f"✅ Found {len(rules)} total rules")
        print()
        
        if rules:
            print("Sample rules:")
            for i, rule in enumerate(rules[:3], 1):  # Show first 3
                print(f"\n{i}. Feature Type: {rule['feature_type']}")
                print(f"   Backend Pattern: {rule['backend_pattern']}")
                print(f"   Frontend Pattern: {rule['frontend_pattern']}")
                print(f"   Integration Points: {rule['integration_points']}")
        else:
            print("⚠️  No rules found in database")
    
    except Exception as e:
        print(f"❌ ERROR: {e}")
        import traceback
        traceback.print_exc()
        return
    
    print()
    print("-" * 70)
    print()
    
    # Test 2: Get specific feature type
    if rules:
        first_feature_type = rules[0]['feature_type']
        print(f"[TEST 2] Querying specific feature type: '{first_feature_type}'...")
        try:
            filtered_rules = query_architecture_rules(first_feature_type)
            print(f"✅ Found {len(filtered_rules)} rule(s) for '{first_feature_type}'")
            
            if filtered_rules:
                rule = filtered_rules[0]
                print(f"\nRule details:")
                print(f"  ID: {rule['id']}")
                print(f"  Feature Type: {rule['feature_type']}")
                print(f"  Backend: {rule['backend_pattern']}")
                print(f"  Frontend: {rule['frontend_pattern']}")
        
        except Exception as e:
            print(f"❌ ERROR: {e}")
    
    print()
    print("=" * 70)
    print("✅ Database query tests complete!")
    print("=" * 70)


if __name__ == '__main__':
    main()
