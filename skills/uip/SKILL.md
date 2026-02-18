---
name: uip
description: Unified Input Protocol. 어떤 채널(텔레그램, 직접 입력, URL)로 들어온 신호든 signal_capture 규격으로 표준화.
tools:
  - WebFetch
  - Bash
  - Grep
version: 2.0.0
updated: 2026-02-18
---

# Unified Input Protocol (UIP)

입력 채널이 달라도 동일한 파이프라인 입구(knowledge/signals/)로 수렴시키는 표준화 프로토콜.

## 채널별 처리

### 텔레그램 (자동)
- telegram_secretary.py가 수신 → knowledge/signals/ 직접 저장
- 에이전트 개입 불필요

### 유튜브 URL
1. WebFetch로 메타데이터 추출
2. 트랜스크립트 획득 (youtube-transcript-api)
3. signal_capture 스킬로 저장:
   - type: "youtube_video"
   - video_id, title, transcript 포함

### 텍스트 직접 입력
1. WOOHWAHAE 관점에서 재구성
2. signal_capture 스킬로 저장:
   - type: "text_insight"

### URL (아티클/웹페이지)
1. WebFetch로 본문 추출
2. 핵심 인사이트만 선별 (전문 저장 금지)
3. signal_capture 스킬로 저장:
   - type: "url_content"
   - source_url 포함

## 공통 원칙

- 단순 요약 금지 → WOOHWAHAE 브랜드 관점에서 재해석
- 중복 signal_id 확인 후 저장
- 저장 후 추가 액션 불필요 — Orchestrator가 자동 픽업
