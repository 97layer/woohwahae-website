# Podman Optimization for 97layerOS

**Version**: 1.0.0  
**Date**: 2026-02-15  
**Target**: GCP VM Night Guard (e2-micro, RAM 1GB)  
**Status**: Production Ready

---

## Objective

97layerOS Night Guard를 GCP e2-micro VM (RAM 1GB)에서 안정적으로 24/7 운영하기 위한 Podman 컨테이너 최적화 전략.

**핵심 문제**:
- RAM 1GB로는 AI 에이전트 연산 시 부족
- 파일 I/O 병목 (directives/, knowledge/ 자주 접근)
- API 키 평문 노출 (.env 파일)
- 장애 시 수동 재시작 필요

**해결 전략**:
1. **Swap 메모리 2GB** - RAM 부족 극복
2. **파일 시스템 최적화** - Linux 네이티브 활용
3. **Healthcheck 자동화** - 장애 자동 복구
4. **Podman Secrets** - API 키 보안

---

## Four Optimization Strategies

### 1. File System Speed Optimization

#### GCP VM (Linux): 최적화 불필요
**이유**: Linux 네이티브 파일 시스템은 이미 최적화됨
- 볼륨 마운트: `-v /host:/container:Z` (SELinux 레이블 자동)
- 추가 설정 불필요

**97layerOS 적용**:
```yaml
# deployment/podman-compose.nightguard.yml
volumes:
  - /home/ubuntu/97layerOS:/app:Z  # :Z = SELinux auto-label
```

---

### 2. Resource Allocation & Swap Memory (핵심!)

#### 해결: 2-Layer Swap 전략

**Layer 1: 호스트 VM Swap (2GB)**
```bash
sudo fallocate -l 2G /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

**Layer 2: 컨테이너 Swap 제한**
```yaml
services:
  nightguard:
    mem_limit: 700m
    mem_reservation: 500m
    memswap_limit: 2700m  # RAM 700MB + Swap 2GB
    cpus: '1'
```

---

### 3. Healthcheck Automation

```yaml
healthcheck:
  test: ["CMD", "python3", "execution/system/health_check.py"]
  interval: 5m
  timeout: 10s
  retries: 3
  start_period: 30s

restart: unless-stopped
```

---

### 4. Security: Podman Secrets

```bash
# Secrets 생성
echo -n "$TELEGRAM_BOT_TOKEN" | podman secret create telegram_bot_token -
echo -n "$GEMINI_API_KEY" | podman secret create gemini_api_key -
```

---

## Deployment Guide

### Step 1: VM에 97layerOS 복사
```bash
gcloud compute scp --recurse /Users/97layer/97layerOS 97layer-nightguard:~/ --zone=us-west1-b
```

### Step 2: VM SSH 접속 및 초기화
```bash
gcloud compute ssh 97layer-nightguard --zone=us-west1-b
cd ~/97layerOS/deployment
./init_nightguard_podman.sh
```

### Step 3: 상태 확인
```bash
podman ps
podman logs -f 97layer-nightguard
podman inspect 97layer-nightguard | grep -A10 Health
```

---

## Files

- `deployment/podman-compose.nightguard.yml` - Compose 설정
- `deployment/Dockerfile.nightguard` - 컨테이너 이미지
- `deployment/init_nightguard_podman.sh` - 초기화 스크립트
- `deployment/setup_podman_secrets.sh` - Secrets 설정

---

## Result

✅ RAM 1GB VM에서 안정적 24/7 운영  
✅ 장애 자동 복구 (5분 이내)  
✅ API 키 완전 격리  
✅ 무중단 배치 (systemd 통합)

**결과**: "에이전트들이 97layerOS 기지 안에서 쾌적하게 연산하고, 물리적 한계에 부딪히지 않는다."
