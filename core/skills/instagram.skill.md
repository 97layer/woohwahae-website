# Instagram Content Skill

## Skill ID
`instagram_v1`

## Purpose
WOOHWAHAE의 철학을 Instagram 플랫폼에 최적화하여 발행하되, 알고리즘이 아닌 진정성 중심의 콘텐츠 전략을 실행한다.

## Core Philosophy
> "Anti-Algorithm: 도달률이 아닌 도달의 질, 팔로워 수가 아닌 연결의 깊이"

## Rules

### 1. 72시간 규칙 (Imperfect Publishing)

#### Mandatory Deadline
```python
DRAFT_MAX_AGE = 72  # hours

def enforce_72h_rule(draft):
    age_hours = (now() - draft.created_at).total_seconds() / 3600

    if age_hours >= 72:
        force_publish(draft)  # 완벽하지 않아도 발행
        log_publish("72h_rule_enforced")
```

#### Quality Levels
```yaml
Perfect (100%):   Never shipped (완벽은 영원히 완성 안됨)
Good Enough (80%): Ship it! (72시간 내)
Minimal Viable (60%): Better than nothing (72시간 경과 시)
```

### 2. Content Types & Frequency

#### Primary Content
```yaml
Magazine (월간):
  - 형식: Carousel (5-10 slides)
  - 내용: 깊이 있는 주제 탐구
  - 주기: 월 1회
  - 제작 기간: 1-3주

Insight (주간):
  - 형식: Single Image + Long Caption
  - 내용: 일상의 철학적 관찰
  - 주기: 주 1-2회
  - 제작 기간: 3-5일

Moment (비정기):
  - 형식: Story
  - 내용: 순간의 포착, 과정의 기록
  - 주기: 주 3-5회
  - 제작 기간: 즉시
```

#### Content Mix
```
70% - Philosophical Insights
20% - Behind the Scenes
10% - Service/Product Information
```

### 3. Caption Structure

#### Template
```markdown
[The Hook]
(1-2 sentences, 질문/정의/역설)

[Manuscript]
(2-4 paragraphs, 개인 → 보편)

[Brand Bridge]
(1 paragraph, 자연스러운 연결)

[Afterglow]
(1 sentence, 열린 질문)

---
(No hashtags or minimal, contextual only)
```

#### Character Limits
```yaml
Hook: 50-100 characters
Manuscript: 500-800 characters
Total: 800-1200 characters (Instagram optimal)
```

#### Example Caption
```
미니멀리즘은 덜어냄이 아니라 본질의 발견이다.

8평 반지하 원룸에서 나는 비로소 '나'를 만났다. 넓은 공간도, 많은 물건도 필요하지 않았다. 온전히 내 것인 작은 공간, 그 안의 정숙한 고독.

WOOHWAHAE는 바로 이런 공간입니다. 과하지 않지만 온전한, 조용하지만 확실한.

혹시 당신도 채우기에 지쳐 있지는 않은가?
```

### 4. Visual Requirements

#### Must Have
- [ ] Whitespace ratio ≥ 60%
- [ ] Monochrome or low saturation (< 20%)
- [ ] Natural lighting
- [ ] Single focal point
- [ ] 1080x1080 or 1080x1350 dimensions

#### Must Not Have
- [ ] Filters (Instagram filters 금지)
- [ ] Text overlay (excessive)
- [ ] Multiple focal points
- [ ] High saturation colors
- [ ] Cluttered composition

### 5. Hashtag Strategy

#### Anti-Algorithm Approach
```python
# ❌ OLD WAY (Algorithm Gaming)
hashtags = [
    "#미니멀라이프", "#데일리룩", "#소통해요",
    "#좋아요반사", "#선팔하면맞팔", "#인친"
]  # 30개 해시태그

# ✅ NEW WAY (Contextual, Minimal)
hashtags = [
    "#본질", "#덜어냄"
]  # 0-3개, 맥락에 맞을 때만
```

#### Hashtag Rules
- **Maximum**: 3개
- **Relevance**: 콘텐츠와 직접 관련
- **No Gaming**: 인기 태그 남용 금지
- **Preference**: 해시태그 없음 (가장 이상적)

### 6. Engagement Philosophy

#### What We DON'T Chase
- ❌ Follower count
- ❌ Like count
- ❌ Reach metrics
- ❌ Algorithm favorability
- ❌ Viral posts

#### What We VALUE
- ✅ Meaningful comments (질 높은 대화)
- ✅ Saves (다시 보고 싶은 콘텐츠)
- ✅ Shares (친구에게 보여주고 싶은)
- ✅ DMs (깊은 공감과 연결)
- ✅ Long-term relationships

#### Engagement Response
```python
def respond_to_comment(comment):
    if is_meaningful(comment):  # 3단어 이상, 생각이 담김
        personalized_response = craft_thoughtful_reply(comment)
        reply_within_24h(personalized_response)
    else:
        # 단순 이모지, "좋아요" 등은 응답 안 함
        pass
```

### 7. Posting Schedule

#### Anti-Algorithm Timing
```python
# ❌ Algorithm-Optimized
post_at = "prime_time"  # 오후 6-9시, 알고리즘 선호

# ✅ Human-Centered
post_at = "when_ready"  # 완성되면, 72시간 내
```

#### Preferred Windows (참고용)
```yaml
Morning: 08:00-10:00 (조용한 시작)
Evening: 20:00-22:00 (하루의 성찰)
Weekend: Anytime (여유로운 시간)
```

### 8. Story Strategy

#### Purpose
- 과정의 기록 (behind the scenes)
- 일상의 순간 (authentic moments)
- 시간의 흐름 (time-lapse of work)

#### Content Ideas
```yaml
Process Shots:
  - 헤어컷 과정 (결과가 아닌 과정)
  - 공간 정리 (before → after)
  - 사색의 순간 (책, 차, 창밖)

Micro Insights:
  - 짧은 질문 (5-10 words)
  - 시각적 은유 (이미지 + 1문장)
  - 일상의 철학 (소소한 관찰)

Interaction:
  - 질문 스티커 (깊은 질문만)
  - 투표 (철학적 선택)
  - Quiz (브랜드 철학 테스트)
```

#### Story Frequency
```
Daily: 1-3 stories
Highlights: 월 1개 주제별 모음
Ephemeral: 24시간 후 사라짐 (archive만)
```

## Validation Criteria

### Pre-Publish Checklist
- [ ] **MBQ Check**: Meaning + Brand + Quality
- [ ] **72h Rule**: 72시간 내 작성 완료
- [ ] **Visual**: Design guide 준수
- [ ] **Caption**: Brand voice 일치
- [ ] **Hashtags**: 0-3개, 맥락 적합
- [ ] **CTA**: 강제적 호소 없음

### MBQ Validation
```python
def mbq_check(content):
    """Meaning, Brand, Quality"""

    # M: Meaning (철학 5개 중 1개+)
    meaning = check_philosophical_depth(content)

    # B: Brand (Aesop tone 70%+)
    brand = check_brand_voice(content)

    # Q: Quality (구조적 완결성)
    quality = check_structural_integrity(content)

    return meaning and brand and quality
```

### Post-Publish Analysis
```python
def analyze_performance(post_id):
    """진정성 지표 측정"""

    metrics = {
        "saves_ratio": saves / impressions,
        "comment_depth": avg_comment_length,
        "share_rate": shares / reach,
        "dm_conversations": dm_count,
        "dwell_time": avg_time_spent
    }

    # 높은 saves + 긴 댓글 = 성공
    authenticity_score = (
        metrics['saves_ratio'] * 0.4 +
        metrics['comment_depth'] * 0.3 +
        metrics['dwell_time'] * 0.3
    )

    return authenticity_score
```

## Examples

### ❌ BAD: Algorithm-Driven Post
```
[이미지: 화려한 배경, 텍스트 오버레이]

Caption:
여러분~ 오늘의 꿀팁!💕
팔로우하고 좋아요 눌러주세요!!
지금 DM 주시면 특별 할인 🎁

#미니멀라이프 #데일리 #소통 #좋반 #선팔
#인친 #맞팔 #일상 #힐링 #감성 #분위기
[... 30개 해시태그]
```

**문제점**:
- 알고리즘 게이밍
- 과도한 CTA
- 해시태그 스팸
- 브랜드 정체성 없음

### ✅ GOOD: Brand-Aligned Post
```
[이미지: 여백 70%, 모노크롬, 자연광]

Caption:
완벽을 추구하지만, 불완전함을 수용한다.

이 모순이 나를 가장 잘 설명하는 문장이다.
28권의 일기장, 5만 장의 사진,
모두 완벽을 향한 기록이지만
결국 불완전한 순간들의 집합이다.

WOOHWAHAE도 그렇다.
완벽한 헤어를 추구하지만,
당신의 불완전한 자연스러움을 존중한다.

당신은 어떤 불완전함을 가지고 있는가?
```

**우수한 이유**:
- 역설로 시작 (The Hook)
- 개인 → 보편 전환
- 자연스러운 브랜드 연결
- 열린 질문으로 마무리
- 해시태그 없음

### ✅ GOOD: Story Example
```
[이미지: 헤어컷 진행 중, 흑백]

Text Overlay:
"과정은 결과보다 정직하다"

[스티커: 질문]
"당신이 가장 소중히 여기는 일상의 의식은?"
```

## Integration Points

### For CE (Chief Editor)
```python
from libs.skill_loader import SkillLoader

instagram = SkillLoader.load("instagram_v1")

# 발행 준비 검증
ready_to_publish = instagram.validate_post({
    "caption": caption_text,
    "image": image_path,
    "created_at": draft_timestamp
})

if ready_to_publish['passed']:
    instagram.publish(post_data)
```

### For AD (Art Director)
```python
# Instagram 시각 규격 체크
visual_ok = instagram.check_visual_requirements(image_path)

if not visual_ok['passed']:
    suggestions = instagram.suggest_visual_fixes(image_path)
```

### For Quality Gate
```python
def instagram_quality_gate(post):
    ig_skill = SkillLoader.load("instagram_v1")

    checks = {
        "72h_rule": ig_skill.check_72h_rule(post),
        "mbq": ig_skill.mbq_validation(post),
        "visual": ig_skill.visual_validation(post.image),
        "caption": ig_skill.caption_validation(post.caption)
    }

    return all(checks.values())
```

## Tools & Scripts

### Publishing Tool
```bash
# 발행 준비된 콘텐츠 확인
python libs/skills/instagram/check_ready.py

# 72시간 초과 콘텐츠 강제 발행
python libs/skills/instagram/enforce_72h.py

# 실제 발행
python libs/skills/instagram/publisher.py --post-id 123
```

### Analytics Tool
```bash
# 진정성 지표 분석
python libs/skills/instagram/authenticity_analyzer.py --period 30d

# Output:
# Authenticity Score: 0.78 (High)
# - Saves Ratio: 0.15 (Excellent)
# - Comment Depth: 42 chars (Good)
# - Dwell Time: 8.2s (High)
```

### Caption Generator
```bash
# AI 초안 생성 (brand_voice 스킬 적용)
python libs/skills/instagram/caption_generator.py --topic "minimal_life"

# 브랜드 보이스 검증
python libs/skills/instagram/caption_validator.py draft_caption.txt
```

## Content Calendar

### Planning Approach
```python
# ❌ Algorithm-Optimized
calendar = plan_by_optimal_posting_times()

# ✅ Human-Centered
calendar = plan_by_meaningful_themes()
```

### Monthly Theme Example
```yaml
2026-02:
  Theme: "덜어냄의 미학"
  Magazine: "8평 반지하의 낭만" (Feb 15)
  Insights:
    - Week 1: "공간의 여백"
    - Week 2: "소유의 무게"
    - Week 3: "본질의 발견"
    - Week 4: "고독의 가치"
  Stories: Daily moments of minimalism
```

## Common Pitfalls

### 1. "완벽 함정"
❌ 완벽해질 때까지 발행 안 함
✅ 72시간 내 무조건 발행

### 2. "알고리즘 함정"
❌ 최적 시간대, 해시태그 연구
✅ 준비되면 발행, 맥락 우선

### 3. "숫자 함정"
❌ 팔로워 증가율, 도달률 집착
✅ 댓글의 깊이, 저장 비율 주목

### 4. "트렌드 함정"
❌ Reels 알고리즘, 바이럴 챌린지
✅ 시간을 초월하는 콘텐츠

## Success Metrics (Redefined)

```python
success_metrics = {
    # ❌ Vanity Metrics (무시)
    "follower_count": "irrelevant",
    "like_count": "irrelevant",
    "reach": "irrelevant",

    # ✅ Authenticity Metrics (추적)
    "saves_per_post": "> 10%",
    "avg_comment_length": "> 30 chars",
    "dm_conversations": "> 5 per week",
    "dwell_time": "> 5 seconds",
    "returning_viewers": "> 60%"
}
```

## Version History

- **v1.0** (2026-02-15): Initial skill creation
  - Anti-algorithm philosophy defined
  - 72-hour rule enforced
  - MBQ validation integrated
  - Authenticity metrics established

---

> "Instagram은 도구다. 목적이 아니라 수단이다. 진정성이 먼저다." — 97layerOS
