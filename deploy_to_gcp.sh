#!/bin/bash
# 97layerOS GCP Deployment Script
# Ver 3.0 - Clean Architecture
# Usage: ./deploy_to_gcp.sh [GCP_VM_NAME] [GCP_ZONE]

set -e  # Exit on error

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
GCP_VM_NAME=${1:-97layer-vm}
GCP_ZONE=${2:-us-west1-b}
REMOTE_DIR="97layerOS"

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}97layerOS GCP Deployment${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "VM: $GCP_VM_NAME"
echo "Zone: $GCP_ZONE"
echo ""

# Step 1: Verify local structure
echo -e "${YELLOW}[1/6] Verifying local structure...${NC}"
if [ ! -d "core" ] || [ ! -f "requirements.txt" ]; then
    echo -e "${RED}Error: Invalid directory structure. Run from 97layerOS root.${NC}"
    exit 1
fi
echo "✅ Structure OK"
echo ""

# Step 2: Create backup on remote
echo -e "${YELLOW}[2/6] Creating remote backup...${NC}"
gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
    if [ -d ~/$REMOTE_DIR ]; then
        cd ~/$REMOTE_DIR
        timestamp=\$(date +%Y%m%d_%H%M%S)
        tar -czf ~/97layer-backup-\$timestamp.tar.gz . 2>/dev/null || true
        echo \"✅ Backup created: ~/97layer-backup-\$timestamp.tar.gz\"
    else
        echo \"✅ No existing installation, skipping backup\"
    fi
"
echo ""

# Step 3: Sync code
echo -e "${YELLOW}[3/6] Syncing code to GCP VM...${NC}"
rsync -avz --delete \
    --exclude='.venv' \
    --exclude='__pycache__' \
    --exclude='*.pyc' \
    --exclude='.git' \
    --exclude='archive' \
    --exclude='logs' \
    --exclude='.DS_Store' \
    --exclude='.infra/cache' \
    --exclude='.infra/tmp' \
    . $GCP_VM_NAME:~/$REMOTE_DIR/
echo "✅ Code synced"
echo ""

# Step 4: Install dependencies
echo -e "${YELLOW}[4/6] Installing Python dependencies...${NC}"
gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
    cd ~/$REMOTE_DIR

    # Create venv if not exists
    if [ ! -d .venv ]; then
        python3.11 -m venv .venv
        echo \"✅ Virtual environment created\"
    fi

    # Install/update dependencies
    source .venv/bin/activate
    pip install --upgrade pip -q
    pip install -r requirements.txt -q

    # Verify imports
    python3 -c 'from core.daemons.telegram_secretary import TelegramSecretary' && echo \"✅ Dependencies installed\"
"
echo ""

# Step 5: Setup systemd service
echo -e "${YELLOW}[5/6] Setting up systemd service...${NC}"
gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
    # Get username
    username=\$(whoami)

    # Replace placeholder in service file
    sed \"s/USERNAME_PLACEHOLDER/\$username/g\" ~/$REMOTE_DIR/.infra/systemd/97layer-telegram.service > /tmp/97layer-telegram.service

    # Install service
    sudo mv /tmp/97layer-telegram.service /etc/systemd/system/97layer-telegram.service
    sudo systemctl daemon-reload

    echo \"✅ Systemd service configured\"
"
echo ""

# Step 6: Restart service
echo -e "${YELLOW}[6/6] Restarting Telegram bot...${NC}"
gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
    # Create logs directory
    mkdir -p ~/$REMOTE_DIR/logs

    # Enable and restart service
    sudo systemctl enable 97layer-telegram 2>/dev/null || true
    sudo systemctl restart 97layer-telegram

    # Wait a moment
    sleep 2

    # Check status
    if sudo systemctl is-active --quiet 97layer-telegram; then
        echo \"✅ Telegram bot is running\"
    else
        echo \"❌ Service failed to start. Check logs:\"
        sudo systemctl status 97layer-telegram --no-pager -l
        exit 1
    fi
"
echo ""

# Final status
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Deployment complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Next steps:"
echo "1. Test bot: Send /status to your Telegram bot"
echo "2. View logs: gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command='journalctl -u 97layer-telegram -f'"
echo "3. Monitor memory: gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command='free -h'"
echo ""
