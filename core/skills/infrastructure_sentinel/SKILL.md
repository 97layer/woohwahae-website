---
name: infrastructure_sentinel
description: GCP VM systemd 서비스 상태 모니터링 + 자동 재시작 판단 + 이상 알림.
argument-hint: "[check|restart|status]"
user-invocable: true
context: fork
tools:
  - Bash
version: 2.1.0
updated: 2026-02-23
---

# Infrastructure Sentinel Skill

GCP VM의 3개 서비스 상태를 감시하고 이상 감지 시 보고.

## 감시 대상

| 서비스 | 역할 |
|--------|------|
| 97layer-telegram.service | 텔레그램 봇 (사용자 인터페이스) |
| 97layer-ecosystem.service | SA/AD/CE 에이전트 + Pipeline Orchestrator |
| 97layer-gardener.service | 자율 최적화 에이전트 |

## 상태 확인 명령

```bash
ssh 97layer-vm "systemctl status 97layer-telegram 97layer-ecosystem 97layer-gardener --no-pager"
ssh 97layer-vm "journalctl -u 97layer-ecosystem -n 50 --no-pager"
```

## 재시작 명령

```bash
ssh 97layer-vm "sudo systemctl restart 97layer-ecosystem"
```

## 배포 경로

- VM 앱 경로: /home/skyto5339_gmail_com/97layerOS/
- Static IP: 136.109.201.201
- SSH alias: 97layer-vm

## 이상 패턴

- `activating (auto-restart)` → 크래시 루프. 로그 즉시 확인
- 오래된 pipeline_orchestrator 로그 → _scan_new_signals 미작동 의심
- UTF-8 decode warning → knowledge/signals/ 내 macOS ._ 파일 → 무시 가능 (graceful handling)
