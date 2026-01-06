@echo off
cd desktop\backend-go
echo Loading environment variables...
set CLIP_PROVIDER=local
set CLIP_LOCAL_URL=http://localhost:8000
echo Starting Backend with Image/Multimodal Support...
echo CLIP_PROVIDER=%CLIP_PROVIDER%
echo CLIP_LOCAL_URL=%CLIP_LOCAL_URL%
echo.
go run ./cmd/server
