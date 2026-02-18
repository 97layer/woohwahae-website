---
name: instagram_content_curator
description: WOOHWAHAE 브랜드 정체성 기반 인스타그램 콘텐츠 생성. CE→CD 파이프라인 아웃풋 기준 준수.
tools:
  - Grep
  - Read
  - Bash
version: 2.0.0
updated: 2026-02-18
---

# Instagram Content Curator Skill

97layerOS 파이프라인 최종 산출물(CE 아웃풋)을 인스타그램 게시용 포맷으로 변환.

## CE 아웃풋 구조 (input)

```json
{
  "instagram_caption": "...",
  "hashtags": ["#slowlife", ...],
  "archive_essay": "..."
}
```

## 포맷 기준

- 캡션: 3~5문장. 명사/동사 중심. 여백 활용
- 해시태그: 10~15개. 브랜드 태그 포함 (#woohwahae, #slowlife)
- 이모지 사용 금지
- 볼드 최소화 — 텍스트 자체의 리듬감 우선
- 언어: 한국어

## 이미지 소스 우선순위

1. 순호 제공 실사 사진 (최우선)
2. Imagen API 생성 (미니멀, 자연광)
3. Unsplash fallback

## CD 승인 기준 (approved: true 조건)

- WOOHWAHAE 5 Pillars 부합 여부
- 캡션 밀도 (과도한 설명 없음)
- 브랜드 톤앤매너 일치

## 참조

- IDENTITY.md: WOOSUNHO Editor 정체성
- directives/CD_SUNHO.md: 최종 승인 기준
