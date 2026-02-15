#!/usr/bin/env python3
"""
YouTube Analyzer - Anti-Gravity Protocol Implementation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Extracts YouTube transcripts and generates 3 multi-modal assets:
1. Audio Overview (Podcast-style 2-host dialogue)
2. Visual Deck (Slide presentation with image prompts)
3. Mind Map (Mermaid.js knowledge visualization)

Principles:
- Source Grounding: Transcript saved to knowledge/sources/youtube/
- Multi-modal Synthesis: 3 output formats
- MCP Connector: Integrates with GDrive, NotebookLM, Asset Manager

Container-First:
- Core code: Host (this file)
- Runtime dependencies: Container (youtube-transcript-api, Gemini API)
- Execution: Podman container with mounted knowledge/

Usage:
    python3 execution/system/youtube_analyzer.py --url "https://youtu.be/xxxxx"
    python3 execution/system/youtube_analyzer.py --video-id "xxxxx"
"""

import os
import sys
import json
import re
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Tuple
from urllib.parse import urlparse, parse_qs
import argparse

# Project root setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Import 97layerOS core systems
try:
    from execution.system.asset_manager import AssetManager
    from execution.system.parallel_orchestrator import ParallelOrchestrator
    PHASE1_AVAILABLE = True
except ImportError:
    PHASE1_AVAILABLE = False
    print("⚠️  Phase 1 systems not available. Running in standalone mode.")

# Import AI engine
try:
    from libs.ai_engine import AIEngine
    AI_ENGINE_AVAILABLE = True
except ImportError:
    AI_ENGINE_AVAILABLE = False
    print("⚠️  AIEngine not available. Multi-modal synthesis will be limited.")


class YouTubeAnalyzer:
    """
    Anti-Gravity YouTube Analysis Engine

    Workflow:
    1. Extract transcript → knowledge/sources/youtube/
    2. Generate Audio Overview (Podcast Script)
    3. Generate Visual Deck (Slide Presentation)
    4. Generate Mind Map (Mermaid Diagram)
    5. Register assets → Asset Manager
    6. Sync to Google Drive (optional)
    """

    def __init__(self):
        self.sources_dir = PROJECT_ROOT / 'knowledge' / 'sources' / 'youtube'
        self.audio_dir = PROJECT_ROOT / 'knowledge' / 'assets' / 'audio'
        self.decks_dir = PROJECT_ROOT / 'knowledge' / 'assets' / 'decks'
        self.maps_dir = PROJECT_ROOT / 'knowledge' / 'assets' / 'maps'

        # Create directories
        for dir_path in [self.sources_dir, self.audio_dir, self.decks_dir, self.maps_dir]:
            dir_path.mkdir(parents=True, exist_ok=True)

        # Initialize Phase 1 systems if available
        if PHASE1_AVAILABLE:
            self.asset_manager = AssetManager()
            self.orchestrator = ParallelOrchestrator()

        if AI_ENGINE_AVAILABLE:
            self.ai_engine = AIEngine()

    def extract_video_id(self, url: str) -> Optional[str]:
        """Extract YouTube video ID from URL"""

        # Direct video ID
        if len(url) == 11 and url.isalnum():
            return url

        # Parse URL
        parsed = urlparse(url)

        # youtu.be format
        if parsed.netloc in ['youtu.be', 'www.youtu.be']:
            return parsed.path.lstrip('/')

        # youtube.com format
        if parsed.netloc in ['youtube.com', 'www.youtube.com', 'm.youtube.com']:
            if parsed.path == '/watch':
                query = parse_qs(parsed.query)
                return query.get('v', [None])[0]
            elif parsed.path.startswith('/embed/'):
                return parsed.path.split('/embed/')[1].split('?')[0]
            elif parsed.path.startswith('/v/'):
                return parsed.path.split('/v/')[1].split('?')[0]

        return None

    def fetch_transcript(self, video_id: str) -> Optional[Dict]:
        """
        Fetch YouTube transcript using youtube-transcript-api

        Note: This function requires the Container environment with
        youtube-transcript-api installed. On host, we'll use a fallback.
        """

        try:
            # Attempt container-based extraction
            from youtube_transcript_api import YouTubeTranscriptApi

            transcript_list = YouTubeTranscriptApi.get_transcript(video_id, languages=['ko', 'en'])

            # Format transcript with timestamps
            formatted_transcript = []
            for entry in transcript_list:
                formatted_transcript.append({
                    'timestamp': entry['start'],
                    'duration': entry['duration'],
                    'text': entry['text'],
                    'formatted_time': self._format_timestamp(entry['start'])
                })

            # Get video metadata (title requires additional API call)
            metadata = {
                'video_id': video_id,
                'url': f'https://youtu.be/{video_id}',
                'extracted_at': datetime.now().isoformat(),
                'transcript_language': 'ko' if any('가' <= c <= '힣' for entry in transcript_list for c in entry['text']) else 'en',
                'duration_seconds': sum(entry['duration'] for entry in transcript_list),
                'entry_count': len(transcript_list)
            }

            return {
                'metadata': metadata,
                'transcript': formatted_transcript,
                'raw_text': ' '.join([entry['text'] for entry in formatted_transcript])
            }

        except ImportError:
            print("⚠️  youtube-transcript-api not available. Running in Container mode is required.")
            print("   To install in Container:")
            print("   podman run -it --rm 97layer-os:latest pip install youtube-transcript-api")
            return None

        except Exception as e:
            print(f"❌ Transcript extraction failed: {e}")
            return None

    def _format_timestamp(self, seconds: float) -> str:
        """Format seconds to MM:SS or HH:MM:SS"""
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)

        if hours > 0:
            return f"{hours:02d}:{minutes:02d}:{secs:02d}"
        return f"{minutes:02d}:{secs:02d}"

    def save_transcript(self, video_id: str, transcript_data: Dict) -> Path:
        """Save transcript to knowledge/sources/youtube/"""

        output_path = self.sources_dir / f'{video_id}.json'

        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(transcript_data, f, ensure_ascii=False, indent=2)

        print(f"✅ Transcript saved: {output_path}")
        return output_path

    def generate_audio_overview(self, video_id: str, transcript_data: Dict) -> Path:
        """
        Generate Audio Overview (Podcast Script)

        Format: 2-host conversational dialogue (5-7 minutes)
        Structure: Opening (30s) → Core (4-5min) → Closing (30s)
        """

        metadata = transcript_data['metadata']
        raw_text = transcript_data['raw_text']
        transcript_entries = transcript_data['transcript']

        # Prepare prompt for AI generation
        prompt = f"""당신은 97layerOS의 Chief Editor입니다. Anti-Gravity 프로토콜에 따라 YouTube 영상을 Podcast 형식으로 요약하세요.

**Source**: {metadata['url']}
**Duration**: {metadata['duration_seconds']}초
**Language**: {metadata['transcript_language']}

**Transcript**:
{raw_text[:3000]}  # First 3000 chars for context

---

**Requirements**:

1. **Format**: 2-Host Podcast Dialogue (Host A, Host B)
2. **Duration**: 5-7 minutes (~900-1400 words)
3. **Structure**:
   - Opening (30 seconds): Hook + Context
   - Core Discussion (4-5 minutes): Key insights with timestamps
   - Closing (30 seconds): Summary + Open question

4. **Style**:
   - Conversational, natural flow
   - Aesop benchmark 70%+ (절제된, 본질적)
   - 97layer Brand Voice (thoughtful, authentic, precise)

5. **Citations**:
   - Use [MM:SS] timestamps for all quotes
   - Example: "As mentioned at [03:42], the key insight is..."

6. **Brand Connection**:
   - Link to 2+ of 5 Brand Pillars (Authenticity, Practicality, Elegance, Precision, Innovation)
   - Slow Life philosophy evident

---

**Output Format**:

```markdown
# Audio Script: [Title]

**Duration**: ~5-7 minutes
**Source**: {metadata['url']}

---

## Opening (30 seconds)

**Host A**: [Opening hook]

**Host B**: [Echo + Context]

---

## Core Discussion (4-5 minutes)

**Host A**: [Key Point 1 with [MM:SS] citation]

**Host B**: [Clarification + Example]

**Host A**: [Key Point 2 with [MM:SS] citation]

**Host B**: [Connection to Brand Philosophy]

---

## Closing (30 seconds)

**Host A**: [Summary + Open Question]

**Host B**: [Afterglow - 여운]

---

## Source Citations

- [MM:SS] YouTube: "Title" - Key insight
- [MM:SS] YouTube: "Title" - Supporting detail
```
"""

        # Generate using AI Engine
        if AI_ENGINE_AVAILABLE:
            audio_script = self.ai_engine.generate(
                prompt=prompt,
                model_preference="gemini",  # Gemini for long context
                max_tokens=2000
            )
        else:
            # Fallback: Template-based generation
            audio_script = self._generate_audio_template(video_id, transcript_data)

        # Save to knowledge/assets/audio/
        output_path = self.audio_dir / f'{video_id}_overview.md'
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(audio_script)

        print(f"✅ Audio Overview generated: {output_path}")
        return output_path

    def generate_visual_deck(self, video_id: str, transcript_data: Dict) -> Path:
        """
        Generate Visual Deck (Slide Presentation)

        Format: 8-12 slides with image prompts
        Style: Minimalist, 60%+ whitespace, monochrome
        """

        metadata = transcript_data['metadata']
        raw_text = transcript_data['raw_text']

        prompt = f"""당신은 97layerOS의 Art Director입니다. Anti-Gravity 프로토콜에 따라 Visual Deck을 생성하세요.

**Source**: {metadata['url']}
**Transcript**: {raw_text[:2000]}

---

**Requirements**:

1. **Total Slides**: 8-12
2. **Style**: Minimalist, 60%+ whitespace, monochrome
3. **Each Slide**:
   - Visual Prompt (Midjourney/DALL-E style)
   - Text: 1-3 bullet points maximum
   - Source citation with timestamp

4. **Visual Identity**:
   - Natural light
   - Monochrome palette
   - Noto Serif CJK KR + Crimson Text typography
   - Clean, elegant compositions

---

**Output Format**:

```markdown
# Visual Deck: [Title]

**Total Slides**: 10
**Style**: Minimalist, Monochrome
**Source**: {metadata['url']}

---

## Slide 1: Cover

**Visual Prompt**: "Minimalist abstract composition, white space, natural light, monochrome photography style --ar 16:9 --v 6"

**Text**:
- Title: [Main Title]
- Subtitle: [One-line essence]

---

## Slide 2: Hook

**Visual Prompt**: "[Specific concept visual]"

**Text**:
- Key Quote: "[Opening statement]"
- Source: [[MM:SS]]

[... Continue for all slides ...]
```
"""

        if AI_ENGINE_AVAILABLE:
            visual_deck = self.ai_engine.generate(
                prompt=prompt,
                model_preference="gemini",
                max_tokens=2000
            )
        else:
            visual_deck = self._generate_deck_template(video_id, transcript_data)

        output_path = self.decks_dir / f'{video_id}_deck.md'
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(visual_deck)

        print(f"✅ Visual Deck generated: {output_path}")
        return output_path

    def generate_mind_map(self, video_id: str, transcript_data: Dict) -> Path:
        """
        Generate Mind Map (Mermaid Diagram)

        Format: Radial or Hierarchical Mermaid.js
        Structure: Core concept → Branches → Citations → Brand connections
        """

        metadata = transcript_data['metadata']
        raw_text = transcript_data['raw_text']

        prompt = f"""당신은 97layerOS의 Strategy Analyst입니다. Anti-Gravity 프로토콜에 따라 Mind Map을 생성하세요.

**Source**: {metadata['url']}
**Transcript**: {raw_text[:2000]}

---

**Requirements**:

1. **Format**: Mermaid.js diagram (graph TD or graph LR)
2. **Structure**:
   - Core Concept (중앙)
   - 3-5 Main Branches
   - Sub-concepts with citations
   - Brand Pillars connection nodes

3. **Validation**:
   - Must be valid Mermaid.js syntax
   - Include source citations as nodes
   - Highlight brand connections with distinct styling

---

**Output Format**:

```markdown
# Mind Map: [Title]

**Structure**: Radial
**Source**: {metadata['url']}

```mermaid
graph TD
    A[Core Concept] --> B[Branch 1: Key Idea]
    A --> C[Branch 2: Key Idea]
    A --> D[Branch 3: Key Idea]

    B --> B1[Sub-concept 1]
    B --> B2[Sub-concept 2]
    B1 --> B1a["Source: [[MM:SS]]"]

    C --> C1[Sub-concept 1]
    C --> C2[Sub-concept 2]
    C2 --> C2a["Source: [[MM:SS]]"]

    D --> D1[Sub-concept 1]
    D --> D2[Brand Connection]
    D2 --> D2a["5 Pillars: Authenticity + Precision"]

    style A fill:#f9f9f9,stroke:#333,stroke-width:2px
    style D2 fill:#e8f4f8,stroke:#0066cc,stroke-width:2px
```
```
"""

        if AI_ENGINE_AVAILABLE:
            mind_map = self.ai_engine.generate(
                prompt=prompt,
                model_preference="claude",  # Claude better for structured output
                max_tokens=1500
            )
        else:
            mind_map = self._generate_map_template(video_id, transcript_data)

        output_path = self.maps_dir / f'{video_id}_map.md'
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(mind_map)

        print(f"✅ Mind Map generated: {output_path}")
        return output_path

    def _generate_audio_template(self, video_id: str, transcript_data: Dict) -> str:
        """Fallback template for audio overview"""
        metadata = transcript_data['metadata']

        return f"""# Audio Script: YouTube Analysis

**Duration**: ~5-7 minutes
**Source**: {metadata['url']}

---

## Opening (30 seconds)

**Host A**: Today we're exploring insights from a fascinating video.

**Host B**: Let's dive into the key takeaways.

---

## Core Discussion (4-5 minutes)

**Host A**: The video presents several interesting perspectives...

[Note: Full AI-generated content requires AIEngine integration]

---

## Closing (30 seconds)

**Host A**: These insights connect well with our philosophy of thoughtful living.

**Host B**: What resonates with you?

---

## Source Citations

- [{metadata['url']}]
"""

    def _generate_deck_template(self, video_id: str, transcript_data: Dict) -> str:
        """Fallback template for visual deck"""
        metadata = transcript_data['metadata']

        return f"""# Visual Deck: YouTube Analysis

**Total Slides**: 10
**Style**: Minimalist, Monochrome
**Source**: {metadata['url']}

---

## Slide 1: Cover

**Visual Prompt**: "Minimalist abstract composition, white space, natural light --ar 16:9"

**Text**:
- Title: Video Insights
- Source: YouTube

[Note: Full AI-generated content requires AIEngine integration]
"""

    def _generate_map_template(self, video_id: str, transcript_data: Dict) -> str:
        """Fallback template for mind map"""
        metadata = transcript_data['metadata']

        return f"""# Mind Map: YouTube Analysis

**Structure**: Radial
**Source**: {metadata['url']}

```mermaid
graph TD
    A[Core Concept] --> B[Key Idea 1]
    A --> C[Key Idea 2]
    A --> D[Key Idea 3]

    B --> B1["Source: {metadata['url']}"]
    C --> C1["Source: {metadata['url']}"]
    D --> D1["Brand Connection"]

    style A fill:#f9f9f9,stroke:#333,stroke-width:2px
    style D1 fill:#e8f4f8,stroke:#0066cc,stroke-width:2px
```

[Note: Full AI-generated content requires AIEngine integration]
"""

    def analyze(self, url: str) -> Dict[str, Path]:
        """
        Complete Anti-Gravity Analysis Workflow

        Returns:
            Dict with paths to generated assets:
            {
                'source': Path to transcript JSON,
                'audio': Path to audio overview,
                'deck': Path to visual deck,
                'map': Path to mind map
            }
        """

        print("=" * 70)
        print("🛸 Anti-Gravity YouTube Analyzer")
        print("=" * 70)

        # Step 1: Extract video ID
        video_id = self.extract_video_id(url)
        if not video_id:
            print(f"❌ Invalid YouTube URL: {url}")
            return {}

        print(f"📹 Video ID: {video_id}")
        print(f"🔗 URL: https://youtu.be/{video_id}")
        print()

        # Step 2: Fetch transcript (Source Grounding)
        print("⏳ [1/5] Extracting transcript...")
        transcript_data = self.fetch_transcript(video_id)

        if not transcript_data:
            print("❌ Transcript extraction failed. Cannot proceed.")
            return {}

        source_path = self.save_transcript(video_id, transcript_data)
        print()

        # Step 3: Generate Audio Overview
        print("⏳ [2/5] Generating Audio Overview (Podcast Script)...")
        audio_path = self.generate_audio_overview(video_id, transcript_data)
        print()

        # Step 4: Generate Visual Deck
        print("⏳ [3/5] Generating Visual Deck (Slide Presentation)...")
        deck_path = self.generate_visual_deck(video_id, transcript_data)
        print()

        # Step 5: Generate Mind Map
        print("⏳ [4/5] Generating Mind Map (Mermaid Diagram)...")
        map_path = self.generate_mind_map(video_id, transcript_data)
        print()

        # Step 6: Register assets (if Phase 1 available)
        if PHASE1_AVAILABLE:
            print("⏳ [5/5] Registering assets...")

            self.asset_manager.register_asset(
                path=str(audio_path),
                asset_type='audio_overview',
                source=f'youtube:{video_id}'
            )

            self.asset_manager.register_asset(
                path=str(deck_path),
                asset_type='visual_deck',
                source=f'youtube:{video_id}'
            )

            self.asset_manager.register_asset(
                path=str(map_path),
                asset_type='mind_map',
                source=f'youtube:{video_id}'
            )

            print("✅ Assets registered")
        else:
            print("⏳ [5/5] Asset registration skipped (Phase 1 not available)")

        print()
        print("=" * 70)
        print("✅ Anti-Gravity Analysis Complete")
        print("=" * 70)
        print(f"📁 Source: {source_path}")
        print(f"🎙️  Audio: {audio_path}")
        print(f"🎨 Deck: {deck_path}")
        print(f"🗺️  Map: {map_path}")
        print()

        return {
            'source': source_path,
            'audio': audio_path,
            'deck': deck_path,
            'map': map_path
        }


def main():
    """CLI entry point"""
    parser = argparse.ArgumentParser(
        description="YouTube Analyzer - Anti-Gravity Protocol",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Analyze YouTube URL
    python3 execution/system/youtube_analyzer.py --url "https://youtu.be/xxxxx"

    # Analyze by video ID
    python3 execution/system/youtube_analyzer.py --video-id "xxxxx"

    # Test with container
    podman run -v $(pwd)/knowledge:/app/knowledge \\
        97layer-os:latest \\
        python3 execution/system/youtube_analyzer.py --url "https://youtu.be/xxxxx"
        """
    )

    parser.add_argument(
        '--url',
        type=str,
        help='YouTube video URL (e.g., https://youtu.be/xxxxx)'
    )

    parser.add_argument(
        '--video-id',
        type=str,
        help='YouTube video ID (e.g., xxxxx)'
    )

    args = parser.parse_args()

    if not args.url and not args.video_id:
        parser.print_help()
        sys.exit(1)

    # Determine URL
    url = args.url if args.url else f'https://youtu.be/{args.video_id}'

    # Run analysis
    analyzer = YouTubeAnalyzer()
    results = analyzer.analyze(url)

    if results:
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == '__main__':
    main()
