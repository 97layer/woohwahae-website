#!/bin/bash
# 97LAYER OS Cloud Provisioning Script (GCP/Ubuntu)
# 이 스크립트는 구글 클라우드 서버에서 97layerOS 환경을 최단 시간에 구축합니다.

echo "◈ 97LAYER OS: Cloud Setup Starting..."

# 1. 시스템 업데이트 및 필수 패키지 설치
sudo apt update && sudo apt upgrade -y
sudo apt install -y python3-pip python3-venv git curl rclone

# 2. 프로젝트 디렉토리 준비 (사용자 홈 디렉토리)
mkdir -p ~/97layerOS
cd ~/97layerOS

# 3. 가상환경 생성
python3 -m venv .venv
source .venv/bin/activate

# 4. 필수 라이브러리 설치 (최소 요구사항)
pip install google-generativeai python-dotenv

# 5. rclone 설정 가이드 (Google Drive 연동)
echo "◈ 구글 드라이브 연동을 위해 'rclone config'를 실행해야 합니다."
echo "◈ 원격 이름은 'gdrive'로 설정하십시오."

# 6. systemd 서비스 등록 (자동 재시작 설정을 위한 템플릿 파일 생성)
cat <<EOF > 97layer_telegram.service
[Unit]
Description=97LAYER OS Telegram Daemon
After=network.target

[Service]
ExecStart=$(pwd)/.venv/bin/python $(pwd)/execution/telegram_daemon.py
WorkingDirectory=$(pwd)
Restart=always
User=$(whoami)
Environment=PYTHONPATH=$(pwd)

[Install]
WantedBy=multi-user.target
EOF

echo "◈ 준비 완료. 맥북의 97layerOS 폴더를 이 서버로 복제한 후,"
echo "◈ 'sudo cp 97layer_telegram.service /etc/systemd/system/' 를 실행하십시오."
