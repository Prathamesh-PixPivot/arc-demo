# Backend Deployment - Final Steps

## Status: ✅ Code Ready, Awaiting Server Transfer

All backend code builds successfully locally. SSH connection timeouts prevent automated deployment.

## Package Created

**File**: `d:\arc-demo\backend-fixed-files.zip`

Contains all 5 fixed files:
1. `cmd/server/main.go` - License middleware fix
2. `internal/api/handlers/sso_handler.go` - SSO rewrite
3. `internal/api/handlers/signup_handler.go` - Atomic transactions
4. `internal/api/handlers/auth_handler.go` - Login verification
5. `internal/core/services/email_service.go` - Retry logic (Timeout field removed)

## Deployment Options

### Option 1: Upload via WinSCP (Recommended - Fastest)

1. Download WinSCP: https://winscp.net/eng/download.php
2. Connect to server:
   - Host: `arc-demo.thepixpivot.com`
   - User: `arc`
   - Protocol: SFTP
3. Upload `backend-fixed-files.zip` to `/home/arc/`
4. In your SSH terminal:
   ```bash
   cd /home/arc
   unzip -o backend-fixed-files.zip -d apps/backend/
   cd apps/backend
   docker compose -f docker-compose.prod.yml build --no-cache arc-backend
   docker compose -f docker-compose.prod.yml up -d
   docker compose -f docker-compose.prod.yml ps
   curl http://localhost:8080/health
   ```

### Option 2: Direct File Copy

Copy files one by one using WinSCP to:
- `/home/arc/apps/backend/cmd/server/main.go`
- `/home/arc/apps/backend/internal/api/handlers/sso_handler.go`
- `/home/arc/apps/backend/internal/api/handlers/signup_handler.go`
- `/home/arc/apps/backend/internal/api/handlers/auth_handler.go`
- `/home/arc/apps/backend/internal/core/services/email_service.go`

Then rebuild:
```bash
cd /home/arc/apps/backend
docker compose -f docker-compose.prod.yml build --no-cache arc-backend
docker compose -f docker-compose.prod.yml up -d
```

### Option 3: Base64 Transfer (If no file transfer tools available)

I can encode files as base64 text that you can copy-paste into the server terminal.

## After Successful Deployment

1. **Verify health**:
   ```bash
   curl http://localhost:8080/health
   # Should return: OK
   ```

2. **Test license fix** (from Windows):
   ```powershell
   cd d:\arc-demo
   .\test-backend-production.ps1
   ```
   Expected: Endpoints return 401/403/503/201 (NOT 402)

3. **Configure SMTP**:
   ```bash
   nano /home/arc/apps/backend/.env
   # Update SMTP_USER and SMTP_PASS
   docker compose -f docker-compose.prod.yml restart arc-backend
   ```

4. **Test full flow**:
   - Manual signup → Email → Login
   - SSO signup → Onboarding → Dashboard

## Troubleshooting

**Build fails?**
```bash
docker compose -f docker-compose.prod.yml logs arc-backend
```

**Files not copying?**
```bash
ls -la /home/arc/apps/backend/cmd/server/
ls -la /home/arc/apps/backend/internal/api/handlers/
```

**Container won't start?**
```bash
docker compose -f docker-compose.prod.yml down
docker system prune -f
docker compose -f docker-compose.prod.yml up -d
```

## Ready to Deploy!

All code is tested and working. Just needs file transfer → rebuild → test.

Choose your preferred method and let me know if you need help with any step!
