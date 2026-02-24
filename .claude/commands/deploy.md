---
description: GCP VM 배포 — 전체 or 특정 파일/서비스 타겟
---

# /deploy — VM 배포

`skills/deploy/SKILL.md` 규칙에 따라 배포 실행.

## 사용법

```
/deploy              → 전체 배포 (scripts/deploy/deploy.sh)
/deploy ecosystem    → ecosystem 서비스 재시작만
/deploy telegram     → telegram 서비스 재시작만
/deploy gardener     → gardener 서비스 재시작만
/deploy backend      → woohwahae-backend 재시작만
/deploy core/agents/gardener.py  → 단일 파일 scp + 재시작
```

## 실행

`$ARGUMENTS` 파싱 후 `skills/deploy/SKILL.md` 흐름 실행.
