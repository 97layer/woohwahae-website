---
name: signal_capture
description: 어떤 채널(텔레그램/유튜브/URL/텍스트)로 들어온 신호든 표준화 후 knowledge/signals/에 저장. 파이프라인 입구.
tools:
  - WebFetch
  - Glob
  - Grep
  - Write
  - Bash
version: 3.0.0
updated: 2026-02-18
note: uip 스킬 흡수 통합
---

# Signal Capture Skill (구 UIP 통합본)

SA→AD→CE→CD 파이프라인의 입구. 채널 표준화 + 저장을 단일 스킬로 처리.

## 채널별 처리 (구 UIP)

### 텔레그램 (자동)
- telegram_secretary.py가 수신 → knowledge/signals/ 직접 저장
- 에이전트 개입 불필요

### 유튜브 URL
1. WebFetch로 메타데이터 추출
2. 트랜스크립트 획득 (youtube-transcript-api)
3. type: "youtube_video" + video_id, title, transcript 포함 후 저장

### URL (아티클/웹페이지)
1. WebFetch로 본문 추출
2. 핵심 인사이트만 선별 (전문 저장 금지)
3. type: "url_content" + source_url 포함 후 저장

### 텍스트 직접 입력
1. WOOHWAHAE 관점에서 재구성
2. type: "text_insight" 로 저장

## 출력 규격 (knowledge/signals/<signal_id>.json)

```json
{
  "signal_id": "text_YYYYMMDD_HHMMSS",
  "type": "text_insight | youtube_video | url_content",
  "status": "captured",
  "content": "핵심 인사이트 (나열 금지, 구조적 재구성)",
  "captured_at": "2026-02-18T00:00:00",
  "from_user": "97layer",
  "metadata": {}
}
```

## 타입별 추가 필드

| type | 추가 필드 |
|------|---------|
| youtube_video | video_id, title, transcript |
| url_content | source_url |
| text_insight | (없음) |

## 실행 순서

1. 중복 확인: Grep(signal_id, "knowledge/signals/")
2. 채널 판별 → 타입 결정
3. WOOHWAHAE 관점(미니멀리즘, 슬로우라이프, 본질)에서 내용 재구성
4. knowledge/signals/<signal_id>.json 저장
5. 자동 처리: Pipeline Orchestrator 30초 폴링 → SA 태스크 생성

## 핵심 규칙

- status는 반드시 "captured" — orchestrator가 이 값만 픽업
- signal_id 형식: text_YYYYMMDD_HHMMSS / youtube_YYYYMMDD_HHMMSS
- 단순 요약 금지
- 구버전 경로 knowledge/inbox/ 사용 금지
