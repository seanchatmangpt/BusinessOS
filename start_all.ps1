# Script para iniciar Backend com suporte a Image Search
# PowerShell script

Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 60) -ForegroundColor Cyan
Write-Host "Iniciando BusinessOS Backend com Image/Multimodal Support" -ForegroundColor Green
Write-Host "=" -NoNewline -ForegroundColor Cyan
Write-Host ("=" * 60) -ForegroundColor Cyan
Write-Host ""

# Definir variáveis de ambiente
$env:CLIP_PROVIDER = "local"
$env:CLIP_LOCAL_URL = "http://localhost:8000"

Write-Host "Variáveis de ambiente configuradas:" -ForegroundColor Yellow
Write-Host "  CLIP_PROVIDER = $env:CLIP_PROVIDER" -ForegroundColor Gray
Write-Host "  CLIP_LOCAL_URL = $env:CLIP_LOCAL_URL" -ForegroundColor Gray
Write-Host ""

# Ir para diretório do backend
Set-Location -Path "C:\Users\Pichau\Desktop\BusinessOS-main-dev\desktop\backend-go"

Write-Host "Iniciando servidor..." -ForegroundColor Yellow
Write-Host ""

# Iniciar servidor
go run .\cmd\server
