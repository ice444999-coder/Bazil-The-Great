#!/usr/bin/env pwsh
# Quick Test Runner for OpenAI Function Calling Integration
# Run this script to test SOLACE's new capabilities

Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
Write-Host "🤖 SOLACE OpenAI Function Calling Test" -ForegroundColor Cyan
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
Write-Host ""

# Check if API key is set
if (-not $env:OPENAI_API_KEY) {
    Write-Host "❌ ERROR: OPENAI_API_KEY not set" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please set your OpenAI API key:" -ForegroundColor Yellow
    Write-Host "  `$env:OPENAI_API_KEY = `"sk-proj-...`"" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host "✅ OPENAI_API_KEY found" -ForegroundColor Green
Write-Host ""

# Change to agent_swarm directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

Write-Host "📂 Working directory: $PWD" -ForegroundColor Cyan
Write-Host ""

# Ask user which test to run
Write-Host "Choose test to run:" -ForegroundColor Yellow
Write-Host "  [1] Start WebSocket Server (test_websocket_server.py)" -ForegroundColor White
Write-Host "  [2] Run OpenAI Chat Test (test_openai_chat.py)" -ForegroundColor White
Write-Host "  [3] Run Both (separate terminals)" -ForegroundColor White
Write-Host ""

$choice = Read-Host "Enter choice (1-3)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "🚀 Starting WebSocket Server..." -ForegroundColor Green
        Write-Host "   Press Ctrl+C to stop" -ForegroundColor Yellow
        Write-Host ""
        python test_websocket_server.py
    }
    
    "2" {
        Write-Host ""
        Write-Host "🧪 Running OpenAI Chat Test..." -ForegroundColor Green
        Write-Host ""
        python test_openai_chat.py
    }
    
    "3" {
        Write-Host ""
        Write-Host "Opening two terminals..." -ForegroundColor Green
        Write-Host ""
        Write-Host "Terminal 1: WebSocket Server" -ForegroundColor Cyan
        Start-Process pwsh -ArgumentList "-NoExit", "-Command", "cd '$PWD'; Write-Host '🚀 Starting WebSocket Server...' -ForegroundColor Green; python test_websocket_server.py"
        
        Start-Sleep -Seconds 2
        
        Write-Host "Terminal 2: OpenAI Chat Test" -ForegroundColor Cyan
        Start-Process pwsh -ArgumentList "-NoExit", "-Command", "cd '$PWD'; Write-Host '🧪 Running OpenAI Chat Test...' -ForegroundColor Green; Start-Sleep -Seconds 3; python test_openai_chat.py"
        
        Write-Host ""
        Write-Host "✅ Both terminals opened!" -ForegroundColor Green
        Write-Host "   Check the new windows for output" -ForegroundColor Yellow
    }
    
    default {
        Write-Host ""
        Write-Host "❌ Invalid choice" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 69) -ForegroundColor Cyan
