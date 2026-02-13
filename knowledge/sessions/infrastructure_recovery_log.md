---
type: log
status: active
dependencies: [execution/technical_daemon.py, execution/telegram_daemon.py]
last_updated: 2026-02-12
---

# 97LAYER 인프라 복구 및 최적화 기록

## 조치 사항 요약

1. **가상환경(venv) 격리 및 안정화**
    * 구글 드라이브 동기화 폴더 내의 권한 잠금 문제를 해결하기 위해 로컬 임시 경로(`/tmp/venv_97layer`)에 가상환경을 재구축했습니다.
    * `requests`, `schedule`, `google-generativeai` 등 핵심 종속성을 완벽히 설치했습니다.

2. **MCP 서버 체계 현대화**
    * NPM 레지스트리에서 404 오류를 발생시키던 기존 패키지들을 검증된 대안으로 교체했습니다.
    * `fetch`: `mcp-server-fetch-typescript` (Node.js 기반)
    * `search`: `brave-search` (공식 대체제)
    * `context7`: 제공된 API Key를 사용하여 공식 Upstash MCP 서버 연동 완료.

3. **AI 엔진 및 설정 표준화**
    * `libs/core_config.py`에 `AI_MODEL_CONFIG`를 추가하여 모델 관리를 중앙집중화했습니다.
    * `libs/ai_engine.py`를 수정하여 `GEMINI_API_KEY`와 `GOOGLE_API_KEY`를 모두 지원하며, 설정 객체 유무에 관계없이 안전하게 초기화되도록 보강했습니다.

4. **데몬 프로세스 정상화**
    * `Technical` 및 `Telegram` 데몬을 새로운 환경에서 재가동했으며, 정상적으로 하트비트와 로그를 생성함을 확인했습니다.

## 향후 관리 지침

* **동기화 방어**: 프로젝트 루트에 `venv` 폴더를 직접 생성하지 말고, 데몬 실행 시에는 항상 `/tmp` 또는 지정된 로컬 경로를 참조하십시오.
* **API Key 관리**: `google-search` 기능을 사용하려면 `mcp_config.json`의 `BRAVE_API_KEY`를 실제 키로 업데이트해야 합니다.
