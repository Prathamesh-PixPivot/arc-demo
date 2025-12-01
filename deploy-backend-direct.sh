#!/bin/bash

set -e  # Exit on error

echo "🚀 Deploying Backend with License Fix..."
echo "=========================================="

# Configuration
REMOTE_USER="arc"
REMOTE_HOST="arc-demo.thepixpivot.com"
REMOTE_DIR="/home/arc/apps/backend"
LOCAL_DIR="./apps/backend"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "\n${YELLOW}Step 1: Syncing modified files to server...${NC}"

# Transfer only the modified Go files and critical files
echo "Transferring main.go (license fix)..."
rsync -avz "${LOCAL_DIR}/cmd/server/main.go" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/cmd/server/"

echo "Transferring sso_handler.go (rewrite)..."
rsync -avz "${LOCAL_DIR}/internal/api/handlers/sso_handler.go" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/internal/api/handlers/"

echo "Transferring signup_handler.go (atomic transactions)..."
rsync -avz "${LOCAL_DIR}/internal/api/handlers/signup_handler.go" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/internal/api/handlers/"

echo "Transferring auth_handler.go (verification checks)..."
rsync -avz "${LOCAL_DIR}/internal/api/handlers/auth_handler.go" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/internal/api/handlers/"

echo "Transferring email_service.go (retry logic)..."
rsync -avz "${LOCAL_DIR}/internal/core/services/email_service.go" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/internal/core/services/"

echo "Transferring .env.production..."
rsync -avz "${LOCAL_DIR}/.env.production" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/.env"

echo -e "${GREEN}✅ Files transferred${NC}"

echo -e "\n${YELLOW}Step 2: Building and restarting backend...${NC}"

ssh "${REMOTE_USER}@${REMOTE_HOST}" << 'ENDSSH'
cd /home/arc/apps/backend

echo "Stopping existing containers..."
docker compose -f docker-compose.prod.yml down 2>/dev/null || true

echo "Building backend with latest changes..."
docker compose -f docker-compose.prod.yml build --no-cache arc-backend

echo "Starting services..."
docker compose -f docker-compose.prod.yml up -d

echo "Waiting for services to start..."
sleep 5

echo ""
echo "Container Status:"
docker compose -f docker-compose.prod.yml ps

echo ""
echo "Testing health endpoint..."
sleep 2
curl -s http://localhost:8080/health || echo "Health check failed"

echo ""
echo "Recent logs:"
docker compose -f docker-compose.prod.yml logs --tail=30 arc-backend

ENDSSH

echo -e "\n${GREEN}==================================${NC}"
echo -e "${GREEN}🎉 Deployment Complete!${NC}"
echo -e "${GREEN}==================================${NC}"
echo ""
echo "Next steps:"
echo "1. Run test script: .\test-backend-production.ps1"
echo "2. Configure SMTP in .env if needed"
echo "3. Test authentication flows"
