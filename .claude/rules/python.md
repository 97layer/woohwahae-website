---
globs: core/**/*.py
---

# Python 코딩 규칙 (core/ 내 모든 .py)

- 로깅: f-string 금지. lazy formatting 사용
  - ❌ `logger.info(f"처리: {item}")`
  - ✅ `logger.info("처리: %s", item)`
- type hints 권장 (함수 시그니처)
- docstring: 신규 함수에만. 기존 함수는 건드리지 않음
- import 순서: stdlib → third-party → local (isort 기준)
- 에러 처리: 빈 except 금지. 구체적 예외만 catch
