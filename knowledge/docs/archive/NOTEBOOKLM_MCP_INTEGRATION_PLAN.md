# 🛸 NotebookLM MCP 통합 계획

> **목표**: Anti-Gravity 프로토콜의 프로덕션 구현 - NotebookLM의 28개 도구를 LAYER OS 5-Agent 시스템에 통합

---

## 📋 Executive Summary

### 현재 상황
- ✅ **Phase 3 완료**: Anti-Gravity Skill 정의, DIY youtube_analyzer.py, Telegram 통합
- ⚠️  **한계**: youtube-transcript-api 의존, AI 생성 품질 불안정, 수동 자산 생성

### NotebookLM MCP 가치 제안
- **28개 통합 도구**: source_add_*, audio_create, mindmap_create, notebook_query 등
- **Google Gemini 기반**: 검증된 RAG 시스템, 고품질 Multi-modal 출력
- **MCP 표준**: Claude Desktop, Cline, 기타 MCP 클라이언트와 즉시 호환
- **자동화 연쇄**: 사령관 시나리오 완벽 구현 가능

### 전략 결정
**Dual-Engine Architecture (이중 엔진)**:
- 🥇 **Primary Engine**: NotebookLM MCP (80% 케이스 - 고품질, 빠름)
- 🛡️  **Fallback Engine**: DIY youtube_analyzer.py (20% 케이스 - 인증 실패, 커스터마이징)

---

## 🎯 Phase 4 Implementation Roadmap

### Phase 4.1: NotebookLM MCP 설치 및 인증 (30분)

#### Step 1: 포드맨 컨테이너에 설치
```bash
# Podman 컨테이너 진입
podman exec -it 97layer-os bash

# Python 3.11+ 확인
python3 --version

# notebooklm-mcp-cli 설치
pip install notebooklm-mcp-cli

# 설치 확인
nlm --version
```

#### Step 2: 맥북 로컬에서 인증 (GUI 필요)
```bash
# 맥북 로컬 터미널에서 실행
pip install notebooklm-mcp-cli

# Chrome 브라우저로 Google 로그인 프롬프트
nlm login

# 성공 시 쿠키 파일 생성됨
# 위치: ~/.notebooklm/cookies.json
```

#### Step 3: 쿠키를 Podman 컨테이너로 복사
```bash
# 맥북 로컬에서 실행
podman cp ~/.notebooklm/cookies.json 97layer-os:/root/.notebooklm/

# 컨테이너 내부에서 확인
podman exec -it 97layer-os bash
ls -la /root/.notebooklm/cookies.json

# 테스트: 노트북 목록 조회
nlm notebook_list
```

**예상 출력**:
```json
{
  "notebooks": [
    {
      "id": "notebook_abc123",
      "title": "My Research",
      "created_at": "2026-02-16T..."
    }
  ]
}
```

---

### Phase 4.2: NotebookLM Bridge 구현 (1시간)

**목표**: 28개 도구 중 핵심 8개를 Python으로 래핑

#### 파일: `execution/system/notebooklm_bridge.py`

```python
#!/usr/bin/env python3
"""
NotebookLM MCP Bridge - LAYER OS Wrapper

28개 도구 중 Anti-Gravity 핵심 8개 래핑:
1. notebook_create/list (Foundation)
2. source_add_url/text/file (Source Grounding)
3. notebook_query (RAG)
4. audio_create (Multi-modal)
5. mindmap_create (Multi-modal)

Container-First:
- CLI 호출: subprocess로 nlm 명령 실행
- 인증: Podman 내부 ~/.notebooklm/cookies.json
- 에러 핸들링: 인증 실패 → DIY fallback 트리거
"""

import subprocess
import json
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime


class NotebookLMBridge:
    """NotebookLM MCP CLI Wrapper"""

    def __init__(self):
        self.cli_command = "nlm"

        # 인증 확인
        if not self._check_auth():
            raise RuntimeError(
                "NotebookLM 인증 필요. "
                "맥북에서 'nlm login' 실행 후 쿠키 복사하세요."
            )

    def _check_auth(self) -> bool:
        """인증 상태 확인"""
        try:
            result = subprocess.run(
                [self.cli_command, "notebook_list"],
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.returncode == 0
        except Exception:
            return False

    def _run_command(self, args: List[str]) -> Dict[str, Any]:
        """CLI 명령 실행 및 JSON 파싱"""
        try:
            result = subprocess.run(
                [self.cli_command] + args,
                capture_output=True,
                text=True,
                timeout=60
            )

            if result.returncode != 0:
                raise RuntimeError(f"CLI Error: {result.stderr}")

            return json.loads(result.stdout)

        except json.JSONDecodeError:
            # 텍스트 응답인 경우
            return {"output": result.stdout, "type": "text"}

        except Exception as e:
            raise RuntimeError(f"NotebookLM Bridge 오류: {e}")

    # === Foundation Tools ===

    def create_notebook(self, title: str) -> str:
        """새 노트북 생성"""
        result = self._run_command(["notebook_create", "--title", title])
        return result.get("notebook_id")

    def list_notebooks(self) -> List[Dict]:
        """노트북 목록 조회"""
        result = self._run_command(["notebook_list"])
        return result.get("notebooks", [])

    # === Source Grounding Tools ===

    def add_source_url(self, notebook_id: str, url: str) -> str:
        """URL 소스 추가 (YouTube, Web)"""
        result = self._run_command([
            "source_add_url",
            "--notebook-id", notebook_id,
            "--url", url
        ])
        return result.get("source_id")

    def add_source_text(self, notebook_id: str, text: str, title: str) -> str:
        """텍스트 소스 추가"""
        result = self._run_command([
            "source_add_text",
            "--notebook-id", notebook_id,
            "--title", title,
            "--text", text
        ])
        return result.get("source_id")

    def add_source_file(self, notebook_id: str, file_path: Path) -> str:
        """파일 소스 추가 (PDF, DOCX)"""
        result = self._run_command([
            "source_add_file",
            "--notebook-id", notebook_id,
            "--file", str(file_path)
        ])
        return result.get("source_id")

    # === RAG Tool ===

    def query_notebook(self, notebook_id: str, query: str) -> str:
        """노트북 소스 기반 질의 (RAG)"""
        result = self._run_command([
            "notebook_query",
            "--notebook-id", notebook_id,
            "--query", query
        ])
        return result.get("answer", result.get("output", ""))

    # === Multi-modal Synthesis Tools ===

    def create_audio(self, notebook_id: str, output_path: Optional[Path] = None) -> Path:
        """Audio Overview 생성 (Podcast)"""
        args = ["audio_create", "--notebook-id", notebook_id]

        if output_path:
            args.extend(["--output", str(output_path)])

        result = self._run_command(args)

        # CLI가 파일 경로 반환
        audio_file = result.get("audio_file", result.get("output_file"))
        return Path(audio_file)

    def create_mindmap(self, notebook_id: str) -> str:
        """Mind Map 생성 (Mermaid.js)"""
        result = self._run_command([
            "mindmap_create",
            "--notebook-id", notebook_id
        ])
        return result.get("mermaid_code", result.get("output", ""))


# === Anti-Gravity Workflow ===

def anti_gravity_youtube(url: str) -> Dict[str, Any]:
    """
    Anti-Gravity YouTube 분석 (NotebookLM 엔진)

    Workflow:
    1. 노트북 생성
    2. YouTube URL 소스 추가
    3. 3가지 질의 (요약, 인사이트, 브랜드 연결)
    4. Audio Overview + Mind Map 생성
    5. 결과 반환
    """
    bridge = NotebookLMBridge()

    # Step 1: 노트북 생성
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    notebook_id = bridge.create_notebook(f"YouTube Analysis {timestamp}")

    # Step 2: 소스 추가
    source_id = bridge.add_source_url(notebook_id, url)

    # Step 3: RAG 질의 (3가지)
    summary = bridge.query_notebook(
        notebook_id,
        "이 영상의 핵심 메시지를 3줄로 요약해주세요."
    )

    insights = bridge.query_notebook(
        notebook_id,
        "이 영상에서 가장 독창적인 인사이트는 무엇인가요? "
        "Aesop 스타일로 절제되고 본질적인 언어로 답해주세요."
    )

    brand_connection = bridge.query_notebook(
        notebook_id,
        "이 내용이 다음 5가지 브랜드 철학 중 어디에 연결되나요? "
        "1) Authenticity 2) Practicality 3) Elegance 4) Precision 5) Innovation"
    )

    # Step 4: Multi-modal 자산 생성
    audio_path = bridge.create_audio(notebook_id)
    mindmap_mermaid = bridge.create_mindmap(notebook_id)

    return {
        "notebook_id": notebook_id,
        "source_id": source_id,
        "summary": summary,
        "insights": insights,
        "brand_connection": brand_connection,
        "audio_file": audio_path,
        "mindmap": mindmap_mermaid
    }
```

---

### Phase 4.3: youtube_analyzer.py 리팩터링 (30분)

**목표**: Dual-Engine 아키텍처 구현

#### 변경사항

```python
# execution/system/youtube_analyzer.py

class YouTubeAnalyzer:
    def __init__(self, engine: str = "auto"):
        """
        engine 옵션:
        - "auto": NotebookLM 시도 → 실패 시 DIY
        - "notebooklm": NotebookLM 전용 (실패 시 에러)
        - "diy": DIY 전용
        """
        self.engine = engine

        # NotebookLM Bridge 초기화 시도
        if engine in ["auto", "notebooklm"]:
            try:
                from execution.system.notebooklm_bridge import NotebookLMBridge
                self.notebooklm = NotebookLMBridge()
                self.notebooklm_available = True
            except Exception as e:
                self.notebooklm_available = False
                if engine == "notebooklm":
                    raise RuntimeError(f"NotebookLM 엔진 초기화 실패: {e}")

    def analyze(self, url: str) -> Dict[str, Path]:
        """Dual-Engine 분석"""

        # Engine 선택 로직
        if self.engine == "auto":
            if self.notebooklm_available:
                try:
                    return self._analyze_notebooklm(url)
                except Exception as e:
                    logger.warning(f"NotebookLM 실패, DIY로 fallback: {e}")
                    return self._analyze_diy(url)
            else:
                return self._analyze_diy(url)

        elif self.engine == "notebooklm":
            return self._analyze_notebooklm(url)

        else:  # "diy"
            return self._analyze_diy(url)

    def _analyze_notebooklm(self, url: str) -> Dict[str, Path]:
        """NotebookLM MCP 엔진"""
        from execution.system.notebooklm_bridge import anti_gravity_youtube

        result = anti_gravity_youtube(url)

        # 결과를 표준 포맷으로 변환
        return {
            "source": self._save_notebooklm_source(result),
            "audio": result["audio_file"],
            "deck": self._convert_to_deck(result),
            "map": self._save_mindmap(result["mindmap"])
        }

    def _analyze_diy(self, url: str) -> Dict[str, Path]:
        """기존 DIY 엔진 (youtube-transcript-api)"""
        # 기존 로직 유지...
        pass
```

---

### Phase 4.4: Telegram 통합 업그레이드 (30분)

#### 변경사항: `/youtube` 명령어 개선

```python
# execution/daemons/telegram_secretary.py

async def youtube_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
    """
    /youtube <URL> [--engine notebooklm|diy|auto]

    예시:
    /youtube https://youtu.be/xxxxx
    /youtube https://youtu.be/xxxxx --engine notebooklm
    """

    # 파라미터 파싱
    args = context.args
    url = args[0] if args else None
    engine = "auto"  # 기본값

    if "--engine" in args:
        idx = args.index("--engine")
        engine = args[idx + 1]
        url = args[0]

    # 엔진 표시
    engine_emoji = {
        "notebooklm": "🤖 NotebookLM (Google Gemini)",
        "diy": "🛠️  DIY (youtube-transcript-api)",
        "auto": "🛸 Auto (NotebookLM → DIY fallback)"
    }

    await update.message.reply_text(
        f"🛸 **Anti-Gravity 프로토콜 시작**\n\n"
        f"🔧 Engine: {engine_emoji.get(engine, engine)}\n"
        f"🔗 URL: {url}\n\n"
        f"⏳ 처리 중..."
    )

    # 분석 실행
    analyzer = YouTubeAnalyzer(engine=engine)
    results = await asyncio.to_thread(analyzer.analyze, url)

    # 결과 전송 (엔진에 따라 메시지 커스터마이징)
    if analyzer.notebooklm_available and engine != "diy":
        response = "✅ **NotebookLM MCP로 분석 완료**\n\n"
        response += "🎙️  Audio Overview: Google Gemini 생성\n"
        response += "🗺️  Mind Map: NotebookLM RAG 기반\n"
    else:
        response = "✅ **DIY 엔진으로 분석 완료**\n\n"
        response += "🎙️  Audio Overview: Claude/Gemini 생성\n"
        response += "🗺️  Mind Map: 자체 알고리즘\n"

    # ...
```

---

### Phase 4.5: End-to-End 테스트 (1시간)

#### 테스트 시나리오

**Scenario 1: NotebookLM 엔진 (Happy Path)**
```bash
# Telegram에서
/youtube https://youtu.be/blWbJOEheSA

# 예상 결과:
# ✅ NotebookLM MCP로 분석 완료
# 🎙️  Audio Overview: [파일 경로]
# 🗺️  Mind Map: [Mermaid 다이어그램]
# 📊 Brand Connection: Authenticity + Precision
```

**Scenario 2: DIY Fallback (인증 실패)**
```bash
# NotebookLM 쿠키 삭제 후
/youtube https://youtu.be/blWbJOEheSA

# 예상 결과:
# ⚠️  NotebookLM 실패, DIY로 fallback
# ✅ DIY 엔진으로 분석 완료
```

**Scenario 3: 수동 엔진 선택**
```bash
/youtube https://youtu.be/blWbJOEheSA --engine diy
/youtube https://youtu.be/blWbJOEheSA --engine notebooklm
```

#### 품질 검증 체크리스트

- [ ] **Audio Overview**:
  - NotebookLM: Google Gemini 음성 품질
  - DIY: Claude/Gemini 텍스트 품질
  - 비교: Aesop Score ≥ 70%

- [ ] **Mind Map**:
  - NotebookLM: RAG 기반 구조적 정확성
  - DIY: 자체 알고리즘 커버리지
  - 비교: Brand Pillars 연결 여부

- [ ] **Ralph Loop MBQ**:
  - Meaning: 5 Pillars 중 2개+ 연결
  - Brand: Aesop Benchmark 70%+
  - Quality: 구조적 완결성

---

### Phase 4.6: Parallel Orchestrator 확장 (선택사항, 2시간)

**목표**: 5-Agent Junction Protocol 완전 통합

#### 파일: `execution/system/parallel_orchestrator.py`

```python
class ParallelOrchestrator:

    async def process_youtube_notebooklm(self, url: str) -> Dict[str, Any]:
        """
        NotebookLM MCP + 5-Agent Junction Protocol

        Junction 5단계:
        1. Capture (SA): source_add_url
        2. Connect (SA): notebook_query (기존 지식 연결)
        3. Meaning (CE): notebook_query (Aesop 스타일 재구성)
        4. Manifest (AD): audio_create + mindmap_create
        5. Cycle (CD + TD): Ralph Loop MBQ 검증
        """

        # Step 1: Capture (SA)
        sa_task = asyncio.create_task(
            self._agent_capture_notebooklm(url)
        )

        # Step 2: Connect (SA 병렬)
        connect_task = asyncio.create_task(
            self._agent_connect_notebooklm(url)
        )

        sa_result, connect_result = await asyncio.gather(sa_task, connect_task)

        # Step 3: Meaning (CE 순차)
        ce_result = await self._agent_meaning_notebooklm(
            notebook_id=sa_result["notebook_id"]
        )

        # Step 4: Manifest (AD 순차)
        ad_result = await self._agent_manifest_notebooklm(
            notebook_id=sa_result["notebook_id"]
        )

        # Step 5: Cycle (CD + TD)
        cd_validation = await self._agent_cycle_ralph(
            audio=ad_result["audio"],
            mindmap=ad_result["mindmap"]
        )

        return {
            "status": "success",
            "assets": ad_result,
            "validation": cd_validation,
            "quality_score": cd_validation["ralph_score"]
        }
```

---

## 🎯 Success Criteria

### Phase 4 완료 조건

1. ✅ **설치 확인**: Podman 내부에서 `nlm notebook_list` 성공
2. ✅ **인증 확인**: 쿠키 파일 복사 완료, 만료 없음
3. ✅ **Bridge 구현**: 8개 핵심 도구 래핑 완료, 단위 테스트 통과
4. ✅ **Dual-Engine**: NotebookLM → DIY fallback 정상 작동
5. ✅ **Telegram 통합**: `/youtube` 명령어로 엔드투엔드 성공
6. ✅ **품질 검증**: Ralph Loop MBQ 통과 (Aesop Score ≥ 70%)

### 프로덕션 준비 체크리스트

- [ ] NotebookLM 쿠키 자동 갱신 메커니즘 (선택사항)
- [ ] 에러 로깅 및 모니터링 (Telegram 알림)
- [ ] 사용량 제한 (NotebookLM API rate limit 고려)
- [ ] 비용 추적 (NotebookLM 무료 티어 확인)

---

## 🚨 Risks & Mitigation

### Risk 1: 인증 만료
- **리스크**: Google 쿠키 만료 시 NotebookLM 작동 중단
- **완화**:
  - Dual-Engine으로 DIY fallback 자동 전환
  - Telegram으로 "쿠키 갱신 필요" 알림
  - 쿠키 만료 전 자동 재인증 스크립트 (cron)

### Risk 2: API 변경
- **리스크**: Google이 NotebookLM 내부 API 변경
- **완화**:
  - notebooklm-mcp-cli 커뮤니티 모니터링
  - DIY 엔진을 기본 안전망으로 유지
  - 버전 고정 (특정 버전에서 안정화 후)

### Risk 3: 품질 일관성
- **리스크**: NotebookLM 출력이 Brand Voice와 불일치
- **완화**:
  - Ralph Loop MBQ 강제 검증
  - notebook_query에 "Aesop 스타일" 명시
  - CE (Chief Editor) 후처리 단계 추가

---

## 📊 Expected Outcomes

### 정량적 개선

| Metric | Before (DIY) | After (NotebookLM) | Improvement |
|--------|--------------|---------------------|-------------|
| **분석 시간** | 2-3분 | 30-60초 | **3-4배 단축** |
| **Audio 품질** | Claude 생성 (불안정) | Gemini 음성 (일관) | **안정성 +50%** |
| **Mind Map 정확도** | 자체 알고리즘 | RAG 기반 | **정확도 +30%** |
| **Brand Alignment** | 수동 후처리 필요 | Query로 자동 조정 | **자동화 100%** |

### 정성적 가치

- **프로덕션 품질 즉시 확보**: Google 검증된 시스템
- **28개 도구 확장 가능성**: Audio/Map 외에도 Study Guide, Infographic 등
- **Cross-AI Context**: NotebookLM → Claude/Gemini 맥락 공유
- **자동화 연쇄 구현**: 사령관의 "YouTube → 요약 → 자산 → 텔레그램" 완성

---

## 🔄 Next Steps After Phase 4

### Phase 5: 완전 자율 루프
- **Scheduled Analysis**: cron으로 매일 큐레이션된 YouTube 영상 자동 분석
- **Push Notifications**: 텔레그램으로 아침 브리핑 (오디오 + 마인드맵)
- **Asset Library**: knowledge/assets/를 검색 가능한 벡터 DB로 구축

### Phase 6: 확장 적용
- **PDF Research**: 논문, 리포트를 NotebookLM으로 분석
- **Web Article**: 경쟁사 블로그, 뉴스 자동 큐레이션
- **Brand Strategy**: 멀티 소스 통합 인사이트 (YouTube + PDF + Web)

---

## 💭 Philosophical Alignment

> "Anti-Gravity는 무게를 없애는 것이 아니라, 구조를 만드는 것이다."

NotebookLM MCP 통합은:
- **Source Grounding**: 출처 명시로 환각 제거
- **Multi-modal**: 오디오, 비주얼, 다이어그램으로 다층 이해
- **MCP Connector**: AI 간 맥락 공유로 지식 순환

이는 Slow Life 철학과 완벽히 일치:
- 빠르게 소비하지 않고, **구조화된 자산**으로 보존
- 알고리즘에 의존하지 않고, **출처 기반 진실성**
- 일회성 응답이 아닌, **재사용 가능한 지식 제품**

---

> **사령관 승인 대기 중**
> Phase 4.1 설치부터 시작할까요, 아니면 계획 수정이 필요한가요?
