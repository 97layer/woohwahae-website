#!/usr/bin/env python3
"""
97layerOS YouTube Analyzer
Automatically extracts, analyzes, and stores YouTube content

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import re
import json
from pathlib import Path
from typing import Dict, Optional
from datetime import datetime
import logging

# Load .env file
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

try:
    import google.genai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

try:
    from youtube_transcript_api import YouTubeTranscriptApi
    from youtube_transcript_api._errors import TranscriptsDisabled, NoTranscriptFound
    TRANSCRIPT_AVAILABLE = True
except ImportError:
    TRANSCRIPT_AVAILABLE = False

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class YouTubeAnalyzer:
    """
    Analyzes YouTube videos and extracts insights

    Features:
    - Extract video ID from URL
    - Get transcript (Korean/English)
    - Summarize with Gemini
    - Save to knowledge base
    """

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize YouTube Analyzer

        Args:
            api_key: Gemini API key (or use GOOGLE_API_KEY env var)
        """
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        # Initialize Gemini
        if GEMINI_AVAILABLE and self.api_key:
            self.client = genai.Client(api_key=self.api_key)
            self._model_name = 'gemini-2.5-flash'
            self.model = True  # flag: LLM available
            logger.info("✅ Gemini initialized for YouTube analysis")
        else:
            self.client = None
            self.model = None
            logger.warning("⚠️  Gemini not available")

        # Output directories
        self.signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        self.signals_dir.mkdir(parents=True, exist_ok=True)

    def extract_video_id(self, url: str) -> Optional[str]:
        """
        Extract YouTube video ID from URL

        Supports:
        - https://youtu.be/VIDEO_ID
        - https://www.youtube.com/watch?v=VIDEO_ID
        - https://youtube.com/watch?v=VIDEO_ID
        """
        patterns = [
            r'(?:youtu\.be/)([a-zA-Z0-9_-]{11})',
            r'(?:youtube\.com/watch\?v=)([a-zA-Z0-9_-]{11})',
            r'(?:youtube\.com/embed/)([a-zA-Z0-9_-]{11})'
        ]

        for pattern in patterns:
            match = re.search(pattern, url)
            if match:
                return match.group(1)

        return None

    def get_transcript(self, video_id: str) -> Optional[str]:
        """
        Get transcript for YouTube video

        Tries Korean first, then English
        """
        if not TRANSCRIPT_AVAILABLE:
            logger.warning("youtube_transcript_api not installed")
            return None

        try:
            # Create API instance and fetch transcript
            api = YouTubeTranscriptApi()

            # Try Korean first, then English
            try:
                transcript_data = api.fetch(video_id, languages=['ko'])
                logger.info("✅ Found Korean transcript")
            except:
                try:
                    transcript_data = api.fetch(video_id, languages=['en'])
                    logger.info("✅ Found English transcript")
                except:
                    # Try any available transcript
                    transcript_data = api.fetch(video_id)
                    logger.info("✅ Found automatic transcript")

            # Format transcript
            text = ' '.join([snippet.text for snippet in transcript_data.snippets])
            return text

        except TranscriptsDisabled:
            logger.warning("Transcripts disabled for this video")
            return None
        except NoTranscriptFound:
            logger.warning("No transcript found")
            return None
        except Exception as e:
            logger.error(f"Transcript extraction failed: {e}")
            return None

    def analyze_with_gemini(self, transcript: str, video_url: str) -> Dict:
        """
        Analyze transcript with Gemini

        Returns:
            {
                'summary': str,
                'key_insights': List[str],
                'topics': List[str],
                'actionable_items': List[str]
            }
        """
        if not self.model or not transcript:
            return {
                'summary': '분석 불가 (Gemini API 없음)',
                'key_insights': [],
                'topics': [],
                'actionable_items': []
            }

        prompt = f"""다음은 YouTube 영상의 자막입니다. 97layer 브랜드 철학(본질, 절제, 자기긍정)에 맞춰 분석해주세요.

**영상 URL**: {video_url}

**자막**:
{transcript[:8000]}  # Limit to 8000 chars

**분석 요청**:
1. **요약** (3-4문장으로 핵심만)
2. **핵심 인사이트** (3-5개 bullet points)
3. **주요 토픽** (키워드 5개)
4. **실행 가능한 아이디어** (우리가 적용할 수 있는 것들)

간결하고 명확하게 답변해주세요. 과장하지 말고 본질만 추출하세요.
"""

        try:
            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
            text = response.text

            # Parse response (simple parsing)
            return {
                'summary': text[:500],  # First part as summary
                'full_analysis': text,
                'key_insights': self._extract_bullet_points(text),
                'topics': [],  # Could be enhanced with better parsing
                'actionable_items': []
            }

        except Exception as e:
            logger.error(f"Gemini analysis failed: {e}")
            return {
                'summary': '분석 실패',
                'key_insights': [],
                'topics': [],
                'actionable_items': []
            }

    def _extract_bullet_points(self, text: str) -> list:
        """Extract bullet points from text"""
        lines = text.split('\n')
        bullets = []
        for line in lines:
            line = line.strip()
            if line.startswith('-') or line.startswith('•') or line.startswith('*'):
                bullets.append(line[1:].strip())
        return bullets[:5]  # Top 5

    def save_to_knowledge_base(self, analysis: Dict, video_url: str, transcript: str = None) -> str:
        """
        Save analysis to knowledge/signals/

        Returns: Path to saved file
        """
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        video_id = self.extract_video_id(video_url)
        filename = f'youtube_{video_id}_{timestamp}.json'
        filepath = self.signals_dir / filename

        data = {
            'type': 'youtube_video',
            'source': video_url,
            'video_id': video_id,
            'captured_at': datetime.now().isoformat(),
            'transcript': transcript[:2000] + '...' if transcript and len(transcript) > 2000 else transcript,
            'full_transcript_length': len(transcript) if transcript else 0,
            'analysis': analysis,
            'status': 'captured'  # Will be processed by Multi-Agent later
        }

        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

        logger.info(f"✅ Saved to {filepath}")
        return str(filepath)

    def process_url(self, video_url: str) -> Dict:
        """
        Complete processing pipeline for YouTube URL

        Returns:
            {
                'success': bool,
                'video_id': str,
                'transcript': str,
                'analysis': Dict,
                'saved_path': str,
                'error': Optional[str]
            }
        """
        logger.info(f"Processing YouTube URL: {video_url}")

        # Step 1: Extract video ID
        video_id = self.extract_video_id(video_url)
        if not video_id:
            return {
                'success': False,
                'error': 'Invalid YouTube URL'
            }

        logger.info(f"Video ID: {video_id}")

        # Step 2: Get transcript
        transcript = self.get_transcript(video_id)

        if not transcript:
            # Save without transcript (URL + metadata only)
            logger.warning(f"No transcript available, saving URL only")
            analysis = {
                'summary': 'Transcript unavailable (video from cloud IP blocked by YouTube)',
                'key_insights': [],
                'topics': [],
                'note': 'Manual review needed'
            }
            saved_path = self.save_to_knowledge_base(analysis, video_url, transcript=None)
            return {
                'success': True,  # Changed to True - partial success
                'video_id': video_id,
                'transcript': None,
                'analysis': analysis,
                'saved_path': saved_path,
                'warning': 'Transcript unavailable - saved URL for manual review'
            }

        logger.info(f"Transcript length: {len(transcript)} characters")

        # Step 3: Analyze with Gemini
        analysis = self.analyze_with_gemini(transcript, video_url)

        # Step 4: Save to knowledge base
        saved_path = self.save_to_knowledge_base(analysis, video_url, transcript)

        return {
            'success': True,
            'video_id': video_id,
            'transcript': transcript[:500] + '...',  # Preview only
            'analysis': analysis,
            'saved_path': saved_path
        }


# ========================
# Convenience Functions
# ========================

_youtube_analyzer_instance = None


def get_youtube_analyzer() -> YouTubeAnalyzer:
    """Get YouTubeAnalyzer instance (singleton)"""
    global _youtube_analyzer_instance
    if _youtube_analyzer_instance is None:
        _youtube_analyzer_instance = YouTubeAnalyzer()
    return _youtube_analyzer_instance


# ========================
# CLI Testing
# ========================

def main():
    """CLI test interface"""
    import sys

    # Configure logging
    logging.basicConfig(level=logging.INFO)

    if len(sys.argv) < 2:
        print("Usage: python youtube_analyzer.py <youtube_url>")
        print("Example: python youtube_analyzer.py https://youtu.be/DpD0wnGk03s")
        sys.exit(1)

    url = sys.argv[1]

    analyzer = YouTubeAnalyzer()
    result = analyzer.process_url(url)

    if result['success']:
        print("\n✅ YouTube Analysis Complete!")
        print("=" * 60)
        print(f"Video ID: {result['video_id']}")
        print(f"Transcript Preview: {result['transcript'][:200]}...")
        print(f"\nSummary: {result['analysis']['summary'][:300]}...")
        print(f"\nSaved to: {result['saved_path']}")
    else:
        print(f"\n❌ Failed: {result.get('error', 'Unknown error')}")


if __name__ == "__main__":
    main()
