---
name: data_curation
description: knowledge/ 내 지식 자산 정화 + 온톨로지 태깅 + 중복 제거. INTELLIGENCE_QUANTA.md 밀도 유지.
tools:
  - Grep
  - Glob
  - Read
  - Edit
  - Bash
version: 2.0.0
updated: 2026-02-18
---

# Data Curation Skill

흩어진 신호와 지식을 연결하여 구조적 필연성을 가진 온톨로지로 변환.

## 주요 작업

1. **신호 상태 정화**: knowledge/signals/ 내 analyzed/published 신호 중 30일 초과분 아카이빙
2. **온톨로지 태깅**: 파일 상단 YAML 헤더에 context/importance/relationship 추가
3. **중복 제거**: Grep으로 동일 signal_id 탐지 → 최신 1개만 유지
4. **QUANTA 갱신**: knowledge/agent_hub/INTELLIGENCE_QUANTA.md 완료 항목 업데이트

## 실행 패턴 (토큰 최적화)

```
Glob("knowledge/signals/*.json") → 목록 파악
Grep("status.*published", "knowledge/signals/") → 완료 신호 필터
Read(offset, limit) → 대상 파일만 읽기
Edit → 변경 최소 범위만 수정
```

## 아카이빙 경로

- 완료 신호: knowledge/docs/archive/YYYY/MM_month/signals/
- 구버전 knowledge/inbox/ 파일: 이전 후 삭제

## 핵심 규칙

- 원본 훼손 금지. 수정 전 반드시 내용 확인
- 정제 목적: 파일 정리가 아닌 지능 밀도 향상
- QUANTA 수정은 덮어쓰기 (Append 금지)
