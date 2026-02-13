# GCP Migration Action Guide

사용자님이 구글 클라우드 서버(GCP)에서 직접 수행해주셔야 할 핵심 단계입니다.

## 1. 인스턴스 접속

GCP 콘솔에서 생성한 인스턴스의 **[SSH]** 버튼을 클릭하여 터미널을 엽니다.

## 2. 기초 환경 구축 (복사/붙여넣기)

터미널에 아래 명령어를 복사하여 실행하십시오. (자동으로 97layerOS 폴더 구조와 필수 라이브러리를 설치합니다.)

```bash
curl -sSL https://raw.githubusercontent.com/사용자계정/97layerOS/main/execution/setup_cloud.sh | bash
```

*(참고: 아직 GitHub에 코드가 없다면, 맥북 터미널에서 `scp -r ~/97layerOS [GCP_IP]:~/` 명령어로 직접 전송할 수도 있습니다. 어려우시다면 말씀해 주세요.)*

## 3. 구글 드라이브 연동 (rclone)

서버 터미널에서 다음을 입력하고 안내에 따라 'gdrive'라는 이름으로 연동합니다.

```bash
rclone config
```

## 4. 서비스 등록 (자동 재시작)

모든 파일이 서버의 `~/97layerOS`에 위치하고 rclone 설정이 끝났다면:

```bash
sudo cp ~/97layerOS/97layer_telegram.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable 97layer_telegram
sudo systemctl start 97layer_telegram
```

---
**주의**: 현재 `telegram_daemon.py`는 `task_status.json`에 의존하고 있으므로, 맥북의 최초 상태 파일을 반드시 함께 서버로 옮겨야 합니다.
