---
name: data_curation
description: 97layerOS의 지능 자산을 관리하고 온톨로지를 구축하는 스킬입니다. 데이터 간 상관관계를 분석하여 고지능 큐레이션을 수행합니다.
tools:
  - grep_search
  - find_by_name
  - replace_file_content
  - multi_replace_file_content
---

# Data Curation Skill

이 스킬은 흩어진 데이터 조각들을 연결하여 '구조적 필연성'을 가진 지능형 온톨로지로 변환하는 역할을 수행합니다. 향후 독립할 'The Curator' 에이전트의 핵심 엔진입니다.

## 핵심 기능

1. **지능적 인덱싱**: `directives/`와 `knowledge/` 간의 논리적 연결 고리를 찾아 메타데이터로 기록합니다.
2. **온톨로지 태깅**: 파일 상단에 YAML 형식의 지능 태그(Context, Importance, Relationship)를 부여합니다.
3. **데이터 정화**: 중복되거나 노이즈가 섞인 정보를 식별하여 통합하거나 소거합니다.

## 실행 가이드

- **상관관계 탐색**: `grep_search`를 사용하여 공통 키워드를 추출하고, 지식 간의 계층 구조를 연산합니다.
- **태그 부여**: `data_asset_management.md` 지시서에 정의된 표준 형식을 준수하여 파일 헤더를 업데이트합니다.

## 주의사항

- 정제된 원본 데이터의 훼손을 방지하기 위해 `replace_file_content` 사용 시 변경 이력을 명확히 기록하십시오.
- 단순한 파일 정리가 아닌 '지능의 밀도'를 높이는 것에 집중하십시오.

참조: [data_asset_management.md](file:///Users/97layer/97layerOS/directives/data_asset_management.md)
