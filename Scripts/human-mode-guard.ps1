# HUMAN MODE Guard - Forward-Only Header Enforcement
# Integrated with Sentinel - Only protects NEW work, doesn't touch legacy files

param(
    [switch]$Check,      # Just check, don't modify
    [switch]$Inject,     # Inject headers into changed files
    [switch]$Verify      # Run verification on changed files
)

$ErrorActionPreference = "Stop"

# HUMAN MODE header templates by file type
$Headers = @{
    'go' = @'
/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/

'@
    'ps1' = @'
# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md

'@
    'js' = @'
// HUMAN MODE - Truth Protocol Active
// System: Senior CTO-scientist reasoning mode engaged
// Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
// This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md

'@
    'py' = @'
# HUMAN MODE - Truth Protocol Active
# System: Senior CTO-scientist reasoning mode engaged
# Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
# This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md

'@
    'html' = @'
<!-- HUMAN MODE - Truth Protocol Active
     System: Senior CTO-scientist reasoning mode engaged
     Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
     This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md -->

'@
}

# Get changed files in current branch (vs main or HEAD~1)
function Get-ChangedFiles {
    $branch = git branch --show-current
    
    # Try comparing to main first
    $changed = git diff --name-only origin/main...HEAD 2>$null
    
    # If no origin/main, compare to last commit
    if (-not $changed) {
        $changed = git diff --name-only HEAD~1 2>$null
    }
    
    # If still nothing, get staged files
    if (-not $changed) {
        $changed = git diff --name-only --cached 2>$null
    }
    
    return $changed | Where-Object { $_ }
}

# Check if file has HUMAN MODE header
function Test-HasHeader {
    param([string]$FilePath)
    
    if (-not (Test-Path $FilePath)) { return $false }
    
    $firstLines = Get-Content $FilePath -TotalCount 5 -ErrorAction SilentlyContinue
    return ($firstLines -join "`n") -match "HUMAN MODE"
}

# Inject header into file
function Add-HumanModeHeader {
    param([string]$FilePath)
    
    $ext = [System.IO.Path]::GetExtension($FilePath).TrimStart('.')
    
    # Skip if no header template for this type
    if (-not $Headers.ContainsKey($ext)) {
        Write-Host "  SKIP: $FilePath (no template for .$ext)" -ForegroundColor Yellow
        return $false
    }
    
    # Skip if already has header
    if (Test-HasHeader $FilePath) {
        Write-Host "  OK: $FilePath (already protected)" -ForegroundColor Green
        return $true
    }
    
    # Inject header
    $header = $Headers[$ext]
    $content = Get-Content $FilePath -Raw
    $newContent = $header + $content
    
    Set-Content $FilePath -Value $newContent -NoNewline
    Write-Host "  ✅ INJECTED: $FilePath" -ForegroundColor Cyan
    return $true
}

# Calculate SHA256 hash of file
function Get-FileHash256 {
    param([string]$FilePath)
    $hash = Get-FileHash $FilePath -Algorithm SHA256
    return $hash.Hash
}

# Main execution
Write-Host "`n🛡️  HUMAN MODE Guard - Forward-Only Protection`n" -ForegroundColor Magenta

# Get changed files
$changedFiles = Get-ChangedFiles

if (-not $changedFiles) {
    Write-Host "No changed files detected. Nothing to guard." -ForegroundColor Yellow
    exit 0
}

Write-Host "Changed files detected: $($changedFiles.Count)" -ForegroundColor Cyan
$changedFiles | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
Write-Host ""

# Filter to only code files
$codeExtensions = @('go', 'ps1', 'js', 'ts', 'py', 'html', 'css', 'sql')
$codeFiles = $changedFiles | Where-Object { 
    $ext = [System.IO.Path]::GetExtension($_).TrimStart('.')
    $codeExtensions -contains $ext
}

if (-not $codeFiles) {
    Write-Host "No code files in changes. Nothing to guard." -ForegroundColor Yellow
    exit 0
}

Write-Host "Code files to process: $($codeFiles.Count)`n" -ForegroundColor Cyan

# Check mode - just report status
if ($Check) {
    Write-Host "🔍 CHECK MODE - Verifying headers...`n" -ForegroundColor Yellow
    
    $missing = @()
    foreach ($file in $codeFiles) {
        if (-not (Test-Path $file)) { continue }
        
        if (Test-HasHeader $file) {
            Write-Host "  ✅ $file" -ForegroundColor Green
        } else {
            Write-Host "  ❌ $file (missing header)" -ForegroundColor Red
            $missing += $file
        }
    }
    
    if ($missing) {
        Write-Host "`n⚠️  $($missing.Count) files missing HUMAN MODE headers" -ForegroundColor Red
        Write-Host "Run with -Inject to add headers automatically`n" -ForegroundColor Yellow
        exit 1
    } else {
        Write-Host "`n✅ All changed files have HUMAN MODE protection`n" -ForegroundColor Green
        exit 0
    }
}

# Inject mode - add headers
if ($Inject) {
    Write-Host "💉 INJECT MODE - Adding headers...`n" -ForegroundColor Cyan
    
    $injected = 0
    foreach ($file in $codeFiles) {
        if (-not (Test-Path $file)) { continue }
        
        if (Add-HumanModeHeader $file) {
            if (-not (Test-HasHeader $file)) {
                $injected++
            }
        }
    }
    
    Write-Host "`n✅ Injected headers into $injected files" -ForegroundColor Green
    
    # Create verification ledger
    if ($Verify) {
        Write-Host "`n📋 Creating verification ledger...`n" -ForegroundColor Cyan
        
        $ledger = @{
            timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
            branch = git branch --show-current
            protected_files = @{}
        }
        
        foreach ($file in $codeFiles) {
            if (-not (Test-Path $file)) { continue }
            if (Test-HasHeader $file) {
                $ledger.protected_files[$file] = Get-FileHash256 $file
            }
        }
        
        $ledgerPath = "truth_protocol/verified_ledger.json"
        New-Item -ItemType Directory -Force -Path "truth_protocol" | Out-Null
        $ledger | ConvertTo-Json -Depth 10 | Set-Content $ledgerPath
        
        Write-Host "✅ Verification ledger created: $ledgerPath" -ForegroundColor Green
        Write-Host "   Protected files: $($ledger.protected_files.Count)`n" -ForegroundColor Cyan
    }
    
    exit 0
}

# Verify mode - check hashes
if ($Verify) {
    Write-Host "🔐 VERIFY MODE - Checking file integrity...`n" -ForegroundColor Cyan
    
    $ledgerPath = "truth_protocol/verified_ledger.json"
    
    if (-not (Test-Path $ledgerPath)) {
        Write-Host "⚠️  No verification ledger found" -ForegroundColor Yellow
        Write-Host "Run with -Inject -Verify to create one`n" -ForegroundColor Yellow
        exit 0
    }
    
    $ledger = Get-Content $ledgerPath | ConvertFrom-Json
    $tampered = @()
    
    foreach ($file in $codeFiles) {
        if (-not (Test-Path $file)) { continue }
        if (-not $ledger.protected_files.$file) { continue }
        
        $currentHash = Get-FileHash256 $file
        $recordedHash = $ledger.protected_files.$file
        
        if ($currentHash -ne $recordedHash) {
            Write-Host "  ⚠️  $file (hash mismatch - modified)" -ForegroundColor Yellow
            $tampered += $file
        } else {
            Write-Host "  ✅ $file (verified)" -ForegroundColor Green
        }
    }
    
    if ($tampered) {
        Write-Host "`n⚠️  $($tampered.Count) files modified since last verification" -ForegroundColor Yellow
        Write-Host "This is normal for active development. Re-run -Inject -Verify to update ledger`n" -ForegroundColor Cyan
    } else {
        Write-Host "`n✅ All files verified - no tampering detected`n" -ForegroundColor Green
    }
    
    exit 0
}

# Default mode - interactive
Write-Host "📋 INTERACTIVE MODE`n" -ForegroundColor Cyan
Write-Host "Available actions:" -ForegroundColor White
Write-Host "  -Check    : Verify which files have headers" -ForegroundColor Gray
Write-Host "  -Inject   : Add headers to files missing them" -ForegroundColor Gray
Write-Host "  -Verify   : Check file integrity vs ledger" -ForegroundColor Gray
Write-Host ""
Write-Host "Example: .\human-mode-guard.ps1 -Inject -Verify`n" -ForegroundColor Yellow
