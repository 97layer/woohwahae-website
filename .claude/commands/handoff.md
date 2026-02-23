---
description: 세션 종료 핸드오프 — QUANTA 업데이트 + 다음 에이전트를 위한 상태 기록
---

# /handoff — Session Handoff

세션을 안전하게 종료하고 다음 에이전트에게 상태를 인계한다.

## 실행 순서

1. **완료한 작업 요약 작성** (아래 형식)

2. **핸드오프 실행**
```bash
./scripts/session_handoff.sh "<agent-id>" "<작업 요약>" "<다음 태스크1>" "<다음 태스크2>"
```

스크립트 없는 환경:
```bash
python core/system/handoff.py --handoff
```

3. **INTELLIGENCE_QUANTA.md 업데이트** — 새로 생성된 자산, 변경된 상태 반영

4. **세션 기록 저장** (선택)
```
knowledge/docs/sessions/YYYYMMDD_session.md
```

## 핸드오프 요약 형식

```
agent-id: claude-sonnet-[날짜]
작업 요약: [완료한 태스크 1-3줄]
생성 자산: [경로 목록]
다음 태스크:
  - [태스크1]
  - [태스크2]
주의사항: [다음 에이전트가 알아야 할 것]
```

## 주의

handoff 미실행 시 Git Pre-Commit Hook이 커밋을 차단한다.
