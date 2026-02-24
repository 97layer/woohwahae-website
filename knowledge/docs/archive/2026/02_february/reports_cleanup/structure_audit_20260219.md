# 시스템 구조 정밀 진단 보고서 (System Structure Audit)

**작성일**: 2026.02.19
**작성자**: 97LAYER 수석 오케스트레이터

---

## 1. 진단 개요

사용자의 요청("현재 시스템 구조먼저 점검 한다")에 따라, `LAYER OS`의 파일 시스템과 아키텍처를 정밀 분석했습니다.
특히 **'라이프스타일 순환 시스템(Think-Touch-Feel)'**으로의 전환을 위해 현재 구조가 적합한지 검토했습니다.

---

## 2. 현황 분석 (AS-IS Analysis)

### A. Root Pollution (루트 오염)

`website/` 디렉토리에 정체 불명의 파일들이 산재해 있어 유지보수성을 크게 저해하고 있습니다.

- **Unmanaged Images**: `20251114_012857(1).jpg`, `대표심볼.JPG` 등 날짜/한글 파일명의 이미지가 루트에 방치됨 (`assets/img/`로 이동 필요).
- **Loose Scripts**: `add_analytics.py`, `fix_nav.py`, `serve.py`, `update_nav.sh` 등 관리용 스크립트가 웹 소스와 섞여 있음 (`scripts/` 또는 `core/utils/`로 격리 필요).
- **Duplicate Configs**: `CNAME`, `robots.txt`, `sitemap.xml` 등이 분산되어 있어 배포 시 누락 위험 존재.

### B. Structural Inconsistency (구조적 불일치)

`WEBSITE_STRUCTURE.md`에 정의된 이상적인 구조와 실제 구현 간의 괴리가 큽니다.

- **Photography**: 기획상 `/photography/`는 비주얼 아카이브여야 하나, 현재는 단일 파일(`photography.html`)로 존재하며 확장성이 없음.
- **Shop**: `shop.html`은 존재하나 내부 콘텐츠나 결제 로직 연결이 전무함 (Empty Shell).
- **Service**: 단순 `service.html`로 존재. 'Ritual'로서의 예약 시스템을 담기에는 구조가 빈약함.

### C. Core System Logic

- `core/` 디렉토리는 `admin`, `agents`, `bridges`, `daemons`로 나름의 체계를 갖추고 있으나, `website`와의 연동(Static Site Generation 등)이 명확하지 않음.
- `woohwahae_pipeline.py` (유령 파이프라인) 문제는 여전히 해결되지 않음.

---

## 3. 개선 방향 (TO-BE Strategy)

### A. "The Cleanroom Protocol" (루트 대청소)

- **이미지 격리**: 모든 JPG/PNG 파일을 `website/assets/img/legacy/`로 강제 이동 및 링크 수정.
- **스크립트 격리**: 관리용 파이썬/쉘 스크립트를 `core/tools/`로 이동.

### B. "The Trinity Architecture" (삼위일체 구조)

새로운 브랜드 철학(Think-Touch-Feel)을 폴더 구조에 반영하여 직관성을 높입니다.

| 철학 | 기존 폴더 | 변경 제안 | 역할 |
| :--- | :--- | :--- | :--- |
| **Think** | `archive/` | `website/journal/` | 브랜드 저널리즘, 에세이 (텍스트 중심) |
| **Touch** | `shop/` | `website/objects/` | 큐레이션 오브제, 디지털 굿즈 (물성 중심) |
| **Feel** | `service/` | `website/ritual/` | 예약, 공간 안내, 티 세레모니 (경험 중심) |

### C. "Automated Deployment" (파이프라인 복구)

- `core/system/woohwahae_pipeline.py`를 되살려 `journal` 폴더에 자동으로 콘텐츠를 주입하는 로직 구축.

---

## 4. 제언

지금의 구조는 **"정리되지 않은 서랍장"**과 같습니다. 물건은 다 들어있지만, 어디에 무엇이 있는지 찾기 어렵고 새로운 물건(기능)을 넣을 공간이 없습니다.
**'오브젝트(Object)'와 '리추얼(Ritual)'**을 담을 그릇을 준비하려면, 먼저 이 서랍장을 비우고 칸막이를 새로 질러야 합니다.

**우선순위 1**: `website/` 루트의 지저분한 파일들부터 청소(Clean-up)할 것을 권장합니다.

이상.
