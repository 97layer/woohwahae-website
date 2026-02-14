#!/usr/bin/env python3
"""
97LAYER OS - 5-Agent Multimodal System with Agent Hub Integration
완전한 자율 협업 시스템: Agent Hub + Anti-Gravity + Real-time Dashboard

Features:
- 에이전트 간 직접 통신 (Agent Hub)
- 충돌 방지 메커니즘 (Anti-Gravity)
- Junction Protocol 자동화
- 실시간 Dashboard 연동
"""

import os
import sys
import json
import time
import urllib.request
import threading
import queue
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List
from enum import Enum

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Import Agent Hub
from libs.agent_hub import AgentHub, MessageType

# Configuration
BOT_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
GEMINI_KEY = "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
CLAUDE_KEY = os.getenv("ANTHROPIC_API_KEY", "")
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"

# State Files
STATE_FILE = PROJECT_ROOT / "knowledge" / "system_state.json"
SYNAPSE_FILE = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"


class TaskPriority(Enum):
    """작업 우선순위 (Anti-Gravity)"""
    CRITICAL = 1  # CD 최종 판단
    HIGH = 2      # SA 분석
    MEDIUM = 3    # CE 콘텐츠 생성
    LOW = 4       # AD 시각 분석


class StrategyAnalyst:
    """SA - 정보 수집 및 패턴 분석 (Gemini)"""

    def __init__(self, gemini_key: str, hub: AgentHub):
        self.gemini_key = gemini_key
        self.name = "Strategy Analyst (SA)"
        self.hub = hub
        self.agent_key = "SA"

    def analyze_signal(self, content: str, signal_id: str = None) -> Dict[str, Any]:
        """신호 분석 및 패턴 탐지"""
        print(f"[{self.name}] Analyzing signal...")

        # Context7 MCP로 최신 브랜딩 트렌드 참조 가능
        prompt = f"""다음 텍스트를 97layer의 5대 철학 축 관점에서 분석하세요:
1. 고독 (Solitary Essence)
2. 불완전 (Imperfection)
3. 시간 (Time Archive)
4. 선례 (Precedent Setting)
5. 반알고리즘 (Anti-Algorithm)

텍스트: {content[:500]}

JSON 형식으로 응답:
{{"patterns": ["keyword1", "keyword2"], "philosophy_match": "가장 관련된 축", "score": 0-100}}"""

        response = self._call_gemini(prompt)

        # JSON 파싱
        try:
            import re
            json_match = re.search(r'\{.*\}', response, re.DOTALL)
            if json_match:
                analysis = json.loads(json_match.group())
            else:
                analysis = {
                    "patterns": [],
                    "philosophy_match": "unknown",
                    "score": 50,
                    "raw_analysis": response[:200]
                }
        except:
            analysis = {"raw_analysis": response[:200], "score": 50}

        print(f"[{self.name}] Analysis complete: {analysis.get('philosophy_match', 'N/A')} (Score: {analysis.get('score', 0)})")

        # Hub를 통해 CE에게 전달 (점수 60 이상)
        if analysis.get("score", 0) >= 60 and signal_id:
            self.hub.send_message(
                self.agent_key, "CE",
                MessageType.REQUEST,
                {
                    "action": "generate_content",
                    "signal_id": signal_id,
                    "content": content,
                    "analysis": analysis
                }
            )
            print(f"[{self.name}] Forwarded to CE for content generation")

        return analysis

    def _call_gemini(self, prompt: str) -> str:
        """Gemini API 호출"""
        url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent"

        data = json.dumps({
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {"temperature": 0.7, "maxOutputTokens": 500}
        }).encode('utf-8')

        url_with_key = f"{url}?key={self.gemini_key}"
        req = urllib.request.Request(url_with_key, data=data, headers={"Content-Type": "application/json"})

        try:
            with urllib.request.urlopen(req, timeout=15) as resp:
                result = json.loads(resp.read().decode('utf-8'))
            return result['candidates'][0]['content']['parts'][0]['text']
        except Exception as e:
            return f"Error: {e}"


class ArtDirector:
    """AD - 시각 분석 및 디자인 (Gemini Vision)"""

    def __init__(self, gemini_key: str, hub: AgentHub):
        self.gemini_key = gemini_key
        self.name = "Art Director (AD)"
        self.hub = hub
        self.agent_key = "AD"

    def analyze_image(self, image_bytes: bytes, caption: str = "", signal_id: str = None) -> Dict[str, Any]:
        """이미지 멀티모달 분석"""
        print(f"[{self.name}] Analyzing image with Gemini Vision...")

        import base64
        image_base64 = base64.b64encode(image_bytes).decode('utf-8')

        prompt = f"""WOOHWAHAE 비주얼 아이덴티티 관점에서 이미지를 분석하세요:
- 모노크롬 미학 적합성
- 60% 여백 원칙
- 미니멀리즘 구현도
- 브랜드 철학 반영도

{f'캡션: {caption}' if caption else ''}

JSON 형식으로 응답:
{{"aesthetic_score": 0-100, "recommendations": ["제안1", "제안2"], "brand_fit": "high/medium/low"}}"""

        url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent"

        data = json.dumps({
            "contents": [{
                "parts": [
                    {"text": prompt},
                    {"inline_data": {"mime_type": "image/jpeg", "data": image_base64}}
                ]
            }],
            "generationConfig": {"temperature": 0.7, "maxOutputTokens": 500}
        }).encode('utf-8')

        url_with_key = f"{url}?key={self.gemini_key}"
        req = urllib.request.Request(url_with_key, data=data, headers={"Content-Type": "application/json"})

        try:
            with urllib.request.urlopen(req, timeout=20) as resp:
                result = json.loads(resp.read().decode('utf-8'))
            analysis_text = result['candidates'][0]['content']['parts'][0]['text']

            import re
            json_match = re.search(r'\{.*\}', analysis_text, re.DOTALL)
            if json_match:
                analysis = json.loads(json_match.group())
            else:
                analysis = {"raw_analysis": analysis_text[:300]}

            print(f"[{self.name}] Image analysis complete")
            return analysis

        except Exception as e:
            print(f"[{self.name}] Error: {e}")
            return {"error": str(e)}


class ChiefEditor:
    """CE - 콘텐츠 서사 집필 (Gemini)"""

    def __init__(self, gemini_key: str, hub: AgentHub):
        self.gemini_key = gemini_key
        self.name = "Chief Editor (CE)"
        self.hub = hub
        self.agent_key = "CE"

    def generate_content(self, signal: str, sa_analysis: Dict, signal_id: str = None) -> str:
        """Aesop 벤치마크 기반 콘텐츠 생성"""
        print(f"[{self.name}] Generating content...")

        prompt = f"""97layer/WOOHWAHAE 스타일로 인스타그램 캡션을 작성하세요.

입력 신호: {signal[:300]}
SA 분석: {json.dumps(sa_analysis, ensure_ascii=False)}

요구사항:
- 400-800자
- Aesop 톤: 절제되고 지적이며 과장 없음
- Hook → Manuscript → Afterglow 구조
- 이모지 없음, 볼드 없음
- 질문이나 미완성 사고로 끝

한국어로 작성."""

        response = self._call_gemini(prompt)
        print(f"[{self.name}] Content generated ({len(response)} chars)")

        # Hub를 통해 CD에게 승인 요청 (chat_id 포함)
        if signal_id:
            msg_data = {
                    "action": "approve_content",
                    "signal_id": signal_id,
                    "content": response,
                    "metadata": {
                        "original_signal": signal[:200],
                        "sa_analysis": sa_analysis
                    }
                }
            self.hub.send_message(self.agent_key, "CD", MessageType.REQUEST, msg_data)
            print(f"[{self.name}] Sent to CD for approval")

        return response

    def _call_gemini(self, prompt: str) -> str:
        """Gemini API 호출"""
        url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent"

        data = json.dumps({
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {"temperature": 0.8, "maxOutputTokens": 1000}
        }).encode('utf-8')

        url_with_key = f"{url}?key={self.gemini_key}"
        req = urllib.request.Request(url_with_key, data=data, headers={"Content-Type": "application/json"})

        try:
            with urllib.request.urlopen(req, timeout=15) as resp:
                result = json.loads(resp.read().decode('utf-8'))
            return result['candidates'][0]['content']['parts'][0]['text']
        except Exception as e:
            return f"Error generating content: {e}"


class CreativeDirector:
    """CD - 최종 의사결정 (Claude Haiku)"""

    def __init__(self, claude_key: str, gemini_key: str, hub: AgentHub):
        self.claude_key = claude_key if claude_key and "your_" not in claude_key else None
        self.gemini_key = gemini_key
        self.name = "Creative Director (CD)"
        self.hub = hub
        self.agent_key = "CD"
        self.using_claude = bool(self.claude_key)

    def sovereign_judgment(self, content: str, metadata: Dict, signal_id: str = None) -> Dict[str, Any]:
        """Sovereign 최종 판단"""
        print(f"[{self.name}] Making sovereign judgment...")
        print(f"[{self.name}] Engine: {'Claude Haiku' if self.using_claude else 'Gemini (fallback)'}")

        if self.using_claude:
            judgment = self._claude_judgment(content, metadata)
        else:
            judgment = self._gemini_judgment(content, metadata)

        # Hub를 통해 TD에게 결과 전달
        if signal_id:
            self.hub.send_message(
                self.agent_key, "TD",
                MessageType.DECISION,
                {
                    "signal_id": signal_id,
                    "judgment": judgment,
                    "content": content if judgment.get("approved") else None
                }
            )
            print(f"[{self.name}] Decision sent to TD")

        return judgment

    def _claude_judgment(self, content: str, metadata: Dict) -> Dict[str, Any]:
        """Claude Haiku로 판단"""
        try:
            import anthropic
            client = anthropic.Anthropic(api_key=self.claude_key)

            prompt = f"""WOOHWAHAE Creative Director로서 콘텐츠를 최종 승인 판단하세요.

콘텐츠:
{content}

메타데이터:
{json.dumps(metadata, ensure_ascii=False, indent=2)}

MBQ 기준:
1. 철학적 일치성 (97layer 5대 축)
2. 톤 일관성 (Aesop 벤치마크)
3. 구조 완성도 (Hook→Manuscript→Afterglow)
4. 반알고리즘성

JSON 형식으로 응답:
{{
  "approved": true/false,
  "score": 0-100,
  "strengths": ["..."],
  "weaknesses": ["..."],
  "decision": "승인/반려 사유"
}}"""

            response = client.messages.create(
                model="claude-3-haiku-20240307",  # Haiku for cost efficiency
                max_tokens=800,
                temperature=0.3,
                messages=[{"role": "user", "content": prompt}]
            )

            result_text = response.content[0].text

            import re
            json_match = re.search(r'\{.*\}', result_text, re.DOTALL)
            if json_match:
                judgment = json.loads(json_match.group())
            else:
                judgment = {"raw_response": result_text, "approved": False}

            print(f"[{self.name}] Claude judgment: {'APPROVED' if judgment.get('approved') else 'REJECTED'} (Score: {judgment.get('score', 0)})")
            return judgment

        except Exception as e:
            print(f"[{self.name}] Claude error: {e}")
            return self._gemini_judgment(content, metadata)

    def _gemini_judgment(self, content: str, metadata: Dict) -> Dict[str, Any]:
        """Gemini로 대체 판단"""
        url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent"

        prompt = f"""WOOHWAHAE Creative Director로서 콘텐츠를 판단하세요.

콘텐츠: {content[:500]}

JSON 형식: {{"approved": true/false, "score": 0-100, "decision": "이유"}}"""

        data = json.dumps({
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {"temperature": 0.3, "maxOutputTokens": 300}
        }).encode('utf-8')

        url_with_key = f"{url}?key={self.gemini_key}"
        req = urllib.request.Request(url_with_key, data=data, headers={"Content-Type": "application/json"})

        try:
            with urllib.request.urlopen(req, timeout=15) as resp:
                result = json.loads(resp.read().decode('utf-8'))
            response_text = result['candidates'][0]['content']['parts'][0]['text']

            import re
            json_match = re.search(r'\{.*\}', response_text, re.DOTALL)
            if json_match:
                return json.loads(json_match.group())
            return {"approved": False, "raw": response_text, "score": 0}

        except Exception as e:
            return {"approved": False, "error": str(e), "score": 0}


class TechnicalDirector:
    """TD - 전체 시스템 오케스트레이션 + Agent Hub 통합"""

    def __init__(self):
        self.name = "Technical Director (TD)"
        self.agent_key = "TD"

        # Agent Hub 초기화
        self.hub = AgentHub(str(PROJECT_ROOT))

        # 에이전트 초기화 (Hub 주입)
        self.sa = StrategyAnalyst(GEMINI_KEY, self.hub)
        self.ad = ArtDirector(GEMINI_KEY, self.hub)
        self.ce = ChiefEditor(GEMINI_KEY, self.hub)
        self.cd = CreativeDirector(CLAUDE_KEY, GEMINI_KEY, self.hub)

        # Hub에 에이전트 등록
        self.hub.register_agent("SA", self._sa_handler)
        self.hub.register_agent("AD", self._ad_handler)
        self.hub.register_agent("CE", self._ce_handler)
        self.hub.register_agent("CD", self._cd_handler)
        self.hub.register_agent("TD", self._td_handler)

        # Notifier 초기화 (텔레그램 전송용)
        from libs.notifier import Notifier
        self.notifier = Notifier()

        # Anti-Gravity: 작업 큐 (우선순위 기반)
        self.task_queue = queue.PriorityQueue()
        self.active_signals = {}  # signal_id -> lock

        # 통계
        self.stats = {
            "signals_captured": 0,
            "images_analyzed": 0,
            "content_generated": 0,
            "approved": 0,
            "rejected": 0
        }

        print(f"[{self.name}] Initialized with Agent Hub")
        self._update_synapse()

    def _sa_handler(self, message: Dict) -> Any:
        """SA 메시지 핸들러"""
        action = message["data"].get("action")
        if action == "analyze":
            content = message["data"].get("content")
            signal_id = message["data"].get("signal_id")
            # chat_id는 TD에서 직접 관리하므로 더 이상 SA에게 넘길 필요 없음
            return self.sa.analyze_signal(content, signal_id)
        return None

    def _ad_handler(self, message: Dict) -> Any:
        """AD 메시지 핸들러"""
        action = message["data"].get("action")
        if action == "analyze_image":
            image_bytes = message["data"].get("image_bytes")
            caption = message["data"].get("caption", "")
            signal_id = message["data"].get("signal_id")
            return self.ad.analyze_image(image_bytes, caption, signal_id)
        return None

    def _ce_handler(self, message: Dict) -> Any:
        """CE 메시지 핸들러"""
        action = message["data"].get("action")
        if action == "generate_content":
            content = message["data"].get("content")
            analysis = message["data"].get("analysis")
            signal_id = message["data"].get("signal_id")
            return self.ce.generate_content(content, analysis, signal_id)
        return None

    def _cd_handler(self, message: Dict) -> Any:
        """CD 메시지 핸들러"""
        action = message["data"].get("action")
        if action == "approve_content":
            content = message["data"].get("content")
            metadata = message["data"].get("metadata")
            signal_id = message["data"].get("signal_id")
            return self.cd.sovereign_judgment(content, metadata, signal_id)
        return None

    def _td_handler(self, message: Dict) -> Any:
        """TD 메시지 핸들러 (최종 결과 수신)"""
        msg_type = message["type"]
        if msg_type == "decision":
            signal_id = message["data"].get("signal_id")
            judgment = message["data"].get("judgment")
            content = message["data"].get("content")

            if judgment.get("approved"):
                self._save_approved(content, judgment, signal_id)
                self.stats["approved"] += 1
                print(f"[{self.name}] Content APPROVED and saved: {signal_id}")
                
                # 사용자에게 알림 전송 (active_signals에서 chat_id 조회)
                signal_info = self.active_signals.get(signal_id, {})
                chat_id = signal_info.get("chat_id")
                
                if chat_id:
                    briefing = f"◈ *Sovereign Judgment - {signal_id}*\n\n"
                    # content가 없는 경우 (판단만 하는 경우 등) 처리
                    display_content = content if content else judgment.get("decision", "No content provided")
                    briefing += f"{display_content}\n\n"
                    briefing += f"✓ Score: {judgment.get('score', 0)}/100\n"
                    briefing += f"✓ Decision: {judgment.get('decision', 'Approved')}"
                    self.notifier.send_message(int(chat_id), briefing)
            else:
                self.stats["rejected"] += 1
                print(f"[{self.name}] Content REJECTED: {signal_id}")
                
                # 반려 알림 전송
                signal_info = self.active_signals.get(signal_id, {})
                chat_id = signal_info.get("chat_id")
                if chat_id:
                    self.notifier.send_message(int(chat_id), f"⚠️ *Content Rejected - {signal_id}*\n\nReason: {judgment.get('decision', 'Fit criteria not met')}")

            # 작업 완료 - 잠금 해제
            if signal_id in self.active_signals:
                del self.active_signals[signal_id]

        return None

    def process_text_signal(self, text: str, user: str) -> str:
        """텍스트 신호 처리 파이프라인 (Junction Protocol)"""
        print(f"\n{'='*60}")
        print(f"[{self.name}] Processing text signal from {user}")
        print(f"{'='*60}")

        # Signal ID 생성
        signal_id = f"sig-{int(datetime.now().timestamp())}"

        # Anti-Gravity: 중복 처리 방지
        if signal_id in self.active_signals:
            print(f"[{self.name}] Signal {signal_id} already in progress")
            return signal_id

        self.active_signals[signal_id] = {"lock": threading.Lock(), "chat_id": user}

        # Stage 1: Capture
        signal_file = self._save_signal(text, "text", user, signal_id)
        self.stats["signals_captured"] += 1

        # Stage 2: Hub를 통해 SA에게 분석 요청
        self.hub.send_message(
            self.agent_key, "SA",
            MessageType.REQUEST,
            {
                "action": "analyze",
                "content": text,
                "signal_id": signal_id
            }
        )

        self._update_synapse()
        return signal_id

    def process_image_signal(self, image_bytes: bytes, caption: str, user: str) -> str:
        """이미지 신호 처리 파이프라인"""
        print(f"\n{'='*60}")
        print(f"[{self.name}] Processing image signal from {user}")
        print(f"{'='*60}")

        signal_id = f"img-{int(datetime.now().timestamp())}"

        # Stage 1: Capture
        signal_file = self._save_signal(f"[IMAGE] {caption}", "image", user, signal_id)
        self.stats["signals_captured"] += 1
        self.stats["images_analyzed"] += 1

        # Stage 2: Hub를 통해 AD에게 분석 요청
        self.hub.send_message(
            self.agent_key, "AD",
            MessageType.REQUEST,
            {
                "action": "analyze_image",
                "image_bytes": image_bytes,
                "caption": caption,
                "signal_id": signal_id
            }
        )

        self._update_synapse()
        return signal_id

    def _save_signal(self, content: str, signal_type: str, source: str, signal_id: str) -> str:
        """지식 베이스에 신호 저장"""
        signal_dir = PROJECT_ROOT / "knowledge" / "raw_signals"
        signal_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now()
        filename = f"{signal_id}_{signal_type}_{source}.md"

        with open(signal_dir / filename, "w", encoding="utf-8") as f:
            f.write(f"# Raw Signal - {signal_type.upper()}\n\n")
            f.write(f"**Signal ID**: {signal_id}\n")
            f.write(f"**Date**: {timestamp.isoformat()}\n")
            f.write(f"**Source**: {source}\n")
            f.write(f"**Type**: {signal_type}\n\n")
            f.write(f"---\n\n{content}\n")

        return filename

    def _save_approved(self, content: str, judgment: Dict, signal_id: str):
        """승인된 콘텐츠 저장 (Junction Protocol 완료)"""
        publish_dir = PROJECT_ROOT / "knowledge" / "assets" / "ready_to_publish"
        publish_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now()
        filename = f"approved_{signal_id}.md"

        with open(publish_dir / filename, "w", encoding="utf-8") as f:
            f.write(f"# Approved Content\n\n")
            f.write(f"**Signal ID**: {signal_id}\n")
            f.write(f"**Date**: {timestamp.isoformat()}\n")
            f.write(f"**Score**: {judgment.get('score', 0)}/100\n")
            f.write(f"**Decision**: {judgment.get('decision', '')}\n\n")
            f.write(f"---\n\n{content}\n")

        print(f"[{self.name}] Saved to: {filename}")

    def _update_synapse(self):
        """Synapse Bridge 업데이트 (Anti-Gravity 상태 동기화)"""
        synapse_data = {
            "active_agents": {
                "SA": {
                    "role": "Strategy Analyst",
                    "status": "active",
                    "current_task": "Pattern analysis",
                    "last_heartbeat": datetime.now().isoformat()
                },
                "AD": {
                    "role": "Art Director",
                    "status": "active",
                    "current_task": "Visual analysis",
                    "last_heartbeat": datetime.now().isoformat()
                },
                "CE": {
                    "role": "Chief Editor",
                    "status": "active",
                    "current_task": "Content generation",
                    "last_heartbeat": datetime.now().isoformat()
                },
                "CD": {
                    "role": "Creative Director",
                    "status": "active",
                    "current_task": "Sovereign judgment",
                    "last_heartbeat": datetime.now().isoformat()
                },
                "TD": {
                    "role": "Technical Director",
                    "status": "active",
                    "current_task": "Orchestration",
                    "last_heartbeat": datetime.now().isoformat()
                }
            },
            "collaboration_mode": "Active",
            "synapse_status": "Synchronized",
            "active_signals": len(self.active_signals),
            "stats": self.stats,
            "last_update": datetime.now().isoformat()
        }

        SYNAPSE_FILE.parent.mkdir(parents=True, exist_ok=True)
        with open(SYNAPSE_FILE, 'w', encoding='utf-8') as f:
            json.dump(synapse_data, f, indent=2, ensure_ascii=False)

        # system_state.json도 업데이트
        self._update_system_state()

    def _update_system_state(self):
        """System State 업데이트"""
        state_data = {
            "system_status": "OPERATIONAL",
            "last_update": datetime.now().isoformat(),
            "agents": {
                "SA": {"status": "ACTIVE", "last_heartbeat": datetime.now().isoformat()},
                "AD": {"status": "ACTIVE", "last_heartbeat": datetime.now().isoformat()},
                "CE": {"status": "ACTIVE", "last_heartbeat": datetime.now().isoformat()},
                "CD": {"status": "ACTIVE", "last_heartbeat": datetime.now().isoformat()},
                "TD": {"status": "ACTIVE", "last_heartbeat": datetime.now().isoformat()}
            },
            "stats": self.stats
        }

        STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
        with open(STATE_FILE, 'w', encoding='utf-8') as f:
            json.dump(state_data, f, indent=2, ensure_ascii=False)


class FiveAgentBot:
    """5인 체계 텔레그램 봇 (Agent Hub 통합)"""

    def __init__(self):
        self.td = TechnicalDirector()
        self.offset = None

    def get_updates(self):
        """텔레그램 업데이트 가져오기"""
        url = f"{BASE_URL}/getUpdates?timeout=30"
        if self.offset:
            url += f"&offset={self.offset}"

        try:
            with urllib.request.urlopen(url, timeout=35) as response:
                data = json.loads(response.read().decode('utf-8'))
                return data.get("result", []) if data.get("ok") else []
        except Exception as e:
            print(f"[ERROR] Get updates: {e}")
            return []

    def send_message(self, chat_id: int, text: str):
        """메시지 전송"""
        url = f"{BASE_URL}/sendMessage"
        data = json.dumps({
            "chat_id": chat_id,
            "text": text[:4000],
            "parse_mode": "Markdown"
        }).encode('utf-8')

        try:
            req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"})
            with urllib.request.urlopen(req, timeout=10):
                return True
        except Exception as e:
            print(f"[ERROR] Send message: {e}")
            return False

    def handle_command(self, message: Dict):
        """명령어 처리"""
        text = message.get("text", "")
        chat_id = message["chat"]["id"]

        if text == "/start":
            welcome = """🦋 *97LAYER OS - 5-Agent Hub System*

*5인 체계 (Agent Hub 통합):*
• Strategy Analyst (SA) - 패턴 분석
• Art Director (AD) - 이미지 분석
• Chief Editor (CE) - 콘텐츠 생성
• Creative Director (CD) - 최종 판단
• Technical Director (TD) - 오케스트레이션

*Features:*
✅ Agent Hub - 에이전트 간 직접 통신
✅ Anti-Gravity - 충돌 방지 메커니즘
✅ Junction Protocol - 자동화 파이프라인
✅ Real-time Dashboard - http://localhost:8000

메시지나 이미지를 보내면 5인 체계가 자율적으로 작동합니다!"""
            self.send_message(chat_id, welcome)

        elif text == "/status":
            status = f"""📊 *System Status*

*Stats:*
• Signals: {self.td.stats['signals_captured']}
• Images: {self.td.stats['images_analyzed']}
• Approved: {self.td.stats['approved']}
• Rejected: {self.td.stats['rejected']}

*Agents (via Hub):*
• SA: ✅ Gemini
• AD: ✅ Gemini Vision
• CE: ✅ Gemini
• CD: {'✅ Claude Haiku' if self.td.cd.using_claude else '⚠️ Gemini'}
• TD: ✅ Active

*Hub Stats:*
• Messages routed: {self.td.hub.stats['messages_routed']}
• Active signals: {len(self.td.active_signals)}

Dashboard: http://localhost:8000"""
            self.send_message(chat_id, status)

    def handle_text(self, message: Dict):
        """텍스트 메시지 처리"""
        text = message.get("text", "")
        if text.startswith("/"):
            self.handle_command(message)
            return

        chat_id = message["chat"]["id"]
        user = message["from"].get("username", message["from"].get("first_name", "User"))

        self.send_message(chat_id, "🔄 5-Agent Hub processing...")

        # TD가 Junction Protocol 파이프라인 실행 (chat_id 전달)
        signal_id = self.td.process_text_signal(text, chat_id)

        response = f"✓ Signal captured: `{signal_id}`\n\nAgents are collaborating via Hub:\nSA → CE → CD → TD"
        self.send_message(chat_id, response)

    def handle_photo(self, message: Dict):
        """이미지 처리 (멀티모달)"""
        chat_id = message["chat"]["id"]
        user = message["from"].get("username", message["from"].get("first_name", "User"))
        caption = message.get("caption", "")

        self.send_message(chat_id, "📷 AD analyzing via Hub...")

        try:
            photo = message["photo"][-1]
            file_id = photo["file_id"]

            file_info_url = f"{BASE_URL}/getFile?file_id={file_id}"
            with urllib.request.urlopen(file_info_url, timeout=10) as resp:
                file_data = json.loads(resp.read().decode('utf-8'))

            if not file_data.get("ok"):
                self.send_message(chat_id, "❌ Failed to get image")
                return

            file_path = file_data["result"]["file_path"]
            file_url = f"https://api.telegram.org/file/bot{BOT_TOKEN}/{file_path}"

            with urllib.request.urlopen(file_url, timeout=20) as resp:
                image_bytes = resp.read()

            # TD가 이미지 파이프라인 실행
            signal_id = self.td.process_image_signal(image_bytes, caption, user)

            response = f"✅ *Image Analysis Complete*\n\nSignal ID: `{signal_id}`\nSaved to knowledge base."
            self.send_message(chat_id, response)

        except Exception as e:
            print(f"[ERROR] Photo processing: {e}")
            self.send_message(chat_id, f"❌ Error: {str(e)[:200]}")

    def run(self):
        """봇 실행"""
        print("=" * 60)
        print("97LAYER OS - 5-AGENT HUB INTEGRATED SYSTEM")
        print("=" * 60)
        print(f"Started: {datetime.now()}")
        print(f"\nAgents:")
        print(f"  • SA (Strategy Analyst): Gemini Pattern Analysis")
        print(f"  • AD (Art Director): Gemini Vision")
        print(f"  • CE (Chief Editor): Gemini Content Generation")
        print(f"  • CD (Creative Director): {'Claude Haiku' if self.td.cd.using_claude else 'Gemini'}")
        print(f"  • TD (Technical Director): Hub Orchestration")
        print(f"\nFeatures:")
        print(f"  ✅ Agent Hub - Direct communication")
        print(f"  ✅ Anti-Gravity - Conflict prevention")
        print(f"  ✅ Junction Protocol - Automated pipeline")
        print(f"  ✅ Dashboard - http://localhost:8000")
        print("\nPress Ctrl+C to stop\n")

        while True:
            try:
                updates = self.get_updates()

                for update in updates:
                    self.offset = update["update_id"] + 1

                    if "message" in update:
                        message = update["message"]

                        if "text" in message:
                            self.handle_text(message)
                        elif "photo" in message:
                            self.handle_photo(message)

                time.sleep(0.5)

            except KeyboardInterrupt:
                print("\n\nShutting down...")
                break
            except Exception as e:
                print(f"[ERROR] Main loop: {e}")
                time.sleep(5)

        print(f"\nFinal Stats:")
        for key, value in self.td.stats.items():
            print(f"  {key}: {value}")
        print(f"Hub messages routed: {self.td.hub.stats['messages_routed']}")


def main():
    """Main entry"""
    # Kill existing
    os.system("pkill -f five_agent 2>/dev/null")
    os.system("pkill -f telegram_daemon 2>/dev/null")
    time.sleep(2)

    # Run 5-agent hub system
    bot = FiveAgentBot()
    bot.run()


if __name__ == "__main__":
    main()
