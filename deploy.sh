#!/bin/bash
#
# 97layerOS One-Command Deployment Script
# Usage: ./deploy.sh
#
# This script:
# 1. Creates deployment package
# 2. Uploads to GCP VM via SCP
# 3. Extracts and restarts service
# 4. Verifies deployment
#
# Prerequisites:
# - SSH key configured for GCP VM
# - ~/.ssh/config entry for 97layer-vm
#

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VM_IP="34.136.109.201"
VM_USER="skyto5339_gmail_com"
VM_KEY="${HOME}/.ssh/google_compute_engine"
VM_HOST="${VM_USER}@${VM_IP}"
VM_PATH="/home/${VM_USER}/97layerOS"
DEPLOY_PACKAGE="97layer-deploy-$(date +%Y%m%d-%H%M%S).tar.gz"

# SSH Command Alias
SSH_CMD="ssh -i ${VM_KEY} -o ConnectTimeout=10"
SCP_CMD="scp -i ${VM_KEY} -q"

echo -e "${BLUE}🚀 97layerOS One-Command Deployment (Deep RAG V6)${NC}"
echo -e "${BLUE}====================================${NC}\n"

# Step 1: Check SSH connection
echo -e "${YELLOW}[1/6]${NC} Checking SSH connection to ${VM_IP}..."
if ${SSH_CMD} ${VM_HOST} "echo 'SSH OK'" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ SSH connection OK${NC}\n"
else
    echo -e "${RED}❌ SSH connection failed to ${VM_IP}${NC}"
    echo -e "${YELLOW}Troubleshooting:${NC}"
    echo "1. Verify VM is RUNNING in GCP Console"
    echo "2. Check if IP ${VM_IP} is still correct"
    echo "3. Ensure ${VM_KEY} exists"
    exit 1
fi

# Step 2: Create deployment package
echo -e "${YELLOW}[2/6]${NC} Creating deployment package..."
cd "${PROJECT_ROOT}"

tar -czf "/tmp/${DEPLOY_PACKAGE}" \
    core/ \
    directives/ \
    knowledge/docs/ \
    knowledge/agent_hub/ \
    requirements.txt \
    start_telegram.sh \
    start_monitor.sh \
    deployment/ \
    --exclude='*.pyc' \
    --exclude='__pycache__' \
    --exclude='.git' \
    --exclude='.env' \
    2>&1 | grep -v "Ignoring unknown extended header" || true

PACKAGE_SIZE=$(du -h "/tmp/${DEPLOY_PACKAGE}" | cut -f1)
echo -e "${GREEN}✅ Package created: ${DEPLOY_PACKAGE} (${PACKAGE_SIZE})${NC}\n"

# Step 3: Upload to VM
echo -e "${YELLOW}[3/6]${NC} Uploading to GCP VM..."
${SCP_CMD} "/tmp/${DEPLOY_PACKAGE}" ${VM_HOST}:~/
echo -e "${GREEN}✅ Upload complete${NC}\n"

# Step 4: Extract on VM
echo -e "${YELLOW}[4/6]${NC} Extracting on VM..."
${SSH_CMD} ${VM_HOST} << 'ENDSSH'
    set -e
    cd ~/97layerOS

    # Backup old version
    if [ -d "core/" ]; then
        echo "  📦 Backing up current version..."
        tar -czf ~/97layer-backup-$(date +%Y%m%d-%H%M%S).tar.gz core/ directives/ knowledge/ 2>/dev/null || true
    fi

    # Extract new version
    echo "  📂 Extracting new version..."
    tar -xzf ~/*-deploy-*.tar.gz 2>&1 | grep -v "Ignoring unknown extended header" || true

    # Clean up
    rm ~/*-deploy-*.tar.gz

    echo "  ✅ Extraction complete"
ENDSSH
echo -e "${GREEN}✅ Extraction complete${NC}\n"

# Step 5: Restart service
echo -e "${YELLOW}[5/6]${NC} Restarting Telegram bot service..."
${SSH_CMD} ${VM_HOST} << 'ENDSSH'
    set -e
    sudo sed -i 's/telegram_secretary_v4.py/telegram_secretary_v6.py/g' /etc/systemd/system/97layer-telegram.service
    sudo sed -i 's/telegram_secretary_v5.py/telegram_secretary_v6.py/g' /etc/systemd/system/97layer-telegram.service
    sudo systemctl daemon-reload
    sudo systemctl restart 97layer-telegram
    sleep 2
    sudo systemctl status 97layer-telegram --no-pager | head -10
ENDSSH
echo -e "${GREEN}✅ Service restarted${NC}\n"

# Step 6: Verify deployment
echo -e "${YELLOW}[6/6]${NC} Verifying deployment..."
sleep 3

VERIFICATION=$(${SSH_CMD} ${VM_HOST} << 'ENDSSH'
    # Check if bot is running
    if pgrep -f "telegram_secretary" > /dev/null; then
        echo "bot_running=true"
    else
        echo "bot_running=false"
    fi

    # Check service status
    if systemctl is-active --quiet 97layer-telegram; then
        echo "service_active=true"
    else
        echo "service_active=false"
    fi

    # Get memory usage
    ps aux | grep telegram_secretary | grep -v grep | awk '{print "memory_mb="$6/1024}' | head -1
ENDSSH
)

BOT_RUNNING=$(echo "$VERIFICATION" | grep "bot_running" | cut -d'=' -f2)
SERVICE_ACTIVE=$(echo "$VERIFICATION" | grep "service_active" | cut -d'=' -f2)
MEMORY_MB=$(echo "$VERIFICATION" | grep "memory_mb" | cut -d'=' -f2)

if [ "$BOT_RUNNING" = "true" ] && [ "$SERVICE_ACTIVE" = "true" ]; then
    echo -e "${GREEN}✅ Deployment successful!${NC}"
    echo -e "   🤖 Bot: Running"
    echo -e "   ⚙️  Service: Active"
    if [ -n "$MEMORY_MB" ]; then
        echo -e "   💾 Memory: ${MEMORY_MB}MB"
    fi
else
    echo -e "${RED}⚠️  Deployment completed but bot may not be running${NC}"
    echo -e "   🤖 Bot: ${BOT_RUNNING}"
    echo -e "   ⚙️  Service: ${SERVICE_ACTIVE}"
    echo -e "\n${YELLOW}Check logs:${NC}"
    echo -e "   ${SSH_CMD} ${VM_HOST} 'sudo journalctl -u 97layer-telegram -n 50'"
fi

# Clean up local package
rm "/tmp/${DEPLOY_PACKAGE}"

echo -e "\n${BLUE}====================================${NC}"
echo -e "${GREEN}🎉 Deployment complete!${NC}"
echo -e "${BLUE}====================================${NC}\n"

echo -e "${YELLOW}Next steps:${NC}"
echo "1. Test bot in Telegram: Send a message or /status"
echo "2. Check logs: ${SSH_CMD} ${VM_HOST} 'sudo journalctl -u 97layer-telegram -f'"
echo "3. Monitor: ${SSH_CMD} ${VM_HOST} 'python3 97layerOS/core/system/monitor_dashboard.py'"
