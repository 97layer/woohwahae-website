---
description: 모닝 리포트 생성 → knowledge/reports/morning_YYYYMMDD.md
---

# /morning — Morning Report Generator

오늘 날짜의 모닝 리포트를 생성한다.

## 실행 전 확인

```bash
cat knowledge/agent_hub/state.md
cat knowledge/system/work_lock.json
```

## 리포트 구조

`knowledge/reports/morning_YYYYMMDD.md` 에 작성 (오늘 날짜 적용):

```markdown
---
type: morning-report
created: YYYY-MM-DD
status: raw
tags: [모닝리포트]
---

# Morning Report — YYYYMMDD

## 어제의 핵심 성과

## 오늘의 우선순위 (TOP 3)

## 시스템 상태
- QUANTA 최신 업데이트:
- 활성 서비스:
- 알림 사항:

## THE CYCLE 현재 위치
Input → Store → Connect → Generate → Publish → Input again
현재: [해당 단계 표시]
```

## 완료 후

state.md 업데이트 후 handoff 실행.
