@echo off

echo ==========================
echo Starting Firmware Server
echo ==========================

if not exist firmwares (
    mkdir firmwares
)

if not exist flashers (
    mkdir flashers
)

echo Building...
set CGO_ENABLED=1
go build -o firmware-server.exe

if errorlevel 1 (
    echo.
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo Starting server...
echo.

firmware-server.exe

pause