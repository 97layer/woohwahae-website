---
name: deploy
description: GCP VM 배포. 전체 또는 특정 파일/서비스만 타겟 배포.
argument-hint: "[all|telegram|ecosystem|gardener|backend|<file_path>]"
user-invocable: true
context: fork
tools:
  - Bash
  - Glob
version: 1.0.0
updated: 2026-02-24
---

# Deploy Skill

로컬 코드 → GCP VM 배포. 인자에 따라 범위 결정.

## 인자별 동작

| 인자 | 동작 |
|------|------|
| (없음) / `all` | 전체 배포 (`core/scripts/deploy/deploy.sh`) |
| `telegram` | 97layer-telegram 서비스만 재시작 |
| `ecosystem` | 97layer-ecosystem 서비스만 재시작 |
| `gardener` | 97layer-gardener 서비스만 재시작 |
| `backend` | woohwahae-backend 서비스만 재시작 |
| `<file_path>` | 단일 파일 scp 후 관련 서비스 재시작 |

## 환경 상수

```
VM_HOST: skyto5339_gmail_com@136.109.201.201
VM_PATH: /home/skyto5339_gmail_com/97layerOS
SSH_KEY: ~/.ssh/google_compute_engine
SSH: ssh -i ~/.ssh/google_compute_engine -o StrictHostKeyChecking=no
SCP: scp -i ~/.ssh/google_compute_engine -o StrictHostKeyChecking=no
```

## 실행 흐름

### 전체 배포
```bash
bash core/scripts/deploy/deploy.sh
```

### 서비스 재시작만
```bash
ssh -i ~/.ssh/google_compute_engine -o StrictHostKeyChecking=no \
    skyto5339_gmail_com@136.109.201.201 \
    "sudo systemctl restart 97layer-<service> && systemctl is-active 97layer-<service>"
```

### 단일 파일 scp + 재시작
1. 로컬 경로에서 파일 확인
2. VM 경로 매핑: 로컬 `/Users/97layer/97layerOS/<rel>` → VM `/home/skyto5339_gmail_com/97layerOS/<rel>`
3. scp 전송
4. 파일 경로로 서비스 추론:
   - `core/daemons/telegram_*` → 97layer-telegram
   - `core/agents/*` or `core/system/*` → 97layer-ecosystem
   - `core/agents/gardener*` → 97layer-gardener
   - `website/backend/*` → woohwahae-backend
   - 기타 → 97layer-ecosystem (기본값)
5. 서비스 재시작 + 상태 확인

## 배포 후 필수 확인

```bash
ssh -i ~/.ssh/google_compute_engine skyto5339_gmail_com@136.109.201.201 \
    "systemctl is-active 97layer-telegram 97layer-ecosystem 97layer-gardener"
```

## 핵심 규칙

- 배포 전 로컬 변경사항 커밋 여부 확인 (git status)
- .env 파일 절대 scp 금지
- 배포 후 반드시 서비스 상태 출력
- 실패 시 마지막 10줄 로그 출력: `journalctl -u <service> -n 10 --no-pager`
