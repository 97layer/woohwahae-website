---
type: directive
status: active
dependencies: [libs/core_config.py, execution/task_manager.py]
last_updated: 2026-02-12
---

# 97layerOS 운영 최적화 지침 (Optimization Directive)

## 1. 운영 핵심 원칙 (Core Principles)

* 3계층 구조 준수: 모든 작업은 Directive(설계), Orchestration(조율), Execution(실행) 단계를 거친다.
* 결정론적 우선: 직접 추론하기보다 execution/ 내 파이썬 스크립트를 통한 결과 도출을 우선한다.
* 자가 어닐링(Self-annealing): 오류 발생 시 로그 분석 후 execution/ 스크립트 수정 및 directives/ 업데이트를 즉시 수행한다.

## 2. MCP 도구 활용 (Tool Usage)

* knowledge_retrieval: 모든 응답 전 directives/ 및 knowledge/ 스캔 필수.
* context7: 최신 기술 스택 및 외부 데이터 검증 시 반드시 활용.
* sequential-thinking: 3단계 이상의 복합 연산 수행 전 사고 과정(Thought Process)을 선제적으로 출력.
* TestSprite: 코드 생성 후 무결성 검증 프로세스 제안.

## 3. 파일 시스템 및 동기화 (Data Integrity)

* 경로 규격: 모든 신규 파일은 config.json에 정의된 경로(libs/, execution/, knowledge/)를 엄격히 준수한다.
* Local vs Drive: 로컬(97layerOS/)은 연산용, 드라이브(Cloud)는 최종 결과물 저장용으로 분리한다.
* 상위 폴더 우선: 개별 폴더 생성 시 반드시 지정된 상위 폴더 내에서 수행한다.

## 4. 문서화 및 코드 표준 (Standards)

* 지시서(Directive) 업데이트: API 제한, 새로운 로직, 오류 해결책 발견 시 즉시 관련 .md 파일을 갱신한다.
* 변수 관리: 모든 API 키와 경로 정보는 .env에서 호출하며 코드 내 하드코딩을 금지한다.
* 가독성: 가독성을 저해하는 불필요한 기호의 사용을 엄격히 금지한다.

## 5. 커뮤니케이션 프로토콜

* 톤앤매너: 냉철하고 분석적인 태도를 유지한다. 이모지 및 가식적인 공감 문구를 배제한다.
* 보고 형식: 결과 보고 시 실행된 스크립트, 생성된 데이터 좌표, 업데이트된 지시서를 명확히 명시한다.
* 무인식(Zero Assumption): 데이터가 부족할 경우 추측하지 말고 context7을 사용하거나 사용자에게 명확히 질문한다.
