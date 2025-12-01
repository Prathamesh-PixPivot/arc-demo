# Direct Server Deployment - Copy-Paste Commands

Since SCP keeps timing out, use these commands directly in your SSH session on the server.

## Step 1: Create Temporary Directory

```bash
cd /home/arc
mkdir -p temp-backend-deploy
cd temp-backend-deploy
```

## Step 2: Create Files Using Cat

Run these commands one by one in your server terminal:

### File 1: email_service.go (FIXED - No Timeout field)

```bash
cat > email_service.go << 'EOF'
```

**Then in Windows, open:** `d:\arc-demo\apps\backend\internal\core\services\email_service.go`

**Copy the ENTIRE file contents and paste into the terminal**

**Then type:** `EOF` and press Enter

### File 2-5: Repeat for other files

Same process for:
- `sso_handler.go` 
- `signup_handler.go`
- `auth_handler.go`
- `main.go`

## Step 3: Move Files to Correct Locations

```bash
cd /home/arc/temp-backend-deploy

# Copy to correct locations
cp email_service.go /home/arc/apps/backend/internal/core/services/
cp sso_handler.go /home/arc/apps/backend/internal/api/handlers/
cp signup_handler.go /home/arc/apps/backend/internal/api/handlers/
cp auth_handler.go /home/arc/apps/backend/internal/api/handlers/
cp main.go /home/arc/apps/backend/cmd/server/
```

## Step 4: Rebuild

```bash
cd /home/arc/apps/backend
docker compose -f docker-compose.prod.yml build --no-cache arc-backend
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8080/health
```

## Alternative: I'll Show You File Contents

Would you like me to display the complete content of each file here so you can copy-paste into nano directly on the server?
