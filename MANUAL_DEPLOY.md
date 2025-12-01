# Manual Deployment Instructions

## Issue
SSH connection keeps timing out. Automated deployment not working.

## Solution
Manually copy files to server via your existing SSH session.

## Files to Deploy

All these files build successfully locally. Copy them to the server:

### 1. License Fix
**File**: `cmd/server/main.go`
**Location on server**: `/home/arc/apps/backend/cmd/server/main.go`

### 2. SSO Handler Rewrite  
**File**: `internal/api/handlers/sso_handler.go`
**Location on server**: `/home/arc/apps/backend/internal/api/handlers/sso_handler.go`

### 3. Atomic Transactions
**File**: `internal/api/handlers/signup_handler.go`
**Location on server**: `/home/arc/apps/backend/internal/api/handlers/signup_handler.go`

### 4. Login Verification
**File**: `internal/api/handlers/auth_handler.go`
**Location on server**: `/home/arc/apps/backend/internal/api/handlers/auth_handler.go`

### 5. Email Service (FIXED - no Timeout field)
**File**: `internal/core/services/email_service.go`
**Location on server**: `/home/arc/apps/backend/internal/core/services/email_service.go`

## Deployment Steps

### Option 1: Use WinSCP or FileZilla
1. Open WinSCP/FileZilla
2. Connect to `arc-demo.thepixpivot.com` as user `arc`
3. Navigate to `/home/arc/apps/backend`
4. Upload each file to its corresponding location
5. Run rebuild commands (see below)

### Option 2: Copy-Paste in Nano
Since you're already SSH'd into the server:

```bash
cd /home/arc/apps/backend

# For each file - edit and paste content:
nano internal/core/services/email_service.go
# Delete all content (Ctrl+K repeatedly)
# Paste new content from Windows
# Save: Ctrl+X, Y, Enter

# Repeat for other files...
```

## After Files Are Copied

```bash
cd /home/arc/apps/backend

# Rebuild
docker compose -f docker-compose.prod.yml build --no-cache arc-backend

# If build succeeds:
docker compose -f docker-compose.prod.yml up -d

# Verify:
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8080/health
```

## Test After Deployment

From Windows:
```powershell
cd d:\arc-demo
.\test-backend-production.ps1
```

Expected: All endpoints return 401/403/503/201 (NOT 402)
