---
name: infrastructure_sentinel
description: 97layerOS의 시스템 건전성 및 자기 치유(Self-healing)를 관리하는 스킬입니다. 가상 환경 루프 및 노드 모듈 누수를 탐지하고 소거합니다.
tools:
  - run_command
  - command_status
---

# Infrastructure Sentinel Skill

이 스킬은 97layerOS의 '순수 지능 스테이트'를 유지하기 위해 인프라를 정기적으로 정제하는 역할을 수행합니다.

## 핵심 기능

1. **지속적 정제(Continuous Sanitization)**: `venv`, `.venv`, `node_modules` 등 시스템 노이즈를 탐지하고 삭제합니다.
2. **데몬 관리**: `snapshot_daemon.py`의 가동 상태를 감시하고 필요시 재시작합니다.
3. **리소스 최적화**: 임시 잠금 파일 및 캐시 찌꺼기를 정리하여 에이전트 연산 속도를 확보합니다.

## 실행 도구 및 방법

- **Sentinel 가동**: `execution/snapshot_daemon.py`를 실행하여 1시간 주기 자동 정제 실행.
- **수동 정제**: `find` 및 `rm` 명령어를 조합하여 즉각적인 환경 숙청(Purge) 수행.

## 주의사항

- 루트 디렉토리 내의 가상 환경은 무조건 삭제 대상입니다.
- 정제 전 중요 설정 파일(`.env`, `token.json`)의 존재 여부를 반드시 확인하십시오.

참조: [infrastructure_sentinel.md](file:///Users/97layer/97layerOS/directives/infrastructure_sentinel.md)
