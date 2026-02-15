# 97layerOS Skills System

## Overview

Skills are **reusable capability modules** that consolidate scattered directives into focused, actionable units. Each skill defines clear rules, validation criteria, and examples for specific domains.

## Philosophy

> "파편화는 혼돈이다. 스킬은 통합이다."

Instead of having dozens of scattered directive files, we consolidate domain expertise into **5 core skills** that are:

- **Loadable**: Dynamic injection into agent workflows
- **Validatable**: Automated quality checks
- **Versioned**: Clear evolution tracking
- **Actionable**: Specific, implementable rules

## Core Skills

### 1. brand_voice.skill.md
**Purpose**: WOOHWAHAE 브랜드 톤앤매너 구현

**Key Rules**:
- Aesop Benchmark 70%+
- The Hook Formula (정의/역설/질문)
- 절제된 언어, 금지어 사용 금지
- 여운 남기는 구조

**Validation**: Aesop score, banned word check, structure analysis

---

### 2. design_guide.skill.md
**Purpose**: 미니멀 디자인 철학 실현

**Key Rules**:
- 여백 60%+ (Whitespace ratio)
- 모노크롬 우선 (채도 < 20%)
- 자연광 사용
- 단일 포커스

**Validation**: Whitespace analyzer, saturation checker, layout validator

---

### 3. instagram.skill.md
**Purpose**: Anti-Algorithm Instagram 전략

**Key Rules**:
- 72시간 규칙 (Imperfect Publishing)
- 해시태그 0-3개
- MBQ 검증 (Meaning + Brand + Quality)
- 진정성 지표 우선

**Validation**: 72h check, MBQ validator, authenticity score

---

### 4. infra_ops.skill.md
**Purpose**: 하이브리드 인프라 운영 전문지식

**Key Rules**:
- Podman 우선 (rootless containers)
- Zero-cost cloud optimization
- Health monitoring + auto-healing
- Backup 전략

**Validation**: Service health checks, resource monitoring, security audit

---

### 5. pattern_recognition.skill.md
**Purpose**: 신호 포착과 패턴 인식

**Key Rules**:
- 무조건 저장 (No filtering at capture)
- 5-stage processing pipeline
- 개인 → 보편 전환
- Connection graph building

**Validation**: Pattern strength, signal count, insight depth

---

## Usage

### CLI Interface

```bash
# List all skills
python libs/skill_loader.py list

# Load a skill
python libs/skill_loader.py load brand_voice_v1

# Validate content
python libs/skill_loader.py validate brand_voice_v1 content.md

# Inject skills to prompt
python libs/skill_loader.py inject brand_voice_v1 design_guide_v1
```

### Python API

```python
from libs.skill_loader import SkillLoader, SkillValidator

# Load skill
skill = SkillLoader.load("brand_voice_v1")

print(f"Skill: {skill.name}")
print(f"Purpose: {skill.purpose}")
print(f"Rules: {len(skill.rules)}")

# Validate content
validator = SkillValidator(skill)
result = validator.validate(content_text)

if result['passed']:
    print("✓ Content passes skill validation")
else:
    print("✗ Issues found:")
    for issue in result['issues']:
        print(f"  - {issue}")

    print("\nSuggestions:")
    for suggestion in result['suggestions']:
        print(f"  - {suggestion}")
```

### Integration with Agents

```python
# In agent workflow
from libs.skill_loader import inject_skills_to_prompt

# CE (Chief Editor) with brand voice skill
skills = ["brand_voice_v1", "instagram_v1"]
enhanced_prompt = inject_skills_to_prompt(skills)

response = gemini.generate(
    prompt=f"{enhanced_prompt}\n\n{task_description}"
)
```

### Quality Gate Integration

```python
# In execution/system/quality_gate.py
from libs.skill_loader import validate_content

def content_quality_gate(content_path):
    """Validate content against all relevant skills"""

    with open(content_path) as f:
        content = f.read()

    # Check brand voice
    brand_result = validate_content("brand_voice_v1", content)

    # Check Instagram rules if applicable
    if is_instagram_content(content_path):
        ig_result = validate_content("instagram_v1", content)
    else:
        ig_result = {"passed": True}

    return brand_result['passed'] and ig_result['passed']
```

## Skill Format Specification

Every skill MUST follow this format:

```markdown
# [Skill Name]

## Skill ID
`skill_id_v1`

## Purpose
[Clear purpose statement in Korean]

## Core Philosophy
> "Quote that captures essence"

## Rules

### 1. [Rule Category]
[Detailed explanation]

#### [Subsection]
```python
# Code example if applicable
```

### 2. [Next Rule Category]
...

## Validation Criteria

### Automated Checks
- [ ] **Check Name**: Description (threshold)
- [ ] **Check Name**: Description

### Manual Review
- [ ] **Review Point**: What to check
- [ ] **Review Point**: What to check

## Examples

### ❌ BAD: [Example Title]
```
[Bad example content]
```

**문제점**:
- Issue 1
- Issue 2

### ✅ GOOD: [Example Title]
```
[Good example content]
```

**우수한 이유**:
- Reason 1
- Reason 2

## Integration Points

### For [Agent Role]
```python
# How this agent uses the skill
```

## Tools & Scripts

### [Tool Name]
```bash
# Command examples
```

## Common Pitfalls

### 1. "[Pitfall Name]"
❌ What not to do
✅ What to do instead

## Version History

- **v1.0** (YYYY-MM-DD): Initial creation
  - Feature 1
  - Feature 2

---

> "Quote about the skill" — 97layerOS
```

## Directory Structure

```
core/skills/
├── README.md                     # This file
├── brand_voice.skill.md          # 브랜드 보이스
├── design_guide.skill.md         # 디자인 가이드
├── instagram.skill.md            # Instagram 전략
├── infra_ops.skill.md            # 인프라 운영
└── pattern_recognition.skill.md  # 패턴 인식

libs/
└── skill_loader.py               # Skill loading system
```

## Skill Lifecycle

### 1. Creation
```bash
# Create new skill from template
cp core/skills/_TEMPLATE.skill.md core/skills/new_skill.skill.md
# Edit and commit
git add core/skills/new_skill.skill.md
git commit -m "skill: Add new_skill v1.0"
```

### 2. Evolution
Skills evolve through version updates (NOT new files):

```markdown
## Version History

- **v1.2** (2026-03-01): Enhanced validation
  - Added automated tone analysis
  - Improved example gallery

- **v1.1** (2026-02-20): Minor improvements
  - Clarified rule #3
  - Added integration example

- **v1.0** (2026-02-15): Initial creation
```

### 3. Integration
Skills are automatically discovered and loaded by `skill_loader.py`. No registration needed.

### 4. Validation
```bash
# Validate skill format
python libs/skill_loader.py load new_skill_v1

# Expected output:
# Skill: New Skill
# ID: new_skill_v1
# Purpose: ...
# Rules: 10
# Version: v1.0
```

## Best Practices

### DO ✅
- **Single Source of Truth**: One skill file per domain
- **Version Up**: Update existing file, commit with version bump
- **Clear Rules**: Specific, actionable, measurable
- **Examples**: Show both good and bad
- **Validation**: Automate where possible

### DON'T ❌
- **Fragmentation**: No `brand_voice_v1.md`, `brand_voice_v2.md`, `brand_voice_final.md`
- **Vagueness**: No "be creative" or "make it good"
- **Theory Only**: Include practical examples and code
- **Static**: Skills should evolve with system

## Future Skills (Planned)

```yaml
Phase 2 (Q2 2026):
  - video_content.skill.md      # YouTube 콘텐츠
  - community_mgmt.skill.md     # 커뮤니티 관리
  - analytics.skill.md          # 데이터 분석

Phase 3 (Q3 2026):
  - collaboration.skill.md      # 협업 프로토콜
  - crisis_mgmt.skill.md        # 위기 대응
  - scaling.skill.md            # 스케일링 전략
```

## Metrics

Track skill effectiveness:

```python
# In execution/progress_analyzer.py
def skill_effectiveness_report():
    """Measure skill impact"""

    metrics = {
        "brand_voice_v1": {
            "usage_count": 145,
            "pass_rate": 0.82,
            "avg_aesop_score": 0.76,
            "impact": "High - consistent brand voice"
        },
        "instagram_v1": {
            "usage_count": 89,
            "pass_rate": 0.91,
            "72h_compliance": 0.95,
            "impact": "High - improved publishing cadence"
        }
    }

    return metrics
```

## Contributing

### Adding a New Skill

1. **Identify Need**: What domain expertise is scattered?
2. **Research**: Gather existing directives on topic
3. **Consolidate**: Extract core rules and principles
4. **Formalize**: Write in skill format
5. **Validate**: Test with real content
6. **Integrate**: Add to agent workflows
7. **Iterate**: Improve based on usage

### Improving Existing Skill

1. **Gather Feedback**: What's working? What's not?
2. **Analyze**: Review validation failures
3. **Update**: Enhance rules, examples, validation
4. **Version Bump**: Update version history
5. **Document**: Note what changed and why
6. **Test**: Validate against historical content

---

## Philosophy

> "Skills는 도구가 아니라 지혜의 결정체다.
> 파편화된 지식을 통합하고, 재사용 가능하게 만들고, 검증 가능하게 한다.
> 매 사용마다 더 나아지는, 살아있는 시스템이다."

— 97layerOS Skill System Manifesto

---

**Version**: 1.0
**Created**: 2026-02-15
**Status**: Active
