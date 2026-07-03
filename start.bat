@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo [1/3] Building frontend...
cd web
call npm run build --silent
if %errorlevel% neq 0 (
    echo Frontend build failed!
    pause
    exit /b %errorlevel%
)
cd ..

echo [2/3] Building backend...
go build -o server.exe ./cmd/server
if %errorlevel% neq 0 (
    echo Backend build failed!
    pause
    exit /b %errorlevel%
)

echo [3/3] Starting server...
server.exe

pause
