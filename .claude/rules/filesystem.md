---
globs: "**/*"
alwaysApply: true
---

# 4축 구조 강제 — 파일 생성/이동 시 자동 적용

## 허용 경로 (이 외 모든 경로에 파일 생성 금지)

| 파일 유형 | 허용 경로 |
|-----------|-----------|
| Python (.py) | `core/agents/`, `core/system/`, `core/daemons/`, `core/scripts/`, `core/admin/`, `core/skills/`, `core/tests/` |
| Markdown (.md) | `directives/`, `knowledge/` 하위 |
| HTML/CSS/JS | `website/` 하위 |
| JSON (데이터) | `knowledge/signals/`, `knowledge/corpus/`, `knowledge/system/`, `knowledge/clients/`, `knowledge/service/`, `knowledge/reports/` |
| 설정 파일 | `.claude/`, `.github/`, 루트의 CLAUDE.md, README.md, .ai_rules만 허용 |

## 절대 금지

- 루트(/)에 .md, .py, .json 파일 생성
- `scripts/` (루트 레벨) — 반드시 `core/scripts/`에 생성
- `tests/` (루트 레벨) — 반드시 `core/tests/`에 생성
- 금지 파일명: SESSION_SUMMARY_*, WAKEUP_REPORT*, DEPLOY_*, NEXT_STEPS*, temp_*, untitled_*

## 위반 시 행동

1. 즉시 작업 중단
2. `directives/MANIFEST.md` 재참조
3. 올바른 경로에 재생성
4. 잘못 생성된 파일 삭제
