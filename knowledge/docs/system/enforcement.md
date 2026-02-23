# LAYER OS 강제 집행 레이어
# Source: CLAUDE.md ENFORCEMENT LAYERS (이전됨)
# Last Updated: 2026-02-23

## 레이어 구조

| Layer | 메커니즘 | 위치 |
|-------|---------|------|
| 1 | Git Pre-Commit Hook | `.git/hooks/pre-commit` |
| 2 | GitHub Actions CI/CD | `.github/workflows/session-integrity.yml` |
| 3 | Bootstrap Script | `scripts/session_bootstrap.sh` |
| 4 | Handoff Script | `scripts/session_handoff.sh` |

**handoff 미실행 시 커밋 차단.**

## 규칙

- Layer 1 (Git Hook): 커밋 시 세션 핸드오프 완료 여부 검증
- Layer 2 (CI/CD): 세션 무결성 자동 검사
- Layer 3 (Bootstrap): 세션 시작 시 QUANTA + lock + cache 자동 확인
- Layer 4 (Handoff): 세션 종료 시 다음 에이전트를 위한 상태 기록
