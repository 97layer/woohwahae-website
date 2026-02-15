# Token Optimization Protocol (토큰 최적화 프로토콜)

## 목적

AI 에이전트가 개발 작업 시 토큰 소비를 최소화하면서도 높은 품질의 결과물을 생산하기 위한 시스템 차원의 운영 원칙.

---

## 1. 핵심 원칙

### A. Query Before Read (읽기 전 탐색)

**절대 원칙**: 파일 전체를 읽기 전에 반드시 구조 파악

```bash
# ❌ 나쁜 습관
Read(file_path="large_module.py")  # 5000 lines → 20,000+ tokens wasted

# ✅ 올바른 습관
Glob(pattern="**/*module*.py")     # Find candidates first
Grep(pattern="class TargetClass", output_mode="files_with_matches")
Grep(pattern="class TargetClass", output_mode="content", -n=True, -A=10)
Read(file_path="module.py", offset=145, limit=30)  # Only relevant section
```

**예상 절약**: 파일당 평균 15,000 토큰 절약

### B. Snippet-First Architecture (스니펫 우선 구조)

전체 파일 대신 관련 코드 조각만 추출하여 컨텍스트로 사용.

**구현 방법**:
```python
from execution.system.token_optimizer import TokenOptimizer

optimizer = TokenOptimizer()

# 1. 파일이 큰지 판단
if optimizer.should_use_snippet(content, threshold_tokens=1000):
    # 2. 키워드 기반 스니펫 추출
    snippet = optimizer.extract_relevant_snippets(
        content=file_content,
        keywords=["authentication", "jwt", "verify"],
        context_lines=3
    )
    # 3. 스니펫만 AI에게 전달
    result = ai_engine.query(snippet)
```

**예상 절약**: 컨텍스트당 평균 5,000-10,000 토큰 절약

### C. Cache-First Strategy (캐시 우선 전략)

동일하거나 유사한 요청은 반드시 캐시 확인.

**구현 순서**:
```python
optimizer = TokenOptimizer()

# 1. 캐시 확인
cached = optimizer.get_cached_response(prompt, max_age_hours=24)

if cached:
    return cached  # 즉시 반환, 토큰 0 소비

# 2. 캐시 미스면 AI 호출
response = ai_engine.query(prompt)

# 3. 응답 캐시
optimizer.cache_response(prompt, response, metadata={'context': 'dev_task'})
```

**예상 절약**: 캐시 히트율 30% 가정 시, 전체 토큰의 30% 절약

### D. Conversation Compression (대화 압축)

장기 대화는 주기적으로 압축하여 컨텍스트 윈도우 관리.

**구현**:
```python
# 대화 이력이 긴 경우
compressed_history = optimizer.compress_conversation_history(
    messages=conversation_messages,
    max_messages=10  # 최근 10개 + 요약
)

# 압축된 히스토리로 AI 호출
response = ai_engine.query(new_prompt, history=compressed_history)
```

**예상 절약**: 대화당 평균 5,000-20,000 토큰 절약

---

## 2. 개발 작업별 최적화 전략

### A. 코드 리뷰

```markdown
1. Glob으로 리뷰 대상 파일 목록 확인
2. Grep으로 변경된 함수/클래스명 추출
3. 변경된 부분 주변만 Read (offset + limit)
4. 리뷰 결과 캐시 (같은 커밋은 재리뷰 불필요)
```

**토큰 절약**: 80%+

### B. 버그 수정

```markdown
1. Grep으로 에러 메시지 관련 코드 위치 파악
2. 해당 파일의 관련 함수만 Read
3. 수정 후 diff만 확인 (전체 파일 재확인 불필요)
```

**토큰 절약**: 70%+

### C. 새 기능 개발

```markdown
1. 기존 유사 기능 검색 (Grep + Glob)
2. 관련 파일의 구조만 파악 (클래스/함수 시그니처)
3. 필요한 의존성만 확인
4. 구현 후 테스트 결과만 분석
```

**토큰 절약**: 60%+

### D. 리팩토링

```markdown
1. 리팩토링 대상 파일/함수 목록화
2. 의존성 그래프만 추출 (execution/system/dependency_analyzer.py 사용)
3. 변경 영향 범위만 Read
4. 리팩토링 결과를 diff로 확인
```

**토큰 절약**: 70%+

---

## 3. 실행 도구

### 사용 가능한 스크립트

| 스크립트 | 용도 | 토큰 절약 |
|---------|------|-----------|
| `execution/system/token_optimizer.py` | 캐싱, 스니펫 추출, 대화 압축 | 50-80% |
| `execution/system/thought_process.py` | 자율 사고 루프 (캐시 활용) | 30% |
| `libs/ai_engine.py` | AI 호출 시 자동 캐싱 | 20-40% |

### 통합 방법

기존 스크립트에 TokenOptimizer 통합:

```python
# 기존 코드
from libs.ai_engine import AIEngine
ai = AIEngine()
response = ai.query(large_prompt)

# 최적화된 코드
from execution.system.token_optimizer import TokenOptimizer
from libs.ai_engine import AIEngine

optimizer = TokenOptimizer()
ai = AIEngine()

# 캐시 확인
cached = optimizer.get_cached_response(large_prompt)
if cached:
    response = cached
else:
    # 스니펫 추출 후 쿼리
    if optimizer.should_use_snippet(large_prompt):
        snippet = optimizer.extract_relevant_snippets(
            large_prompt,
            keywords=["target", "keywords"]
        )
        response = ai.query(snippet)
    else:
        response = ai.query(large_prompt)

    # 캐시 저장
    optimizer.cache_response(large_prompt, response)
```

---

## 4. 모니터링 및 측정

### 최적화 리포트 확인

```bash
python execution/system/token_optimizer.py report
```

**출력 예시**:
```
==================================================
TOKEN OPTIMIZATION REPORT
==================================================
cache_hits          : 156
cache_writes        : 98
tokens_saved        : 245,830
cache_hit_rate      : 61.42%
total_requests      : 254
last_updated        : 2026-02-15T14:32:10
==================================================
```

### 캐시 관리

```bash
# 전체 캐시 삭제
python execution/system/token_optimizer.py clear

# 24시간 이상된 캐시만 삭제
python execution/system/token_optimizer.py clear 24
```

---

## 5. 에이전트 행동 규칙

### 개발 작업 시작 시

1. **탐색 단계**
   - Glob/Grep으로 파일 구조 파악
   - 캐시된 분석 결과 확인
   - 필요한 파일 목록만 추출

2. **분석 단계**
   - 큰 파일은 스니펫 추출
   - 관련 없는 코드는 무시
   - 구조/시그니처만 필요하면 헤더만 Read

3. **실행 단계**
   - 스크립트 우선 사용 (deterministic execution)
   - AI 쿼리 전 캐시 확인
   - 결과는 요약해서 저장

4. **완료 단계**
   - 최적화 통계 업데이트
   - 재사용 가능한 결과는 캐시
   - 불필요한 임시 파일 정리

### 금지 사항

❌ **절대 하지 말 것**:
- 파일 전체를 무작정 Read
- 동일한 파일을 여러 번 Read (캐시 활용)
- 대화 이력을 무제한 유지
- 디버그 출력을 AI 컨텍스트에 포함
- 같은 분석을 반복 수행

✅ **반드시 할 것**:
- Grep/Glob 먼저 사용
- 캐시 확인 먼저
- 스니펫 추출 우선
- 대화 이력 주기적 압축
- 최적화 통계 모니터링

---

## 6. 성공 지표

### 목표 KPI

| 지표 | 목표 | 측정 방법 |
|-----|------|----------|
| 캐시 히트율 | 40%+ | `token_optimizer.py report` |
| 평균 요청당 토큰 수 | 2,000 이하 | AI 로그 분석 |
| 파일 읽기당 평균 토큰 | 1,000 이하 | Read 호출 모니터링 |
| 일일 총 토큰 소비 | 기존 대비 50% 감소 | 통합 대시보드 |

### 주간 리뷰

매주 최적화 리포트 확인:
```bash
# 통계 확인
python execution/system/token_optimizer.py report

# 가장 많이 호출된 쿼리 분석
cat .tmp/token_cache/*.json | jq '.metadata'

# 캐시 효율 개선 방안 도출
```

---

## 7. 통합 예시

### 실제 개발 시나리오

**사용자 요청**: "인증 모듈에서 JWT 검증 로직을 찾아서 버그 수정해줘"

**최적화된 실행 흐름**:

```python
# 1단계: 파일 탐색 (토큰 거의 0)
files = glob("**/auth*.py")  # → 3 files found

# 2단계: 키워드 검색 (토큰 < 100)
matches = grep("jwt.*verify", output_mode="content", -n=True)
# → Found in: auth/jwt_handler.py:145-167

# 3단계: 관련 부분만 읽기 (토큰 ~500)
code = read("auth/jwt_handler.py", offset=140, limit=30)

# 4단계: 스니펫 추출 (토큰 ~300)
snippet = optimizer.extract_relevant_snippets(
    code,
    keywords=["verify", "jwt", "decode"],
    context_lines=3
)

# 5단계: 캐시 확인 후 AI 쿼리
prompt = f"Find bug in this JWT verification code:\n{snippet}"
cached = optimizer.get_cached_response(prompt)

if not cached:
    response = ai.query(prompt)  # 토큰 ~800
    optimizer.cache_response(prompt, response)
else:
    response = cached  # 토큰 0

# 총 소비: ~1,600 tokens (기존 대비 90% 절약)
# 기존 방식: auth 모듈 전체 읽기 → ~15,000 tokens
```

---

## 8. 자가 개선 (Self-Annealing)

### 학습 루프

1. **측정**: 매일 토큰 사용량 추적
2. **분석**: 비효율적인 패턴 발견
3. **개선**: 캐시 전략, 스니펫 추출 규칙 업데이트
4. **검증**: 개선 후 토큰 절약률 측정

### Directive 업데이트

새로운 최적화 기법 발견 시:
1. 이 directive 업데이트
2. 관련 스크립트 개선
3. 다른 에이전트에게 전파

---

## 요약

**3가지 핵심 습관**:
1. **Query Before Read**: Grep/Glob 먼저, Read는 최소한으로
2. **Cache Everything**: 동일 요청은 절대 재실행 금지
3. **Snippet-First**: 큰 파일은 관련 부분만 추출

**기대 효과**:
- 평균 60-80% 토큰 절약
- 응답 속도 2-3배 향상 (캐시 히트 시)
- 비용 절감 및 지속 가능한 AI 운영

---

**생성일**: 2026-02-15
**버전**: 1.0
**작성자**: 97LAYER System
**연관 스크립트**: `execution/system/token_optimizer.py`
