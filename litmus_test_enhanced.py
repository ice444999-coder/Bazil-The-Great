import subprocess
import time
import requests
import os
import re
import tempfile

PROJECT_DIR = "C:\\ARES_Workspace\\ARES_API"  # Your root
SERVER_URL = "http://localhost:8080"
ENDPOINTS = [
    "/", "/dashboard.html", "/trading.html", "/api/bazil/rewards", "/api/bazil/findings", "/solace/ws"  # Add more as needed
]
EXPECTED_LOGS = ["System healthy", "Heal triggered", "Patch applied", "Verified—system healed"]  # Patterns for success
ERROR_PATTERNS = [r"err(or)?", "failed", "deviation", "mismatch", "undefined", "missing import"]  # For wiring issues

def run_command(cmd, cwd=PROJECT_DIR, timeout=60):
    try:
        process = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=timeout)
        if process.returncode != 0:
            raise ValueError(f"Command {' '.join(cmd)} failed with code {process.returncode}\nStdout: {process.stdout}\nStderr: {process.stderr}")
        return process.stdout + process.stderr
    except Exception as e:
        return f"Exception: {str(e)}"

def check_wiring_errors(code_dir):
    errors = []
    # Use PowerShell Select-String instead of grep on Windows
    ps_cmd = ["powershell", "-Command", f"Get-ChildItem -Path '{code_dir}' -Recurse -Include *.go | Select-String -Pattern '(undefined:|missing import|wiring|deviation|panic|error)' | Select-Object -First 50"]
    grep_out = run_command(ps_cmd)
    if grep_out and "No such file" not in grep_out and grep_out.strip():
        errors.append(f"Wiring patterns found:\n{grep_out}")
    
    # Run go vet
    vet_out = run_command(["go", "vet", "./..."])
    if "error" in vet_out.lower():
        errors.append(f"Go vet errors:\n{vet_out}")
    
    # Run golangci-lint for static analysis (skip if not installed)
    try:
        lint_out = run_command(["golangci-lint", "run", "--no-config", "--fast"], timeout=30)
        if "error" in lint_out.lower() or "issue" in lint_out.lower():
            errors.append(f"Golangci-lint issues:\n{lint_out}")
    except:
        print("  (golangci-lint not installed, skipping)")
    
    return errors

def start_server():
    # Kill existing processes on port 8080
    try:
        run_command(["powershell", "-Command", "$proc = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess; if ($proc) { Stop-Process -Id $proc -Force }"], timeout=5)
    except:
        pass
    
    # Start server using go run (since we may not have ares_api.exe built yet)
    proc = subprocess.Popen(["go", "run", "cmd/main.go"], cwd=PROJECT_DIR, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    time.sleep(10)  # Longer wait for startup
    return proc

def test_endpoints():
    errors = []
    for endpoint in ENDPOINTS:
        try:
            resp = requests.get(SERVER_URL + endpoint, timeout=10)
            if resp.status_code != 200:
                errors.append(f"Endpoint {endpoint} failed: status {resp.status_code}")
            if "error" in resp.text.lower() or not resp.text:
                errors.append(f"Endpoint {endpoint} has empty/error content: {resp.text[:200]}...")
        except Exception as e:
            errors.append(f"Endpoint {endpoint} inaccessible: {str(e)}")
    return errors

def check_logs(proc, duration=30):
    errors = []
    start_time = time.time()
    log_content = ""
    while time.time() - start_time < duration:
        log_content += proc.stdout.read().decode('utf-8', errors='ignore')
        log_content += proc.stderr.read().decode('utf-8', errors='ignore')
        time.sleep(1)
    
    # Check success
    missing = [p for p in EXPECTED_LOGS if p not in log_content]
    if missing:
        errors.append(f"Missing expected logs: {', '.join(missing)}")
    
    # Check errors
    for pattern in ERROR_PATTERNS:
        matches = re.findall(pattern, log_content, re.IGNORECASE)
        if matches:
            errors.append(f"Error patterns found: {matches}")
    
    return errors, log_content

def inject_fault_and_test_heal():
    errors = []
    solace_file = os.path.join(PROJECT_DIR, "internal", "solace", "solace_agent.go")
    backup = solace_file + ".bak"
    
    # Check if file exists
    if not os.path.exists(solace_file):
        errors.append(f"File not found for fault injection: {solace_file}")
        return errors
    
    os.rename(solace_file, backup)
    
    # Inject unused import
    with open(backup, 'r', encoding='utf-8') as f:
        content = f.read()
    injected = 'import "unused/pkg" // Injected fault\n' + content
    with open(solace_file, 'w', encoding='utf-8') as f:
        f.write(injected)
    
    print("Fault injected: Unused import in solace_agent.go")
    time.sleep(5)  # Give time for loop to detect
    
    # Check if healed (check logs if they exist)
    try:
        post_logs = run_command(["powershell", "-Command", "Get-Content server.log -ErrorAction SilentlyContinue | Select-String 'healed'"])
        if "healed" not in post_logs:
            errors.append("Healing did not trigger or succeed")
    except:
        errors.append("Cannot verify healing - no server.log found")
    
    # Clean up
    os.rename(backup, solace_file)
    return errors

def run_litmus_test():
    print("Starting Enhanced Litmus Test for ARES Wiring...")
    errors = []

    # Step 1: Code wiring check
    print("Step 1: Scanning for wiring errors...")
    code_errors = check_wiring_errors(os.path.join(PROJECT_DIR, "internal"))
    errors.extend(code_errors)

    # Step 2: Tidy, build, test
    print("Step 2: Tidy dependencies...")
    tidy_out = run_command(["go", "mod", "tidy"])
    if "error" in tidy_out.lower():
        errors.append(f"Tidy failed:\n{tidy_out}")

    print("Building app...")
    build_out = run_command(["go", "build", "-o", "ares_api.exe", "cmd/main.go"])
    if "error" in build_out.lower():
        errors.append(f"Build failed:\n{build_out}")

    print("Running unit/integration tests...")
    test_out = run_command(["go", "test", "./..."])
    if "FAIL" in test_out:
        errors.append(f"Go test failures:\n{test_out}")

    # Step 3: Runtime server test
    print("Step 3: Starting server...")
    proc = start_server()
    try:
        endpoint_errors = test_endpoints()
        errors.extend(endpoint_errors)

        # Step 4: Log monitoring
        print("Step 4: Monitoring logs...")
        log_errors, logs = check_logs(proc, duration=60)  # Longer for heal
        errors.extend(log_errors)
        print(f"Logs excerpt:\n{logs[-1000:]}")  # Last 1000 chars for debug

        # Step 5: Fault injection and heal test
        print("Step 5: Injecting fault and testing heal...")
        heal_errors = inject_fault_and_test_heal()
        errors.extend(heal_errors)

    finally:
        proc.terminate()

    # Report
    if errors:
        print("\nERRORS FOUND (Wiring/Function Issues):")
        for err in errors:
            print(f"- {err}")
        print("\nFix these and rerun. Paste full output back for patches.")
    else:
        print("\nALL TESTS PASS - System wired 100% perfectly!")

if __name__ == "__main__":
    run_litmus_test()
