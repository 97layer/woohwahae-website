# Pattern Recognition Skill

## Skill ID
`pattern_recognition_v1`

## Purpose
원시 신호(raw signals)에서 의미 있는 패턴을 포착하고, 개인적 경험을 보편적 인사이트로 전환하는 능력을 체계화한다.

## Core Philosophy
> "모든 것이 신호다. 패턴은 숨어있지 않고, 연결되기를 기다린다."

## Architecture

### 5-Stage Signal Processing
```
Raw Signals → Pattern Detection → Connection → Insight → Content
    (SA)          (SA)              (SA)        (CE)      (CE+AD)
```

### Signal Types
```yaml
Input Sources:
  - telegram: 개인 대화, 생각의 파편
  - youtube: 영상 분석, 댓글 패턴
  - web: 트렌드, 철학적 글
  - clipboard: 순간의 포착
  - drive: 문서, 이미지

Signal Categories:
  - philosophical: 철학적 질문
  - personal: 개인적 경험
  - observational: 일상의 관찰
  - cultural: 문화 트렌드
  - technical: 기술 인사이트
```

## Rules

### 1. Capture Everything (무조건 저장)

#### No Filtering at Capture
```python
def capture_signal(raw_input):
    """필터링 없이 무조건 저장"""

    # ❌ BAD: Filtering at capture
    if is_important(raw_input):  # NO!
        save(raw_input)

    # ✅ GOOD: Save everything
    signal_id = generate_id()
    save_raw_signal(signal_id, raw_input)
    return signal_id
```

#### Signal ID Format
```python
# rs-{timestamp}_{source}.md
examples = [
    "rs-1771117789_telegram.md",
    "rs-1771117790_youtube_dQw4w9WgXcQ.md",
    "rs-1771117791_clipboard.md",
]
```

#### Storage Location
```
knowledge/raw_signals/
├── rs-001_telegram.md
├── rs-002_youtube.md
├── rs-003_clipboard.md
└── ...
```

### 2. Pattern Detection (패턴 발견)

#### Pattern Types

##### Temporal Patterns (시간 패턴)
```python
def detect_temporal_pattern(signals):
    """시간 기반 패턴"""

    patterns = []

    # 반복 (Repetition)
    if appears_multiple_times(theme, window="7d"):
        patterns.append({
            "type": "repetition",
            "theme": theme,
            "frequency": count,
            "insight": "이 주제가 계속 떠오른다"
        })

    # 주기 (Cycle)
    if has_regular_interval(theme):
        patterns.append({
            "type": "cycle",
            "theme": theme,
            "interval": "weekly",
            "insight": "주기적으로 고민하는 주제"
        })

    return patterns
```

##### Semantic Patterns (의미 패턴)
```python
def detect_semantic_pattern(signals):
    """의미 기반 패턴"""

    # 키워드 추출
    keywords = extract_keywords(signals)

    # 클러스터링
    clusters = cluster_by_meaning(keywords)

    # 주제 추출
    themes = []
    for cluster in clusters:
        if len(cluster) >= 3:  # 최소 3개 신호
            theme = extract_common_theme(cluster)
            themes.append({
                "theme": theme,
                "signals": cluster,
                "strength": len(cluster),
                "insight": synthesize_insight(cluster)
            })

    return themes
```

##### Emergent Patterns (떠오르는 패턴)
```python
def detect_emergent_pattern(signals, window="30d"):
    """최근 떠오르는 새로운 패턴"""

    recent = filter_by_date(signals, window)
    historical = filter_by_date(signals, f"-{window}")

    recent_themes = extract_themes(recent)
    historical_themes = extract_themes(historical)

    # 새롭게 등장한 주제
    emergent = []
    for theme in recent_themes:
        if theme not in historical_themes:
            emergent.append({
                "theme": theme,
                "first_seen": theme.earliest_signal,
                "growth_rate": theme.frequency,
                "insight": "새롭게 관심이 생긴 주제"
            })

    return emergent
```

### 3. Connection Building (연결 구축)

#### Connection Types
```yaml
Direct:
  Example: "미니멀리즘" → "덜어냄"
  Method: Keyword matching

Associative:
  Example: "8평 반지하" → "고독" → "본질 탐구"
  Method: Semantic similarity

Metaphorical:
  Example: "여백 60%" → "삶의 여유"
  Method: Conceptual mapping

Temporal:
  Example: "코로나" → "고독" → "철학"
  Method: Time-based causality
```

#### Connection Graph
```python
class ConnectionGraph:
    """신호 간 연결 그래프"""

    def __init__(self):
        self.nodes = {}  # signal_id → signal
        self.edges = []  # (signal_a, signal_b, weight, type)

    def add_connection(self, signal_a, signal_b, connection_type):
        """두 신호 연결"""

        weight = calculate_connection_strength(
            signal_a,
            signal_b,
            connection_type
        )

        if weight > 0.5:  # Threshold
            self.edges.append({
                "from": signal_a.id,
                "to": signal_b.id,
                "weight": weight,
                "type": connection_type
            })

    def find_clusters(self):
        """연결된 신호 클러스터 찾기"""
        return community_detection(self.edges)

    def shortest_path(self, start, end):
        """두 신호 사이 최단 경로"""
        return dijkstra(self.edges, start, end)
```

#### Connection Strength
```python
def calculate_connection_strength(signal_a, signal_b, type):
    """연결 강도 계산 (0-1)"""

    if type == "direct":
        # 키워드 중복도
        overlap = keyword_overlap(signal_a, signal_b)
        return overlap

    elif type == "semantic":
        # 의미적 유사도 (embeddings)
        similarity = cosine_similarity(
            embed(signal_a.text),
            embed(signal_b.text)
        )
        return similarity

    elif type == "temporal":
        # 시간적 근접성
        time_diff = abs(signal_a.timestamp - signal_b.timestamp)
        proximity = exp(-time_diff / (7 * 86400))  # 7일 반감기
        return proximity

    else:
        return 0.0
```

### 4. Insight Generation (인사이트 생성)

#### Personal to Universal
```python
def transform_to_universal(personal_pattern):
    """개인적 패턴 → 보편적 인사이트"""

    # 1. 개인적 경험 추출
    personal = extract_personal_story(personal_pattern)

    # 2. 핵심 감정/경험 식별
    core_experience = identify_core_emotion(personal)

    # 3. 보편화
    universal = {
        "personal": personal,
        "core": core_experience,
        "universal": generalize(core_experience),
        "question": generate_reflective_question(core_experience)
    }

    return universal

# Example
personal_pattern = {
    "signals": ["8평 반지하", "혼자", "일기 28권"],
    "theme": "고독 속 자아 발견"
}

insight = transform_to_universal(personal_pattern)
# {
#   "personal": "8평 반지하에서 혼자 일기를 쓰며...",
#   "core": "고독 속에서 진정한 자아를 발견함",
#   "universal": "많은 사람들이 혼자만의 시간을 통해 자신을 발견한다",
#   "question": "당신은 언제 가장 '나'다운가?"
# }
```

#### Insight Validation
```python
def validate_insight(insight):
    """인사이트 품질 검증"""

    criteria = {
        "depth": has_philosophical_depth(insight),
        "universality": is_universally_relatable(insight),
        "authenticity": feels_authentic(insight),
        "actionability": provides_reflection_point(insight),
    }

    # 모두 True여야 통과
    return all(criteria.values()), criteria
```

### 5. Pattern Persistence (패턴 저장)

#### Pattern Document Format
```markdown
---
id: pattern_20260215_minimal_life
created: 2026-02-15
theme: 미니멀리즘과 본질 탐구
strength: 0.85
signals: [rs-001, rs-015, rs-042, rs-078]
---

# Pattern: 미니멀리즘과 본질 탐구

## Signals (4개)
- rs-001: "8평 반지하의 삶"
- rs-015: "물건을 버리니 마음도 가벼워짐"
- rs-042: "여백의 미학"
- rs-078: "덜어냄의 용기"

## Connections
- 공간의 물리적 제약 → 정신적 자유
- 소유의 감소 → 본질로의 회귀
- 외부 소음 제거 → 내면의 목소리 발견

## Insight
개인적: 좁은 공간에서 오히려 자유를 느꼈다
보편적: 제약이 본질을 발견하게 한다

## Potential Content
Title: "8평 반지하의 낭만"
Hook: "좁은 공간이 넓은 자유를 가르쳐주었다"
Core: 물리적 여유보다 정신적 여유
```

## Validation Criteria

### Pattern Quality Checks
- [ ] **Signal Count**: 최소 3개 이상의 신호
- [ ] **Connection Strength**: 평균 연결 강도 > 0.5
- [ ] **Temporal Span**: 신호들이 시간적으로 분산됨
- [ ] **Depth**: 표면적이지 않은 깊이 있는 주제
- [ ] **Universality**: 개인을 넘어 보편적 공감 가능

### Insight Quality Checks
- [ ] **Personal Story**: 구체적 개인 경험 포함
- [ ] **Core Emotion**: 명확한 핵심 감정/경험
- [ ] **Universal Bridge**: 보편적 전환 성공
- [ ] **Reflective Question**: 생각을 유도하는 질문
- [ ] **Brand Alignment**: WOOHWAHAE 철학과 일치

## Examples

### ❌ BAD: Weak Pattern
```yaml
Theme: "좋은 날씨"
Signals: 2개 (rs-010, rs-011)
Connection: "둘 다 날씨 언급"
Insight: "날씨가 좋으면 기분이 좋다"
```

**문제점**:
- 신호 수 부족
- 피상적 연결
- 뻔한 인사이트
- 보편화 실패

### ✅ GOOD: Strong Pattern
```yaml
Theme: "고독과 자아 발견"
Signals: 7개 (rs-001, rs-015, rs-023, rs-042, rs-056, rs-071, rs-089)
Connections:
  - Temporal: 3개월간 반복적 출현
  - Semantic: "혼자", "고독", "나", "본질" 키워드 클러스터
  - Emergent: 코로나 이후 강화된 주제

Insight:
  Personal: |
    8평 반지하 원룸에서 혼자 보낸 시간이
    나를 가장 '나'답게 만들었다.

  Universal: |
    많은 이들이 혼자만의 시간을 통해
    자신의 본질을 발견한다.
    고독은 두려움이 아닌 선물이다.

  Question: |
    당신은 언제 가장 '나'다운가?
    마지막으로 혼자만의 시간을 가진 게 언제인가?

Potential: Magazine "고독한 본질 탐구자"
```

## Integration Points

### For SA (Strategy Analyst)
```python
from libs.skill_loader import SkillLoader

pattern_skill = SkillLoader.load("pattern_recognition_v1")

# 신호 수집
signals = collect_recent_signals(days=30)

# 패턴 탐지
patterns = pattern_skill.detect_patterns(signals)

# 인사이트 생성
for pattern in patterns:
    if pattern['strength'] > 0.7:
        insight = pattern_skill.generate_insight(pattern)
        save_insight(insight)
```

### For CE (Chief Editor)
```python
# 콘텐츠 생성을 위한 인사이트 조회
insights = pattern_skill.get_mature_insights(min_strength=0.8)

for insight in insights:
    # 콘텐츠 초안 생성
    draft = create_content_from_insight(insight)
```

### For Quality Gate
```python
def pattern_quality_gate(pattern):
    """패턴 품질 검증"""

    pr_skill = SkillLoader.load("pattern_recognition_v1")

    validation = pr_skill.validate_pattern(pattern)

    if not validation['passed']:
        log_rejection(pattern, validation['issues'])
        return False

    return True
```

## Tools & Scripts

### Pattern Detector
```bash
# 최근 신호에서 패턴 탐지
python libs/skills/pattern_recognition/detector.py --days 30

# Output:
# Found 5 patterns:
# 1. "고독과 자아 발견" (strength: 0.85, signals: 7)
# 2. "미니멀리즘" (strength: 0.78, signals: 5)
# 3. "시간의 기록" (strength: 0.72, signals: 6)
# ...
```

### Connection Visualizer
```bash
# 패턴 연결 그래프 시각화
python libs/skills/pattern_recognition/visualizer.py pattern_123

# Generates: pattern_123_graph.html (interactive visualization)
```

### Insight Generator
```bash
# 패턴에서 인사이트 생성
python libs/skills/pattern_recognition/insight_gen.py pattern_123

# Output saved to: knowledge/patterns/insight_pattern_123.md
```

## Advanced Techniques

### Ontology Transform
```python
def ontology_transform(raw_signal):
    """원시 신호를 존재론적 구조로 변환"""

    # 1. Entity extraction
    entities = extract_entities(raw_signal)

    # 2. Relationship mapping
    relationships = map_relationships(entities)

    # 3. Concept hierarchy
    hierarchy = build_concept_tree(entities, relationships)

    # 4. Philosophical dimensions
    dimensions = {
        "temporal": extract_time_dimension(raw_signal),
        "spatial": extract_space_dimension(raw_signal),
        "emotional": extract_emotion_dimension(raw_signal),
        "existential": extract_existential_themes(raw_signal),
    }

    return {
        "entities": entities,
        "relationships": relationships,
        "hierarchy": hierarchy,
        "dimensions": dimensions,
    }
```

### Recursive Pattern Discovery
```python
def recursive_pattern_discovery(patterns, depth=0, max_depth=3):
    """패턴의 패턴 발견 (재귀적)"""

    if depth >= max_depth:
        return patterns

    # 현재 레벨 패턴들 간의 패턴 찾기
    meta_patterns = detect_patterns_in_patterns(patterns)

    if len(meta_patterns) > 0:
        # 더 깊이 들어감
        return recursive_pattern_discovery(
            meta_patterns,
            depth + 1,
            max_depth
        )
    else:
        return patterns
```

## Common Pitfalls

### 1. "확증 편향 함정"
❌ 원하는 패턴만 찾으려 함
✅ 데이터가 말하게 둠

### 2. "과잉 해석 함정"
❌ 우연을 패턴으로 착각
✅ 통계적 유의성 확인

### 3. "조급함 함정"
❌ 신호 2-3개로 패턴 선언
✅ 충분한 신호 누적 (최소 5개+)

### 4. "복잡성 함정"
❌ 지나치게 복잡한 연결
✅ 단순하고 명확한 패턴

## Learning & Evolution

```python
class PatternLearning:
    """패턴 인식 스킬의 자기 학습"""

    def learn_from_feedback(self, pattern_id, outcome):
        """결과 피드백으로 학습"""

        pattern = load_pattern(pattern_id)

        if outcome == "published_successful":
            # 성공한 패턴의 특성 학습
            features = extract_features(pattern)
            add_to_positive_examples(features)

        elif outcome == "rejected":
            # 실패한 패턴 분석
            issues = analyze_failure(pattern)
            update_detection_rules(issues)

    def adjust_thresholds(self):
        """임계값 자동 조정"""

        performance = analyze_historical_performance()

        if performance['false_positive_rate'] > 0.3:
            # 너무 많은 오탐 → 임계값 높임
            self.connection_threshold += 0.05

        if performance['false_negative_rate'] > 0.2:
            # 놓치는 패턴 많음 → 임계값 낮춤
            self.connection_threshold -= 0.05
```

## Version History

- **v1.0** (2026-02-15): Initial skill creation
  - 5-stage signal processing pipeline
  - Multiple pattern detection algorithms
  - Personal-to-universal transformation
  - Connection graph building
  - Recursive pattern discovery

---

> "패턴은 언제나 있었다. 우리가 보지 못했을 뿐이다." — 97layerOS
