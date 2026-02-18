---
name: signal_capture
description: 외부 유입 정보(URL, 텍스트, 유튜브, 사용자 생각)를 knowledge/signals/에 JSON 신호로 저장. 파이프라인 입구.
tools:
  - WebFetch
  - Glob
  - Grep
  - Write
  - Bash
version: 2.0.0
updated: 2026-02-18
---

# Signal Capture Skill

SA→AD→CE→CD 파이프라인의 입구. 신호가 여기서 생성되어야 체인이 시작된다.

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
2. 신호 구성: type 결정 → WOOHWAHAE 관점에서 내용 재구성
3. 저장: knowledge/signals/<signal_id>.json
4. 자동 처리: Pipeline Orchestrator 30초 폴링 → SA 태스크 자동 생성

## 핵심 규칙

- status는 반드시 "captured" — orchestrator가 이 값만 픽업
- signal_id 형식: text_YYYYMMDD_HHMMSS 또는 youtube_YYYYMMDD_HHMMSS
- 단순 요약 금지. WOOHWAHAE 철학(미니멀리즘, 슬로우라이프, 본질)과의 연결 지점 중심
- 구버전 경로 knowledge/inbox/ 사용 금지 → 현재는 knowledge/signals/ 사용
