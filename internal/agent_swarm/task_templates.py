# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
"""
Task Templates Library
Reusable task definitions - never write SQL manually again!
"""

TEMPLATES = {
    "code_quality_zero_tolerance": {
        "task_type": "code_refactoring",
        "priority": 10,
        "description": """CRITICAL: Fix ALL code quality issues. Zero tolerance policy.

TARGETS:
- {file_count} files to fix
- {warning_count} warnings to eliminate  
- {error_count} errors to fix

PHASES:
1. SENTINEL: Comprehensive audit of all violations
2. ARCHITECT: Design refactoring strategy
3. FORGE: Implement fixes with testing
4. SENTINEL: Validate ZERO warnings/errors
5. SOLACE: Final approval

SUCCESS CRITERIA:
- ZERO errors ✓
- ZERO warnings ✓
- All functionality preserved ✓
- Visual appearance identical ✓
- Browser compatibility verified ✓

Quality Standard: PRODUCTION READY - ZERO DEFECTS
Expected Duration: {duration} minutes""",
        "context": {
            "strict_mode": True,
            "max_warnings": 0,
            "max_errors": 0,
            "testing_required": True
        }
    },
    
    "ui_fix": {
        "task_type": "ui_building",
        "priority": 8,
        "description": """Fix UI component: {component_name}

Issues to fix:
{issues}

Requirements:
- Maintain current functionality
- Improve user experience
- Follow design system
- Add proper error handling

Expected Duration: {duration} minutes""",
        "context": {
            "framework": "react",
            "testing_required": True
        }
    },
    
    "bug_fix": {
        "task_type": "debugging",
        "priority": 9,
        "description": """Bug Fix: {bug_title}

Description: {bug_description}

Steps to reproduce:
{reproduction_steps}

Expected behavior: {expected}
Actual behavior: {actual}

Agents:
1. SENTINEL: Reproduce and diagnose
2. ARCHITECT: Design solution
3. FORGE: Implement fix
4. SENTINEL: Verify resolved

Expected Duration: {duration} minutes""",
        "context": {
            "severity": "high",
            "testing_required": True
        }
    },
    
    "feature_implementation": {
        "task_type": "feature_development",
        "priority": 7,
        "description": """Implement Feature: {feature_name}

Requirements:
{requirements}

Technical Specifications:
{tech_specs}

Acceptance Criteria:
{acceptance_criteria}

Agents:
1. ARCHITECT: Design architecture
2. FORGE: Implement backend + frontend
3. SENTINEL: Test all scenarios
4. SOLACE: Review and approve

Expected Duration: {duration} minutes""",
        "context": {
            "testing_required": True,
            "documentation_required": True
        }
    }
}


def format_template(template_name, **kwargs):
    """Format a template with parameters"""
    if template_name not in TEMPLATES:
        raise ValueError(f"Template '{template_name}' not found")
    
    template = TEMPLATES[template_name].copy()
    
    # Format description with kwargs
    template['description'] = template['description'].format(**kwargs)
    
    # Merge context if provided
    if 'context' in kwargs:
        template['context'].update(kwargs['context'])
        del kwargs['context']
    
    # Add file_paths if provided
    if 'file_paths' in kwargs:
        template['file_paths'] = kwargs['file_paths']
    
    return template
