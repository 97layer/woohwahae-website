---
name: intelligence_backup
description: 핵심 지능 자산(knowledge/, directives/, core/) 스냅샷 → Google Drive 동기화.
argument-hint: "[full|incremental]"
user-invocable: true
context: fork
tools:
  - Bash
version: 2.1.0
updated: 2026-02-23
---

# Intelligence Backup Skill

핵심 자산을 보존하고 GDrive에 동기화. rclone 기반.

## 동기화 대상 (allowlist)

- knowledge/ (signals, agent_hub, docs, system)
- directives/
- core/ (agents, system 코드)

## 실행 명령

```bash
# 로컬 → GDrive 동기화 (sync_drive.sh 사용)
bash core/scripts/sync_drive.sh

# 수동 스냅샷 (긴급 시)
tar --exclude='.git' --exclude='node_modules' --exclude='__pycache__' \
    -czf /tmp/LAYER OS_$(date +%Y%m%d_%H%M%S).tar.gz \
    knowledge/ directives/ core/ CLAUDE.md .ai_rules
```

## GDrive 경로

- rclone remote: gdrive (인증 완료 2026-02-17)
- 동기화 대상: gdrive:97layerOS/

## 우선순위

1. state.md — 세션 연속성 앵커
2. knowledge/signals/ — 파이프라인 입력 데이터
3. directives/ — 에이전트 행동 지침
4. core/ — 실행 코드

## 주의

- .env, token.json 등 시크릿 파일은 GDrive 동기화 제외
- node_modules, .git, __pycache__ 항상 제외
