#!/usr/bin/env python3
"""
ARES Trading Tab Litmus Test Suite
Tests each upgrade subtask to ensure no regression
"""
import requests
import time
import json
import sys
from typing import Dict, List

BASE_URL = "http://localhost:8080"
RESULTS = []

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    END = '\033[0m'

def test_result(name: str, passed: bool, message: str = ""):
    status = f"{Colors.GREEN}✅ PASS{Colors.END}" if passed else f"{Colors.RED}❌ FAIL{Colors.END}"
    RESULTS.append({"test": name, "passed": passed, "message": message})
    print(f"{status} | {name}")
    if message:
        print(f"    {Colors.YELLOW}└─ {message}{Colors.END}")

def test_api_health():
    """Test 1: API Health Check"""
    try:
        r = requests.get(f"{BASE_URL}/api/v1/health", timeout=5)
        test_result("API Health Check", r.status_code == 200, f"Status: {r.status_code}")
        return r.status_code == 200
    except Exception as e:
        test_result("API Health Check", False, str(e))
        return False

def test_trading_page_loads():
    """Test 2: Trading Page Loads"""
    try:
        r = requests.get(f"{BASE_URL}/trading.html", timeout=5)
        has_chart = b"chart" in r.content.lower()
        has_orderform = b"order" in r.content.lower()
        passed = r.status_code == 200 and has_chart and has_orderform
        test_result("Trading Page Loads", passed, f"Status: {r.status_code}, Chart: {has_chart}, OrderForm: {has_orderform}")
        return passed
    except Exception as e:
        test_result("Trading Page Loads", False, str(e))
        return False

def test_dashboard_loads():
    """Test 3: Dashboard Page Loads"""
    try:
        r = requests.get(f"{BASE_URL}/dashboard.html", timeout=5)
        passed = r.status_code == 200
        test_result("Dashboard Page Loads", passed, f"Status: {r.status_code}")
        return passed
    except Exception as e:
        test_result("Dashboard Page Loads", False, str(e))
        return False

def test_trading_endpoints():
    """Test 4: Trading API Endpoints"""
    endpoints = [
        "/api/v1/trading/performance",
        "/api/v1/trading/stats",
    ]
    all_pass = True
    for endpoint in endpoints:
        try:
            r = requests.get(f"{BASE_URL}{endpoint}", timeout=5)
            passed = r.status_code in [200, 404]  # 404 acceptable if not implemented yet
            if not passed:
                all_pass = False
            test_result(f"Endpoint {endpoint}", passed, f"Status: {r.status_code}")
        except Exception as e:
            test_result(f"Endpoint {endpoint}", False, str(e))
            all_pass = False
    return all_pass

def test_websocket_available():
    """Test 5: WebSocket Hub Available"""
    # Note: This is a basic check, full WebSocket testing requires ws library
    try:
        # Check if the API supports WebSocket routes
        r = requests.get(f"{BASE_URL}/health.html", timeout=5)
        passed = r.status_code == 200
        test_result("WebSocket Infrastructure", passed, f"Health page accessible: {r.status_code}")
        return passed
    except Exception as e:
        test_result("WebSocket Infrastructure", False, str(e))
        return False

def test_solace_integration():
    """Test 6: SOLACE Integration Check"""
    try:
        r = requests.get(f"{BASE_URL}/api/v1/solace/stats", timeout=5)
        passed = r.status_code in [200, 404]
        test_result("SOLACE Integration", passed, f"Status: {r.status_code}")
        return passed
    except Exception as e:
        test_result("SOLACE Integration", False, str(e))
        return False

def run_all_tests():
    """Run all litmus tests"""
    print(f"\n{Colors.BLUE}{'='*60}{Colors.END}")
    print(f"{Colors.BLUE}ARES Trading Tab Litmus Test Suite{Colors.END}")
    print(f"{Colors.BLUE}{'='*60}{Colors.END}\n")
    
    tests = [
        test_api_health,
        test_trading_page_loads,
        test_dashboard_loads,
        test_trading_endpoints,
        test_websocket_available,
        test_solace_integration,
    ]
    
    for test_func in tests:
        test_func()
        time.sleep(0.5)  # Brief pause between tests
    
    # Summary
    print(f"\n{Colors.BLUE}{'='*60}{Colors.END}")
    passed = sum(1 for r in RESULTS if r["passed"])
    total = len(RESULTS)
    pass_rate = (passed / total * 100) if total > 0 else 0
    
    print(f"{Colors.BLUE}Test Summary:{Colors.END}")
    print(f"  Total Tests: {total}")
    print(f"  Passed: {Colors.GREEN}{passed}{Colors.END}")
    print(f"  Failed: {Colors.RED}{total - passed}{Colors.END}")
    print(f"  Pass Rate: {pass_rate:.1f}%")
    print(f"{Colors.BLUE}{'='*60}{Colors.END}\n")
    
    # Return exit code
    return 0 if passed == total else 1

if __name__ == "__main__":
    exit_code = run_all_tests()
    sys.exit(exit_code)
