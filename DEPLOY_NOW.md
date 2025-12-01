# Quick Deployment Commands

## For immediate deployment of license fix:

### If using Docker Compose:
```bash
cd d:\arc-demo\apps\backend
docker-compose down
docker-compose build
docker-compose up -d
```

### If using Kubernetes:
```bash
cd d:\arc-demo
# Build and push new image
docker build -t your-registry/arc-backend:latest ./apps/backend
docker push your-registry/arc-backend:latest

# Force K8s to pull new image
kubectl rollout restart deployment/arc-backend
kubectl rollout status deployment/arc-backend
```

### If running on remote server:
```bash
# SSH to server
ssh user@arc-demo.thepixpivot.com

# Pull latest code
cd /path/to/arc-demo/apps/backend
git pull origin main

# Rebuild
go build -o bin/server cmd/server/main.go

# Restart
pm2 restart arc-backend
# OR
sudo systemctl restart arc-backend
```

## Verify Fix Applied:
```powershell
# Run this from your local machine
cd d:\arc-demo
.\test-backend-production.ps1
```

**Success = Status codes are 401, 403, 503, or 201 (NOT 402)**

## If Still Getting 402 Errors:
1. Check if new code is actually deployed (verify pod/container age)
2. Ensure environment variables are loaded
3. Check logs for startup errors
4. Force restart deployment
