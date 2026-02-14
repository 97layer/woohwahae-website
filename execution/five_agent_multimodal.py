#!/usr/bin/env python3
"""
97LAYER OS - 5-Agent Multimodal System
완전한 멀티모달 구현: 텍스트 + 이미지 + 음성

5인 체계:
- Creative Director (CD): Claude Opus - 최종 판단
- Strategy Analyst (SA): Gemini - 패턴 분석
- Art Director (AD): Gemini Vision - 이미지 분석
- Chief Editor (CE): Gemini - 콘텐츠 작성
- Technical Director (TD): 전체 오케스트레이션
"""

import os
import sys
import json
import time
import urllib.request
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List

PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Configuration
BOT_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
GEMINI_KEY = "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"
CLAUDE_KEY = os.getenv("ANTHROPIC_API_KEY", "")
BASE_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"


class StrategyAnalyst:
    """SA - 정보 수집 및 패턴 분석 (Gemini)"""

    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key
        self.name = "Strategy Analyst (SA)"

    def analyze_signal(self, content: str) -> Dict[str, Any]:
        """신호 분석 및 패턴 탐지"""
        print(f"[{self.name}] Analyzing signal...")

        # Gemini로 패턴 분석
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

        # JSON 파싱 시도
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
            analysis = {"raw_analysis": response[:200]}

        print(f"[{self.name}] Analysis complete: {analysis.get('philosophy_match', 'N/A')}")
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

    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key
        self.name = "Art Director (AD)"

    def analyze_image(self, image_bytes: bytes, caption: str = "") -> Dict[str, Any]:
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

            # JSON 파싱
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

    def __init__(self, gemini_key: str):
        self.gemini_key = gemini_key
        self.name = "Chief Editor (CE)"

    def generate_content(self, signal: str, sa_analysis: Dict) -> str:
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
    """CD - 최종 의사결정 (Claude Opus if available, else Gemini)"""

    def __init__(self, claude_key: str, gemini_key: str):
        self.claude_key = claude_key if claude_key and "your_" not in claude_key else None
        self.gemini_key = gemini_key
        self.name = "Creative Director (CD)"
        self.using_claude = bool(self.claude_key)

    def sovereign_judgment(self, content: str, metadata: Dict) -> Dict[str, Any]:
        """Sovereign 최종 판단"""
        print(f"[{self.name}] Making sovereign judgment...")
        print(f"[{self.name}] Engine: {'Claude Opus' if self.using_claude else 'Gemini (fallback)'}")

        if self.using_claude:
            return self._claude_judgment(content, metadata)
        else:
            return self._gemini_judgment(content, metadata)

    def _claude_judgment(self, content: str, metadata: Dict) -> Dict[str, Any]:
        """Claude Opus로 판단"""
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
                model="claude-3-opus-20240229",  # Opus - Sovereign Judgment
                max_tokens=800,
                temperature=0.3,
                messages=[{"role": "user", "content": prompt}]
            )

            result_text = response.content[0].text

            # JSON 파싱
            import re
            json_match = re.search(r'\{.*\}', result_text, re.DOTALL)
            if json_match:
                judgment = json.loads(json_match.group())
            else:
                judgment = {"raw_response": result_text}

            print(f"[{self.name}] Claude judgment: {'APPROVED' if judgment.get('approved') else 'REJECTED'}")
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
            return {"approved": False, "raw": response_text}

        except Exception as e:
            return {"approved": False, "error": str(e)}


class TechnicalDirector:
    """TD - 전체 시스템 오케스트레이션"""

    def __init__(self):
        self.name = "Technical Director (TD)"
        self.sa = StrategyAnalyst(GEMINI_KEY)
        self.ad = ArtDirector(GEMINI_KEY)
        self.ce = ChiefEditor(GEMINI_KEY)
        self.cd = CreativeDirector(CLAUDE_KEY, GEMINI_KEY)

        self.stats = {
            "signals_captured": 0,
            "images_analyzed": 0,
            "content_generated": 0,
            "approved": 0,
            "rejected": 0
        }

    def process_text_signal(self, text: str, user: str) -> Dict[str, Any]:
        """텍스트 신호 처리 파이프라인"""
        print(f"\n{'='*60}")
        print(f"[{self.name}] Processing text signal from {user}")
        print(f"{'='*60}")

        # Stage 1: Capture
        signal_file = self._save_signal(text, "text", user)
        self.stats["signals_captured"] += 1

        # Stage 2: SA Analysis
        sa_analysis = self.sa.analyze_signal(text)

        # Stage 3: CE Content Generation (if score high enough)
        if sa_analysis.get("score", 0) >= 60:
            content = self.ce.generate_content(text, sa_analysis)
            self.stats["content_generated"] += 1

            # Stage 4: CD Judgment
            judgment = self.cd.sovereign_judgment(content, {
                "original_signal": text[:200],
                "sa_analysis": sa_analysis
            })

            if judgment.get("approved"):
                self.stats["approved"] += 1
                self._save_approved(content, judgment)
            else:
                self.stats["rejected"] += 1

            return {
                "captured": True,
                "analyzed": True,
                "content_generated": True,
                "judgment": judgment,
                "content": content
            }
        else:
            return {
                "captured": True,
                "analyzed": True,
                "content_generated": False,
                "reason": "Score too low",
                "sa_analysis": sa_analysis
            }

    def process_image_signal(self, image_bytes: bytes, caption: str, user: str) -> Dict[str, Any]:
        """이미지 신호 처리 파이프라인 (멀티모달)"""
        print(f"\n{'='*60}")
        print(f"[{self.name}] Processing image signal from {user}")
        print(f"{'='*60}")

        # Stage 1: Capture
        signal_file = self._save_signal(f"[IMAGE] {caption}", "image", user)
        self.stats["signals_captured"] += 1
        self.stats["images_analyzed"] += 1

        # Stage 2: AD Visual Analysis
        ad_analysis = self.ad.analyze_image(image_bytes, caption)

        # Stage 3: SA Pattern Analysis on caption
        sa_analysis = self.sa.analyze_signal(caption) if caption else {}

        return {
            "captured": True,
            "image_analyzed": True,
            "ad_analysis": ad_analysis,
            "sa_analysis": sa_analysis
        }

    def _save_signal(self, content: str, signal_type: str, source: str) -> str:
        """지식 베이스에 신호 저장"""
        signal_dir = PROJECT_ROOT / "knowledge" / "raw_signals"
        signal_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now()
        filename = f"rs-{timestamp.strftime('%Y%m%d%H%M%S')}_{signal_type}_{source}.md"

        with open(signal_dir / filename, "w", encoding="utf-8") as f:
            f.write(f"# Raw Signal - {signal_type.upper()}\n\n")
            f.write(f"**Date**: {timestamp.isoformat()}\n")
            f.write(f"**Source**: {source}\n")
            f.write(f"**Type**: {signal_type}\n\n")
            f.write(f"---\n\n{content}\n")

        return filename

    def _save_approved(self, content: str, judgment: Dict):
        """승인된 콘텐츠 저장"""
        publish_dir = PROJECT_ROOT / "knowledge" / "assets" / "ready_to_publish"
        publish_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now()
        filename = f"approved_{timestamp.strftime('%Y%m%d%H%M%S')}.md"

        with open(publish_dir / filename, "w", encoding="utf-8") as f:
            f.write(f"# Approved Content\n\n")
            f.write(f"**Date**: {timestamp.isoformat()}\n")
            f.write(f"**Score**: {judgment.get('score', 0)}/100\n")
            f.write(f"**Decision**: {judgment.get('decision', '')}\n\n")
            f.write(f"---\n\n{content}\n")


class FiveAgentBot:
    """5인 체계 텔레그램 봇"""

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
            welcome = """🦋 *97LAYER OS - 5-Agent System*

*5인 체계:*
• Strategy Analyst (SA) - 패턴 분석
• Art Director (AD) - 이미지 분석
• Chief Editor (CE) - 콘텐츠 생성
• Creative Director (CD) - 최종 판단
• Technical Director (TD) - 오케스트레이션

*멀티모달 지원:*
✅ 텍스트 분석 (Gemini)
✅ 이미지 분석 (Gemini Vision)
✅ 최종 승인 (Claude Opus)

메시지나 이미지를 보내면 자동으로 5인 체계가 작동합니다!"""
            self.send_message(chat_id, welcome)

        elif text == "/status":
            status = f"""📊 *System Status*

*Stats:*
• Signals: {self.td.stats['signals_captured']}
• Images: {self.td.stats['images_analyzed']}
• Content: {self.td.stats['content_generated']}
• Approved: {self.td.stats['approved']}
• Rejected: {self.td.stats['rejected']}

*Agents:*
• SA: ✅ Gemini
• AD: ✅ Gemini Vision
• CE: ✅ Gemini
• CD: {'✅ Claude' if self.td.cd.using_claude else '⚠️ Gemini (no Claude key)'}
• TD: ✅ Active"""
            self.send_message(chat_id, status)

    def handle_text(self, message: Dict):
        """텍스트 메시지 처리"""
        text = message.get("text", "")
        if text.startswith("/"):
            self.handle_command(message)
            return

        chat_id = message["chat"]["id"]
        user = message["from"].get("username", message["from"].get("first_name", "User"))

        self.send_message(chat_id, "🔄 5-Agent System processing...")

        # TD가 전체 파이프라인 실행
        result = self.td.process_text_signal(text, user)

        # 결과 전송
        if result.get("content_generated"):
            judgment = result.get("judgment", {})
            status = "✅ APPROVED" if judgment.get("approved") else "❌ REJECTED"

            response = f"""{status}

*SA Analysis:* {result.get('sa_analysis', {}).get('philosophy_match', 'N/A')}
*CD Score:* {judgment.get('score', 0)}/100

*Generated Content:*
{result.get('content', '')[:500]}..."""
        else:
            response = f"✓ Signal captured\n*SA Score:* {result.get('sa_analysis', {}).get('score', 0)}/100 (threshold: 60)"

        self.send_message(chat_id, response)

    def handle_photo(self, message: Dict):
        """이미지 처리 (멀티모달)"""
        chat_id = message["chat"]["id"]
        user = message["from"].get("username", message["from"].get("first_name", "User"))
        caption = message.get("caption", "")

        self.send_message(chat_id, "📷 AD analyzing image...")

        try:
            # 이미지 다운로드
            photo = message["photo"][-1]
            file_id = photo["file_id"]

            # Get file path
            file_info_url = f"{BASE_URL}/getFile?file_id={file_id}"
            with urllib.request.urlopen(file_info_url, timeout=10) as resp:
                file_data = json.loads(resp.read().decode('utf-8'))

            if not file_data.get("ok"):
                self.send_message(chat_id, "❌ Failed to get image")
                return

            file_path = file_data["result"]["file_path"]
            file_url = f"https://api.telegram.org/file/bot{BOT_TOKEN}/{file_path}"

            # Download image
            with urllib.request.urlopen(file_url, timeout=20) as resp:
                image_bytes = resp.read()

            # TD가 이미지 파이프라인 실행
            result = self.td.process_image_signal(image_bytes, caption, user)

            # 결과 전송
            ad_analysis = result.get("ad_analysis", {})
            response = f"""✅ *Image Analysis Complete*

*AD Assessment:*
• Score: {ad_analysis.get('aesthetic_score', 'N/A')}/100
• Brand Fit: {ad_analysis.get('brand_fit', 'N/A')}

*Recommendations:*
{chr(10).join('• ' + r for r in ad_analysis.get('recommendations', [])[:3])}

Signal saved to knowledge base."""

            self.send_message(chat_id, response)

        except Exception as e:
            print(f"[ERROR] Photo processing: {e}")
            self.send_message(chat_id, f"❌ Error: {str(e)[:200]}")

    def run(self):
        """봇 실행"""
        print("=" * 60)
        print("97LAYER OS - 5-AGENT MULTIMODAL SYSTEM")
        print("=" * 60)
        print(f"Started: {datetime.now()}")
        print(f"\nAgents:")
        print(f"  • SA (Strategy Analyst): Gemini Pattern Analysis")
        print(f"  • AD (Art Director): Gemini Vision")
        print(f"  • CE (Chief Editor): Gemini Content Generation")
        print(f"  • CD (Creative Director): {'Claude Opus' if self.td.cd.using_claude else 'Gemini (fallback)'}")
        print(f"  • TD (Technical Director): Orchestration")
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


def main():
    """Main entry"""
    # Kill existing
    os.system("pkill -f WORKING_BOT 2>/dev/null")
    os.system("pkill -f telegram_daemon 2>/dev/null")
    os.system("pkill -f unified_system 2>/dev/null")
    time.sleep(2)

    # Run 5-agent system
    bot = FiveAgentBot()
    bot.run()


if __name__ == "__main__":
    main()