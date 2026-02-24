# 97LAYER 시스템 가치 평가 및 수치화 보고서

**작성일**: 2026.02.19
**작성자**: 97LAYER 수석 오케스트레이터

---

## 1. Executive Summary (요약)

**"잠재력은 높으나 현금화 불가능한 자산과 높은 이자율의 기술 부채가 공존"**

현재 시스템은 **약 5.7만 라인(SLOC)**의 코드 베이스와 **10개의 브랜드 아티클**을 보유하고 있으나, **112개의 미해결 과제(TODO)**와 **44개의 하드코딩된 설정**으로 인해 **운영 리스크(Liability)**가 자산 가치를 상쇄하고 있습니다.

- **추정 자산 가치**: **$3,500** (단순 개발 공수 기준)
- **추정 기술 부채**: **-$2,000** (리팩토링 및 리스크 비용)
- **순자산 가치 (NAV)**: **$1,500 (잠정)** + **α (브랜드 프리미엄)**

---

## 2. 자산 가치 평가 (Assets Valuation)

### A. 유형 자산 (Tangible Technical Assets)

1. **Codebase Volume**: ~57,355 Lines (Total)
    - 유의미한 코어 로직(`core/`): 약 5,000 Lines 추정
    - 웹사이트 스타일링(`style.css` 등): 약 3,000 Lines
    - *평가*: 1인 개발 기준 **약 2-3개월 분량의 지적 노동**이 투입됨.
2. **Content Assets**:
    - **Issues**: 10건 (Manifesto, Slow Life 등) → 브랜드 철학이 담긴 고유 자산.
    - **Signals**: 72건 (Knowledge/signals) → 원천 데이터 확보 상태.
3. **Infrastructure**:
    - Docker 컨테이너화 완료 (비록 배포 방식은 불안정하나 명세서는 존재).

### B. 무형 자산 (Intangible Brand Assets)

- **Identity**: "WOOHWAHAE"라는 브랜드 네러티브와 "Remove the Noise"라는 철학적 기반.
- **Domain**: `woohwahae.kr` (평가액 산정 불가, 브랜드 핵심).

---

## 3. 부채 평가 (Liability Assessment)

### A. 기술 부채 (Technical Debt)

1. **Explicit Debt (명시적 부채)**:
    - **TODOs**: 112개 (코드 곳곳에 산재한 "나중에 할 것들").
    - **FIXMEs**: 16개 (즉시 고쳐야 할 잠재적 버그).
    - **Hardcoded IPs**: 44건 (`136.109.201.201` 등).
    - *비용 환산*: 이를 해결하기 위해 최소 **40시간(1주)**의 전임 개발자 공수 필요.

### B. 운영 리스크 (Operational Risk)

- **Bus Factor = 1**: 97LAYER(1인) 부재 시 시스템 운영 불가.
- **Pipeline Failure**: `woohwahae_pipeline.py` 작동 불능 → 콘텐츠 자동 수급 중단.
- **Data Fragility**: `SYSTEM_ENFORCEMENT_REPORT.md` 유령 파일 사태로 증명된 파일 시스템 신뢰도 저하.
- **Asset Void**: 관리되지 않는 이미지 자산 (`asset_registry.json` Empty).

---

## 4. 수치화 지표 (Key Metrics)

| 지표 (Metric) | 수치 (Value) | 해석 |
| :--- | :--- | :--- |
| **Total Lines of Code** | **57,355** | 라이브러리/중복 포함. 실제 코어는 10% 미만 추정. |
| **Core Documentation** | **7** Files | `WEBSITE_STRUCTURE.md` 등. 설계 문서는 양호. |
| **Commit Velocity** | **86** Commits | 초기 구축 단계. 아직 안정화되지 않음. |
| **Debt Ratio** | **High** | 코드 라인 대비 TODO 비율이 높음 (약 1/500). |
| **Asset Utilization** | **Low** | 수집된 Signal(72개) 대비 발행 Issue(10개) 전환율 낮음. |

---

## 5. 결론 및 제언

**"지금은 빚을 갚고 자산을 유동화할 때"**

1. **부채 상환 우선**: 44개의 하드코딩된 IP를 환경변수로 분리하는 것이 시급합니다. 이것만 해결해도 보안 리스크 자산 가치는 급상승합니다.
2. **자산 유동화**: 72개의 수집된 Signal을 가공하여 아티클 수를 10개에서 50개로 늘리면, 시스템의 실질적 가치(트래픽, 브랜드 파워)가 폭발적으로 증가합니다.
3. **파이프라인 정상화**: '죽은 코드'를 살려 자동화 수익(콘텐츠 생산)을 창출하십시오.

현재 이 시스템은 **"다듬어지지 않은 원석(Rough Diamond)"** 상태입니다. 세공(Refactoring) 없이는 돌덩이에 불과하나, 제대로 깎으면 보석이 됩니다.

이상.
