@echo off

echo ==========================
echo Starting Firmware Server
echo ==========================

set ADMIN_TOKEN=super_secret_admin_token

if not exist firmwares (
    mkdir firmwares
)

go build -o firmware-server.exe

firmware-server.exe

pause