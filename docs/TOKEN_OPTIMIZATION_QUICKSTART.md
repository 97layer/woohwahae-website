# Token Optimization Quick Start Guide

> **목표**: AI 에이전트가 개발 작업 시 토큰 소비를 60-80% 줄이는 방법

---

## 📋 TL;DR (Too Long; Didn't Read)

```bash
# 1. 주간 리포트 확인
python execution/system/weekly_optimization_monitor.py

# 2. 자가 개선 실행 (매주)
python execution/system/self_annealing_optimizer.py run

# 3. 캐시 정리 (선택)
python execution/system/token_optimizer.py clear 72
```

**핵심 4가지 규칙:**
1. 파일 전체 읽기 금지 → Grep 먼저
2. 캐시 먼저 확인
3. 큰 파일은 스니펫만 추출
4. 의존성 분석으로 필요한 파일만 읽기

---

## 🚀 즉시 사용 가능한 도구

### 1. TokenOptimizer (캐싱 & 스니펫)

```python
from execution.system.token_optimizer import TokenOptimizer

optimizer = TokenOptimizer()

# 캐시 확인
cached = optimizer.get_cached_response(prompt)
if cached:
    return cached  # 토큰 0 소비

# 스니펫 추출
if optimizer.should_use_snippet(file_content):
    snippet = optimizer.extract_relevant_snippets(
        file_content,
        keywords=["function_name", "class_name"],
        context_lines=3
    )
    # snippet만 AI에게 전달 → 60-80% 토큰 절약
```

### 2. DependencyAnalyzer (구조 파악)

```python
from execution.system.dependency_analyzer import DependencyAnalyzer

analyzer = DependencyAnalyzer()

# 파일 전체 읽지 않고 구조만 파악
summary = analyzer.get_file_summary("path/to/file.py")
# 출력: 클래스명, 함수 시그니처, import 목록만

# 변경 영향 범위 분석
affected = analyzer.find_affected_files("target_file.py")
# 영향받는 파일만 읽으면 됨
```

### 3. 주간 모니터링

```bash
# 최적화 현황 확인
python execution/system/weekly_optimization_monitor.py

# 출력:
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📊 WEEKLY TOKEN OPTIMIZATION REPORT
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Overall Grade: A (Very Good)
# Cache Hit Rate: 61.42%
# Tokens Saved: 245,830
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 4. 자가 개선 시스템

```bash
# 비효율 패턴 자동 감지 및 학습
python execution/system/self_annealing_optimizer.py run

# 출력:
# 🔄 SELF-ANNEALING OPTIMIZATION CYCLE
# Large prompts found:     3
# Repeated queries:        12
# Learnings generated:     2
# Directive updated:       Yes
```

---

## 🎯 실전 예시

### 예시 1: 버그 수정

**❌ 기존 방식 (비효율)**
```python
# 파일 전체 읽기 (15,000 tokens)
code = read_file("auth/handler.py")
result = ai.query(f"Find bug in:\n{code}")
```

**✅ 최적화 방식 (1,500 tokens)**
```python
from execution.system.token_optimizer import TokenOptimizer

optimizer = TokenOptimizer()

# 1. 버그 위치 찾기
grep("error.*authentication", "auth/*.py")  # 100 tokens
# → auth/handler.py:145 발견

# 2. 관련 부분만 읽기
code = read_file("auth/handler.py", offset=140, limit=20)  # 500 tokens

# 3. 스니펫 추출
snippet = optimizer.extract_relevant_snippets(
    code,
    keywords=["authentication", "verify", "error"],
    context_lines=3
)  # 300 tokens

# 4. 캐시 확인 후 AI 쿼리
prompt = f"Fix bug in:\n{snippet}"
cached = optimizer.get_cached_response(prompt)
if not cached:
    result = ai.query(prompt)  # 600 tokens
    optimizer.cache_response(prompt, result)
else:
    result = cached  # 0 tokens

# 총: 1,500 tokens (90% 절약)
```

### 예시 2: 리팩토링

**❌ 기존 방식 (비효율)**
```python
# 모든 관련 파일 읽기 (50,000 tokens)
for file in all_python_files:
    code = read_file(file)
    dependencies = ai.query(f"Find dependencies in {code}")
```

**✅ 최적화 방식 (5,000 tokens)**
```python
from execution.system.dependency_analyzer import DependencyAnalyzer

analyzer = DependencyAnalyzer()

# 1. 의존성 그래프 캐시에서 로드
graph = analyzer.load_cached_graph()
if not graph:
    # 처음만 생성 (전체 프로젝트 한번만 분석)
    graph = analyzer.build_dependency_graph()
    analyzer.cache_graph(graph)

# 2. 리팩토링 대상 파일 영향 범위만 추출
affected = analyzer.find_affected_files("target_module.py")
# → 3개 파일만 영향받음

# 3. 영향받는 파일의 요약만 읽기
for file in affected:
    summary = analyzer.get_file_summary(file)
    # AI에게 전체 코드 대신 요약만 전달

# 총: 5,000 tokens (90% 절약)
```

---

## 📊 모니터링 대시보드

### 매주 월요일 체크리스트

```bash
# 1. 주간 리포트
python execution/system/weekly_optimization_monitor.py

# 확인 사항:
# - Cache Hit Rate: 40% 이상 유지?
# - Tokens Saved: 지난주 대비 증가?
# - Overall Grade: B+ 이상?

# 2. 자가 개선 실행
python execution/system/self_annealing_optimizer.py run

# 3. 권장사항 확인 및 적용
# 리포트에 나온 HIGH priority 항목 먼저 처리

# 4. 오래된 캐시 정리
python execution/system/token_optimizer.py clear 72  # 72시간 이상
```

---

## 🔧 기존 코드에 통합하기

### AI Engine 통합 (이미 완료)

[ai_engine.py](../libs/ai_engine.py)에 TokenOptimizer가 자동으로 통합되었습니다:

```python
# libs/ai_engine.py
from execution.system.token_optimizer import TokenOptimizer

class AIEngine:
    def __init__(self):
        self.optimizer = TokenOptimizer()

    def generate_thought(self, prompt):
        # 자동으로 캐시 확인
        cached = self.optimizer.get_cached_response(prompt)
        if cached:
            return cached

        # AI 호출
        response = self._call_api(prompt)

        # 자동으로 캐싱
        self.optimizer.cache_response(prompt, response)
        return response
```

**설정 불필요 → 즉시 사용 가능**

### 새 스크립트 작성 시

```python
#!/usr/bin/env python3
"""
Your New Script
"""
from execution.system.token_optimizer import TokenOptimizer
from execution.system.dependency_analyzer import DependencyAnalyzer

def your_function():
    optimizer = TokenOptimizer()
    analyzer = DependencyAnalyzer()

    # 1. 의존성 확인
    affected = analyzer.find_affected_files("target.py")

    # 2. 요약만 읽기
    for file in affected:
        summary = analyzer.get_file_summary(file)

        # 3. 캐시 확인
        cached = optimizer.get_cached_response(f"analyze:{file}")
        if not cached:
            # AI 처리
            result = ai.query(summary)
            optimizer.cache_response(f"analyze:{file}", result)
        else:
            result = cached
```

---

## 📈 성공 지표

### 목표 KPI

| 지표 | 목표 | 현재 상태 확인 |
|-----|------|--------------|
| 캐시 히트율 | 40%+ | `weekly_optimization_monitor.py` |
| 평균 요청당 토큰 | 2,000 이하 | 리포트의 `tokens_saved` 참고 |
| 주간 절약 토큰 | 50,000+ | 리포트 참고 |
| Overall Grade | B+ 이상 | 리포트 참고 |

### 등급 기준

- **A+ (90+)**: Excellent - 최적화 완벽
- **A (80-89)**: Very Good - 잘 유지 중
- **B+ (70-79)**: Good - 개선 여지 있음
- **B (60-69)**: Fair - 최적화 필요
- **C (50-59)**: Needs Improvement - 즉시 개선 필요
- **D (<50)**: Poor - 시스템 점검 필요

---

## 🛠️ 트러블슈팅

### 문제 1: 캐시 히트율이 30% 미만

**원인**: 쿼리 패턴이 너무 다양하거나 프롬프트가 매번 달라짐

**해결**:
```python
# 프롬프트 표준화
# ❌ BAD: 매번 다른 프롬프트
f"Find bugs in this code: {code} at {datetime.now()}"

# ✅ GOOD: 표준화된 프롬프트
f"Find bugs in:\n{code_snippet}"
```

### 문제 2: 토큰 절약이 적음

**원인**: 여전히 큰 파일을 전체 읽고 있음

**해결**:
```bash
# 자가 개선 실행
python execution/system/self_annealing_optimizer.py run

# Large prompts 확인
python execution/system/self_annealing_optimizer.py analyze

# 큰 프롬프트 발견 시 스니펫 추출으로 변경
```

### 문제 3: Overall Grade가 C 이하

**원인**: 최적화 도구를 사용하지 않음

**해결**:
1. `directives/token_optimization_protocol.md` 다시 읽기
2. 4가지 핵심 규칙 준수 확인
3. 코드에 TokenOptimizer 통합
4. 1주일 후 재측정

---

## 📚 더 알아보기

- **상세 가이드**: [directives/token_optimization_protocol.md](../directives/token_optimization_protocol.md)
- **시스템 규칙**: [CLAUDE.md](../CLAUDE.md#token-optimization-critical)
- **스크립트 소스**:
  - [token_optimizer.py](../execution/system/token_optimizer.py)
  - [dependency_analyzer.py](../execution/system/dependency_analyzer.py)
  - [weekly_optimization_monitor.py](../execution/system/weekly_optimization_monitor.py)
  - [self_annealing_optimizer.py](../execution/system/self_annealing_optimizer.py)

---

## 🎓 5분 튜토리얼

### Step 1: 현재 상태 확인
```bash
python execution/system/token_optimizer.py report
```

### Step 2: 테스트 실행
```bash
python execution/system/token_optimizer.py test
```

### Step 3: 첫 주간 리포트
```bash
python execution/system/weekly_optimization_monitor.py
```

### Step 4: 코드에 적용
```python
from execution.system.token_optimizer import TokenOptimizer
optimizer = TokenOptimizer()

# 기존 코드 전:
# result = ai.query(large_prompt)

# 기존 코드 후:
snippet = optimizer.extract_relevant_snippets(large_prompt, keywords=["target"])
cached = optimizer.get_cached_response(snippet)
result = cached if cached else ai.query(snippet)
```

### Step 5: 1주 후 재확인
```bash
python execution/system/weekly_optimization_monitor.py
# 토큰 절약량 확인!
```

---

**시작하세요!** 5분만 투자하면 앞으로 수개월간 60-80% 토큰을 절약할 수 있습니다.

**질문/이슈**: 이 프로젝트는 자가 개선 시스템입니다. 문제 발견 시:
1. `execution/system/self_annealing_optimizer.py run` 실행
2. 자동으로 학습하고 directive 업데이트
3. 시스템이 더 강해집니다

---

**작성**: 2026-02-15
**버전**: 1.0
**작성자**: 97LAYER System
