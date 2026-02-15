# Design Guide Skill

## Skill ID
`design_guide_v1`

## Purpose
WOOHWAHAE의 시각 정체성을 구현하여 모든 비주얼 콘텐츠에서 "여백 60%+, 본질만 남기기" 철학을 실현한다.

## Core Philosophy
> "장식이 아닌 구조, 색이 아닌 형태, 가득함이 아닌 비어있음"

## Rules

### 1. 여백 60%+ (Whitespace Rule)

#### Canvas Composition
```
Total Canvas: 100%
├── Content: 35-40%
└── Whitespace: 60-65%
```

#### Measurement Method
```python
def calculate_whitespace_ratio(image):
    total_pixels = image.width * image.height
    content_pixels = count_non_white_pixels(image)
    whitespace_ratio = 1 - (content_pixels / total_pixels)

    return whitespace_ratio >= 0.60  # Must be 60%+
```

#### Visual Formula
```
[============== Canvas ==============]
[=====Content=====]         [Whitespace]
     40%                        60%
```

### 2. 모노크롬 우선 (Monochrome Priority)

#### Color Palette
```yaml
Primary:
  - Pure White: #FFFFFF
  - Pure Black: #000000
  - Gray Scale: #EEEEEE → #333333 (10 steps)

Accent (제한적 사용):
  - Deep Navy: #1A2332
  - Warm Gray: #8B8680
  - Cream: #F5F4F0

Forbidden:
  - 형광색 (Neon colors)
  - 고채도 (High saturation > 50%)
  - 원색 (Primary colors: red, blue, yellow)
```

#### Saturation Rule
```python
def validate_color_saturation(image):
    hsv = convert_to_hsv(image)
    avg_saturation = calculate_avg_saturation(hsv)

    # 평균 채도 < 20%
    return avg_saturation < 0.20
```

### 3. 자연광 우선 (Natural Light Priority)

#### Lighting Principles
- **Golden Hour**: 일출 후 1시간, 일몰 전 1시간
- **Soft Light**: 확산된 자연광 (직사광선 회피)
- **Shadows**: 그림자를 제거하지 않고 활용
- **Minimal Flash**: 인공 조명 최소화

#### Validation
```python
lighting_checklist = {
    "natural_light": True,      # 자연광 사용
    "harsh_shadows": False,     # 강한 그림자 없음
    "artificial_glow": False,   # 인공 빛 없음
    "golden_hour": True         # 가능하면 골든아워
}
```

### 4. Typography Hierarchy

#### Font System
```yaml
Korean:
  Primary: "Noto Serif CJK KR"
  Weight: 400 (Regular), 700 (Bold)
  Usage: Body text, Headers

English:
  Primary: "Crimson Text"
  Weight: 400, 600, 700
  Usage: Body text, Headers

Fallback:
  System: -apple-system, sans-serif
```

#### Size Hierarchy
```css
h1: 2.5rem   /* 40px - Rare use */
h2: 2.0rem   /* 32px - Section titles */
h3: 1.5rem   /* 24px - Sub-sections */
p:  1.0rem   /* 16px - Body text */
small: 0.875rem /* 14px - Caption */
```

#### Line Height
```css
body {
  line-height: 1.8;  /* 180% - 가독성과 여백 */
}

h1, h2, h3 {
  line-height: 1.4;  /* 140% - 제목은 좀 더 긴밀 */
}
```

### 5. Layout Principles

#### Grid System
```
12-column grid with wide gutters
Column: 60px
Gutter: 40px (넓은 여백)
Margin: 80px (모바일 24px)
```

#### Spacing Scale
```python
spacing = {
    'xs': '4px',
    'sm': '8px',
    'md': '16px',
    'lg': '32px',
    'xl': '64px',
    'xxl': '128px'
}

# 여백은 항상 크게
prefer_large_spacing = ['lg', 'xl', 'xxl']
```

### 6. Image Treatment

#### Style Guidelines
- **Composition**: 중심 오프셋 (Rule of Thirds)
- **Focus**: 단일 피사체, 배경 블러 허용
- **Texture**: 자연스러운 질감 보존
- **Editing**: 최소한의 보정 (색감 조정 자제)

#### Forbidden Effects
- ❌ 비네팅 (Vignette)
- ❌ 과도한 필터
- ❌ HDR 효과
- ❌ 인스타그램 필터

#### Approved Adjustments
- ✅ Exposure correction (±0.3 EV)
- ✅ Contrast adjustment (minimal)
- ✅ Crop to improve composition
- ✅ Desaturation (toward B&W)

## Validation Criteria

### Automated Checks
- [ ] **Whitespace Ratio** ≥ 60%
- [ ] **Color Saturation** < 20% (average)
- [ ] **Font Usage**: Approved fonts only
- [ ] **File Size**: < 500KB (web optimization)
- [ ] **Dimensions**: 1080x1080 or 1080x1350 (Instagram)

### Manual Review
- [ ] **Minimalism**: 불필요한 요소 제거됨
- [ ] **Balance**: 시각적 균형 확보
- [ ] **Breathing Room**: 답답하지 않음
- [ ] **Brand Fit**: WOOHWAHAE 정체성 부합

### The 5-Second Test
```python
def five_second_test(image):
    """
    5초 안에 다음이 느껴지는가?
    """
    return {
        "calm": True,        # 평온함
        "elegant": True,     # 우아함
        "spacious": True,    # 여유로움
        "authentic": True,   # 진정성
        "minimal": True      # 미니멀함
    }
```

## Examples

### ❌ BAD: Cluttered Design
```
[ 화려한 배경 + 큰 텍스트 + 여러 색상 + 테두리 + 이모지 ]
- 여백 < 30%
- 고채도 컬러 사용
- 과도한 요소
- 복잡한 구성
```

**문제점**:
- 시선이 분산됨
- 브랜드 정체성 흐릿
- 알고리즘 집착 (시선 끌기)

### ✅ GOOD: Minimal Design
```
[                                    ]
[                                    ]
[          본질로 돌아가는            ]
[             여정                   ]
[                                    ]
[                                    ]
```

**우수한 이유**:
- 여백 75%+
- 모노크롬 (검정 텍스트, 흰 배경)
- 중심 정렬, 균형감
- 본질에 집중

### ✅ GOOD: Product Photo
```
Composition:
- Subject: 단일 제품 (면도칼, 빗 등)
- Background: Pure white or soft gray
- Lighting: Soft natural light from left
- Shadow: Soft, natural shadow on right
- Whitespace: 70%
```

### ✅ GOOD: Typography Example
```css
.content {
  font-family: 'Noto Serif CJK KR';
  font-size: 16px;
  line-height: 1.8;
  letter-spacing: 0.02em;
  color: #333333;

  /* 여백 강조 */
  margin-bottom: 32px;
  padding: 64px 80px;
}

.title {
  font-size: 32px;
  font-weight: 700;
  line-height: 1.4;
  margin-bottom: 48px;  /* 넉넉한 여백 */
}
```

## Integration Points

### For AD (Art Director)
```python
from libs.skill_loader import SkillLoader

skill = SkillLoader.load("design_guide_v1")

# 이미지 검증
validation = skill.validate_image(image_path)

if not validation['passed']:
    print(f"Failed checks: {validation['failed']}")

    # 자동 개선 제안
    suggestions = skill.suggest_improvements(image_path)
```

### For CE (Chief Editor)
```python
# 텍스트 레이아웃 검증
layout_validation = skill.validate_typography(html_content)

# 여백 충분한가?
# 폰트 규칙 준수하는가?
# 가독성 확보되었는가?
```

### For Quality Gate
```python
def visual_quality_check(asset_path):
    design_guide = SkillLoader.load("design_guide_v1")

    checks = {
        "whitespace": design_guide.check_whitespace(asset_path),
        "color": design_guide.check_color_saturation(asset_path),
        "dimensions": design_guide.check_dimensions(asset_path),
        "file_size": design_guide.check_file_size(asset_path)
    }

    return all(checks.values())
```

## Tools & Scripts

### Whitespace Analyzer
```bash
python libs/skills/design_guide/whitespace_analyzer.py image.jpg
# Output:
# Whitespace Ratio: 0.68 (68%) ✓ PASS
# Content Area: 32%
# Recommendation: Optimal balance
```

### Color Saturation Checker
```bash
python libs/skills/design_guide/saturation_checker.py image.jpg
# Output:
# Average Saturation: 0.15 (15%) ✓ PASS
# High Saturation Pixels: 2%
# Recommendation: Excellent monochrome adherence
```

### Layout Validator
```bash
python libs/skills/design_guide/layout_validator.py design.html
# Output:
# Font Usage: ✓ PASS (Approved fonts only)
# Spacing: ✓ PASS (Consistent scale)
# Hierarchy: ✓ PASS (Clear structure)
```

### Image Optimizer
```bash
python libs/skills/design_guide/image_optimizer.py input.jpg output.jpg
# Actions:
# - Resize to 1080x1080
# - Compress to < 500KB
# - Apply subtle desaturation
# - Preserve quality
```

## Design Templates

### Instagram Post Template
```yaml
Canvas: 1080x1080px
Layout:
  - Top Margin: 200px (18.5%)
  - Bottom Margin: 200px (18.5%)
  - Side Margins: 120px (11% each)
  - Content Area: 840x680px (57%)

Typography:
  - Title: 40px, Noto Serif CJK KR Bold
  - Body: 24px, Noto Serif CJK KR Regular
  - Line Height: 1.8

Colors:
  - Background: #FFFFFF
  - Text: #333333
  - Accent: #8B8680 (if needed)
```

### Story Template
```yaml
Canvas: 1080x1920px
Layout:
  - Safe Area: 1080x1680px (avoid top/bottom 120px)
  - Top Margin: 300px (15.6%)
  - Bottom Margin: 300px (15.6%)
  - Side Margins: 80px (7.4% each)

Focus Area: Center 920x1320px
Whitespace: 65%+
```

## Common Pitfalls

### 1. "채우기 함정"
❌ 빈 공간을 채워야 한다는 강박
✅ 비어있음이 메시지다

### 2. "컬러 함정"
❌ 눈길을 끌기 위한 밝은 색상
✅ 모노크롬의 절제된 우아함

### 3. "복잡성 함정"
❌ 더 많은 요소 = 더 좋은 디자인
✅ 최소한의 요소 = 명확한 메시지

### 4. "트렌드 함정"
❌ 인스타그램 트렌드 따라하기
✅ 시대를 초월하는 클래식

## Design Principles Summary

```python
principles = {
    "Less is More": "덜어낼수록 강력하다",
    "Space Speaks": "여백이 말한다",
    "Light Matters": "빛이 중요하다",
    "Structure over Decoration": "장식보다 구조",
    "Timeless not Trendy": "트렌드보다 시간성"
}
```

## Continuous Learning

디자인 스킬도 피드백을 통해 진화합니다:

```python
def learn_from_performance(post_id):
    # 높은 engagement + 낮은 bounce rate = 좋은 디자인
    metrics = get_visual_metrics(post_id)

    if metrics['dwell_time'] > 5.0 and metrics['saves'] > avg:
        design = extract_design_patterns(post_id)
        add_to_best_practices(design)
```

## Version History

- **v1.0** (2026-02-15): Initial skill creation
  - Consolidated from: IDENTITY.md, visual_identity_guide.md
  - Added automated validation tools
  - Created design templates
  - Integrated with skill_loader system

---

> "디자인은 보이는 것이 아니라 느껴지는 것이다." — 97layerOS
