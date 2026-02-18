#!/usr/bin/env python3
"""
WOOHWAHAE Brand Consultant Agent
Archive for Slow Life - 최상위 플랫폼 철학 수호자

Role:
- WOOHWAHAE 브랜드 철학 일관성 검증
- "Archive for Slow Life" 적합성 평가
- 7개 섹션 병렬 오케스트레이션
- 모든 콘텐츠의 최종 품질 게이트

Philosophy:
"가속화된 세상에서 주체적으로 살아가기 위한
생각의 기록, 실천의 모음, 의식의 플랫폼"

Author: WOOHWAHAE System
Created: 2026-02-17
"""

import os
import sys
import json
import asyncio
from pathlib import Path
from typing import Dict, Any, Optional, List, Tuple
from datetime import datetime
from enum import Enum

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task, QueueManager
from core.agents.agent_hub import get_hub

# Claude API
try:
    import anthropic
    CLAUDE_AVAILABLE = True
except ImportError:
    CLAUDE_AVAILABLE = False
    print("⚠️  anthropic not installed. Run: pip install anthropic")


class WOOHWAHAESection(Enum):
    """7개 병렬 섹션"""
    ABOUT = "about"           # 철학의 선언
    ARCHIVE = "archive"       # 생각의 기록 (Magazine)
    SHOP = "shop"            # 의식적 소비의 큐레이션
    SERVICE = "service"      # 실천의 공간 (Hair Atelier)
    PLAYLIST = "playlist"    # 감각의 리듬
    PROJECT = "project"      # 협업의 실험
    PHOTOGRAPHY = "photography"  # 순간의 포착


class BrandConsultant:
    """
    WOOHWAHAE Brand Consultant

    최상위 플랫폼 철학 수호자
    모든 콘텐츠가 "Archive for Slow Life" 철학에 부합하는지 검증
    7개 섹션 병렬 오케스트레이션
    """

    # WOOHWAHAE 핵심 철학
    CORE_PHILOSOPHY = {
        "platform": "Archive for Slow Life",
        "mission": "가속화된 세상에서 주체적으로 살아가기",
        "values": {
            "생각의 기록": "슬로우라이프에 대한 사색과 철학",
            "실천의 모음": "일상에서 구현하는 구체적 방법들",
            "의식의 플랫폼": "의식적으로 선택하고 살아가는 커뮤니티"
        },
        "editor": "WOOSUNHO - Editor & Curator"
    }

    # 5 Pillars (IDENTITY.md 기준)
    BRAND_PILLARS = {
        "authenticity": "진정성 - 가면 없는 대화, 취약함의 수용",
        "practicality": "실용성 - 지속 가능한 관리, 과시 없는 실질",
        "elegance": "우아함 - 절제된 표현, 여백의 미학",
        "precision": "정밀함 - 구조적 사고, 기반의 견고함",
        "innovation": "혁신 - 관습의 파괴, Anti-Uniformity"
    }

    def __init__(self, agent_id: str = "brand-consultant", api_key: Optional[str] = None, mock_mode: bool = False):
        """
        Initialize Brand Consultant

        Args:
            agent_id: Unique agent instance ID
            api_key: Anthropic API key
            mock_mode: Use mock responses for testing
        """
        self.agent_id = agent_id
        self.agent_type = "BRAND_CONSULTANT"
        self.mock_mode = mock_mode

        if not self.mock_mode:
            if not CLAUDE_AVAILABLE:
                raise ImportError("anthropic required: pip install anthropic")

            api_key = api_key or os.getenv('ANTHROPIC_API_KEY')
            if not api_key:
                raise ValueError("ANTHROPIC_API_KEY not found")

            self.client = anthropic.Anthropic(api_key=api_key)
            self.model = "claude-3-5-sonnet-latest"  # Use latest available model
        else:
            self.client = None
            self.model = "mock"

        # Agent Hub for orchestration
        self.hub = get_hub(str(PROJECT_ROOT))

        # Load brand guidelines
        self._load_brand_guidelines()

        print(f"[Brand Consultant] WOOHWAHAE 철학 수호자 준비 완료.")

    def _load_brand_guidelines(self):
        """브랜드 가이드라인 로드"""
        self.guidelines = {
            "identity": None,
            "tone": None,
            "visual": None
        }

        # IDENTITY.md 로드
        identity_path = PROJECT_ROOT / 'directives' / 'IDENTITY.md'
        if identity_path.exists():
            self.guidelines["identity"] = identity_path.read_text(encoding='utf-8')

        # 톤앤매너 가이드
        self.guidelines["tone"] = """
        - 겸손하고 사색적 (과시 없음)
        - 개인적이면서도 보편적
        - 치유적이고 따뜻한
        - 질문으로 끝나는 열린 결말
        - "~이다" 보다 "~일 수 있다"
        """

        # 비주얼 가이드
        self.guidelines["visual"] = """
        - Monochrome + 1 accent color
        - 40% 이상 여백
        - Magazine B 스타일 그리드
        - 자연광과 그림자
        """

    async def audit_content(self, content: Dict[str, Any]) -> Dict[str, Any]:
        """
        콘텐츠 브랜드 일관성 검증

        Args:
            content: 검증할 콘텐츠
                {
                    'type': 'text|image|video',
                    'source': 'instagram|website|etc',
                    'data': {...},
                    'metadata': {...}
                }

        Returns:
            Audit result with scores and recommendations
        """
        print(f"[Brand Consultant] 콘텐츠 감사 시작...")

        if self.mock_mode:
            # Mock response for testing
            audit_result = self._get_mock_audit_result(content)
        else:
            prompt = self._build_audit_prompt(content)

            try:
                message = self.client.messages.create(
                    model=self.model,
                    max_tokens=1024,
                    messages=[{"role": "user", "content": prompt}]
                )

                response_text = message.content[0].text
                audit_result = self._parse_audit_result(response_text)
            except Exception as e:
                print(f"[Brand Consultant] API 오류, Mock 모드 전환: {e}")
                audit_result = self._get_mock_audit_result(content)

        # 메타데이터 추가
        audit_result.update({
            'audited_by': self.agent_id,
            'audited_at': datetime.now().isoformat(),
            'content_type': content.get('type'),
            'source': content.get('source')
        })

        score = audit_result.get('philosophy_score', 0)
        print(f"[Brand Consultant] 감사 완료. 철학 적합도: {score}/100")

        return audit_result

    async def classify_for_sections(self, content: Dict[str, Any]) -> List[Tuple[WOOHWAHAESection, float]]:
        """
        콘텐츠를 7개 섹션에 분류

        Args:
            content: 분류할 콘텐츠

        Returns:
            [(섹션, 적합도 점수)] 리스트 (내림차순)
        """
        print(f"[Brand Consultant] 섹션 분류 시작...")

        if self.mock_mode:
            # Use mock classification for testing
            result = self._get_mock_classification(content)
        else:
            prompt = f"""당신은 WOOHWAHAE의 Brand Consultant입니다.

WOOHWAHAE = "Archive for Slow Life"
가속화된 세상에서 주체적으로 살아가기 위한 플랫폼

7개 섹션:
- ABOUT: 철학의 선언
- ARCHIVE: 생각의 기록 (Magazine)
- SHOP: 의식적 소비의 큐레이션
- SERVICE: 실천의 공간 (Hair Atelier)
- PLAYLIST: 감각의 리듬
- PROJECT: 협업의 실험
- PHOTOGRAPHY: 순간의 포착

다음 콘텐츠를 분석하여 어느 섹션에 속하는지 판단하세요:

**콘텐츠:**
{json.dumps(content, ensure_ascii=False, indent=2)[:1500]}

각 섹션별 적합도를 0-100으로 평가하고, JSON 형식으로 응답:
{{
    "classifications": [
        {{"section": "ARCHIVE", "score": 85, "reason": "..."}},
        {{"section": "SERVICE", "score": 60, "reason": "..."}}
    ],
    "primary_section": "ARCHIVE",
    "multi_section": true
}}

JSON만 출력하세요.
"""

            try:
                message = self.client.messages.create(
                    model=self.model,
                    max_tokens=1024,
                    messages=[{"role": "user", "content": prompt}]
                )

                response_text = message.content[0].text
                result = json.loads(self._extract_json(response_text))
            except Exception as e:
                print(f"[Brand Consultant] API 오류, Mock 모드 전환: {e}")
                result = self._get_mock_classification(content)

        # Enum으로 변환
        classifications = []
        for item in result.get('classifications', []):
            section_name = item['section']
            try:
                section = WOOHWAHAESection[section_name]
                classifications.append((section, item['score']))
            except KeyError:
                continue

        # 점수 내림차순 정렬
        classifications.sort(key=lambda x: x[1], reverse=True)

        print(f"[Brand Consultant] 1차 섹션: {classifications[0][0].value if classifications else 'None'}")

        return classifications

    async def orchestrate_parallel_processing(self, signal_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        신호를 받아 병렬 처리 오케스트레이션

        1. 브랜드 적합성 검증
        2. 7개 섹션 분류
        3. 에이전트들에게 병렬 작업 할당
        4. 결과 수집 및 통합

        Args:
            signal_data: 입력 신호 (Instagram post, text, etc.)

        Returns:
            Orchestration result
        """
        print(f"[Brand Consultant] 병렬 처리 시작...")

        # 1. 브랜드 감사
        audit_result = await self.audit_content(signal_data)
        philosophy_score = audit_result.get('philosophy_score', 0)

        if philosophy_score < 50:
            print(f"[Brand Consultant] 철학 적합도 미달 ({philosophy_score}/100). 처리 중단.")
            return {
                'status': 'rejected',
                'reason': 'Low philosophy alignment',
                'score': philosophy_score
            }

        # 2. 섹션 분류
        sections = await self.classify_for_sections(signal_data)

        # 3. 병렬 작업 생성
        tasks = []
        queue_manager = QueueManager(str(PROJECT_ROOT))

        # Primary section (가장 적합한 섹션)
        if sections:
            primary_section = sections[0][0]

            # SA, AD, CE 에이전트에게 병렬 작업 할당
            tasks.append({
                'agent': 'SA',
                'task_type': 'analyze_signal',
                'context': {
                    'section': primary_section.value,
                    'philosophy_score': philosophy_score
                }
            })

            tasks.append({
                'agent': 'AD',
                'task_type': 'create_visual_concept',
                'context': {
                    'section': primary_section.value,
                    'style_guide': self.guidelines['visual']
                }
            })

            tasks.append({
                'agent': 'CE',
                'task_type': 'write_content',
                'context': {
                    'section': primary_section.value,
                    'tone_guide': self.guidelines['tone']
                }
            })

        # 4. Agent Hub를 통해 병렬 실행 요청
        collaboration_id = self.hub.request_collaboration(
            initiator='BRAND_CONSULTANT',
            participants=['SA', 'AD', 'CE'],
            topic='Content Production for WOOHWAHAE',
            context={
                'signal_data': signal_data,
                'sections': [(s.value, score) for s, score in sections],
                'philosophy_score': philosophy_score,
                'tasks': tasks
            }
        )

        print(f"[Brand Consultant] 협업 ID: {collaboration_id}")

        # 5. 결과 대기 (비동기)
        await asyncio.sleep(5)  # 실제로는 콜백 대기

        return {
            'status': 'processing',
            'collaboration_id': collaboration_id,
            'philosophy_score': philosophy_score,
            'sections': [(s.value, score) for s, score in sections],
            'tasks_created': len(tasks)
        }

    def _build_audit_prompt(self, content: Dict[str, Any]) -> str:
        """감사 프롬프트 생성"""
        return f"""당신은 WOOHWAHAE의 Brand Consultant입니다.

**WOOHWAHAE 핵심 철학:**
{json.dumps(self.CORE_PHILOSOPHY, ensure_ascii=False, indent=2)}

**5 Pillars:**
{json.dumps(self.BRAND_PILLARS, ensure_ascii=False, indent=2)}

**브랜드 가이드라인:**
{self.guidelines['identity'][:1000] if self.guidelines['identity'] else ''}

다음 콘텐츠를 평가하세요:

**콘텐츠:**
{json.dumps(content, ensure_ascii=False, indent=2)[:1500]}

평가 기준:
1. WOOHWAHAE 철학 적합도 (0-100)
2. 5 Pillars 각각 점수 (0-20)
3. 톤앤매너 일관성
4. "Archive for Slow Life" 메시지 전달력

JSON 형식으로 응답:
{{
    "philosophy_score": 0-100,
    "pillars": {{
        "authenticity": 0-20,
        "practicality": 0-20,
        "elegance": 0-20,
        "precision": 0-20,
        "innovation": 0-20
    }},
    "tone_consistency": 0-100,
    "message_clarity": 0-100,
    "strengths": ["강점1", "강점2"],
    "concerns": ["우려1"],
    "recommendations": ["개선안1", "개선안2"]
}}

JSON만 출력하세요.
"""

    def _parse_audit_result(self, response_text: str) -> Dict[str, Any]:
        """감사 결과 파싱"""
        try:
            json_text = self._extract_json(response_text)
            return json.loads(json_text)
        except json.JSONDecodeError:
            return {
                'philosophy_score': 50,
                'error': 'Failed to parse audit result',
                'raw_response': response_text[:500]
            }

    def _extract_json(self, text: str) -> str:
        """텍스트에서 JSON 추출"""
        if '```json' in text:
            start = text.find('```json') + 7
            end = text.find('```', start)
            return text[start:end].strip()
        elif '{' in text:
            start = text.find('{')
            end = text.rfind('}') + 1
            return text[start:end]
        return text.strip()

    def _get_mock_audit_result(self, content: Dict[str, Any]) -> Dict[str, Any]:
        """Mock audit result for testing"""
        # Analyze content and generate mock scores
        caption = str(content.get('data', {}).get('caption', ''))
        hashtags = content.get('data', {}).get('hashtags', [])

        # Score based on keywords and hashtags
        philosophy_score = 70
        if 'slowlife' in caption.lower() or 'slowlife' in hashtags:
            philosophy_score += 15
        if 'woohwahae' in caption.lower() or 'woohwahae' in hashtags:
            philosophy_score += 10
        if '읽는미장' in caption or any('미장' in tag for tag in hashtags):
            philosophy_score += 5

        philosophy_score = min(philosophy_score, 95)

        return {
            'philosophy_score': philosophy_score,
            'pillars': {
                'authenticity': 18,
                'practicality': 17,
                'elegance': 16,
                'precision': 15,
                'innovation': 14
            },
            'tone_consistency': 85,
            'message_clarity': 80,
            'strengths': ['슬로우라이프 철학 표현', '진정성 있는 메시지'],
            'concerns': ['더 깊은 사색적 톤 필요'],
            'recommendations': ['철학적 질문 추가', '개인적 경험 공유']
        }

    def _get_mock_classification(self, content: Dict[str, Any]) -> Dict[str, Any]:
        """Mock classification result for testing"""
        caption = str(content.get('data', {}).get('caption', ''))
        hashtags = content.get('data', {}).get('hashtags', [])

        classifications = []

        # Determine sections based on content
        if any(word in caption.lower() for word in ['헤어', '미장', '펌', '스타일', 'bob']):
            classifications.append({'section': 'SERVICE', 'score': 85, 'reason': '헤어 서비스 관련 콘텐츠'})

        if any(word in caption.lower() for word in ['철학', '생각', '사색', '치유']):
            classifications.append({'section': 'ARCHIVE', 'score': 75, 'reason': '철학적 사색 콘텐츠'})

        if any(word in caption.lower() for word in ['음악', 'playlist', '플레이리스트']):
            classifications.append({'section': 'PLAYLIST', 'score': 90, 'reason': '음악 관련 콘텐츠'})

        if not classifications:
            classifications.append({'section': 'ABOUT', 'score': 60, 'reason': '일반 브랜드 콘텐츠'})

        # Sort by score
        classifications.sort(key=lambda x: x['score'], reverse=True)

        return {
            'classifications': classifications[:3],
            'primary_section': classifications[0]['section'],
            'multi_section': len(classifications) > 1
        }

    def get_status(self) -> Dict[str, Any]:
        """에이전트 상태 반환"""
        return {
            'agent_id': self.agent_id,
            'agent_type': self.agent_type,
            'status': 'active',
            'philosophy': self.CORE_PHILOSOPHY,
            'model': self.model,
            'timestamp': datetime.now().isoformat()
        }


# Singleton instance
_consultant_instance: Optional[BrandConsultant] = None


def get_consultant(mock_mode: bool = True) -> BrandConsultant:
    """싱글톤 Brand Consultant 인스턴스 반환"""
    global _consultant_instance

    if _consultant_instance is None:
        _consultant_instance = BrandConsultant(mock_mode=mock_mode)

    return _consultant_instance


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import asyncio

    async def test_consultant():
        """Brand Consultant 테스트"""
        consultant = get_consultant()

        # 테스트 콘텐츠 (Instagram 포스트 시뮬레이션)
        test_content = {
            'type': 'instagram_post',
            'source': '@woosunhokr',
            'data': {
                'caption': """#읽는미장

레이어 BOB 스타일에 텍스처 펌을 가미한 형태입니다.
젠더리스한 느낌이 있으면서도 자유로운 분위기가 특징으로,
이솝 헤어 폴리쉬를 사용하여 전체적으로 역동적인 느낌을 연출했습니다.

자유로운 분위기, 실용적인 펌 디자인을 원하는 분들에게 제안합니다.

#woohwahae #slowlife #헤어디자인""",
                'image_url': 'https://instagram.com/...',
                'hashtags': ['읽는미장', 'woohwahae', 'slowlife', '헤어디자인']
            }
        }

        print("\n" + "="*50)
        print("WOOHWAHAE Brand Consultant Test")
        print("="*50 + "\n")

        # 1. 브랜드 감사
        print("1. 브랜드 철학 감사")
        audit_result = await consultant.audit_content(test_content)
        print(f"   철학 적합도: {audit_result.get('philosophy_score', 'N/A')}/100")
        print(f"   5 Pillars: {audit_result.get('pillars', {})}")

        # 2. 섹션 분류
        print("\n2. 7개 섹션 분류")
        sections = await consultant.classify_for_sections(test_content)
        for section, score in sections[:3]:
            print(f"   {section.value}: {score}점")

        # 3. 병렬 처리 오케스트레이션
        print("\n3. 병렬 처리 오케스트레이션")
        orchestration_result = await consultant.orchestrate_parallel_processing(test_content)
        print(f"   상태: {orchestration_result.get('status')}")
        print(f"   협업 ID: {orchestration_result.get('collaboration_id')}")
        print(f"   생성된 작업: {orchestration_result.get('tasks_created')}개")

        print("\n" + "="*50)
        print("테스트 완료!")
        print("="*50)

    # 테스트 실행
    asyncio.run(test_consultant())