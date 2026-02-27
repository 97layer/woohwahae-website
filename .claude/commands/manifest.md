---
description: SYSTEM.md §10 기반 배치 규칙 조회
---

# /manifest — 서재 배치 규칙 조회

파일 생성 전 반드시 확인한다.

## 실행 순서

1. **배치 규칙 읽기 (SYSTEM.md §10)**

```bash
cat directives/SYSTEM.md | grep -A 50 '§10'
```

1. **산출물 배치 규칙 요약 출력**

| 산출물 유형 | 저장 경로 | 명명 규칙 |
|---|---|---|
| 신호 (원시) | knowledge/signals/ | {type}_{YYYYMMDD}_{HHMMSS}.json |
| 신호 (분석) | knowledge/corpus/entries/ | entry_{signal_id}.json |
| 에세이 HTML | website/archive/essay-{NNN}-{slug}/ | index.html |
| 리포트 | knowledge/reports/ | {type}_{YYYYMMDD}.md |
| 자산 등록 | knowledge/system/asset_registry.json | append |
| 에이전트 제안 | knowledge/agent_hub/council_room.md | append |

## 규칙

- 위 테이블에 없는 경로에 파일 생성 금지
- 루트(/)에 .md 파일 생성 금지
- filesystem_cache.json 미확인 상태로 파일 생성 금지
