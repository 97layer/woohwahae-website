# 97layerOS GCP Deployment Guide

> **목적**: Clean Architecture (Ver 3.0) GCP VM 배포 가이드
> **대상**: Google Cloud e2-micro (Always Free Tier)
> **갱신**: 2026-02-16

---

## 🎯 Deployment Overview

### System Requirements
- **GCP VM**: e2-micro (0.25-2 vCPU, 1GB RAM, 30GB disk)
- **OS**: Debian 12 or Ubuntu 22.04 LTS
- **Python**: 3.11+
- **Podman**: 4.0+ (optional, for container isolation)

### Resource Usage (Expected)
```
Telegram Bot: ~100MB RAM (24/7)
Total Usage: ~150MB / 1GB (85% free)
```

---

## 📦 Step 1: GCP VM Setup

### 1.1. Create e2-micro VM (if not exists)
```bash
# On local machine (gcloud CLI)
gcloud compute instances create 97layer-vm \
  --zone=us-west1-b \
  --machine-type=e2-micro \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --boot-disk-size=30GB \
  --tags=http-server,https-server
```

### 1.2. SSH into VM
```bash
gcloud compute ssh 97layer-vm --zone=us-west1-b
```

---

## 🔧 Step 2: VM Environment Setup

### 2.1. Install Python 3.11+
```bash
sudo apt update
sudo apt install -y python3.11 python3.11-venv python3-pip git
```

### 2.2. Install Podman (optional)
```bash
sudo apt install -y podman
```

### 2.3. Create project directory
```bash
mkdir -p ~/97layerOS
cd ~/97layerOS
```

---

## 📥 Step 3: Deploy Code

### 3.1. Clone from Git (recommended)
```bash
cd ~/97layerOS
git clone <YOUR_GIT_REPO_URL> .
```

### 3.2. Or: rsync from local machine
```bash
# On local machine
rsync -avz --exclude='.venv' --exclude='__pycache__' \
  --exclude='.git' --exclude='archive' \
  ~/97layerOS/ 97layer-vm:~/97layerOS/
```

### 3.3. Verify structure
```bash
ls -la ~/97layerOS
# Should see: core/, directives/, knowledge/, requirements.txt
```

---

## 🐍 Step 4: Python Environment

### 4.1. Create virtual environment
```bash
cd ~/97layerOS
python3.11 -m venv .venv
source .venv/bin/activate
```

### 4.2. Install dependencies
```bash
pip install --upgrade pip
pip install -r requirements.txt
```

### 4.3. Verify imports
```bash
python3 -c "from core.daemons.telegram_secretary import TelegramSecretary; print('✅ Imports OK')"
```

---

## 🔐 Step 5: Environment Variables

### 5.1. Create .env file
```bash
cd ~/97layerOS
nano .env
```

### 5.2. Add credentials
```bash
# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here

# Claude API (for Creative Director only)
ANTHROPIC_API_KEY=your_claude_api_key_here

# Gemini API (free tier)
GOOGLE_API_KEY=your_gemini_api_key_here

# NotebookLM (copy from macOS)
# Already configured via notebooklm-mcp-cli login
```

### 5.3. Set permissions
```bash
chmod 600 .env
```

---

## 📋 Step 6: NotebookLM MCP Setup

### 6.1. Install NotebookLM CLI
```bash
source .venv/bin/activate
pip install notebooklm-mcp-cli==0.3.2
```

### 6.2. Copy credentials from macOS
```bash
# On local machine
scp -r ~/.notebooklm-mcp-cli/ 97layer-vm:~/.notebooklm-mcp-cli/

# On GCP VM, verify
nlm notebook list
# Should show existing notebooks
```

---

## 🚀 Step 7: Start Telegram Bot

### 7.1. Test run (foreground)
```bash
cd ~/97layerOS
source .venv/bin/activate
python3 core/daemons/telegram_secretary.py
```

### 7.2. Verify bot responds to /status

---

## 🔄 Step 8: Systemd Service (24/7 operation)

### 8.1. Create systemd service
```bash
sudo nano /etc/systemd/system/97layer-telegram.service
```

### 8.2. Service configuration
```ini
[Unit]
Description=97layerOS Telegram Executive Secretary
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/home/YOUR_USERNAME/97layerOS
Environment="PATH=/home/YOUR_USERNAME/97layerOS/.venv/bin"
ExecStart=/home/YOUR_USERNAME/97layerOS/.venv/bin/python3 core/daemons/telegram_secretary.py
Restart=always
RestartSec=10

# Logging
StandardOutput=append:/home/YOUR_USERNAME/97layerOS/logs/telegram.log
StandardError=append:/home/YOUR_USERNAME/97layerOS/logs/telegram.error.log

[Install]
WantedBy=multi-user.target
```

### 8.3. Enable and start service
```bash
# Replace YOUR_USERNAME with actual username
sudo systemctl daemon-reload
sudo systemctl enable 97layer-telegram
sudo systemctl start 97layer-telegram
```

### 8.4. Check status
```bash
sudo systemctl status 97layer-telegram
```

### 8.5. View logs
```bash
# Real-time logs
journalctl -u 97layer-telegram -f

# Or check log files
tail -f ~/97layerOS/logs/telegram.log
```

---

## 📊 Step 9: Monitoring & Verification

### 9.1. Check bot status
```bash
# Send Telegram message: /status
# Should respond with system status
```

### 9.2. Monitor memory usage
```bash
# Check VM memory
free -h

# Check bot process
ps aux | grep telegram_secretary
```

### 9.3. Monitor API costs
```bash
# Check logs for Claude API calls
grep "Claude API" ~/97layerOS/logs/telegram.log | wc -l

# Creative Director should be < 300 calls/month for $10 budget
```

---

## 🛠️ Maintenance Commands

### Start/Stop/Restart
```bash
sudo systemctl start 97layer-telegram
sudo systemctl stop 97layer-telegram
sudo systemctl restart 97layer-telegram
```

### View logs
```bash
journalctl -u 97layer-telegram -n 100  # Last 100 lines
journalctl -u 97layer-telegram --since "1 hour ago"
```

### Update code
```bash
cd ~/97layerOS
git pull
sudo systemctl restart 97layer-telegram
```

### Debugging
```bash
# Stop service
sudo systemctl stop 97layer-telegram

# Run manually in foreground
cd ~/97layerOS
source .venv/bin/activate
python3 core/daemons/telegram_secretary.py
```

---

## 🚨 Troubleshooting

### Bot not responding
1. Check systemd status: `sudo systemctl status 97layer-telegram`
2. Check logs: `journalctl -u 97layer-telegram -n 50`
3. Verify .env: `cat .env | grep TELEGRAM_BOT_TOKEN`
4. Test imports: `python3 -c "from core.daemons.telegram_secretary import TelegramSecretary"`

### Memory issues (OOM)
1. Check memory: `free -h`
2. Reduce bot features if needed
3. Consider upgrading to e2-small (but costs $)

### NotebookLM not working
1. Check credentials: `nlm notebook list`
2. Re-login if expired: `nlm login`
3. Copy new credentials from macOS

### Claude API errors
1. Check API key: `echo $ANTHROPIC_API_KEY`
2. Verify balance: Check Anthropic dashboard
3. Rate limit: Should be < 10 calls/hour for CD

---

## 💰 Cost Monitoring

### Expected Costs
- **GCP VM**: $0/month (Always Free e2-micro)
- **Claude API**: ~$10/month (Creative Director only)
- **Gemini API**: $0/month (Free with Pro subscription)
- **Total**: ~$10/month ✅

### Usage Limits (to stay under $10)
- **Claude Sonnet 4.5**: < 300 calls/month (~10 calls/day)
- **Creative Director role**: Final decisions only
- **Other agents**: Use Gemini (free)

### Monitor usage
```bash
# Count Claude API calls today
grep "Creative Director" ~/97layerOS/logs/telegram.log | grep "$(date +%Y-%m-%d)" | wc -l
```

---

## 🎯 Next Steps After Deployment

1. **Test all Telegram commands**:
   - /status, /report, /analyze, /signal, /youtube

2. **Verify integrations**:
   - NotebookLM MCP
   - Google Drive sync
   - Asset tracking

3. **Monitor for 1 week**:
   - Memory usage (should stay < 200MB)
   - API costs (should stay < $10/month)
   - Bot stability (uptime %)

4. **Decide on Phase 6**:
   - If stable → Proceed with VM Ecosystem
   - If issues → Fix before expansion

---

## 📝 Rollback Plan

If deployment fails:

```bash
# 1. Stop service
sudo systemctl stop 97layer-telegram

# 2. Restore from backup (if needed)
cd ~/97layerOS
tar -xzf archive/2026-02-pre-refactor/backup_20260216_020059.tar.gz

# 3. Or git reset
git reset --hard <previous_commit>

# 4. Restart service
sudo systemctl start 97layer-telegram
```

---

> **슬로우 라이프 원칙**: 배포는 급하지 않게, 각 단계 검증 후 진행. 문제 발생 시 롤백 후 재시도.
