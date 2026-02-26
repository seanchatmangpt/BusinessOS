@echo off
cd /d %~dp0

echo Starting backend server for testing...

if "%DATABASE_URL%"=="" (echo ERROR: DATABASE_URL not set & exit /b 1)
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
