# LAYER OS 코드 설계 규칙
# Source: CLAUDE.md SYSTEM DESIGN RULES (이전됨)
# Last Updated: 2026-02-23

> "애플 칩처럼 — 겉은 단순, 속은 치밀. 보이지 않는 곳이 더 견고해야 한다."

---

## 1. 환경변수 원칙 (ENV FIRST)

- **모든 설정값은 `.env`에서만 읽는다.** 코드 안 하드코딩 금지.
- 필수 env 목록: `core/system/env_validator.py` 에 중앙 관리.
- **에이전트 시작 시 반드시 `validate_env()` 호출** — 누락 시 FATAL 종료.
- 도메인/경로 변경 시 `.env` 한 줄만 수정하면 전체 반영되는 구조 유지.

```python
# ✅ GOOD
from core.system.env_validator import get_site_base_url, validate_env
validate_env("sa_agent")
url = get_site_base_url()

# ❌ BAD
url = "https://woohwahae.kr"  # 하드코딩 금지
```

---

## 2. 로그 원칙 (EXPLICIT LOGGING)

- 조용히 실패하는 코드 금지. **실패는 반드시 명시적으로 출력.**
- 중요 이벤트(태스크 시작/완료/실패)는 `flush=True`로 즉시 출력.
- HTML 저장, API 호출 등 부수 작업은 `try/except` 분리 — 본 작업이 죽지 않도록.
- 로그 파일: `.infra/logs/{service}.log` / `.infra/logs/{service}.error.log`

```python
# ✅ GOOD
try:
    self._save_html(result)
except Exception as e:
    print(f"[WARN] HTML 저장 실패 (무시): {e}", flush=True)

# ❌ BAD
result = do_something()  # 실패해도 모름
```

---

## 3. 경로 원칙 (PATH CONSISTENCY)

- `PROJECT_ROOT`는 `env_validator.get_project_root()` 단일 진입점 사용.
- 상대경로 직접 사용 금지 — 파일 위치에 따라 결과가 달라짐.
- 경로 구성: `PROJECT_ROOT / 'website' / 'archive'` 형태 (pathlib 우선).

---

## 4. 디버깅 원칙 (DEBUGGABLE DESIGN)

- 에러 메시지에 **컨텍스트 포함** — "실패" 아닌 "어디서, 무엇이, 왜 실패".
- VM 직접 패치 후 로컬 동기화 필수 — 소스 불일치 방지.
- 새 기능 추가 시 실패 경로(unhappy path)를 먼저 설계.
