# WOOHWAHAE Platform System
**Archive for Slow Life**

가속화된 세상에서 주체적으로 살아가기 위한 생각의 기록, 실천의 모음, 의식의 플랫폼

## System Architecture

### 1. Core Philosophy
WOOHWAHAE는 단순한 브랜드가 아닌 **최상위 플랫폼**입니다.
- **Mission**: 가속화된 세상에서 주체적으로 살아가기
- **Vision**: Archive for Slow Life
- **Editor & Curator**: WOOSUNHO

### 2. 7 Parallel Sections
모든 섹션은 하나의 철학을 다르게 표현하는 병렬 구조:

1. **ABOUT** - 철학의 선언
2. **ARCHIVE** - 생각의 기록 (Magazine)
3. **SHOP** - 의식적 소비의 큐레이션
4. **SERVICE** - 실천의 공간 (Hair Atelier)
5. **PLAYLIST** - 감각의 리듬
6. **PROJECT** - 협업의 실험
7. **PHOTOGRAPHY** - 순간의 포착

### 3. System Components

#### Brand Consultant Agent (`core/agents/brand_consultant.py`)
- **Role**: WOOHWAHAE 철학 수호자
- **Features**:
  - 콘텐츠 철학 점수 평가 (0-100)
  - 5 Pillars 기준 검증
  - 7개 섹션 자동 분류
  - Mock mode for testing (API 없이 작동)
- **Usage**:
```python
from core.agents.brand_consultant import BrandConsultant

# Initialize with mock mode for testing
consultant = BrandConsultant(mock_mode=True)

# Audit content
result = await consultant.audit_content(content)
philosophy_score = result['philosophy_score']

# Classify into sections
sections = await consultant.classify_for_sections(content)
```

#### Content Manager (`core/utils/content_manager.py`)
- **Role**: 콘텐츠 입력 및 관리
- **Features**:
  - Manual content input (Instagram API 대체)
  - Content storage and retrieval
  - Processing pipeline integration
- **Usage**:
```python
from core.utils.content_manager import ContentManager

manager = ContentManager()

# Add manual content
content_id = manager.add_manual_content(
    title="#읽는미장",
    content="헤어 작업 설명...",
    content_type="service",
    hashtags=["woohwahae", "slowlife"]
)

# Process pending content
pending = manager.get_pending_content()
```

#### Instagram Crawler (`core/utils/instagram_crawler.py`)
- **Role**: Instagram 콘텐츠 자동 수집
- **Status**: Authentication required (Manual input recommended)
- **Alternative**: Use Content Manager for manual input

#### Website (`website/`)
- **Design**: Magazine B inspired minimalist design
- **Structure**:
  - `index.html` - Main platform page with 7 sections
  - `assets/css/magazine.css` - White minimalist design system
  - Grid-based layout with numbered sections

### 4. Design System

**Magazine B Style Guidelines**:
- Background: Pure white (#FFFFFF)
- Text: Pure black (#000000)
- Accent: Magazine B Red (#FF0000)
- Typography: Clean sans-serif
- Layout: Grid-based with 40%+ white space
- Sections: Numbered boxes (01-07)

### 5. Testing

#### Run Pipeline Test
```bash
# Test with mock data
python3 tests/test_pipeline.py

# Full pipeline test
python3 tests/test_full_pipeline.py
```

#### Expected Output
- Philosophy Score: 70-95/100 for aligned content
- 5 Pillars: Each scored 0-20
- Section Classification: Automatic assignment to 1 of 7 sections

### 6. Content Processing Pipeline

```
Instagram/Manual Input
        ↓
Content Manager (Storage)
        ↓
Brand Consultant (Validation)
        ↓
    /   |   \
Audit  Score  Classify
        ↓
7 Sections Distribution
        ↓
Website Display
```

### 7. Installation & Setup

```bash
# Install dependencies
pip install anthropic  # For Claude API (optional)
pip install instaloader  # For Instagram (optional)
pip install beautifulsoup4  # For web scraping

# Set environment variable (if using Claude API)
export ANTHROPIC_API_KEY="your-key-here"

# Run tests
python3 tests/test_full_pipeline.py
```

### 8. Current Status

✅ **Completed**:
- Brand Consultant Agent with mock mode
- Content Manager for manual input
- 7-section classification system
- Magazine B style website design
- Full pipeline testing

⚠️ **Requires Authentication**:
- Instagram API (use manual input instead)
- Claude API (use mock mode for testing)

### 9. Usage Example

```python
# Complete workflow example
import asyncio
from core.agents.brand_consultant import BrandConsultant
from core.utils.content_manager import ContentManager

async def process_content():
    # Initialize
    manager = ContentManager()
    consultant = BrandConsultant(mock_mode=True)

    # Add content
    manager.add_manual_content(
        title="슬로우라이프 실천",
        content="일상에서 천천히 살아가는 방법...",
        hashtags=["woohwahae", "slowlife"]
    )

    # Process
    for item in manager.get_pending_content():
        content = manager.prepare_for_brand_consultant(item['id'])

        # Audit
        result = await consultant.audit_content(content)
        print(f"Philosophy Score: {result['philosophy_score']}/100")

        # Classify
        sections = await consultant.classify_for_sections(content)
        print(f"Primary Section: {sections[0][0].value}")

        # Update status
        manager.update_content_status(item['id'], 'processed', result)

# Run
asyncio.run(process_content())
```

### 10. Philosophy Verification

All content must align with:

**Core Values**:
1. **생각의 기록** - 슬로우라이프에 대한 사색과 철학
2. **실천의 모음** - 일상에서 구현하는 구체적 방법들
3. **의식의 플랫폼** - 의식적으로 선택하고 살아가는 커뮤니티

**5 Pillars**:
1. **Authenticity** (진정성) - 가면 없는 대화
2. **Practicality** (실용성) - 지속 가능한 관리
3. **Elegance** (우아함) - 절제된 표현
4. **Precision** (정밀함) - 구조적 사고
5. **Innovation** (혁신) - 관습의 파괴

---

## Support

**Editor & Curator**: WOOSUNHO
**Instagram**: @woosunhokr
**Platform**: WOOHWAHAE - Archive for Slow Life

---

*"WOOHWAHAE는 브랜드가 아니라 삶의 방식입니다."*