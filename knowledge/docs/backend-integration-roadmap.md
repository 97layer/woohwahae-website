# 백엔드 + 풀스택 하네스 통합 로드맵
> 작성: 2026-03-04 | 검토: Claude + Gemini (ready 확인 완료: 2026-03-04T02:05:17Z)
> 상태: **PENDING — Phase 0 게이트 미통과 시 구현 금지**

---

## 목표

1. 분산 백엔드(Flask/FastAPI/ecommerce)를 단일 Gateway 진입점으로 통합
2. 하네스(Queue/Orchestrator/Telegram/Agents) 제어면을 API화
3. 운영 스크립트/검증 게이트를 단일 경로로 통합

---

## 확정 아키텍처

- **Canonical runtime**: FastAPI Gateway (`core/backend/main.py`)
- **도메인 흡수**: cms(Flask) / upload(photo_upload) / commerce(ecommerce)
- **Harness control plane**: `/harness/status`, `/queue/pending`, `POST /queue/task`, `/healthz`
- **에이전트 실행축**: 현행 파일 큐 + pipeline_orchestrator.py 유지 (브로커 전환 범위 제외)

---

## 신규/정규화 API

| 경로 | 메서드 | 설명 |
|------|--------|------|
| `/healthz` | GET | gateway + queue + orchestrator + domain 서비스 종합 헬스 |
| `/harness/status` | GET | agents / lock / queue depth / orchestrator 루프 상태 |
| `/queue/pending` | GET | pending task 요약 목록 |
| `/queue/task` | POST | 운영 태스크 주입 (권한 필요) |
| `/cms/*` | GET/POST | 기존 Flask CMS API 호환 경로 |
| `/upload/*` | GET/POST | photo_upload 호환 경로 |
| `/commerce/*` | GET/POST | ecommerce API 프록시/이관 경로 |

---

## 구현 워크스트림

| WS | 파일 | 내용 |
|----|------|------|
| WS-A | `core/backend/main.py`, `core/backend/config.py` | Gateway 표준화: 앱 초기화/인증/예외/로깅 공통화 |
| WS-B | `core/backend/app.py` → main.py 하위 라우터 | Flask CMS 단계 이관, 인증/세션 Gateway 통일 |
| WS-C | `core/backend/photo_upload.py`, `ecommerce/*` | include/mount 1차 결합 후 공통 auth/logging 적용 |
| WS-D | `pipeline_orchestrator.py`, `queue_manager.py`, `main.py` | Harness Control API 추가 |
| WS-E | `start_harness_fullstack.sh`, `session_bootstrap.sh`, `harness_doctor.py` | 운영 단일화: Gateway 포함 기동/헬스/닥터 표준화 |
| WS-F | `core/system/agent_logger.py` | 관측/감사: correlation id, 감사로그, 장애 분류 코드 |

---

## 단계별 롤아웃

### Phase 0 (필수 게이트)
- 같은 태스크 문구로 Plan Council preflight **연속 2회** 실행
- **통과 조건**: 두 번 모두 `status=ready` + `models_used=["claude","gemini"]`
- 미통과 시: 구현 금지, 설계 문서/리스크 정리만 진행
- 검증 명령:
  ```bash
  python3 core/system/plan_council.py \
    --task "백엔드 로직 및 시스템 풀스택 하네스 구조 통합플랜 기획 — Flask/FastAPI/ecommerce 단일 Gateway 통합" \
    --mode preflight
  ```

### Phase 1: 비파괴 Gateway 도입
- 기존 서비스 유지, Gateway 경유 가능 상태 확보
- 성공 기준: `/healthz`, `/harness/status` 정상 응답

### Phase 2: 기능 이관 + 호환 유지
- CMS/Upload/Commerce 기능 parity 확보
- 레거시 경로: deprecation header 포함 유지

### Phase 3: 운영 경로 단일화
- 기동/배포/헬스체크 단일 명령 체계 확정
- 성공 기준: 단일 명령으로 전체 스택 기동

### Phase 4: 레거시 정리
- 독립 Flask 런타임 종료
- 문서/런북 전환 완료

---

## 승인 기준 (Go/No-Go)

1. Phase 0 게이트 통과 (ready + dual-model 연속 2회)
2. `session_bootstrap.sh` = `READY`
3. `harness_doctor.py --run-tests` = `PASS`
4. 핵심 API 회귀 테스트 0 fail
5. 운영 스크립트 단일 경로에서 기동/중지/헬스체크 성공

---

## 주요 리스크

- **Gateway SPOF**: 단일 진입점 장애 시 전체 다운 — 복구 경로 사전 설계 필수
- 인증 방식 불일치로 인한 통합 복잡도 증가
- 3개 서비스 간 API 버전 호환성 문제
- 하네스 제어 API 권한 관리 미흡 시 보안 취약점
- 운영 중단 없는 마이그레이션 경로 부재

---

## 가정 및 기본값

1. 통합 범위: 전 범위
2. 전환 방식: 단계적 전환
3. Canonical backend: FastAPI Gateway
4. 메시징: 현행 파일 큐 유지 (브로커 전환 이번 범위 제외)
5. Plan Council 결과는 `timestamp + models_used + status` 3종 없으면 유효로 간주하지 않음
