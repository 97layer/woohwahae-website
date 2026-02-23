---
description: INTELLIGENCE_QUANTA.md 빠른 갱신
---

# /quanta — QUANTA 갱신

현재 상태를 INTELLIGENCE_QUANTA.md에 기록한다.

## 실행 순서

1. **현재 QUANTA 읽기**
```bash
cat knowledge/agent_hub/INTELLIGENCE_QUANTA.md
```

2. **📍 현재 상태 (CURRENT STATE) 섹션 갱신**

아래 형식으로 갱신:
```
### [날짜 시간] Session Update - {agent-id}

**완료한 작업**:
- ✅ {작업1}
- ✅ {작업2}

**다음 단계**:
- ⏳ {태스크1}
- ⏳ {태스크2}

**업데이트 시간**: {ISO 8601}
```

3. **✅ 완료된 작업 섹션에 추가** (해당 시)

4. **🎯 다음 작업 섹션 업데이트** (해당 시)

## 주의
- QUANTA는 덮어쓰기 대상 (최신 상태만 유지)
- 갱신 시 이전 CURRENT STATE 블록 교체
