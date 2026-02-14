---
id: diagnostic-20260214
context: Layer 2 Mercenary(Claude Opus 4.6)가 전체 코드베이스 정밀 분석 후 도출한 구조적 진단
links:
  - libs/core_config.py
  - libs/agent_router.py
  - libs/ai_engine.py
  - libs/gardener.py
  - libs/memory_manager.py
  - libs/context_router.py
density: 8
last_curated: 2026-02-14
---

# 97layerOS 구조 진단 — 2026-02-14

기반 평가: 설계 A, 구현 C+. 핵심 미접합 3건.

## CRITICAL

### 명명 이행 중단
brand_constitution.md는 Sovereign/Scout/Architect/Artisan/Narrator로 이행 완료.
core_config.py AGENT_CREW와 agent_router.py AGENT_REGISTRY는 구 체계(CD/TD/AD/CE/SA) 유지.
directives/agents/에 신규 파일만 존재, 구 파일은 archive/로 이동.
agent_router.py L86-95에서 FileNotFoundError silent fail. 페르소나 미로드 상태로 운영 중.

### 모델 분기 미구현
AI_MODEL_CONFIG에 gemini-2.0-flash 단일. Flash/Pro 분기 코드 없음.
AIEngine.generate_response()에 model_type 파라미터 부재.
context_router.py L27에서 model_type="flash" 전달 시 TypeError.

### Gardener 크래시
gardener.py L60: get_recent_context() 호출하나 MemoryManager에 해당 메서드 없음.
/evolve 시 AttributeError. 자가진화 비활성.

## HIGH

route() 로직: keyword_match 결과를 ai_classify가 무조건 덮어씀 (L165-174).
_ai_classify 이중 정의: L176과 L205에 동명 메서드. 두 번째가 첫 번째 덮어씀.
core_config.py 토큰 하드코딩: TELEGRAM_CONFIG.BOT_TOKEN 평문 노출.
bare except 9곳: ai_engine.py L85/L128, memory_manager.py L73 등.

## 수정 순서

즉시: 명명 통일, 토큰 환경변수 전환.
시급: 모델 분기 구현, Gardener 메서드 수정.
정리: route 체이닝, classify 중복 제거, bare except, f-string 로깅.
