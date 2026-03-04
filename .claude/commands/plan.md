---
description: 실행 전 계획 협의 — /plan-council 별칭
---

# /plan — Preflight Planning Council Alias

`/plan`은 `/plan-council`의 축약 별칭이다.

```bash
bash core/scripts/plan_dispatch.sh "$ARGUMENTS" --manual
```

## 규칙

1. 출력에서 반드시 `steps`, `risks`, `checks`를 먼저 확인한다.
2. `status: degraded`면 리스크를 먼저 공유한 뒤 실행한다.
3. `decision: needs_clarification`면 구현을 중단하고 요구사항을 구체화한다.
4. 결과 로그는 `knowledge/system/plan_council_reports.jsonl`에 자동 기록된다.
5. 내부적으로 `core/system/plan_council.py --task "$ARGUMENTS" --mode preflight` 경로를 사용한다.
