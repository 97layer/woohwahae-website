#!/usr/bin/env python3
"""
97layerOS Image Analyzer
Automatically analyzes images and extracts insights

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import json
from pathlib import Path
from typing import Dict, Optional
from datetime import datetime
import logging
import base64

# Load .env file
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class ImageAnalyzer:
    """
    Analyzes images and extracts insights using Gemini Vision

    Features:
    - Analyze images with Gemini Vision
    - Extract text, objects, scenes
    - Generate insights
    - Save to knowledge base
    """

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Image Analyzer

        Args:
            api_key: Gemini API key (or use GOOGLE_API_KEY env var)
        """
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        # Initialize Gemini Vision
        if GEMINI_AVAILABLE and self.api_key:
            genai.configure(api_key=self.api_key)
            self.model = genai.GenerativeModel('gemini-1.5-flash')
            logger.info("✅ Gemini Vision initialized")
        else:
            self.model = None
            logger.warning("⚠️  Gemini not available")

        # Output directories
        self.signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
        self.images_dir = self.signals_dir / 'images'
        self.images_dir.mkdir(parents=True, exist_ok=True)

    def analyze_image(self, image_path: str) -> Dict:
        """
        Analyze image with Gemini Vision

        Returns:
            {
                'description': str,
                'objects': List[str],
                'text_content': str,
                'insights': List[str],
                'brand_relevance': str
            }
        """
        if not self.model:
            return {
                'description': '분석 불가 (Gemini API 없음)',
                'objects': [],
                'text_content': '',
                'insights': [],
                'brand_relevance': ''
            }

        try:
            # Read image
            with open(image_path, 'rb') as f:
                image_data = f.read()

            prompt = """이 이미지를 97layer 브랜드 철학(본질, 절제, 자기긍정)에 맞춰 분석해주세요.

**분석 요청**:
1. **이미지 설명** (2-3문장)
2. **주요 객체/요소** (5개 이하)
3. **텍스트 내용** (이미지에 텍스트가 있다면)
4. **핵심 인사이트** (3개 bullet points)
5. **브랜드 연관성** (97layer에 어떻게 적용할 수 있는지)

간결하고 명확하게 답변해주세요.
"""

            # Create image part
            image_part = {
                'mime_type': self._get_mime_type(image_path),
                'data': image_data
            }

            response = self.model.generate_content([prompt, image_part])
            text = response.text

            return {
                'description': text[:500],
                'full_analysis': text,
                'objects': [],
                'text_content': '',
                'insights': self._extract_bullet_points(text),
                'brand_relevance': ''
            }

        except Exception as e:
            logger.error(f"Gemini Vision analysis failed: {e}")
            return {
                'description': f'분석 실패: {str(e)}',
                'objects': [],
                'text_content': '',
                'insights': [],
                'brand_relevance': ''
            }

    def _get_mime_type(self, file_path: str) -> str:
        """Get MIME type from file extension"""
        ext = Path(file_path).suffix.lower()
        mime_types = {
            '.jpg': 'image/jpeg',
            '.jpeg': 'image/jpeg',
            '.png': 'image/png',
            '.gif': 'image/gif',
            '.webp': 'image/webp'
        }
        return mime_types.get(ext, 'image/jpeg')

    def _extract_bullet_points(self, text: str) -> list:
        """Extract bullet points from text"""
        lines = text.split('\n')
        bullets = []
        for line in lines:
            line = line.strip()
            if line.startswith('-') or line.startswith('•') or line.startswith('*'):
                bullets.append(line[1:].strip())
        return bullets[:5]

    def save_image_and_analysis(self, image_path: str, analysis: Dict) -> str:
        """
        Save image and analysis to knowledge/signals/images/

        Returns: Path to saved JSON file
        """
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        image_filename = Path(image_path).name
        base_name = Path(image_path).stem

        # Copy image to signals directory
        import shutil
        new_image_path = self.images_dir / f"{base_name}_{timestamp}{Path(image_path).suffix}"
        shutil.copy(image_path, new_image_path)

        # Save analysis
        json_filename = f"image_{base_name}_{timestamp}.json"
        json_path = self.images_dir / json_filename

        data = {
            'type': 'image',
            'source': str(image_path),
            'saved_image': str(new_image_path),
            'captured_at': datetime.now().isoformat(),
            'analysis': analysis,
            'status': 'captured'
        }

        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

        logger.info(f"✅ Saved image to {new_image_path}")
        logger.info(f"✅ Saved analysis to {json_path}")
        return str(json_path)

    def process_image(self, image_path: str) -> Dict:
        """
        Complete processing pipeline for image

        Returns:
            {
                'success': bool,
                'image_path': str,
                'analysis': Dict,
                'saved_path': str,
                'error': Optional[str]
            }
        """
        logger.info(f"Processing image: {image_path}")

        if not Path(image_path).exists():
            return {
                'success': False,
                'error': f'Image not found: {image_path}'
            }

        # Analyze image
        analysis = self.analyze_image(image_path)

        # Save image and analysis
        saved_path = self.save_image_and_analysis(image_path, analysis)

        return {
            'success': True,
            'image_path': image_path,
            'analysis': analysis,
            'saved_path': saved_path
        }


# ========================
# Convenience Functions
# ========================

_image_analyzer_instance = None


def get_image_analyzer() -> ImageAnalyzer:
    """Get ImageAnalyzer instance (singleton)"""
    global _image_analyzer_instance
    if _image_analyzer_instance is None:
        _image_analyzer_instance = ImageAnalyzer()
    return _image_analyzer_instance


# ========================
# CLI Testing
# ========================

def main():
    """CLI test interface"""
    import sys

    # Configure logging
    logging.basicConfig(level=logging.INFO)

    if len(sys.argv) < 2:
        print("Usage: python image_analyzer.py <image_path>")
        print("Example: python image_analyzer.py ~/Downloads/screenshot.png")
        sys.exit(1)

    image_path = sys.argv[1]

    analyzer = ImageAnalyzer()
    result = analyzer.process_image(image_path)

    if result['success']:
        print("\n✅ Image Analysis Complete!")
        print("=" * 60)
        print(f"Image: {result['image_path']}")
        print(f"\nDescription: {result['analysis']['description'][:300]}...")
        print(f"\nSaved to: {result['saved_path']}")
    else:
        print(f"\n❌ Failed: {result.get('error', 'Unknown error')}")


if __name__ == "__main__":
    main()
