---
name: signal_capture
description: 외부 유입 정보(URL, 텍스트, 사용자 생각)를 포착하여 knowledge/inbox에 원시 신호(Raw Signal)로 저장하는 스킬입니다.
tools:
  - browser_subagent
  - read_url_content
  - write_to_file
---

# Signal Capture Skill

이 스킬은 97layerOS의 지능 축적 엔진의 입구 역할을 수행합니다. 모든 영감과 데이터를 '배양' 단계로 신속하게 인도합니다.

## 핵심 기능

1. **웹 추출**: 지정된 URL에서 핵심 인사이트와 트랜스크립트를 추출합니다.
2. **원시 신호(Raw Signal) 저장**: 포착된 정보를 `knowledge/inbox/` 폴더에 마크다운 형식으로 저장합니다.
3. **태그 부여 (Raw)**: 저장 시 `density: 1-3` 수준의 원시 지능 태그를 부여합니다.

## 실행 가이드

- **인사이트 요약**: 추출한 정보는 나열하지 말고, 97layerOS의 3계층 아키텍처 관점에서 재구성하십시오.
- **파일 생성**: `knowledge/inbox/raw_[date]_[subject].md` 형식으로 기록하십시오.

## 주의사항

- 인박스에 담는 과정에서는 과도한 정제보다 '빠짐없는 수집'에 집중하십시오. (정제는 `data_curation` 스킬의 몫입니다.)
- 중복된 신호가 이미 존재하는지 `grep_search`로 먼저 확인하십시오.
