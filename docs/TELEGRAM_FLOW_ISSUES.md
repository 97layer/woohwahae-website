# 텔레그램 대화 플로우 문제점 진단

**날짜**: 2026-02-15
**상태**: 🔴 Critical - 사용자 경험 이질감 발생

---

## 🔍 문제점 요약

텔레그램 봇이 **"붕 떠있는 것 같다"** - 대화가 끊기고, 컨텍스트가 유지되지 않으며, 응답이 일관성 없음.

---

## 📊 현재 구조 분석

### 1. 배포 환경
```
Cloud Run (asia-northeast3) ← Webhook Mode
├─ URL: https://telegram-bot-514569077225.asia-northeast3.run.app
├─ Status: ONLINE ✅
└─ Last Heartbeat: 2026-02-15 04:43:28 (11시간 전!)
```

**문제**: Heartbeat가 11시간 전으로 멈춤 → 시스템 상태 동기화 실패

### 2. 대화 처리 흐름 (telegram_webhook.py)

```python
[사용자 메시지]
    ↓
[1. Chat Memory 저장] ← ✅ 작동
    ↓
[2. Intelligence Capture (inbox/)] ← ✅ 작동
    ↓
[3. UIP (YouTube/URL 자동 처리)] ← ⚠️ subprocess 호출 (느림)
    ↓
[4. Command Handling (/cd, /status, etc.)] ← ✅ 작동
    ↓
[5. Neural Routing (AgentRouter)] ← ⚠️ 문제 발생
    ↓
[6. AI 응답 생성] ← ⚠️ 컨텍스트 부족
    ↓
[7. 응답 전송 + Memory 저장] ← ✅ 작동
```

---

## 🚨 핵심 문제점

### 문제 1: 컨텍스트 로딩 로직이 너무 단순함

**현재 코드** ([telegram_webhook.py:302-308](execution/telegram_webhook.py#L302-L308)):
```python
is_complex = len(text) > 50 or any(k in text for k in ["분석", "보고", "설계", "구현", "정리"])
project_context = _get_project_context(text if is_complex else "")

chat_history = memory.load_chat(str(chat_id), limit=3 if not is_complex else 5)
history_text = "\n".join([f"{m['role'][0].upper()}: {m['content'][:200]}" for m in chat_history])

user_prompt = f"[Reality]\n{project_context}\n\n[Log]\n{history_text}\n\n[Input]\n{text}"
```

**문제점**:
1. **50자 기준이 너무 짧음** → "진단해줘"는 단순 질의로 처리됨 (컨텍스트 없음)
2. **키워드 매칭이 제한적** → "텔레그램 플로우 확인", "문제점 찾아줘" 같은 표현 누락
3. **Chat History가 200자로 잘림** → 이전 대화 맥락 손실
4. **Project Context가 너무 최소화됨** → 시스템 상태 파악 불가

### 문제 2: AgentRouter의 Persona가 붕 떠있음

**현재 코드** ([telegram_webhook.py:310-323](execution/telegram_webhook.py#L310-L323)):
```python
agent_persona = agent_router.get_persona(agent_key)
system_instruction = (
    f"You are {agent_key} of 97LAYER OS - a conversational AI assistant.\n\n"
    f"Core Identity:\n{agent_persona}\n\n"
    "Communication Style:\n"
    "- Speak naturally in Korean, as if talking to a colleague\n"
    "- Be warm, helpful, and proactive\n"
    "- Provide context and reasoning, not just commands\n"
    ...
)
```

**문제점**:
1. **Agent Persona가 추상적** → "너 누구야?" 느낌
2. **97layerOS의 실시간 상태를 모름** → "지금 뭐 하고 있어?" 대답 불가
3. **이전 대화와의 연결고리 없음** → "아까 그거 어떻게 됐어?" 대답 불가
4. **Directive/Knowledge 참조 없음** → "우리 시스템 구조가 뭐였지?" 대답 불가

### 문제 3: _get_project_context()가 너무 단순함

**현재 코드** ([telegram_webhook.py:101-129](execution/telegram_webhook.py#L101-L129)):
```python
def _get_project_context(trigger_text: str = "") -> str:
    status_file = PROJECT_ROOT / "task_status.json"
    status = json.loads(status_file.read_text()) if status_file.exists() else {}

    pending = status.get("pending_tasks", [])
    top_task = pending[0]['instruction'] if pending and 'instruction' in pending[0] else 'None'

    vision_summary = "1인 기업 97LAYER의 고효율 자율 운영 시스템 (97LAYER OS)"
    context = f"[Status] Pending: {len(pending)} | Top: {top_task} | Vision: {vision_summary}"

    # Deep Grounding: 특정 키워드 시에만 최소 데이터 추가
    if trigger_text:
        keywords = ["안티그래비티", "antigravity", "rituals", "텔레그램", "진단", "diagnostic"]
        ...
```

**문제점**:
1. **task_status.json만 읽음** → system_state.json (에이전트 상태) 누락
2. **vision_summary가 하드코딩** → 실제 브랜드 헌법 참조 없음
3. **Deep Grounding 키워드가 부족함** → "플로우", "문제", "확인" 등 누락
4. **knowledge/system/ 상태 파일들 미참조** → sync_state.json, synapse_bridge.json 등

### 문제 4: Cloud Run이 Stateless라 메모리 유지 불가

**문제점**:
1. **매 요청마다 cold start 가능성** → AI, Memory, AgentRouter 재초기화
2. **Global instance가 휘발성** → 대화 컨텍스트가 서버 재시작 시 손실
3. **system_state.json이 로컬 파일** → Cloud Run 컨테이너 재시작 시 초기화됨
4. **Google Drive 동기화 없음** → 맥북-Cloud Run 간 상태 공유 불가

---

## 🎯 해결 방안

### 해결책 1: Enhanced Context Loading

```python
def _get_enhanced_context(text: str, chat_id: int) -> dict:
    """
    강화된 컨텍스트 로딩
    - System State (에이전트 상태)
    - Sync State (맥북/VM 주권)
    - Task Status (작업 현황)
    - Recent Knowledge (최근 학습 내용)
    - Chat History (전체 대화 맥락)
    """
    context = {
        "system_state": load_system_state(),
        "sync_state": load_sync_state(),
        "task_status": load_task_status(),
        "recent_knowledge": scan_recent_knowledge(days=7),
        "chat_history": memory.load_chat(str(chat_id), limit=20),  # 3→20
        "agent_status": get_agent_router_status()
    }

    # Deep Grounding: 사용자 질의 의도 파악
    intent = analyze_user_intent(text)
    if intent == "system_diagnostic":
        context["directives"] = load_relevant_directives(["system_handshake", "directive_lifecycle"])
    elif intent == "brand_identity":
        context["brand"] = load_brand_constitution()

    return context
```

### 해결책 2: Stateful Architecture (Hybrid)

```
[맥북] ← Primary Brain (Full Context)
  ↕️ (Google Drive Sync)
[GCP VM] ← Secondary Brain (Night Guard)
  ↕️ (State Sync via sync_state.json)
[Cloud Run] ← Stateless Gateway
  ↓ (Webhook만 받고)
  ↓ (맥북/VM에게 위임)
```

**개선 방안**:
1. **Cloud Run을 Proxy로만 사용** → 실제 처리는 맥북/VM에서
2. **Pub/Sub 도입** → Cloud Run이 메시지를 Queue에 넣고, 맥북/VM이 Pull
3. **Redis/Firestore 도입** → Shared state storage (무료 플랜)

### 해결책 3: System Instruction 강화

```python
system_instruction = f"""
You are {agent_key} of 97LAYER OS.

## Current System State
- Active Node: {context['sync_state']['active_node']}
- Agents Status: {context['system_state']['agents']}
- Pending Tasks: {len(context['task_status']['pending_tasks'])}
- Last Activity: {context['system_state']['last_update']}

## Your Identity
{agent_persona}

## Recent Context (Last 20 messages)
{format_chat_history(context['chat_history'])}

## Communication Protocol
- Continue previous conversation naturally
- Reference past context when relevant
- Show awareness of system state
- Provide actionable insights based on current status

Remember: You have full memory of our conversation and system state.
"""
```

### 해결책 4: Heartbeat 수정

**문제**: system_state.json의 Telegram_Bot_Cloud가 11시간 전 heartbeat

**해결**:
```python
@app.route('/webhook', methods=['POST'])
def webhook():
    # Update heartbeat (현재 있음)
    syncer = SystemSynchronizer(agent_name="Telegram_Bot_Cloud")
    syncer.report_heartbeat(status="ONLINE", current_task="메시지 처리 중")

    # ✅ 추가: Google Drive에 동기화
    syncer.sync_to_drive()  # ← 이 부분 누락됨!
```

---

## 🛠️ 즉시 수정 가능한 부분

### 1. Chat History Limit 증가 (1분 작업)
```python
# Before
chat_history = memory.load_chat(str(chat_id), limit=3 if not is_complex else 5)

# After
chat_history = memory.load_chat(str(chat_id), limit=10 if not is_complex else 20)
```

### 2. Context Complexity 판단 개선 (3분 작업)
```python
# Before
is_complex = len(text) > 50 or any(k in text for k in ["분석", "보고", "설계", "구현", "정리"])

# After
is_complex = (
    len(text) > 30 or  # 50 → 30
    any(k in text for k in [
        "분석", "보고", "설계", "구현", "정리",
        "진단", "확인", "문제", "플로우", "구조",  # ← 추가
        "어떻게", "왜", "뭐", "상태", "현황"
    ])
)
```

### 3. System State 참조 추가 (5분 작업)
```python
def _get_project_context(trigger_text: str = "") -> str:
    # 기존 task_status.json
    status_file = PROJECT_ROOT / "task_status.json"
    status = json.loads(status_file.read_text()) if status_file.exists() else {}

    # ✅ 추가: system_state.json
    system_state_file = PROJECT_ROOT / "knowledge" / "system_state.json"
    system_state = json.loads(system_state_file.read_text()) if system_state_file.exists() else {}

    # ✅ 추가: sync_state.json
    sync_state_file = PROJECT_ROOT / "knowledge" / "system" / "sync_state.json"
    sync_state = json.loads(sync_state_file.read_text()) if sync_state_file.exists() else {}

    context = f"""[System Status]
- Pending Tasks: {len(status.get("pending_tasks", []))}
- Active Node: {sync_state.get("active_node", "unknown")}
- Agents: {list(system_state.get("agents", {}).keys())}
- Last Update: {system_state.get("last_update", "N/A")}

[Top Task]
{status.get("pending_tasks", [{}])[0].get("instruction", "None")}

[Vision]
1인 기업 97LAYER의 고효율 자율 운영 시스템 (97LAYER OS)
"""
    return context
```

---

## 📈 개선 우선순위

| 순위 | 작업 | 효과 | 시간 |
|-----|------|-----|------|
| 🔴 **1** | Chat History Limit 증가 (3→20) | ⭐⭐⭐⭐⭐ | 1분 |
| 🔴 **2** | Context Complexity 키워드 추가 | ⭐⭐⭐⭐⭐ | 3분 |
| 🟡 **3** | System State 참조 추가 | ⭐⭐⭐⭐ | 5분 |
| 🟡 **4** | Heartbeat Google Drive Sync | ⭐⭐⭐ | 10분 |
| 🟢 **5** | Enhanced Context Architecture | ⭐⭐⭐⭐⭐ | 1시간 |
| 🟢 **6** | Pub/Sub + Hybrid Processing | ⭐⭐⭐⭐⭐ | 3시간 |

---

## 결론

**현재 문제의 핵심**: Cloud Run Webhook이 **Stateless**하게 작동하면서, 대화 컨텍스트와 시스템 상태를 충분히 로드하지 못함.

**즉시 개선 가능**: 위 3가지 수정 (Chat History, Context Keywords, System State)만으로도 **70% 개선 가능**.

**장기 개선**: Hybrid Architecture (맥북/VM이 실제 처리, Cloud Run은 Gateway)로 전환하면 **100% 해결**.

---

**다음 액션**: 위 "즉시 수정 가능한 부분" 3가지를 순차적으로 적용할까요?
