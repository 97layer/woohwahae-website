# Brand OS — directives/brand/ 인덱스

> **목적**: WOOHWAHAE 브랜드의 모든 규칙과 기준을 구조화한 문서 모음.
> **권한**: 참조용. 수정 시 CD 승인 필요.
> **갱신**: 2026-02-24

---

## 📘 통합 매뉴얼 (SSOT)

| 파일 | 설명 |
|------|------|
| **`BRAND_MANUAL.md`** | **단일 정본**. 전체 브랜드 규칙 통합 (8 Chapters). 개별 파일과 충돌 시 이 문서 우선. |

---

## 문서 목록

| # | 파일 | 영역 | 내용 |
|---|------|------|------|
| 1 | `foundation.md` | 기반 | 5-Layer OS 모델, 3-Tier 권한, 5 Pillars 상세, 문서 계층 |
| 2 | `story.md` | 서사 | 개인 서사 — 반지하 8평, 헤어디자이너에서 슬로우라이프로 |
| 3 | `audience.md` | 독자 | 타겟 프로필, 페르소나 3종, Anti-Audience 정의 |
| 4 | `voice_tone.md` | 언어 | 어조 스펙트럼, 허용/금지 키워드, 채널별 규칙 |
| 5 | `content_system.md` | 콘텐츠 | 에세이 구조(Hook-Story-Core-Echo), 길이 규칙, 품질 게이트 |
| 6 | `design_tokens.md` | 디자인 | CSS 변수 전체: Colors, Typography, Spacing, Breath, Photography |
| 7 | `experience_map.md` | 경험 | 7개 섹션 IA, 네비게이션, 기술 스택 |
| 8 | `service_ritual.md` | 서비스 | 아틀리에 철학, Client Journey 7단계, 커뮤니케이션 |
| 9 | `teaching.md` | 교육 | 지식 공유 철학, 교육 포맷 3종 |
| 10 | `roadmap.md` | 전략 | 2026~2028 전략, 분기별 마일스톤 |

---

## 사용법

- **IDENTITY.md**: 핵심 선언만 유지. 상세 규칙은 이 디렉토리 참조.
- **에이전트 로딩**: `agent_router.py` → `AGENT_DIRECTIVES`에서 역할별 문서 지정.
- **콘텐츠 생성 시**: CE → `voice_tone.md` + `content_system.md` 필수 참조.
- **디자인 검증 시**: AD → `design_tokens.md` 필수 참조.
- **최종 승인 시**: CD → `foundation.md` + 5 Pillars 기준 대조.

---

**Last Updated**: 2026-02-24
