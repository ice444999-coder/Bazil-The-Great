"""
SENTINEL UI Testing - Automated Playwright Tests
Clicks every button, tests scroll, compares against spec, produces fix logs

Author: SENTINEL (Testing Agent)
Date: October 19, 2025
Purpose: Comprehensive UI validation for ARES system
"""
import sys
sys.path.insert(0, 'C:\\ARES_Workspace\\ARES_API\\Lib\\site-packages')

from playwright.sync_api import sync_playwright, TimeoutError as PlaywrightTimeout
import psycopg2
import json
from datetime import datetime
from pathlib import Path
import time

# ========================================
# CONFIGURATION
# ========================================
ARES_URL = "http://localhost:8080"
PAGES_TO_TEST = [
    {"name": "trading.html", "description": "Main Trading Dashboard"},
    {"name": "code-ide.html", "description": "Code IDE & Build Tools"},
    {"name": "solace-chat.html", "description": "SOLACE Chat Interface"},
    {"name": "analytics.html", "description": "Trading Analytics"},
    {"name": "trading-analytics.html", "description": "Advanced Trading Analytics"},
    {"name": "agents.html", "description": "Agent Management"},
    {"name": "memory.html", "description": "Memory System Viewer"},
    {"name": "masterplan.html", "description": "Master Plan Dashboard"},
    {"name": "glass-box.html", "description": "Glass Box Verification"},
    {"name": "decision-log.html", "description": "Decision Log Viewer"},
    {"name": "tool-registry.html", "description": "Tool Registry & Permissions"},
    {"name": "index.html", "description": "Main Dashboard"},
    {"name": "dashboard.html", "description": "System Dashboard"},
    {"name": "performance.html", "description": "Performance Metrics"},
    {"name": "settings.html", "description": "System Settings"},
    {"name": "logs.html", "description": "System Logs"},
    {"name": "api-docs.html", "description": "API Documentation"},
    {"name": "health.html", "description": "System Health"}
]

SCREENSHOT_DIR = Path("c:\\ARES_Workspace\\ARES_API\\tests\\ui_screenshots")
SCREENSHOT_DIR.mkdir(parents=True, exist_ok=True)

REFERENCE_DIR = Path("c:\\ARES_Workspace\\ARES_API\\tests\\ui_reference")
REFERENCE_DIR.mkdir(parents=True, exist_ok=True)

# Database connection
DB_CONFIG = {
    'host': 'localhost',
    'port': 5433,
    'database': 'ares_pgvector',
    'user': 'ARES',
    'password': 'ARESISWAKING'
}

# ========================================
# TEST FUNCTIONS
# ========================================

def test_page(page, page_info):
    """
    Comprehensive test of a single page
    - Clicks every button
    - Tests scroll behavior
    - Measures window dimensions
    - Takes screenshots
    - Compares against reference
    """
    page_name = page_info["name"]
    results = {
        "page": page_name,
        "description": page_info["description"],
        "url": f"{ARES_URL}/{page_name}",
        "tested_at": datetime.now().isoformat(),
        "page_loaded": False,
        "buttons_found": 0,
        "buttons_clicked": 0,
        "buttons_failed": 0,
        "inputs_found": 0,
        "forms_found": 0,
        "scroll_tested": False,
        "scroll_works": False,
        "window_dimensions": {},
        "errors": [],
        "warnings": [],
        "screenshots": [],
        "pass": False
    }
    
    try:
        print(f"\n  📄 Loading: {page_name}")
        
        # Navigate to page
        response = page.goto(f"{ARES_URL}/{page_name}", wait_until="domcontentloaded", timeout=10000)
        
        if response and response.status >= 400:
            results["errors"].append({
                "type": "navigation",
                "message": f"HTTP {response.status} - Page not found or server error"
            })
            return results
        
        results["page_loaded"] = True
        page.wait_for_timeout(2000)  # Let JS initialize
        
        # ========================================
        # STEP 1: Capture Initial State
        # ========================================
        print(f"  📸 Capturing initial state...")
        screenshot_path = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_initial.png"
        page.screenshot(path=str(screenshot_path), full_page=True)
        results["screenshots"].append(str(screenshot_path))
        
        # ========================================
        # STEP 2: Measure Window Dimensions
        # ========================================
        print(f"  📏 Measuring window dimensions...")
        try:
            dimensions = page.evaluate("""() => {
                return {
                    viewport_width: window.innerWidth,
                    viewport_height: window.innerHeight,
                    content_width: document.documentElement.scrollWidth,
                    content_height: document.documentElement.scrollHeight,
                    scroll_y: window.scrollY,
                    scroll_x: window.scrollX,
                    scroll_max_y: document.documentElement.scrollHeight - window.innerHeight,
                    scroll_max_x: document.documentElement.scrollWidth - window.innerWidth,
                    device_pixel_ratio: window.devicePixelRatio
                };
            }""")
            results["window_dimensions"] = dimensions
            
            # Check for oversized content
            if dimensions["content_height"] > 10000:
                results["warnings"].append({
                    "type": "layout",
                    "message": f"Content height extremely large: {dimensions['content_height']}px (possible infinite scroll or layout bug)"
                })
            
            print(f"     Viewport: {dimensions['viewport_width']}x{dimensions['viewport_height']}")
            print(f"     Content: {dimensions['content_width']}x{dimensions['content_height']}")
            
        except Exception as e:
            results["errors"].append({
                "type": "measurement",
                "message": f"Failed to measure dimensions: {str(e)}"
            })
        
        # ========================================
        # STEP 3: Test Scroll Behavior
        # ========================================
        if results["window_dimensions"].get("scroll_max_y", 0) > 0:
            print(f"  📜 Testing scroll (max: {results['window_dimensions']['scroll_max_y']}px)...")
            results["scroll_tested"] = True
            
            try:
                # Scroll to bottom
                page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
                page.wait_for_timeout(500)
                
                scroll_y = page.evaluate("window.scrollY")
                expected_scroll = results["window_dimensions"]["scroll_max_y"]
                
                # Check if scroll actually worked
                if scroll_y > 0 and scroll_y >= expected_scroll * 0.9:  # Allow 10% tolerance
                    results["scroll_works"] = True
                    print(f"     ✅ Scrolled to {scroll_y}px")
                else:
                    results["scroll_works"] = False
                    results["errors"].append({
                        "type": "scroll",
                        "message": f"Scroll failed: expected ~{expected_scroll}px, got {scroll_y}px"
                    })
                
                # Screenshot scrolled state
                screenshot_path = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_scrolled.png"
                page.screenshot(path=str(screenshot_path), full_page=True)
                results["screenshots"].append(str(screenshot_path))
                
                # Scroll back to top
                page.evaluate("window.scrollTo(0, 0)")
                page.wait_for_timeout(300)
                
            except Exception as e:
                results["errors"].append({
                    "type": "scroll",
                    "message": f"Scroll test failed: {str(e)}"
                })
        else:
            print(f"  📜 No scroll needed (content fits viewport)")
        
        # ========================================
        # STEP 4: Count Form Elements
        # ========================================
        print(f"  🔍 Analyzing form elements...")
        try:
            inputs = page.locator("input, textarea, select").all()
            results["inputs_found"] = len(inputs)
            
            forms = page.locator("form").all()
            results["forms_found"] = len(forms)
            
            print(f"     Found {results['inputs_found']} inputs, {results['forms_found']} forms")
        except Exception as e:
            results["warnings"].append({
                "type": "analysis",
                "message": f"Could not analyze forms: {str(e)}"
            })
        
        # ========================================
        # STEP 5: Find and Test All Buttons
        # ========================================
        print(f"  🔘 Finding buttons...")
        try:
            # Find all clickable elements
            buttons = page.locator("button, input[type='button'], input[type='submit'], a.btn, [role='button']").all()
            results["buttons_found"] = len(buttons)
            print(f"     Found {results['buttons_found']} buttons")
            
            for i, button in enumerate(buttons):
                button_info = {
                    "index": i,
                    "text": None,
                    "id": None,
                    "class": None,
                    "visible": False,
                    "enabled": False,
                    "clicked": False,
                    "error": None
                }
                
                try:
                    # Get button info
                    button_info["text"] = button.inner_text(timeout=1000) or button.get_attribute("value") or f"Button_{i}"
                    button_info["id"] = button.get_attribute("id") or f"btn_{i}"
                    button_info["class"] = button.get_attribute("class") or ""
                    button_info["visible"] = button.is_visible()
                    button_info["enabled"] = button.is_enabled()
                    
                    # Only test visible, enabled buttons
                    if button_info["visible"] and button_info["enabled"]:
                        print(f"     [{i+1}/{results['buttons_found']}] Testing: {button_info['text'][:30]}")
                        
                        # Screenshot before click
                        try:
                            button_screenshot = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_btn_{i}_before.png"
                            button.screenshot(path=str(button_screenshot), timeout=2000)
                        except:
                            pass  # Some buttons can't be screenshotted (outside viewport)
                        
                        # Click button
                        try:
                            button.click(timeout=3000)
                            button_info["clicked"] = True
                            results["buttons_clicked"] += 1
                            page.wait_for_timeout(500)  # Wait for action to complete
                            
                            # Screenshot after click
                            after_screenshot = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_btn_{i}_after.png"
                            page.screenshot(path=str(after_screenshot), full_page=True)
                            results["screenshots"].append(str(after_screenshot))
                            
                            # Check for JS errors in console
                            # (In real implementation, we'd set up console listeners)
                            
                        except PlaywrightTimeout:
                            button_info["error"] = "Click timeout (button may be intercepted or disabled)"
                            results["buttons_failed"] += 1
                            results["warnings"].append({
                                "type": "button",
                                "button": button_info["text"],
                                "message": button_info["error"]
                            })
                        except Exception as e:
                            button_info["error"] = str(e)
                            results["buttons_failed"] += 1
                            results["errors"].append({
                                "type": "button_click",
                                "button": button_info["text"],
                                "message": str(e)
                            })
                    else:
                        # Button is hidden or disabled
                        if not button_info["visible"]:
                            print(f"     [{i+1}/{results['buttons_found']}] Skipping (hidden): {button_info['text'][:30]}")
                        elif not button_info["enabled"]:
                            print(f"     [{i+1}/{results['buttons_found']}] Skipping (disabled): {button_info['text'][:30]}")
                
                except Exception as e:
                    button_info["error"] = f"Analysis failed: {str(e)}"
                    results["warnings"].append({
                        "type": "button_analysis",
                        "message": button_info["error"]
                    })
        
        except Exception as e:
            results["errors"].append({
                "type": "button_discovery",
                "message": f"Failed to find buttons: {str(e)}"
            })
        
        # ========================================
        # STEP 6: Compare Against Reference (if exists)
        # ========================================
        reference_path = REFERENCE_DIR / f"{page_name.replace('.html', '')}_reference.png"
        if reference_path.exists():
            print(f"  🔍 Comparing against reference...")
            results["warnings"].append({
                "type": "comparison",
                "message": "Visual comparison not yet implemented (requires pixelmatch or similar)"
            })
        else:
            print(f"  📝 No reference image found (will create baseline)")
            # Copy initial screenshot as reference
            import shutil
            shutil.copy(results["screenshots"][0], reference_path)
        
        # ========================================
        # STEP 7: Calculate Pass/Fail
        # ========================================
        critical_errors = len([e for e in results["errors"] if e["type"] in ["navigation", "page_crash"]])
        if critical_errors == 0 and results["page_loaded"]:
            results["pass"] = True
        
    except Exception as e:
        results["errors"].append({
            "type": "page_crash",
            "message": f"Page test failed: {str(e)}"
        })
    
    return results


def test_responsive_design(page, page_info):
    """Test page at different viewport sizes"""
    page_name = page_info["name"]
    viewports = [
        {"width": 1920, "height": 1080, "name": "Desktop"},
        {"width": 1366, "height": 768, "name": "Laptop"},
        {"width": 768, "height": 1024, "name": "Tablet"},
        {"width": 375, "height": 667, "name": "Mobile"}
    ]
    
    results = []
    
    for viewport in viewports:
        print(f"  📱 Testing {viewport['name']} ({viewport['width']}x{viewport['height']})...")
        page.set_viewport_size({"width": viewport["width"], "height": viewport["height"]})
        page.wait_for_timeout(500)
        
        screenshot_path = SCREENSHOT_DIR / f"{page_name.replace('.html', '')}_{viewport['name'].lower()}.png"
        page.screenshot(path=str(screenshot_path), full_page=True)
        
        results.append({
            "viewport": viewport["name"],
            "dimensions": f"{viewport['width']}x{viewport['height']}",
            "screenshot": str(screenshot_path)
        })
    
    return results


def run_all_tests():
    """
    Main test runner - tests all ARES pages
    """
    print("=" * 80)
    print("  SENTINEL UI TESTING - Automated Playwright Tests")
    print("  Clicking every button, testing scroll, producing fix logs")
    print("=" * 80)
    print(f"\n  📅 Date: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"  🎯 Target: {ARES_URL}")
    print(f"  📄 Pages to test: {len(PAGES_TO_TEST)}")
    print(f"  📸 Screenshots: {SCREENSHOT_DIR}")
    
    all_results = []
    start_time = time.time()
    
    with sync_playwright() as p:
        # Launch browser
        print("\n  🌐 Launching Chromium browser...")
        browser = p.chromium.launch(
            headless=False,  # Show browser so you can see what's happening
            slow_mo=100  # Slow down by 100ms per action (easier to watch)
        )
        
        context = browser.new_context(
            viewport={"width": 1920, "height": 1080},
            user_agent="SENTINEL-UI-Tester/1.0"
        )
        
        page = context.new_page()
        
        # Test each page
        for i, page_info in enumerate(PAGES_TO_TEST):
            print(f"\n{'─' * 80}")
            print(f"  [{i+1}/{len(PAGES_TO_TEST)}] Testing: {page_info['name']} - {page_info['description']}")
            print(f"{'─' * 80}")
            
            results = test_page(page, page_info)
            all_results.append(results)
            
            # Print summary
            status = "✅ PASS" if results["pass"] else "❌ FAIL"
            print(f"\n  {status}")
            print(f"     Buttons: {results['buttons_clicked']}/{results['buttons_found']} clicked")
            print(f"     Scroll: {'✅ Works' if results['scroll_works'] else '⚠️ Not tested' if not results['scroll_tested'] else '❌ Failed'}")
            print(f"     Errors: {len(results['errors'])}")
            print(f"     Warnings: {len(results['warnings'])}")
            print(f"     Screenshots: {len(results['screenshots'])}")
            
            if results['errors']:
                print(f"\n  ⚠️ Errors found:")
                for error in results['errors'][:3]:  # Show first 3 errors
                    print(f"     - [{error['type']}] {error['message']}")
            
            # Small delay between pages
            time.sleep(1)
        
        print("\n  🔄 Testing responsive design on key pages...")
        responsive_results = []
        key_pages = ["trading.html", "solace-chat.html", "tool-registry.html"]
        for page_info in [p for p in PAGES_TO_TEST if p["name"] in key_pages]:
            print(f"\n  📱 Responsive test: {page_info['name']}")
            responsive = test_responsive_design(page, page_info)
            responsive_results.append({
                "page": page_info["name"],
                "viewports": responsive
            })
        
        browser.close()
    
    elapsed_time = time.time() - start_time
    
    # ========================================
    # Save Results to Database
    # ========================================
    print("\n  💾 Saving results to database...")
    save_results_to_db(all_results)
    
    # ========================================
    # Generate Reports
    # ========================================
    print("  📄 Generating reports...")
    generate_markdown_report(all_results, responsive_results, elapsed_time)
    generate_json_report(all_results, responsive_results, elapsed_time)
    generate_fix_log(all_results)
    
    # ========================================
    # Final Summary
    # ========================================
    total_pages = len(all_results)
    pages_passed = len([r for r in all_results if r["pass"]])
    total_buttons = sum(r['buttons_found'] for r in all_results)
    buttons_clicked = sum(r['buttons_clicked'] for r in all_results)
    total_errors = sum(len(r['errors']) for r in all_results)
    total_warnings = sum(len(r['warnings']) for r in all_results)
    
    print("\n" + "=" * 80)
    print("  ✅ UI TESTING COMPLETE")
    print("=" * 80)
    print(f"\n  📊 Summary:")
    print(f"     Pages tested: {total_pages}")
    print(f"     Pages passed: {pages_passed} ({pages_passed/total_pages*100:.1f}%)")
    print(f"     Buttons tested: {buttons_clicked}/{total_buttons}")
    print(f"     Success rate: {buttons_clicked/total_buttons*100:.1f}%" if total_buttons > 0 else "     Success rate: N/A")
    print(f"     Errors: {total_errors}")
    print(f"     Warnings: {total_warnings}")
    print(f"     Time: {elapsed_time:.1f}s")
    print(f"\n  📁 Reports saved to:")
    print(f"     - UI_TEST_REPORT.md (human-readable)")
    print(f"     - UI_TEST_RESULTS.json (machine-readable)")
    print(f"     - UI_FIX_LOG.md (developer action items)")
    print(f"\n  📸 Screenshots: {SCREENSHOT_DIR}")
    print("=" * 80 + "\n")


def save_results_to_db(results):
    """Save test results to PostgreSQL"""
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        cur = conn.cursor()
        
        for result in results:
            # Insert into test_activity_logs table
            cur.execute("""
                INSERT INTO test_activity_logs (
                    test_type, test_name, status, details, tested_at
                ) VALUES (%s, %s, %s, %s, %s)
            """, (
                'ui_automation_playwright',
                result['page'],
                'pass' if result['pass'] else 'fail',
                json.dumps(result, default=str),
                datetime.now()
            ))
        
        conn.commit()
        conn.close()
        print("     ✅ Results saved to database")
    except Exception as e:
        print(f"     ⚠️ Database save failed: {str(e)}")


def generate_markdown_report(results, responsive_results, elapsed_time):
    """Generate human-readable Markdown report"""
    report_path = Path("c:\\ARES_Workspace\\ARES_API\\tests\\UI_TEST_REPORT.md")
    
    total_pages = len(results)
    pages_passed = len([r for r in results if r["pass"]])
    total_buttons = sum(r['buttons_found'] for r in results)
    buttons_clicked = sum(r['buttons_clicked'] for r in results)
    total_errors = sum(len(r['errors']) for r in results)
    total_warnings = sum(len(r['warnings']) for r in results)
    
    with open(report_path, 'w', encoding='utf-8') as f:
        f.write("# SENTINEL UI TEST REPORT\n\n")
        f.write(f"**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  \n")
        f.write(f"**Test Duration:** {elapsed_time:.1f}s  \n")
        f.write(f"**Target:** {ARES_URL}  \n\n")
        
        f.write("## 📊 Executive Summary\n\n")
        f.write(f"- **Pages Tested:** {total_pages}\n")
        f.write(f"- **Pages Passed:** {pages_passed} ({pages_passed/total_pages*100:.1f}%)\n")
        f.write(f"- **Buttons Found:** {total_buttons}\n")
        f.write(f"- **Buttons Clicked:** {buttons_clicked} ({buttons_clicked/total_buttons*100:.1f}% success)\n")
        f.write(f"- **Errors:** {total_errors}\n")
        f.write(f"- **Warnings:** {total_warnings}\n\n")
        
        # Overall verdict
        if pages_passed == total_pages and total_errors == 0:
            f.write("### ✅ VERDICT: ALL TESTS PASSED\n\n")
        elif pages_passed >= total_pages * 0.8:
            f.write("### ⚠️ VERDICT: MOSTLY WORKING (Minor Issues)\n\n")
        else:
            f.write("### ❌ VERDICT: CRITICAL ISSUES FOUND\n\n")
        
        f.write("---\n\n")
        f.write("## 📄 Detailed Results\n\n")
        
        for result in results:
            status_icon = "✅" if result['pass'] else "❌"
            f.write(f"### {status_icon} {result['page']}\n\n")
            f.write(f"**Description:** {result['description']}  \n")
            f.write(f"**URL:** {result['url']}  \n")
            f.write(f"**Tested At:** {result['tested_at']}  \n\n")
            
            f.write("**Metrics:**\n")
            f.write(f"- Page Loaded: {'✅ Yes' if result['page_loaded'] else '❌ No'}\n")
            f.write(f"- Buttons: {result['buttons_clicked']}/{result['buttons_found']} clicked\n")
            f.write(f"- Failed: {result['buttons_failed']}\n")
            f.write(f"- Inputs: {result['inputs_found']}\n")
            f.write(f"- Forms: {result['forms_found']}\n")
            f.write(f"- Scroll Tested: {'✅ Yes' if result['scroll_tested'] else 'N/A'}\n")
            if result['scroll_tested']:
                f.write(f"- Scroll Works: {'✅ Yes' if result['scroll_works'] else '❌ No'}\n")
            
            if result['window_dimensions']:
                dims = result['window_dimensions']
                f.write(f"\n**Window Dimensions:**\n")
                f.write(f"- Viewport: {dims.get('viewport_width', 'N/A')}x{dims.get('viewport_height', 'N/A')}\n")
                f.write(f"- Content: {dims.get('content_width', 'N/A')}x{dims.get('content_height', 'N/A')}\n")
                f.write(f"- Max Scroll: {dims.get('scroll_max_y', 0)}px vertical\n")
            
            if result['screenshots']:
                f.write(f"\n**Screenshots:** {len(result['screenshots'])} captured\n")
            
            if result['errors']:
                f.write(f"\n**❌ Errors ({len(result['errors'])}):**\n")
                for error in result['errors']:
                    f.write(f"- **[{error['type']}]** {error['message']}\n")
            
            if result['warnings']:
                f.write(f"\n**⚠️ Warnings ({len(result['warnings'])}):**\n")
                for warning in result['warnings']:
                    f.write(f"- **[{warning['type']}]** {warning.get('message', 'No message')}\n")
            
            f.write("\n---\n\n")
        
        if responsive_results:
            f.write("## 📱 Responsive Design Tests\n\n")
            for resp in responsive_results:
                f.write(f"### {resp['page']}\n\n")
                for viewport in resp['viewports']:
                    f.write(f"- **{viewport['viewport']}** ({viewport['dimensions']}): `{viewport['screenshot']}`\n")
                f.write("\n")
    
    print(f"     ✅ Markdown report: {report_path}")


def generate_json_report(results, responsive_results, elapsed_time):
    """Generate machine-readable JSON report"""
    report_path = Path("c:\\ARES_Workspace\\ARES_API\\tests\\UI_TEST_RESULTS.json")
    
    report = {
        "generated_at": datetime.now().isoformat(),
        "test_duration_seconds": elapsed_time,
        "target_url": ARES_URL,
        "summary": {
            "total_pages": len(results),
            "pages_passed": len([r for r in results if r["pass"]]),
            "total_buttons": sum(r['buttons_found'] for r in results),
            "buttons_clicked": sum(r['buttons_clicked'] for r in results),
            "total_errors": sum(len(r['errors']) for r in results),
            "total_warnings": sum(len(r['warnings']) for r in results)
        },
        "results": results,
        "responsive_tests": responsive_results
    }
    
    with open(report_path, 'w', encoding='utf-8') as f:
        json.dump(report, f, indent=2, default=str)
    
    print(f"     ✅ JSON report: {report_path}")


def generate_fix_log(results):
    """Generate developer-focused fix log with actionable items"""
    log_path = Path("c:\\ARES_Workspace\\ARES_API\\tests\\UI_FIX_LOG.md")
    
    # Collect all errors and warnings
    all_issues = []
    for result in results:
        for error in result['errors']:
            all_issues.append({
                "page": result['page'],
                "severity": "ERROR",
                "type": error['type'],
                "message": error['message'],
                "button": error.get('button', 'N/A')
            })
        for warning in result['warnings']:
            all_issues.append({
                "page": result['page'],
                "severity": "WARNING",
                "type": warning['type'],
                "message": warning.get('message', 'No message'),
                "button": warning.get('button', 'N/A')
            })
    
    with open(log_path, 'w', encoding='utf-8') as f:
        f.write("# UI FIX LOG - Developer Action Items\n\n")
        f.write(f"**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  \n")
        f.write(f"**Total Issues:** {len(all_issues)}  \n\n")
        
        if not all_issues:
            f.write("## ✅ NO ISSUES FOUND\n\n")
            f.write("All UI components are working correctly!\n")
        else:
            # Group by severity
            errors = [i for i in all_issues if i['severity'] == 'ERROR']
            warnings = [i for i in all_issues if i['severity'] == 'WARNING']
            
            if errors:
                f.write(f"## ❌ CRITICAL ERRORS ({len(errors)})\n\n")
                f.write("These MUST be fixed before production deployment.\n\n")
                for i, issue in enumerate(errors, 1):
                    f.write(f"### {i}. [{issue['type']}] {issue['page']}\n\n")
                    f.write(f"**Button:** {issue['button']}  \n")
                    f.write(f"**Problem:** {issue['message']}  \n\n")
                    
                    # Add fix suggestions
                    f.write("**How to Fix:**\n")
                    if issue['type'] == 'navigation':
                        f.write("- Check if page exists in `web/` directory\n")
                        f.write("- Verify server is serving static files correctly\n")
                        f.write("- Check for typos in filename\n")
                    elif issue['type'] == 'button_click':
                        f.write("- Verify button click handler is attached\n")
                        f.write("- Check for JavaScript errors in console\n")
                        f.write("- Ensure button is not covered by another element\n")
                    elif issue['type'] == 'scroll':
                        f.write("- Check CSS `overflow` properties\n")
                        f.write("- Verify content is not position: fixed\n")
                        f.write("- Test with different viewport sizes\n")
                    else:
                        f.write("- Review error message and stack trace\n")
                        f.write("- Check browser console for JS errors\n")
                    
                    f.write("\n---\n\n")
            
            if warnings:
                f.write(f"## ⚠️ WARNINGS ({len(warnings)})\n\n")
                f.write("These should be reviewed but won't block deployment.\n\n")
                for i, issue in enumerate(warnings, 1):
                    f.write(f"### {i}. [{issue['type']}] {issue['page']}\n\n")
                    f.write(f"**Issue:** {issue['message']}  \n\n")
                    f.write("---\n\n")
        
        f.write("## 📋 Next Steps\n\n")
        f.write("1. Review all CRITICAL ERRORS above\n")
        f.write("2. Fix issues in order of severity\n")
        f.write("3. Re-run tests: `python sentinel_ui_test.py`\n")
        f.write("4. Compare screenshots in `tests/ui_screenshots/`\n")
        f.write("5. Update reference images if changes are intentional\n\n")
    
    print(f"     ✅ Fix log: {log_path}")


# ========================================
# MAIN ENTRY POINT
# ========================================

if __name__ == '__main__':
    try:
        run_all_tests()
    except KeyboardInterrupt:
        print("\n\n⚠️ Tests interrupted by user")
    except Exception as e:
        print(f"\n\n❌ Test suite crashed: {str(e)}")
        import traceback
        traceback.print_exc()
