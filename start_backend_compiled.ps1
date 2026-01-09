# Start compiled backend with environment variables
Set-Location "C:\Users\Pichau\Desktop\BusinessOS-main-dev\desktop\backend-go"

$env:CLIP_PROVIDER = "local"
$env:CLIP_LOCAL_URL = "http://localhost:8000"

Write-Host "Starting compiled backend..." -ForegroundColor Green
.\server.exe
