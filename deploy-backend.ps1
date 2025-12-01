# PowerShell Deployment Script for Backend
# Run this from Windows PowerShell

$remoteUser = "arc"
$remoteHost = "arc-demo.thepixpivot.com"
$remoteDir = "/home/arc/apps/backend"
$localDir = "d:\arc-demo\apps\backend"

Write-Host "🚀 Deploying Backend with License Fix..." -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

# Step 1: Transfer modified files using SCP
Write-Host "`n📤 Step 1: Transferring modified files..." -ForegroundColor Yellow

$filesToTransfer = @(
    @{Local="$localDir\cmd\server\main.go"; Remote="cmd/server/main.go"},
    @{Local="$localDir\internal\api\handlers\sso_handler.go"; Remote="internal/api/handlers/sso_handler.go"},
    @{Local="$localDir\internal\api\handlers\signup_handler.go"; Remote="internal/api/handlers/signup_handler.go"},
    @{Local="$localDir\internal\api\handlers\auth_handler.go"; Remote="internal/api/handlers/auth_handler.go"},
    @{Local="$localDir\internal\core\services\email_service.go"; Remote="internal/core/services/email_service.go"},
    @{Local="$localDir\.env.production"; Remote=".env"}
)

foreach ($file in $filesToTransfer) {
    if (Test-Path $file.Local) {
        Write-Host "  Transferring $($file.Remote)..." -ForegroundColor Cyan
        scp $file.Local "${remoteUser}@${remoteHost}:${remoteDir}/$($file.Remote)"
    } else {
        Write-Host "  Warning: $($file.Local) not found!" -ForegroundColor Red
    }
}

Write-Host "✅ Files transferred" -ForegroundColor Green

# Step 2: Rebuild and restart backend
Write-Host "`n🔨 Step 2: Rebuilding backend on server..." -ForegroundColor Yellow

$deployCommands = @"
cd /home/arc/apps/backend
echo '🛑 Stopping containers...'
docker compose -f docker-compose.prod.yml down 2>/dev/null || true
echo '🔨 Building backend with latest changes...'
docker compose -f docker-compose.prod.yml build --no-cache arc-backend
echo '🚀 Starting services...'
docker compose -f docker-compose.prod.yml up -d
echo '⏳ Waiting for startup...'
sleep 5
echo ''
echo '📊 Container Status:'
docker compose -f docker-compose.prod.yml ps
echo ''
echo '🏥 Testing health endpoint...'
curl -s http://localhost:8080/health || echo 'Health check waiting...'
echo ''
echo '📜 Recent logs:'
docker compose -f docker-compose.prod.yml logs --tail=30 arc-backend
"@

ssh "${remoteUser}@${remoteHost}" $deployCommands

Write-Host "`n==========================================" -ForegroundColor Green
Write-Host "🎉 Deployment Complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

Write-Host "`nNext: Run test script to verify:" -ForegroundColor Cyan
Write-Host "  .\test-backend-production.ps1" -ForegroundColor White
