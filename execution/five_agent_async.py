#!/usr/bin/env python3
"""
97LAYER OS - 5-Agent Async Parallel System (Phase 2)
진정한 멀티모달 병렬 협업 및 생산성 극대화 엔진

- SA & AD: 비동기 동시 실행 (Parallel Perception)
- CE & CD: 최적화된 시퀀스 처리
- aiohttp 기반 비동기 API 통신
"""

import os
import sys
import json
import asyncio
import aiohttp
import time
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Configuration
BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
GEMINI_KEY = "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
CLAUDE_KEY = os.getenv("ANTHROPIC_API_KEY", "")
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"

# Logging utility for Dashboard
async def report_progress(agent: str, status: str, result: str = ""):
    bridge_file = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"
    try:
        data = {}
        if bridge_file.exists():
            with open(bridge_file, 'r') as f:
                data = json.load(f)
        
        if "activities" not in data:
            data["activities"] = []
            
        data["activities"].append({
            "timestamp": datetime.now().isoformat(),
            "agent": agent,
            "status": status,
            "result": result[:200]
        })
        
        # Keep last 20 activities
        data["activities"] = data["activities"][-20:]
        
        with open(bridge_file, 'w') as f:
            json.dump(data, f, indent=4)
    except:
        pass

class StrategyAnalyst:
    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key
        self.name = "Strategy Analyst (SA)"

    async def analyze_signal(self, session: aiohttp.ClientSession, content: str) -> Dict[str, Any]:
        await report_progress("SA", "Pattern Analysis Started")
        prompt = f"""97layer 철학 분석: {content[:500]}
        JSON 응답: {{"patterns": [], "philosophy_match": "축", "score": 0-100}}"""
        
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key={self.gemini_key}"
        payload = {"contents": [{"parts": [{"text": prompt}]}]}
        
        async with session.post(url, json=payload) as resp:
            data = await resp.json()
            text = data['candidates'][0]['content']['parts'][0]['text']
            
            # Simple extractor
            import re
            json_match = re.search(r'\{.*\}', text, re.DOTALL)
            res = json.loads(json_match.group()) if json_match else {"score": 50}
            
            await report_progress("SA", "Analysis Complete", res.get("philosophy_match", ""))
            return res

class ArtDirector:
    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key
        self.name = "Art Director (AD)"

    async def analyze_image(self, session: aiohttp.ClientSession, image_base64: str, caption: str) -> Dict[str, Any]:
        await report_progress("AD", "Visual Analysis Started")
        prompt = f"""WOOHWAHAE 비주얼 아이덴티티 관점에서 이미지를 분석하세요.
캡션: {caption}

JSON 형식으로 응답하십시오:
{{"aesthetic_score": 0, "brand_fit": "high", "recommendations": ["제안1"]}}"""
        
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key={self.gemini_key}"
        payload = {
            "contents": [{
                "parts": [
                    {"text": prompt},
                    {"inline_data": {"mime_type": "image/jpeg", "data": image_base64}}
                ]
            }]
        }
        
        async with session.post(url, json=payload) as resp:
            data = await resp.json()
            text = data['candidates'][0]['content']['parts'][0]['text']
            
            import re
            json_match = re.search(r'\{.*\}', text, re.DOTALL)
            res = json.loads(json_match.group()) if json_match else {"aesthetic_score": 50}
            
            await report_progress("AD", "Visual Analysis Complete", f"Score: {res.get('aesthetic_score')}")
            return res

class ChiefEditor:
    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key

    async def generate_content(self, session: aiohttp.ClientSession, signal: str, sa_analysis: Dict) -> str:
        await report_progress("CE", "Writing Content")
        prompt = f"97layer 스타일 집필. 신호: {signal}. 분석: {json.dumps(sa_analysis)}"
        
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key={self.gemini_key}"
        payload = {"contents": [{"parts": [{"text": prompt}]}]}
        
        async with session.post(url, json=payload) as resp:
            data = await resp.json()
            res = data['candidates'][0]['content']['parts'][0]['text']
            await report_progress("CE", "Drafting Complete")
            return res

class CreativeDirector:
    def __init__(self, claude_key: str):
        self.claude_key = claude_key

    async def sovereign_judgment(self, session: aiohttp.ClientSession, content: str) -> Dict[str, Any]:
        await report_progress("CD", "Sovereign Judgment Started")
        if not self.claude_key:
            return {"approved": True, "score": 100} # Mock if no key
            
        url = "https://api.anthropic.com/v1/messages"
        headers = {
            "x-api-key": self.claude_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json"
        }
        msg_content = "최종 승인 판단: " + content + ". JSON 응답: {\"approved\": true, \"score\": 90}"
        payload = {
            "model": "claude-3-opus-20240229",
            "max_tokens": 800,
            "messages": [{"role": "user", "content": msg_content}]
        }
        
        async with session.post(url, json=payload, headers=headers) as resp:
            data = await resp.json()
            text = data['content'][0]['text']
            
            import re
            json_match = re.search(r'\{.*\}', text, re.DOTALL)
            res = json.loads(json_match.group()) if json_match else {"approved": True}
            
            await report_progress("CD", "Final Decision Reached", "APPROVED" if res.get("approved") else "REJECTED")
            return res

class TechnicalDirector:
    def __init__(self):
        self.sa = StrategyAnalyst(GEMINI_KEY)
        self.ad = ArtDirector(GEMINI_KEY)
        self.ce = ChiefEditor(GEMINI_KEY)
        self.cd = CreativeDirector(CLAUDE_KEY)

    async def _ask_ai(self, session: aiohttp.ClientSession, prompt: str, system: str = "") -> str:
        """비동기 Gemini 호출 공유 유틸리티"""
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key={GEMINI_KEY}"
        payload = {"contents": [{"parts": [{"text": f"{system}\n\n{prompt}"}]}]}
        async with session.post(url, json=payload) as resp:
            data = await resp.json()
            return data['candidates'][0]['content']['parts'][0]['text']

    async def process_multimodal(self, text: str, image_base64: Optional[str] = None):
        async with aiohttp.ClientSession() as session:
            # Stage 1: Parallel Perception (SA & AD)
            tasks = [self.sa.analyze_signal(session, text)]
            if image_base64:
                tasks.append(self.ad.analyze_image(session, image_base64, text))
            
            print(f"[{datetime.now()}] Parallel Perception Running...")
            results = await asyncio.gather(*tasks)
            sa_res = results[0]
            ad_res = results[1] if image_base64 else None
            
            # Stage 2: Narrative & Judgment
            # 만약 이미지가 있다면 AD의 분석 결과를 CE의 프롬프트에 추가하여 더 풍부한 캡션 생성
            ce_context = text + (f"\n\n[Visual Analysis]: {json.dumps(ad_res)}" if ad_res else "")
            content = await self.ce.generate_content(session, ce_context, sa_res)
            judgment = await self.cd.sovereign_judgment(session, content)
            
            return {
                "sa": sa_res,
                "ad": ad_res,
                "content": content,
                "judgment": judgment
            }

class FiveAgentBot:
    """
    [Deprecated] 텔레그램 직접 수신 기능은 telegram_daemon.py로 통합되었습니다.
    이 클래스는 하위 호환성을 위해 뼈대만 유지하거나 제거될 예정입니다.
    """
    def __init__(self):
        self.td = TechnicalDirector()

async def main():
    print("=" * 60)
    print("97LAYER OS - ASYNC PARALLEL ENGINE (PHASE 2) - STANDBY")
    print("=" * 60)
    
    # 텔레그램 루프를 제거하고 상시 대기 모드 또는 라이브러리 형태로만 유지
    # 외부(telegram_daemon.py)에서 TechnicalDirector를 통해 기능을 사용하도록 권장합니다.
    while True:
        await asyncio.sleep(3600)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass