@echo off
echo ========================================
echo Building ARES Desktop Launcher
echo ========================================

echo.
echo [1/2] Building ARES API Server...
go build -o ares_api.exe -ldflags="-s -w" cmd/main.go
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to build API server
    pause
    exit /b 1
)
echo ✓ ares_api.exe built successfully

echo.
echo [2/2] Building ARES Launcher...
go build -o ARES.exe -ldflags="-s -w -H=windowsgui" cmd/launcher/main.go
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to build launcher
    pause
    exit /b 1
)
echo ✓ ARES.exe built successfully

echo.
echo ========================================
echo Build Complete!
echo ========================================
echo.
echo Files created:
echo   - ares_api.exe (API Server)
echo   - ARES.exe (Desktop Launcher)
echo.
echo To run ARES, simply double-click ARES.exe
echo.
pause
