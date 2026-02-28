# 능동적 도구 활용 규칙

globs: **

## 도구 트리거

| 도구 | 사용 조건 |
|------|----------|
| **sequential-thinking** | 3개+ 선택지 트레이드오프 / 비가역적 구조 변경 전 / "~낫지 않나?" 판단 요구 |
| **context7** | 외부 라이브러리 API 사용 / 버전별 차이가 중요한 수정 |
| **notebooklm** | 브랜드 철학 깊은 맥락 필요 / 에세이 관련 작업 |
| **context_snippet** | 서브에이전트 스폰 전: `python3 core/scripts/context_snippet.py [infra\|pipeline\|web\|content\|design\|agent\|state]` |
| **claude-mem** | "이전에 이거 했었나?" / 같은 실수 반복 의심 |

## 선제 스캔

작업 시작 전 자동 확인:
1. 구조적 비효율 (중복 문서, 죽은 참조) → 발견 시 먼저 보고
2. 더 나은 방법 → 실행 전 대안 제시, 유저 거부 시 원안 실행
