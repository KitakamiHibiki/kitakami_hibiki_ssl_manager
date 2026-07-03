@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo [1/2] Building frontend...
cd web
call npm run build --silent
if %errorlevel% neq 0 (
    echo Frontend build failed!
    pause
    exit /b %errorlevel%
)
cd ..

echo [2/2] Starting server...
go run ./cmd/server

pause
