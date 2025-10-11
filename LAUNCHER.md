# ARES Desktop Launcher

## Overview

ARES.exe is a Windows desktop launcher that provides a seamless one-click experience for running the ARES platform.

## Features

✅ **Single-Click Launch** - Just double-click ARES.exe
✅ **Silent Background Server** - API server runs invisibly in background
✅ **Auto Browser Launch** - Opens ARES UI automatically
✅ **Graceful Shutdown** - Cleanly stops all processes on exit
✅ **No Console Windows** - Professional desktop app experience
✅ **Health Monitoring** - Waits for server to be ready before opening UI

## Files

- **ARES.exe** (5.7 MB) - Desktop launcher executable
- **ares_api.exe** (38 MB) - ARES API server executable
- **ares_server.log** - Server output log (created on first run)

## Usage

### Running ARES

1. Double-click `ARES.exe`
2. Wait a few seconds for the server to start
3. Browser opens automatically to http://localhost:8080
4. Start using ARES!

### Stopping ARES

- Close the browser tab
- ARES continues running in the background
- To fully stop: Press `Ctrl+C` in the console (if visible) or close via Task Manager

### Development Mode

For development, you can still use:
```bash
go run cmd/main.go
```

## Building from Source

```bash
# Build both executables
go build -o ares_api.exe -ldflags="-s -w" cmd/main.go
go build -o ARES.exe -ldflags="-s -w -H=windowsgui" cmd/launcher/main.go
```

Or simply run:
```bash
build.bat
```

## Architecture

```
ARES.exe (Launcher)
   ├─ Starts ares_api.exe (silently)
   ├─ Monitors server health
   ├─ Opens browser when ready
   └─ Handles graceful shutdown

ares_api.exe (API Server)
   ├─ PostgreSQL database
   ├─ Claude AI integration
   ├─ Trading endpoints
   ├─ Memory system
   └─ REST API on :8080
```

## Configuration

The launcher uses the same `.env` file as the API server:
- Database credentials
- Anthropic API key
- Server port (default: 8080)
- Repository path

## Logs

Server logs are written to `ares_server.log` in the installation directory.

## System Requirements

- Windows 10/11
- PostgreSQL running on localhost:5432
- 100MB free disk space
- Internet connection (for Claude API calls)

## Troubleshooting

**Server won't start:**
- Check `ares_server.log` for errors
- Verify PostgreSQL is running
- Check port 8080 is not in use

**Browser doesn't open:**
- Server may still be starting
- Check http://localhost:8080 manually
- Verify firewall allows local connections

**Can't stop ARES:**
- Open Task Manager
- End `ares_api.exe` process
- End `ARES.exe` process

## Desktop Shortcut

To create a desktop shortcut:
1. Right-click `ARES.exe`
2. Select "Create shortcut"
3. Move shortcut to Desktop
4. Rename to "ARES"

## Icon (Future Enhancement)

To add a custom icon:
```bash
go build -o ARES.exe -ldflags="-s -w -H=windowsgui" -ldflags="-X main.icon=ares.ico" cmd/launcher/main.go
```

---

**Built with:** Go 1.x
**Platform:** Windows x64
**License:** MIT
