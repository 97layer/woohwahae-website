# Agent Personas - 5-Agent Multimodal System

## 현역 에이전트 (Production)

### 1. [Strategy Analyst (SA)](strategy_analyst.md)
**역할**: 신호 분석, 브랜드 적합성 판단
**모델**: Gemini Flash
**필수 읽기**: cycle_protocol.md, anti_algorithm_protocol.md
**Multimodal**: 텍스트 분석 (SA+AD 병렬 실행)

### 2. [Art Director (AD)](art_director.md)
**역할**: 이미지 심미 분석, 시각 정체성
**모델**: Gemini Vision
**필수 읽기**: visual_identity_guide.md, aesop_benchmark.md
**Multimodal**: 비전 분석 (SA+AD 병렬 실행)

### 3. [Chief Editor (CE)](chief_editor.md)
**역할**: 멀티모달 콘텐츠 생성
**모델**: Gemini Flash
**필수 읽기**: imperfect_publish_protocol.md, communication_protocol.md
**Multimodal**: SA+AD 결과 통합 → 콘텐츠 생성

### 4. [Creative Director (CD)](creative_director.md)
**역할**: 최종 승인/거부 (Sovereign Judgment)
**모델**: Claude Opus
**필수 읽기**: brand_constitution.md, 97layer_identity.md, woohwahae_identity.md
**Multimodal**: CE 결과 최종 판단

### 5. [Technical Director (TD)](technical_director.md)
**역할**: 오케스트레이션, 인프라, 자동화
**모델**: Gemini Flash
**필수 읽기**: cycle_protocol.md, daemon_workflow.md, sync_protocol.md
**Multimodal**: 전체 파이프라인 관리

## 실험적 역할 (미사용)
- [architect.md](architect.md) - 장기 설계
- [artisan.md](artisan.md) - 실행 장인
- [narrator.md](narrator.md) - 스토리텔링
- [scout.md](scout.md) - 트렌드 탐색
- [sovereign.md](sovereign.md) - 주권자 (CD 대체안?)

## Archive
구버전 파일들은 [archive/](archive/)에 보관됨. 참고만 할 것.

## Agent 추가 시 체크리스트

1. 이 폴더에 `{role_name}.md` 생성
2. 템플릿 사용:
   ```markdown
   ---
   role: [Role Name]
   code: [2-letter code]
   model: [AI Model]
   status: active
   dependencies:
     must_read: [list of directives]
     recommended: [optional directives]
   last_updated: YYYY-MM-DD
   multimodal: true/false
   ---

   # [Role Name] (Code)

   ## I. Identity (독립 인격체)
   ## II. Core Directives (필독 문서)
   ## III. Responsibilities (5단계 사이클 내 역할)
   ## IV. Communication Style
   ```
3. 이 README.md 업데이트
4. 상위 `../README.md` 업데이트
5. `libs/core_config.py`의 `AGENT_CREW` 딕셔너리에 추가
6. Gardener에 등록

---

**Last Updated**: 2026-02-15
**System**: Async 5-Agent Multimodal
**Performance**: 2.5x productivity (11s parallel processing)