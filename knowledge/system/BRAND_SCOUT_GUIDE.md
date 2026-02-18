# Brand Scout Agent — 사용 가이드

> **목적**: 슬로우라이프/미니멀 브랜드 자동 발굴 및 매거진 소재 큐레이션
> **위치**: [core/agents/brand_scout.py](../../core/agents/brand_scout.py)
> **생성**: 2026-02-17

---

## I. 개요

Brand Scout는 **매거진 B 방향 전환**을 위한 브랜드 발굴 시스템입니다.

### 역할

1. **Discovery**: 슬로우라이프 브랜드 자동 발굴 (SNS/웹 크롤링)
2. **Screening**: WOOHWAHAE 5 Pillars 기준 적합성 평가 (0-100점)
3. **Data Gathering**: 승인된 브랜드 심층 데이터 수집
4. **Dossier 생성**: `knowledge/brands/[slug]/` 구조화된 브랜드 자료

---

## II. 사용법

### 1. 후보 브랜드 수동 추가

순호가 발견한 브랜드를 시스템에 등록:

```bash
python core/agents/brand_scout.py --add "제주 Mute" "https://mute.kr" "고요함을 파는 숙소"
```

→ `knowledge/brands/candidates.json`에 추가됨 (status: pending)

### 2. 스크리닝 실행

등록된 후보 브랜드를 WOOHWAHAE 기준으로 평가:

```bash
python core/agents/brand_scout.py --process
```

**평가 기준 (WOOHWAHAE 5 Pillars):**
- Authenticity (진정성): 20점
- Practicality (실용성): 20점
- Elegance (우아함): 20점
- Precision (정밀함): 20점
- Innovation (혁신): 20점

**결과:**
- **80점 이상**: approve (매거진 소재로 승인)
- **50-79점**: review (순호 최종 판단 필요)
- **50점 미만**: reject (부적합)

### 3. 자동 승인 모드

80점 이상 브랜드를 자동으로 승인 + 도시에 생성:

```bash
python core/agents/brand_scout.py --process --auto-approve
```

→ 승인된 브랜드는 즉시 `knowledge/brands/[slug]/`에 도시에 생성

---

## III. 도시에 구조

승인된 브랜드는 다음 구조로 저장됩니다:

```
knowledge/brands/
├── candidates.json           — 후보 큐 (pending/screened/approved)
└── [slug]/                   — 개별 브랜드 도시에
    ├── profile.json          — 메타데이터 + 스크리닝 결과
    ├── raw_content.md        — 수집한 원본 컨텐츠
    └── sa_analysis.json      — SA Agent 전략 분석 (나중에 생성)
```

### `profile.json` 예시

```json
{
  "name": "제주 Mute",
  "url": "https://mute.kr",
  "screening_score": 88,
  "pillars_match": {
    "authenticity": 18,
    "practicality": 17,
    "elegance": 20,
    "precision": 16,
    "innovation": 17
  },
  "decision": "approve",
  "reasoning": "절제된 시각 언어와 고요함을 상품화한 철학이 WOOHWAHAE와 공명. 여백을 활용한 공간 디자인이 우아함 기준 충족.",
  "collected_at": "2026-02-17T14:30:00",
  "status": "pending_sa_analysis"
}
```

### `raw_content.md` 예시

```markdown
# 제주 Mute — Raw Content

**URL**: https://mute.kr
**수집 일시**: 2026-02-17T14:30:00

---

## Website Content

Mute는 제주에서 고요함을 파는 숙소입니다.
불필요한 장식을 제거하고, 오직 본질만 남겼습니다.
빛과 그림자, 바람과 침묵. 그것만으로 충분합니다.

(웹사이트 About, Philosophy 섹션 전체)

---

## Social Media Posts

(Instagram 최근 6개월 포스트 — 미구현)

---

## Press & Reviews

(언론 리뷰 — 미구현)
```

---

## IV. 워크플로우

```
순호 발견
  ↓
Brand Scout --add (후보 등록)
  ↓
Brand Scout --process (스크리닝)
  ↓
[80점 이상] → 자동 도시에 생성
[50-79점] → 순호 최종 판단 (텔레그램 알림)
[50점 미만] → 거부
  ↓
SA Agent 전략 분석 (도시에 → sa_analysis.json)
  ↓
CE Agent 매거진 Issue 편집
  ↓
Magazine Publisher → woohwahae.kr/archive/issue-NN/
```

---

## V. 통합 예정 기능

### Phase 1 (현재)
- ✅ 수동 후보 추가
- ✅ 웹사이트 크롤링 (BeautifulSoup)
- ✅ Gemini 기반 스크리닝
- ✅ 도시에 생성

### Phase 2 (1개월)
- [ ] Instagram 크롤링 (Apify 통합)
- [ ] 언론 리뷰 검색 (Serper.dev API)
- [ ] 텔레그램 알림 (review 등급 브랜드)

### Phase 3 (3개월)
- [ ] Reddit/큐레이션 사이트 자동 발굴
- [ ] SA Agent 자동 연동 (도시에 → 전략 분석)
- [ ] 월간 브랜드 발굴 리포트 (순호에게 제안)

---

## VI. 실행 예시

### 시나리오: 순호가 "제주 Mute"를 발견

```bash
# 1. 후보 등록
python core/agents/brand_scout.py --add "제주 Mute" "https://mute.kr" "고요함을 상품화한 숙소"

# 2. 스크리닝
python core/agents/brand_scout.py --process --auto-approve

# 출력:
# [Scout] 스크리닝 시작: 제주 Mute
# [Scout] 제주 Mute 스크리닝 완료: approve (점수: 88)
# [Scout] 제주 Mute 데이터 수집 시작
# [Scout] 도시에 생성 완료: knowledge/brands/jeju-mute
```

### 결과 확인

```bash
cat knowledge/brands/jeju-mute/profile.json
# → 스크리닝 결과 (88점, approve)

cat knowledge/brands/jeju-mute/raw_content.md
# → 수집한 웹사이트 컨텐츠
```

---

## VII. 텔레그램 통합 (미래)

순호가 텔레그램에서 직접 후보 추가 가능:

```
순호: /scout "제주 Mute" https://mute.kr "고요함을 파는 숙소"

Bot: [Brand Scout] 후보 등록 완료. 스크리닝 시작합니다...
(10초 후)
Bot: ✅ "제주 Mute" — 승인 (88점)
     → knowledge/brands/jeju-mute/ 도시에 생성
     → SA Agent 전략 분석 대기 중

순호: /scout process

Bot: [Brand Scout] 처리 대기 중인 후보: 3개
     1. "서울 어니언" — approve (85점)
     2. "부산 OO" — review (72점) ⚠️ 순호 판단 필요
     3. "강릉 XX" — reject (45점)
```

---

## VIII. 핵심 원칙

### 질 > 양
- 한 달에 1-2개 브랜드만 깊이 다루는 것이 목표
- 80점 이상만 통과 → 브랜드 희소성 유지

### 객관성 유지
- 광고/협찬 브랜드 배제
- 스크리닝 기준 공개 (투명성)
- 점수 조작 금지 (자동 평가만)

### 순호 신뢰
- 자동 시스템이지만 최종 판단은 순호
- review 등급 브랜드는 반드시 확인 요청
- 순호가 거부하면 즉시 도시에 삭제

---

**Last Updated**: 2026-02-17
**Author**: 97layerOS
**Related**: [WEBSITE_STRUCTURE.md](WEBSITE_STRUCTURE.md), [brand_scout.py](../../core/agents/brand_scout.py)

> "Remove the Noise, Reveal the Essence" — WOOHWAHAE
