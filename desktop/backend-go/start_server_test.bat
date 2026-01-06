@echo off
cd /d %~dp0

echo Starting backend server for testing...

set DATABASE_URL=postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30
set SERVER_PORT=8001
set AI_PROVIDER=groq
set GROQ_API_KEY=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

echo Building server...
go build -o server.exe ./cmd/server

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b 1
)

echo Starting server on port 8001...
start /B server.exe

echo Waiting for server to start...
timeout /t 5 /nobreak > nul

curl http://localhost:8001/api/status

echo.
echo Server is running!
