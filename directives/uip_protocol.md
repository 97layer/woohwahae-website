# UIP (Unified Input Protocol) Directive

이 문서는 외부 신호를 97layerOS의 지식 자산으로 변환하는 오케스트레이션 지침서이다.

## 1. 개요

모든 외부 데이터(특히 유튜브 링크)는 이 프로토콜을 거쳐야만 시스템 내 `knowledge/` 계층에 진입할 수 있다.

## 2. 단계별 실행 절차 (SOP)

### Phase 1: 파싱 (Parsing)

1. 입력된 URL이 유튜브 링크인지 확인한다.
2. `execution/youtube_parser.py`를 실행하여 영상의 제목, 설명, 트랜스크립트를 추출한다.
3. 추출 실패 시 사용자에게 알리고 `Gemini Web` 분석 결과 입력을 요청한다.

### Phase 2: 분석 및 필터링 (Analysis & Filtering)

1. 추출된 텍스트에서 '우화해(WOOHWAHAE)' 브랜드 철학과 상충되는 소음(Noise)을 제거한다.
2. 핵심 키워드를 도출하고 기존 지식 베이스와의 연관성을 연산한다.

### Phase 3: 온톨로지 변환 및 저장 (Transformation & Commit)

1. `execution/ontology_transform.py`를 호출하여 표준 마크다운 문서를 생성한다.
2. `ID`를 발급하고 `knowledge/raw_signals/` 경로에 파일을 저장한다.
3. `knowledge/status.json`을 업데이트하여 처리 상태를 기록한다.

## 3. 예외 처리

- 트랜스크립트가 없는 경우: 영상 메타데이터(제목, 설명)만으로 분석을 진행하되, 부족할 경우 사용자 개입을 요청한다.
- 중복 링크: 이미 `knowledge/raw_signals/`에 존재하는 경우 작업을 중단하고 기존 파일을 안내한다.
