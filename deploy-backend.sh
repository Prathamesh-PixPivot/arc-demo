#!/bin/bash

# ARC Backend Deployment Script
# This script deploys the backend to the server

set -e  # Exit on error

echo "🚀 ARC Backend Deployment Script"
echo "=================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
REMOTE_USER="arc"
REMOTE_HOST="arc"
REMOTE_DIR="/home/arc/arc-backend"
LOCAL_DIR="./apps/backend"

# Step 1: Check if .env exists
echo -e "\n${YELLOW}Step 1: Checking environment configuration...${NC}"
if [ ! -f "${LOCAL_DIR}/.env.production" ]; then
    echo -e "${YELLOW}⚠️  .env.production not found!${NC}"
    echo "Creating from template..."
    cp "${LOCAL_DIR}/.env.production.example" "${LOCAL_DIR}/.env.production"
    echo -e "${RED}❌ Please configure .env.production with your actual values before deploying!${NC}"
    exit1
fi
echo -e "${GREEN}✅ Environment configuration found${NC}"

# Step 2: Transfer files via rsync
echo -e "\n${YELLOW}Step 2: Transferring files to server...${NC}"

# Sync Backend
rsync -avz --progress \
    --exclude 'node_modules' \
    --exclude '.git' \
    --exclude 'server.exe' \
    --exclude 'logs/*' \
    --exclude '.env' \
    "${LOCAL_DIR}/" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

# Sync Frontend
echo "Syncing frontend..."
rsync -avz --progress \
    --exclude 'node_modules' \
    --exclude '.next' \
    --exclude '.git' \
    --exclude '.env*.local' \
    "./apps/web/" "${REMOTE_USER}@${REMOTE_HOST}:/home/arc/apps/web/"

echo -e "${GREEN}✅ Files transferred${NC}"

# Step 3: Transfer .env.production as .env
echo -e "\n${YELLOW}Step 3: Deploying production environment...${NC}"
scp "${LOCAL_DIR}/.env.production" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/.env"
echo -e "${GREEN}✅ Environment deployed${NC}"

# Step 4: Build and start containers
echo -e "\n${YELLOW}Step 4: Building and starting containers...${NC}"
ssh "${REMOTE_USER}@${REMOTE_HOST}" << 'ENDSSH'
mkdir -p ~/apps/backend
mkdir -p ~/apps/web

# Move files if they were uploaded to wrong place (for manual run safety)
# But rsync should have handled it if we updated it. 
# Since we are running manually via ssh command in this turn, we will handle it there.
# This script update is for future use.

cd ~/apps/backend

# Stop existing containers (might be running in old location ~/arc-backend, check that too)
if [ -d "~/arc-backend" ]; then
    echo "Stopping containers in old location..."
    cd ~/arc-backend
    docker compose -f docker-compose.prod.yml down || true
    cd ~/apps/backend
fi

# Stop containers in current location
echo "Stopping existing containers..."
docker compose -f docker-compose.prod.yml down || true

# Pull latest images
echo "Pulling latest base images..."
docker compose -f docker-compose.prod.yml pull

# Build new images
echo "Building backend and frontend images..."
docker compose -f docker-compose.prod.yml build --no-cache arc-backend arc-frontend

# Start services
echo "Starting all services..."
docker compose -f docker-compose.prod.yml up -d

# Show running containers
echo ""
echo "Running containers:"
docker ps

echo ""
echo "Waiting for services to be healthy..."
sleep 10

# Check health
docker compose -f docker-compose.prod.yml ps

ENDSSH

echo -e "${GREEN}✅ Containers started${NC}"

# Step 5: Show logs
echo -e "\n${YELLOW}Step 5: Showing recent logs...${NC}"
ssh "${REMOTE_USER}@${REMOTE_HOST}" "cd ~/apps/backend && docker compose -f docker-compose.prod.yml logs --tail=50 arc-backend arc-frontend"

# Final message
echo -e "\n${GREEN}=================================="
echo "🎉 Deployment Complete!"
echo -e "==================================${NC}"
echo ""
echo "Service URLs:"
echo "  - Frontend: http://SERVER_IP:3000"
echo "  - API: http://SERVER_IP:8080"
echo "  - Health: http://SERVER_IP:8080/health"
echo "  - Prometheus: http://SERVER_IP:9090"
echo "  - Grafana: http://SERVER_IP:3000"
echo ""
echo "To view logs:"
echo "  ssh ${REMOTE_USER}@${REMOTE_HOST} 'cd ~/apps/backend && docker compose -f docker-compose.prod.yml logs -f arc-backend arc-frontend'"
echo ""
echo "To stop services:"
echo "  ssh ${REMOTE_USER}@${REMOTE_HOST} 'cd ~/apps/backend && docker compose -f docker-compose.prod.yml down'"
