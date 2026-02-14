#!/bin/bash
# Filename: install_systemd_services.sh
# Purpose: Install systemd services for 97LAYER daemons on GCP
# Instructions: Run this in GCP browser SSH terminal

echo "🔧 97LAYER Systemd Services Installation"
echo "========================================"

# Stop current nohup processes
echo "1️⃣ Stopping existing processes..."
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2

# Copy service files
echo "2️⃣ Installing service files..."
sudo cp ~/97layerOS/97layer_technical.service /etc/systemd/system/
sudo cp ~/97layerOS/97layer_telegram.service /etc/systemd/system/

# Set proper permissions
sudo chmod 644 /etc/systemd/system/97layer_technical.service
sudo chmod 644 /etc/systemd/system/97layer_telegram.service

# Reload systemd
echo "3️⃣ Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable services (auto-start on boot)
echo "4️⃣ Enabling auto-start on boot..."
sudo systemctl enable 97layer_technical.service
sudo systemctl enable 97layer_telegram.service

# Start services
echo "5️⃣ Starting services..."
sudo systemctl start 97layer_technical.service
sudo systemctl start 97layer_telegram.service

# Wait and check status
sleep 3
echo ""
echo "✅ Installation Complete!"
echo ""
echo "📊 Service Status:"
echo "==================="
sudo systemctl status 97layer_technical.service --no-pager -l
echo ""
sudo systemctl status 97layer_telegram.service --no-pager -l

echo ""
echo "📝 Useful Commands:"
echo "==================="
echo "Check logs:        sudo journalctl -u 97layer_technical.service -f"
echo "                   sudo journalctl -u 97layer_telegram.service -f"
echo "Restart service:   sudo systemctl restart 97layer_technical.service"
echo "Stop service:      sudo systemctl stop 97layer_telegram.service"
echo "Service status:    sudo systemctl status 97layer_technical.service"
