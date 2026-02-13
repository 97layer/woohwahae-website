---
type: directive
status: active
dependencies: [task_status.json, libs/core_config.py]
last_updated: 2026-02-12
---

# 97layerOS 시스템 핸드셰이크 지침 (System Handshake Directive)

## 1. 개요 (Objective)

본 지시서는 97layerOS 내의 서로 다른 에이전트(Cursor, Gemini, Claude 등) 및 역할군이 교대하거나 협업할 때, 맥락의 단절 없이 **상태(State)**와 **의도(Intent)**를 완벽하게 전달하기 위한 프로토콜을 정의한다.

## 2. 역할군 정의 (Role Definitions)

| 역할군 | 핵심 책무 | 주요 자산 |
| :--- | :--- | :--- |
| **Strategist (Directive)** | 비즈니스 로직 설계 및 SOP 수립 | `directives/*.md` |
| **Orchestrator (Agent)** | 도구 선택, 워크플로우 제어, 의사결정 | `mcp_config.json`, MCP Tools |
| **Executor (Python)** | 데이터 처리, API 통신, 파일 연산 | `execution/*.py`, `.local_venv` |
| **Validator (QA)** | 결과 검증 및 자가 어닐링 수행 | `sequential-thinking`, Logs |

## 3. 핸드셰이크 프로토콜 (Handover Procedure)

에이전트가 작업을 중단하거나 다른 에이전트에게 넘길 때, 반드시 다음 **'상태 객체(State Object)'**를 갱신해야 한다.

### A. 상태 기록 (State Logging)

모든 에이전트는 세션 종료 전 `knowledge/status.json`에 현재 진행 상황을 다음 규격에 맞춰 기록한다. (참고: 프로젝트 루트의 `task_status.json`과 동기화 상태를 유지할 것)

```json
{
  "task_id": "YYYYMMDD_TASK_NAME",
  "current_phase": "Execution",
  "last_directive": "directives/scrape_news.md",
  "last_action": {
    "tool": "execution/scraper.py",
    "status": "success",
    "output_path": ".tmp/raw_data.json"
  },
  "runtime_env": {
    "venv_path": "/tmp/venv_97layer",
    "daemons_active": ["technical_daemon", "telegram_daemon"]
  },
  "pending_issue": "None",
  "next_step_required": "Data cleaning and validation"
}
```

### B. 컨텍스트 수용 (Context Absorption)

새로운 에이전트가 투입되면 다음 순서로 연산을 시작한다.

1. `knowledge/status.json`을 읽어 마지막 `task_id`와 `next_step_required`를 파악한다.
2. `runtime_env`를 확인하여 로컬 가상환경 및 데몬 프로세스의 생존 여부를 검증한다. (소실 시 즉시 복구 모드 진입)
3. `last_directive`에 명시된 `.md` 파일을 읽어 로직을 동기화한다.
4. `output_path`의 데이터를 검증하여 이전 단계의 실행 결과가 유효한지 확인한다.

## 4. 유기적 협업 규칙 (Inter-Agent Rules)

* **결과물 중심 상호작용**: 에이전트 간의 대화보다 '공유된 파일의 상태'로 소통한다. 결과 파일을 생성하고 `status.json`을 업데이트하는 것이 실질적 상호작용이다.
* **동적 피드백 루프**: `Executor(Python)`가 에러를 발생시키면 `Validator(QA)`는 즉시 `sequential-thinking`을 통해 원인을 분석하고, `Strategist(Directive)`에게 지시서 수정을 요청한다.
* **인프라 종속성 확인**: 매 핸드셰이크 시 `.local_node` 및 권한 잠금 여부를 체크하여 실행 환경의 연속성을 보장한다.

## 5. 자가 어닐링 및 복구 (Self-Annealing & Recovery)

* 만약 이전 에이전트가 남긴 `output_path`에 파일이 없거나 데이터가 손상된 경우, 현재 에이전트는 즉시 **'복구 모드'**로 진입한다.
* 복구 모드에서는 `last_action`을 재실행하고, 실패 원인을 분석하여 `directives/`에 예외 처리 로직을 추가한다.
* 권한 문제(`Operation not permitted`) 발생 시, 즉시 격리된 로컬 경로로 우회하는 전략을 자가 수립한다.
