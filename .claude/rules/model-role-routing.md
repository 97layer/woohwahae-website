# 모델 역할 라우팅 규칙 (Role Routing)

globs: **
source_of_truth: AGENTS.md

## Role Matrix

1. `Claude`  : 기획/리스크 설계 (Plan Council primary planner)
2. `Codex`   : 코드 생성/수정 (implementation owner)
3. `Gemini`  : 검증/비평/회귀 체크 (verification owner)

## Mandatory Flow

1. Plan: `python3 core/system/plan_council.py --task "<intent>" --mode preflight`
2. Code: Codex가 변경안 생성
3. Verify: Gemini critic/validator로 검증
4. Gate: 검증 결과 통과 후만 propose/apply

## Hard Bans

1. Claude가 코드 생성 주체로 동작하는 것 금지
2. Gemini가 최종 실행 결정 주체가 되는 것 금지
3. Plan 없이 코드 생성 시작 금지 (비자명 작업 기준)
