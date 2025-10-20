#!/usr/bin/env python3
"""
Environment Variable Validator
Checks that all required environment variables are set before starting the coordinator.
"""

import os
import sys
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Required environment variables
REQUIRED_VARS = [
    "DB_HOST",
    "DB_PORT", 
    "DB_NAME",
    "DB_USER",
    "DB_PASSWORD",
    "OPENAI_API_KEY"
]

# Optional environment variables (with defaults)
OPTIONAL_VARS = {
    "ANTHROPIC_API_KEY": "Not set (Claude agent disabled)",
    "DEEPSEEK_API_KEY": "Not set (DeepSeek agents disabled)",
    "WEBSOCKET_HOST": "localhost",
    "WEBSOCKET_PORT": "8765"
}

def validate_environment():
    """Validate that all required environment variables are set."""
    print("=" * 70)
    print("🔍 Validating Environment Configuration")
    print("=" * 70)
    
    # Check required variables
    missing = []
    for var in REQUIRED_VARS:
        value = os.getenv(var)
        if not value:
            missing.append(var)
            print(f"❌ {var}: NOT SET")
        else:
            # Mask sensitive values
            if "KEY" in var or "PASSWORD" in var:
                display_value = value[:8] + "..." if len(value) > 8 else "***"
            else:
                display_value = value
            print(f"✅ {var}: {display_value}")
    
    # Check optional variables
    print("\n📋 Optional Configuration:")
    for var, default in OPTIONAL_VARS.items():
        value = os.getenv(var)
        if value:
            if "KEY" in var:
                display_value = value[:8] + "..."
            else:
                display_value = value
            print(f"✅ {var}: {display_value}")
        else:
            print(f"ℹ️  {var}: {default}")
    
    print("=" * 70)
    
    # Exit if any required variables are missing
    if missing:
        print(f"\n❌ ERROR: Missing required environment variables:")
        for var in missing:
            print(f"   - {var}")
        print("\n📝 Next steps:")
        print("   1. Copy .env.example to .env")
        print("   2. Edit .env and add your configuration")
        print("   3. Run this script again")
        print("=" * 70)
        sys.exit(1)
    
    print("✅ All required environment variables are set!")
    print("=" * 70)
    return True

if __name__ == "__main__":
    validate_environment()
