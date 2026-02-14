#!/bin/bash
# GCP 배포 스크립트

echo "🚀 GCP 배포 준비..."
tar --exclude='.git' --exclude='__pycache__' -czf /tmp/deploy.tar.gz .
echo "✅ 배포 파일 준비 완료: /tmp/deploy.tar.gz"
echo ""
echo "📋 GCP 웹 콘솔에서 실행:"
echo "wget http://YOUR_MAC_IP:8000/deploy.tar.gz"
echo "tar -xzf deploy.tar.gz"
echo "./start_system.sh"
