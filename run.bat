@echo off
REM ============================================================
REM  Ignition Firmware Flasher - local Windows dev launcher
REM  Builds the server and runs it on  http://localhost:8200
REM
REM  Run from ANY directory - the script switches to its own.
REM  Extra arguments are forwarded to the server, e.g.:
REM    run.bat -create-admin
REM    run.bat -init-secrets
REM ============================================================

setlocal EnableExtensions
cd /d "%~dp0"

REM ---------- 1. Go must be installed ----------
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not found in PATH. Install Go 1.25+ and retry.
    pause
    exit /b 1
)

REM ---------- 2. .env config ----------
if not exist ".env" (
    echo [WARN] .env not found - creating it from .env_example ...
    if not exist ".env_example" (
        echo [ERROR] .env_example is also missing. Check the project files.
        pause
        exit /b 1
    )
    copy ".env_example" ".env" >nul
)

REM ---------- 3. Data folders ----------
if not exist "firmwares" mkdir "firmwares"
if not exist "flashers"  mkdir "flashers"

REM ---------- 4. Build ----------
echo.
echo Building firmware-server.exe ...
set "CGO_ENABLED=1"
go build -o "firmware-server.exe"
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    pause
    exit /b 1
)

REM ---------- 5. Ensure machine secrets (JWT/AES/HMAC) ----------
REM  No-op if already present; appends only the missing ones.
echo Ensuring machine secrets in .env ...
"firmware-server.exe" -init-secrets
if %errorlevel% neq 0 (
    echo [ERROR] Failed to set up secrets. Run manually:  go run . -init-secrets
    pause
    exit /b 1
)

REM ---------- 6. Run (or forward flags) ----------
if "%~1"=="" (
    echo.
    echo Starting server on http://localhost:8200
    echo Press Ctrl+C in this window to stop it.
    echo.
    "firmware-server.exe"
) else (
    "firmware-server.exe" %*
)

set "SERVER_EXIT=%errorlevel%"
echo.
echo Server exited with code %SERVER_EXIT%.
pause
exit /b %SERVER_EXIT%