---
name: intelligence_backup
description: 97layerOS의 핵심 지능 자산을 아카이빙하고 권한 충돌을 우회하는 백업 전문 스킬입니다. Shadow Copy 및 다중 경로 백업을 처리합니다.
tools:
  - run_command
  - command_status
  - replace_file_content
---

# Intelligence Backup Skill

이 스킬은 97layerOS의 전략 및 지시서 등 핵심 자산을 영구히 보존하고, 협업 에이전트(Web Gemini 등)가 즉시 분석할 수 있는 클린 아카이브를 생성합니다.

## 핵심 기능

1. **Shadow Copy**: OS/클라우드 잠금 에러를 피하기 위해 보안 파일을 `/tmp`로 복제 후 아카이빙합니다.
2. **지능형 필터링**: `node_modules`, `.git` 등 대용량 노이즈를 완벽 배제하여 100MB 미만의 고밀도 스냅샷을 보장합니다.
3. **다중 경로 Fallback**: Google Drive(Primary) -> Local (Fallback) -> Tmp (Safe Harbor) 순으로 가용성을 확보합니다.

## 실행 도구 및 방법

- **정적 스냅샷**: `execution/create_snapshot.py`를 호출하여 수동 백업 수행.
- **자동화 연동**: Sentinel 데몬과 연동되어 1시간 단위로 무중단 실행.

## 결과 확인

- 압축 파일명 형식: `97layerOS_Intelligence_YYYYMMDD_HHMMSS.zip`
- 저장 경로: `/tmp/97layerOS_Snapshots/` (최종 대피소)

참조: [create_snapshot.py](file:///Users/97layer/97layerOS/execution/create_snapshot.py)
