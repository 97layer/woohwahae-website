# WOOHWAHAE Quick Start Guide

## 빠른 시작 (3분)

### 1. 시스템 테스트
```bash
# 전체 파이프라인 테스트
python3 tests/test_full_pipeline.py
```

### 2. 콘텐츠 추가 및 처리
```python
from core.utils.content_manager import ContentManager
from core.agents.brand_consultant import BrandConsultant
import asyncio

async def quick_test():
    # 1. 매니저 초기화
    manager = ContentManager()
    consultant = BrandConsultant(mock_mode=True)

    # 2. 콘텐츠 추가
    content_id = manager.add_manual_content(
        title="테스트 콘텐츠",
        content="슬로우라이프 철학에 대한 생각...",
        hashtags=["woohwahae", "slowlife"]
    )

    # 3. 브랜드 검증
    content = manager.prepare_for_brand_consultant(content_id)
    result = await consultant.audit_content(content)

    # 4. 결과 확인
    print(f"철학 점수: {result['philosophy_score']}/100")

asyncio.run(quick_test())
```

### 3. 웹사이트 확인
```bash
# 웹사이트 파일 위치
ls website/
# index.html - Magazine B 스타일 메인 페이지
# assets/css/magazine.css - 화이트 미니멀 디자인
```

### 4. 주요 파일 위치

**에이전트**:
- `core/agents/brand_consultant.py` - 브랜드 철학 검증

**유틸리티**:
- `core/utils/content_manager.py` - 콘텐츠 관리
- `core/utils/instagram_crawler.py` - Instagram 크롤러

**테스트**:
- `tests/test_full_pipeline.py` - 전체 시스템 테스트
- `tests/test_content.json` - 테스트 데이터

**웹사이트**:
- `website/index.html` - 메인 페이지
- `website/assets/css/magazine.css` - 스타일시트

## 핵심 개념

### WOOHWAHAE = Archive for Slow Life
최상위 플랫폼 (브랜드 X)

### 7개 병렬 섹션
1. ABOUT - 철학의 선언
2. ARCHIVE - 생각의 기록
3. SHOP - 의식적 소비
4. SERVICE - 실천의 공간
5. PLAYLIST - 감각의 리듬
6. PROJECT - 협업의 실험
7. PHOTOGRAPHY - 순간의 포착

### 철학 검증 기준
- **Philosophy Score**: 0-100점
- **5 Pillars**: 각 0-20점
  - Authenticity (진정성)
  - Practicality (실용성)
  - Elegance (우아함)
  - Precision (정밀함)
  - Innovation (혁신)

## 문제 해결

### Instagram API 인증 실패
→ Content Manager의 manual input 사용

### Claude API 키 없음
→ Mock mode 사용 (`mock_mode=True`)

### 콘텐츠 추가 방법
→ `ContentManager.add_manual_content()` 사용

---

**Editor & Curator**: WOOSUNHO
**Platform**: WOOHWAHAE - Archive for Slow Life