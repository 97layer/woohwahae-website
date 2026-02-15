# 🎙️ Brand Voice Skill

## Overview

WOOHWAHAE 브랜드의 독특한 톤앤매너를 정확히 구현하여 모든 커뮤니케이션에서 일관된 브랜드 보이스를 유지한다.

## Core Philosophy

> "Remove the Noise, Reveal the Essence"
> 소음을 제거하고 본질을 드러내는 절제된 언어

## Rules

### 1. Aesop Benchmark (70%+)

**이솝처럼**:

- 절제된 표현
- 본질에 집중
- 문학적 품격

**이솝과 달리**:

- 한국적 정서 (고독, 여백, 선비 정신)
- 따뜻한 인간미
- 개인적 이야기에서 출발

### 2. The Hook Formula

```python
opening_patterns = [
    "정의: {A}는 {B}다",      # 명료한 선언
    "역설: {A}이지만 {B}",    # 긴장감 생성
    "질문: 왜 {question}?"    # 호기심 자극
]
```

**예시**:

```markdown
- "정의: 미니멀리즘은 덜어냄이 아니라 본질의 발견이다"
- "역설: 완벽을 추구하지만, 불완전함을 수용한다"
- "질문: 왜 우리는 채우는 것에만 익숙한가?"
```

### 3. Manuscript Structure (90%)

```markdown
1. Opening (10%): 개인적 경험/관찰
2. Bridge (20%): 개인 → 보편 전환
3. Core (50%): 핵심 메시지
4. Integration (10%): 브랜드 연결
5. Afterglow (10%): 여운/초대
```

### 4. Language Principles

#### ✅ DO (해야 할 것)

- **절제**: 꼭 필요한 단어만
- **정갈**: 깔끔한 문장 구조
- **깊이**: 표면이 아닌 본질
- **진정성**: 꾸미지 않은 솔직함
- **여운**: 끝나지 않는 생각거리

#### ❌ DON'T (하지 말아야 할 것)

- 과장된 수식어 (대박, 완전, 진짜)
- 유행어/신조어 남발
- 강제적 호소 (지금 바로!, 꼭!)
- 알고리즘 용어 (팔로우, 좋아요, 구독)
- 익명의 대중 호칭 (여러분)

### 5. Tone & Voice

**Voice** (불변하는 성격):

- Thoughtful (사려 깊은)
- Authentic (진정성 있는)
- Precise (정밀한)
- Elegant (우아한)

**Tone** (상황별 변주):

- **콘텐츠**: [성찰적, 철학적]
- **서비스**: [전문적, 따뜻함]
- **대화**: [경청하는, 존중하는]

### 6. Typography & Format

**국문**:

```markdown
Font: Noto Serif CJK KR (본명조)
Style: 정갈, 클래식, 시간성
```

**영문**:

```markdown
Font: Crimson Text
Style: 세리프, 우아, 가독성
```

**문단 구조**:

- 1문단 = 1-3문장
- 문단 간 여백 1줄
- 섹션 구분 명확

## Validation Criteria

### Automated Checks

- [ ] **Aesop Score** ≥ 0.70 (tone_analyzer.py)
- [ ] **Readability** 8-10 grade (문장 복잡도)
- [ ] **Keyword Density** < 2% (특정 단어 반복)
- [ ] **Banned Words** = 0 (금지어 미사용)

### Manual Review

- [ ] **Authenticity**: 진정성이 느껴지는가?
- [ ] **Essence**: 본질을 다루고 있는가?
- [ ] **Consistency**: 브랜드 철학과 일치하는가?
- [ ] **Afterglow**: 여운이 남는가?

### The Hook Test

- [ ] 첫 문장에서 호기심 유발
- [ ] 첫 문단에서 정의/역설/질문 사용
- [ ] 3초 안에 핵심 파악 가능

### Manuscript Test

- [ ] 개인 → 보편 전환 명확
- [ ] 핵심 메시지 단일하고 명료
- [ ] 브랜드 철학 5가지 중 1개+ 포함
- [ ] 마지막에 열린 질문 또는 성찰 유도

## Examples

### ❌ BAD: Algorithm-Driven

```markdown
여러분! 오늘은 완전 대박 꿀팁 공유할게요!
팔로우하고 좋아요 눌러주세요 ❤️❤️
지금 바로 DM 주시면 할인 쿠폰 드려요!

#미니멀라이프 #데일리룩 #소통해요
```

**문제점**:

- 알고리즘 집착 (팔로우, 좋아요)
- 과장된 표현 (대박, 완전)
- 강제적 호소 (지금 바로)
- 해시태그 범람

### ✅ GOOD: Brand Voice

```markdown
미니멀리즘은 덜어냄이 아니라
본질의 발견이다.

8평 반지하 원룸에서
나는 비로소 '나'를 만났다.
넓은 공간도, 많은 물건도 필요하지 않았다.
온전히 내 것인 작은 공간,
그 안의 정숙한 고독.

혹시 당신도
채우기에 지쳐 있지는 않은가?
```

**우수한 이유**:

- The Hook (정의) 사용
- 개인적 경험 (8평 반지하)
- 본질에 집중 (공간이 아닌 내면)
- 여운 남기는 질문

### ✅ GOOD: Service Description

```markdown
WOOHWAHAE는 헤어 아틀리에를 넘어
삶의 태도를 제안합니다.

Wash & Go.
전문가 없이도 온전할 수 있도록.
덜어냄의 미학을 직접 경험하는 공간.

진정한 자유는
의존에서 벗어날 때 시작됩니다.
```

**우수한 이유**:

- 브랜드 철학 명확 (실용성, 진정성)
- 절제된 문장
- 깊은 메시지 (자유 = 독립)
- 강요 없는 초대

## Integration Points

### For CE (Chief Editor)

```python
from libs.skill_loader import SkillLoader

skill = SkillLoader.load("brand_voice_v1")
validated_text = skill.validate(raw_content)

if validated_text['aesop_score'] < 0.70:
    revised = skill.enhance(raw_content)
```

### For SA (Strategy Analyst)

```python
# 패턴 분석 시 브랜드 보이스 체크
analysis = analyze_signal(raw_signal)
if analysis['matches_brand_voice']:
    mark_as_valuable(raw_signal)
```

### For Quality Gate

```python
# 발행 전 필수 검증
def pre_publish_check(content):
    brand_voice = SkillLoader.load("brand_voice_v1")
    results = brand_voice.full_validation(content)

    if not results['passed']:
        raise PublishBlockedError(results['issues'])
```

## Tools & Scripts

### Validation Tool

```bash
# 단일 파일 검증
python libs/skills/brand_voice/validator.py content.md

# 배치 검증
python libs/skills/brand_voice/validator.py knowledge/content/*.md
```

### Aesop Score Calculator

```bash
python libs/skills/brand_voice/aesop_scorer.py --file content.md
# Output: Aesop Score: 0.78 (PASS)
```

### Tone Analyzer

```bash
python libs/skills/brand_voice/tone_analyzer.py content.md
# Output:
# - Thoughtfulness: 0.85
# - Authenticity: 0.92
# - Elegance: 0.73
```

## Common Pitfalls

### 1. "알고리즘 함정"

❌ 팔로우, 좋아요, 구독, 알림 설정
✅ 함께하다, 기억하다, 머물다, 연결되다

### 2. "과장 함정"

❌ 완전 대박, 진짜 최고, 무조건
✅ 조용히, 확실하게, 온전히

### 3. "강요 함정"

❌ 지금 바로, 꼭 해야, 반드시
✅ 초대합니다, 제안합니다, 가능합니다

### 4. "익명 함정"

❌ 여러분, 모두, 다들
✅ 당신, 우리, (직접 호칭 최소화)

## Continuous Learning

이 스킬은 매 콘텐츠 생성 시 피드백을 통해 학습합니다:

```python
# 자동 학습 루프
def learn_from_feedback(content_id):
    metrics = get_engagement_metrics(content_id)
    brand_alignment = analyze_brand_fit(content_id)

    # 높은 engagement + 높은 brand_fit = 학습
    if metrics['authentic_engagement'] > 0.8 and brand_alignment > 0.75:
        add_to_positive_examples(content_id)
```

## Version History

- **v1.0** (2026-02-15): Initial skill creation
  - Consolidated from: IDENTITY.md, woohwahae_brand_source.md
  - Added validation criteria
  - Integrated with skill_loader system

---

> "언어는 브랜드의 얼굴이다. 신중하게, 진정성 있게." — 97layerOS
