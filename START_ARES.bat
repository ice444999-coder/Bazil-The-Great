@echo off
title ARES - Adaptive Recognition and Evaluation System

echo ========================================
echo    ARES - Starting with Semantic Memory
echo ========================================
echo.

REM Kill any existing ARES processes on port 8080
echo [1/3] Checking for existing ARES processes...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080') do (
    echo   - Killing process %%a
    taskkill /F /PID %%a >nul 2>&1
)
echo   Done!
echo.

REM Start ARES
echo [2/3] Starting ARES.exe...
echo   - Semantic Memory: ENABLED
echo   - Embedding Model: nomic-embed-text
echo   - Background Workers: 2 (Trading + Embeddings)
echo   - API: http://localhost:8080
echo.
echo [3/3] Server starting...
echo ========================================
echo.

%~dp0ARES.exe

if errorlevel 1 (
    echo.
    echo ========================================
    echo ERROR: ARES failed to start!
    echo Check the error message above.
    echo ========================================
    echo Press any key to close this window...
    pause >nul
) else (
    echo.
    echo ========================================
    echo ARES has stopped.
    echo Press any key to close this window...
    pause >nul
)
