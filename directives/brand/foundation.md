# Brand OS Foundation — 기반 구조

> **소스**: IDENTITY.md §I, §III + SYSTEM.md §I
> **권한**: CD 승인 필수 (FROZEN 등급 내용 포함)

---

## 1. LAYER OS — 5-Layer 모델

LAYER OS는 브랜드를 운영 체제로 추상화한다. 5개 레이어가 수직으로 쌓인다.

| Layer | 이름 | 담당 | 에이전트 |
|-------|------|------|----------|
| L1 | Philosophy | 철학, 방향성, 최종 승인 | CD (순호) |
| L2 | Design | 시각 정체성, 디자인 토큰 | AD |
| L3 | Content | 텍스트 생성, 편집, 신호 분석 | CE + SA |
| L4 | Service | 고객 경험, 아틀리에 운영 | Ritual Module |
| L5 | Business | 수익 추적, 성장 지표 | Growth Module |

횡단 레이어:
- **QA (Ralph)**: STAP 품질 게이트 — CE 산출물 검증
- **Gardener**: 자가 진화 — 군집 성숙도 점검, 개념 진화

---

## 2. 문서 계층

```
IDENTITY.md (핵심 선언, ~100줄)
  └── directives/brand/ (상세 규칙, 11개 파일)
        ├── philosophy.md    ← 브랜드 철학 (3원칙, 동아시아 렌즈)
        ├── foundation.md    ← 이 문서 (5-Layer, 권한, 5 Pillars)
        ├── voice_tone.md
        ├── content_system.md
        ├── design_tokens.md
        └── ...

SYSTEM.md (운영 프로토콜)
  └── directives/agents/ (에이전트별 판단 기준)
        ├── SA.md
        ├── CE.md
        ├── AD.md
        └── CD.md
```

---

## 3. 3-Tier 권한

| Tier | 이름 | 대상 | 수정 조건 |
|------|------|------|-----------|
| FROZEN | 불변 | IDENTITY.md, CD.md | CD 직접 수정만 허용 |
| PROPOSE | 제안 | SA.md, CE.md, AD.md, brand/ | Gardener 제안 → CD 승인 |
| AUTO | 자동 | QUANTA, long_term_memory, signals/ | 에이전트 자동 갱신 |

---

## 4. 5 Pillars — 품질 게이트 (Quality Gate)

> **주의**: 5 Pillars는 산출물을 **평가하는 기준**이다. 브랜드가 향하는 방향은 `philosophy.md` §2 (3원칙) 참조.

WOOHWAHAE 브랜드의 모든 산출물은 이 5가지 기둥으로 평가된다.

### 4.1 진정성 (Authenticity)
- 가면 없는 대화. 취약함을 숨기지 않는다.
- 마케팅 언어 배제. "고객" 대신 "사람", "콘텐츠" 대신 "기록".
- 검증 질문: "이게 진짜 순호의 생각인가?"

### 4.2 실용성 (Practicality)
- Wash & Go 철학 — 지속 가능한 관리.
- 과도한 유지보수가 필요한 것은 본질에 어긋난다.
- 검증 질문: "상대방이 내일부터 실천할 수 있는가?"

### 4.3 우아함 (Elegance)
- 절제된 표현. 여백의 미학. 조용한 고급스러움.
- 빼는 것이 더하는 것보다 어렵다.
- 검증 질문: "이 중 뺄 수 있는 게 있는가?"

### 4.4 정밀함 (Precision)
- 구조적 사고. 기반의 견고함.
- 모호한 표현 금지. 수치와 근거 제시.
- 검증 질문: "애매한 부분이 남아있는가?"

### 4.5 혁신 (Innovation)
- 관습의 파괴. Anti-Uniformity.
- 모두가 하는 방식을 따르지 않는다.
- 검증 질문: "이건 관성인가, 선택인가?"

---

## 5. 금지 사항 (Forbidden)

- 알고리즘 추종 — 숫자보다 한 사람의 진심 어린 공감
- 과도한 장식 — 본질을 가리는 수식어와 복잡성
- 무리 짓기 — 고독의 가치를 훼손하는 소모적 사교
- 속도 압박 — 알고리즘의 업로드 주기에 맞추는 행위

---

## 6. WOOHWAHAE — 첫 번째 인스턴스

LAYER OS는 범용 브랜드 운영 체제다.
WOOHWAHAE는 LAYER OS의 첫 번째 인스턴스다.

```
LAYER OS (The Origin — 기원)
    └── WOOHWAHAE (The First Manifestation — 제1 발현체)
          "Archive for Slow Life"
```

LAYER OS의 구조/파이프라인/에이전트 프레임워크는 브랜드 독립적이다.
WOOHWAHAE 고유 규칙(5 Pillars, 서사, 타겟)은 이 brand/ 디렉토리에 정의된다.

---

**Last Updated**: 2026-02-24
