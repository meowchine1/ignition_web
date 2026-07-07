@echo off

echo ==========================
echo Starting Firmware Server
echo ==========================

set ADMIN_TOKEN=super_secret_admin_token

if not exist firmwares (
    mkdir firmwares
)

if not exist firmware-server.exe (
    echo Creating firmware-server.exe...
    type nul > firmware-server.exe
)

echo Building...
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