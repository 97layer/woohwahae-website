#!/usr/bin/env python3
"""
LAYER OS Chief Editor (CE) Agent
Phase 6.3: NotebookLM 브랜드 가이드 쿼리 연동

Role:
- Content synthesis and editorial direction
- Transform insights + visuals into cohesive narratives
- Brand voice consistency — NotebookLM RAG 기반 실시간 참조
- Final content output (copy, captions, articles)

LLM: Gemini 2.5 Pro (Free tier)
Brand Reference: NotebookLM MCP (쿠키 인증 필요, 없으면 fallback)
Queue: Autonomous task claiming via AgentWatcher
Output: Finalized content ready for CD approval

Author: LAYER OS Technical Director
Updated: 2026-02-16 (Phase 6.3 — NotebookLM 브랜드 연동)
"""

import os
import sys
import json
import logging
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task
from core.system.agent_logger import get_logger

try:
    import google.genai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)

# 브랜드 보이스 로딩 — directive_loader 섹션 단위 추출
def _load_brand_directives() -> str:
    """CE 에이전트 컨텍스트 로드 — 어조/금칙/에세이 규격"""
    try:
        from core.system.directive_loader import load_for_agent
        return load_for_agent("CE", max_total=5000)
    except ImportError:
        pass
    # 최소 fallback
    return (
        "WOOHWAHAE 브랜드 보이스:\n"
        "- 톤: 사색적, 절제된, 밀도 있는\n"
        "- 금지어: 대박, 꿀팁, 핫, 트렌디, 과장 형용사\n"
        "- 허용어: 본질, 기록, 순간, 흔적, 절제, 고요함\n"
        "- 어조: Archive=한다체, Magazine=합니다체"
    )


class ChiefEditor:
    """
    Chief Editor Agent - Content Synthesis & Editorial Direction

    Capabilities:
    - Synthesize SA analysis + AD visuals into cohesive content
    - Write copy in 97layer brand voice (NotebookLM RAG 참조)
    - Create social media captions, articles, newsletters
    - Editorial QA before CD review
    """

    def __init__(self, agent_id: str = "ce-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "CE"
        self._brand_voice_cache: Optional[str] = None

        # AgentLogger 초기화
        self.logger = get_logger("ce", PROJECT_ROOT)
        self.logger.idle("CE Agent 시작")

        if not GEMINI_AVAILABLE:
            raise ImportError("google-generativeai required")

        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found")

        self.client = genai.Client(api_key=api_key)
        self._model_name = 'gemini-2.5-pro'

        # NotebookLM 브릿지 (선택적 — 없어도 동작)
        self.nlm = None
        try:
            from core.system.notebooklm_bridge import get_bridge, is_available
            if is_available():
                self.nlm = get_bridge()
                logger.info("%s: NotebookLM 브랜드 RAG 연결됨", self.agent_id)
            else:
                logger.warning("%s: NotebookLM 미연결 -- fallback 브랜드 보이스 사용", self.agent_id)
        except Exception as e:
            logger.warning("NotebookLM 초기화 실패: %s", e)

        logger.info("CE: 준비됨")

    def _get_brand_voice(self) -> str:
        """
        NotebookLM에서 97layer 브랜드 보이스 참조 가져오기.
        세션 내 첫 호출 시만 쿼리, 이후 캐시 사용.
        NotebookLM 연결 불가 시 fallback 반환.
        """
        if self._brand_voice_cache:
            return self._brand_voice_cache

        if self.nlm:
            try:
                logger.info("%s: NotebookLM 브랜드 보이스 쿼리 중...", self.agent_id)
                result = self.nlm.query_knowledge_base(
                    "97layer 브랜드 보이스와 WOOHWAHAE 톤앤매너 가이드. "
                    "금지어, 허용어, 문장 스타일, 인칭, 시제 규칙을 요약해줘."
                )
                if result and len(result) > 50:
                    self._brand_voice_cache = result
                    logger.info("%s: NotebookLM 브랜드 보이스 캐시 완료 (%d자)", self.agent_id, len(result))
                    return self._brand_voice_cache
            except Exception as e:
                logger.warning("%s: NotebookLM 쿼리 실패, fallback 사용: %s", self.agent_id, e)

        self._brand_voice_cache = _load_brand_directives()
        return self._brand_voice_cache

    def write_content(self, analysis: Dict[str, Any], visual_concept: Dict[str, Any],
                      retry_count: int = 0, feedback: str = "", previous_output: Dict = None) -> Dict[str, Any]:
        """
        SA 분석 + AD 비주얼 컨셉을 기반으로 콘텐츠 작성.
        인스타그램 패키지 + 아카이브 에세이 이중 포맷 생성.

        Args:
            analysis: SA strategic analysis
            visual_concept: AD visual concept
            retry_count: 재작업 횟수 (CD 거절 또는 Ralph 점수 미달)
            feedback: 이전 거절 피드백 (재작업 시)
            previous_output: 이전 결과물 (재작업 참고용)

        Returns:
            {
              instagram_caption, hashtags, archive_essay,
              headline, tone, ...
            }
        """
        signal_id = analysis.get('signal_id', 'unknown')
        retry_msg = " (재작업 %d회차)" % retry_count if retry_count > 0 else ""
        logger.info("CE: %s 초안 작업.%s", signal_id, retry_msg)

        # 브랜드 보이스 참조 (NotebookLM 또는 fallback)
        brand_voice = self._get_brand_voice()
        brand_source = "NotebookLM RAG" if self.nlm else "Brand OS directives"
        logger.info("%s: 브랜드 보이스 출처 — %s", self.agent_id, brand_source)

        # 재작업 컨텍스트
        retry_context = ""
        if retry_count > 0 and feedback:
            retry_context = f"""
**이전 피드백 (반드시 반영):**
{feedback}

"""
        if previous_output:
            retry_context += f"""**이전 출력 (개선 필요):**
- 인스타 캡션: {previous_output.get('instagram_caption', 'N/A')[:100]}
- 에세이 일부: {str(previous_output.get('archive_essay', 'N/A'))[:200]}

"""

        prompt = f"""당신은 Chief Editor(CE)다.
WOOHWAHAE의 SAGE-ARCHITECT 인격으로 콘텐츠를 작성한다.

**핵심 특성 (sage_architect.md)**:
- 관찰자 거리: 가르치지 않고 렌즈만 건넨다
- 열린 결말: 결론을 대신 내리지 않는다
- 본질주의: 과잉을 덜어낸다
- 수평적 동료: "나도 이런 고민을 했다" 구조
- 조용한 밀도: 목소리를 낮추되 발음은 정확하다
- 자기 리듬: 자기 경험 기반

**어미 스펙트럼**: 핵심 철학은 단언("~이다"). 관찰/해석은 열린("~라고 본다", "~일 수 있다"). 정보는 간결("~입니다").

**전략 분석 (SA 제공):**
- 주제: {', '.join(analysis.get('themes', []))}
- 핵심 인사이트: {'; '.join(analysis.get('key_insights', []))}
- 요약: {analysis.get('summary', '')}

**비주얼 컨셉 (AD 제공):**
- 제목: {visual_concept.get('concept_title', '(없음)')}
- 무드: {visual_concept.get('visual_mood', '(없음)')}
- 브랜드 정렬: {visual_concept.get('brand_alignment', '(없음)')}

**브랜드 보이스 가이드:**
{brand_voice}

{retry_context}
위 가이드를 따라 **두 가지 포맷**으로 작성하라:

1. **Instagram 패키지**: 발행 즉시 사용 가능한 형태
2. **Archive Essay**: Hook-Story-Core-Echo 구조의 에세이

아래 JSON 형식으로 반환:
{{
  "instagram_caption": "인스타그램 캡션 (한국어, 150자 이내, 이모지 금지)",
  "hashtags": "#woohwahae #slowlife #아카이브 (관련 해시태그 5-8개)",
  "archive_essay": "아카이브 에세이 (한국어, 500-800자, 한다체, Hook-Story-Core-Echo)",
  "headline": "헤드라인 (10-20자, 명사형)",
  "tone": "contemplative|reflective|grounded 중 하나",
  "brand_voice_source": "{brand_source}"
}}

필수 준수:
- instagram_caption: 150자 이내. 이모지 금지. 메타언급 금지.
- archive_essay: 500자 이상. Echo는 질문/여백으로 맺는다.
- 금지: 느낌표, 이모지, 교훈적 결론, 자기과시, 메타언급
- 금지어: 혁신, 트렌드, 최신, 혁명적, 압도적

JSON만 반환.
"""

        try:
            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
            content_text = response.text

            # JSON 파싱
            if '```json' in content_text:
                json_start = content_text.find('```json') + 7
                json_end = content_text.find('```', json_start)
                json_text = content_text[json_start:json_end].strip()
            elif '```' in content_text:
                json_start = content_text.find('```') + 3
                json_end = content_text.find('```', json_start)
                json_text = content_text[json_start:json_end].strip()
            else:
                json_text = content_text.strip()

            content = json.loads(json_text)

            content.update({
                'signal_id': signal_id,
                'written_by': self.agent_id,
                'written_at': datetime.now().isoformat(),
                'model': self._model_name,
                'brand_voice_source': brand_source,
                'retry_count': retry_count,
                'status': 'draft_for_cd',
            })

            caption_len = len(content.get('instagram_caption', ''))
            essay_len = len(content.get('archive_essay', ''))
            logger.info("CE: 초안 완료. 캡션 %d자, 에세이 %d자.", caption_len, essay_len)
            return content

        except Exception as e:
            logger.error("%s: 콘텐츠 생성 실패: %s", self.agent_id, e)
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload

        logger.info("CE: %s (%s)", task.task_id, task_type)

        if task_type == 'write_content':
            # Orchestrator에서 오는 새 payload 구조 지원
            # payload에 sa_result가 있으면 Orchestrator 경유
            sa_result = payload.get('sa_result', payload.get('analysis', {}))
            visual = payload.get('visual_concept', payload.get('ad_result', {}).get('visual_concept', {}))

            # 재작업 파라미터
            retry_count = payload.get('retry_count', 0)
            feedback = payload.get('feedback', payload.get('cd_feedback', ''))
            previous_output = payload.get('previous_output', None)

            result = self.write_content(
                analysis=sa_result,
                visual_concept=visual,
                retry_count=retry_count,
                feedback=feedback,
                previous_output=previous_output
            )
            # SA 전략 점수를 result에 포함 (Ralph 채점용)
            result['sa_strategic_score'] = sa_result.get('strategic_score', 0)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}

        elif task_type == 'write_corpus_essay':
            # Gardener가 트리거한 corpus 기반 에세이 작성
            # 단일 신호가 아닌 군집 전체 RAG → Magazine B 스타일 롱폼
            result = self._write_corpus_essay(payload)

            # signal_id 정규화 (Orchestrator가 후속 체인에 사용)
            if 'signal_id' not in result:
                result['signal_id'] = payload.get('theme', 'corpus').replace(' ', '_')

            # 발행은 Orchestrator가 CE→Ralph→AD→CD→Publisher 체인으로 처리
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}

        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def _write_corpus_essay(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """
        Corpus 군집 기반 원소스 멀티유즈 콘텐츠 생성.

        하나의 Gemini 호출로 5개 포맷 동시 생성:
          1. archive_essay     — 롱폼 에세이 (800~1200자) → woohwahae.kr/archive/
          2. instagram_caption — 캡션 (150자 이내) → 인스타그램 단일 피드
          3. carousel_slides   — 3~5장 텍스트 슬라이드 → 인스타그램 캐러셀
          4. telegram_summary  — 3줄 요약 → 봇 푸시 알림
          5. pull_quote        — 핵심 1문장 → 웹사이트 히어로/about 인용구

        원칙: 포맷이 다를 뿐 동일한 본질에서 파생. 재가공 아닌 파생.
        """
        theme = payload.get("theme", "")
        rag_context = payload.get("rag_context", [])
        entry_count = payload.get("entry_count", 0)
        instruction = payload.get("instruction", "")
        content_type = payload.get("content_type", "archive")
        content_category = payload.get("content_category", "")

        # AgentLogger: 에세이 작성 시작
        self.logger.think("에세이 작성 중: %s" % theme)

        # essay: archive or essay type → 한다체, 독백, 300-800자
        # journal: magazine or journal type → 합니다체, ~~~하는 법, 1200-3000자
        is_journal = content_type in ("magazine", "journal")

        if is_journal:
            tone_guide = (
                "문체: 합니다체. 형식: ~~~하는 법 / 안내형. "
                "구조: 리드(도입 요약) → 본문(단계별 설명) → 실천(온화한 제안). "
                "길이: 1200-3000자. "
                "마무리: 독자가 실천 가능한 제안. 금지: 자기중심 관찰, 모호한 결론."
            )
            essay_length_spec = "1200~3000자. 리드+본문+실천 구조."
        else:
            tone_guide = (
                "문체: 한다체. 형식: 독백, 관찰자 시점. "
                "구조: Hook(역설/선언) → Story(개인→보편) → Core(통찰) → Echo(Hook 변주). "
                "길이: 300-800자. "
                "마무리: 결론 없이 여운. 열린 질문 또는 여백. 금지: 강한 호소, 행동 유도."
            )
            essay_length_spec = "300~800자. Hook-Story-Core-Echo 구조."

        # RAG 컨텍스트 직렬화
        context_text = ""
        for i, entry in enumerate(rag_context, 1):
            context_text += f"\n[{i}] {entry.get('captured_at', '')[:10]} | {entry.get('signal_type', '')}\n"
            context_text += f"요약: {entry.get('summary', '')}\n"
            insights = entry.get('key_insights', [])
            if insights:
                context_text += f"인사이트: {' / '.join(str(x) for x in insights[:3])}\n"

        category_hint = f"\n카테고리: {content_category}" if content_category else ""

        prompt = f"""너는 WOOHWAHAE의 편집장이다. SAGE-ARCHITECT 인격으로 쓴다.

주제: {theme}{category_hint}
타입: {"Journal (합니다체, 안내형)" if is_journal else "Essay (한다체, 독백형)"}
신호 수: {entry_count}개

**핵심 특성**: 관찰자 거리(가르치지 않음) / 열린 결말(독자 사유 공간) / 본질주의(하나의 생각) / 수평적(교훈 금지) / 조용한 밀도 / 자기 리듬(경험 기반)
**어미 스펙트럼**: 핵심 철학은 단언. 관찰/해석은 열린("~라고 본다", "~일 수 있다"). 정보는 간결.

아래는 이 주제와 관련해 시간을 두고 쌓인 신호들의 요약이다:
{context_text}

이 신호들의 흐름에서 본질을 읽어내고, 아래 5개 포맷을 동시에 만들어라.
모두 같은 본질에서 파생된다. 재가공이 아닌 파생이다.

공통 규칙:
- 한국어
- 이모지 완전 금지, 느낌표 금지
- 볼드, 헤더 사용 금지
- 메타언급 금지 ("이 글에서는", "지금부터")
- 교훈적 결론 금지 ("~해야 한다", "~하자")
- WOOHWAHAE 톤: {tone_guide}

응답 형식 (JSON):
{{
  "essay_title": "제목 (10자 이내, 명사형)",
  "pull_quote": "이 글 전체를 관통하는 핵심 문장 1개 (30자 이내). 웹사이트 히어로에 써도 될 만큼 밀도 있게.",
  "archive_essay": "에세이 본문. {essay_length_spec} 단락 사이 빈 줄.",
  "instagram_caption": "인스타그램 캡션. 150자 이내. 에세이의 핵심을 압축. 마지막 줄은 여백을 주는 한 문장.",
  "carousel_slides": [
    "슬라이드 1: 도입 문장 (30자 이내)",
    "슬라이드 2: 핵심 관찰 (30자 이내)",
    "슬라이드 3: 맥락 또는 역설 (30자 이내)",
    "슬라이드 4: 마무리 또는 질문 (30자 이내)"
  ],
  "telegram_summary": "봇 푸시 알림용 3줄 요약. 각 줄 40자 이내. 첫 줄: 제목. 둘째 줄: 핵심. 셋째 줄: 링크 유도.",
  "theme": "{theme}",
  "content_category": "{content_category}",
  "content_type": "{"journal" if is_journal else "essay"}",
  "entry_count": {entry_count}
}}

JSON만 출력."""

        try:
            import google.genai as genai
            import os, re
            client = genai.Client(api_key=os.getenv('GOOGLE_API_KEY'))
            response = client.models.generate_content(
                model='gemini-2.5-pro',
                contents=[prompt]
            )
            text = response.text.strip()
            match = re.search(r'\{.*\}', text, re.DOTALL)
            if match:
                result = json.loads(match.group())
                formats = [k for k in ['archive_essay', 'instagram_caption', 'carousel_slides',
                                        'telegram_summary', 'pull_quote'] if k in result]
                logger.info("CE: 원소스 멀티유즈 완료 -- %s | 타입: %s | 포맷: %s", theme, content_type, ', '.join(formats))
            else:
                result = {
                    "archive_essay": text,
                    "essay_title": theme,
                    "theme": theme,
                    "entry_count": entry_count
                }

            # ── HTML 저장 ──────────────────────────────────────────
            try:
                self._save_essay_html(result, theme)
            except Exception as html_e:
                # HTML 저장 실패는 에세이 생성 결과에 영향 없음
                logger.warning("CE: HTML 저장 실패 (무시) -- %s", html_e)

            # ── NotebookLM Essay Archive 저장 ──────────────────────
            if self.nlm:
                try:
                    self.nlm.add_essay_to_archive({
                        'essay_title': result.get('essay_title', theme),
                        'theme': theme,
                        'archive_essay': result.get('archive_essay', ''),
                        'pull_quote': result.get('pull_quote', ''),
                        'instagram_caption': result.get('instagram_caption', ''),
                        'issue_num': result.get('issue_num', ''),
                    })
                    logger.info("CE: NotebookLM Essay Archive 저장 완료 -- %s", result.get('essay_title', theme))
                except Exception as nlm_e:
                    logger.warning("CE: NotebookLM 저장 실패 (무시) -- %s", nlm_e)

            return result

        except Exception as e:
            logger.error("CE: corpus 에세이 실패 -- %s", e)
            return {"error": str(e), "theme": theme}

    def _save_essay_html(self, result: dict, theme: str):
        """CE 에세이 결과를 website/archive/essay-NNN-slug/index.html 로 저장."""
        import re as _re
        from datetime import datetime as _dt
        from pathlib import Path as _Path

        # AgentLogger: HTML 저장 시작
        self.logger.write("HTML 생성 중: %s" % theme)

        # env_validator 경유 단일 진입점
        try:
            from core.system.env_validator import get_project_root, get_site_base_url
            PROJECT_ROOT = _Path(get_project_root())
        except Exception:
            PROJECT_ROOT = _Path(__file__).resolve().parent.parent.parent

        instagram_url = os.getenv('INSTAGRAM_URL', 'https://instagram.com/woohwahae')
        archive_dir = PROJECT_ROOT / 'website' / 'archive'

        # ── 이슈 번호 자동 계산 ──
        existing = sorted([
            d for d in archive_dir.iterdir()
            if d.is_dir() and _re.match(r'essay-\d+', d.name)
        ])
        # essay-000 ~ essay-008 형태 모두 포함해서 최댓값 추출
        max_num = 0
        for d in existing:
            m = _re.match(r'essay-0*(\d+)', d.name)
            if m:
                max_num = max(max_num, int(m.group(1)))
        next_num = max_num + 1
        essay_num_str = f"{next_num:03d}"  # 009, 010 ...

        # ── slug 생성 (테마 → 소문자 영문 slug) ──
        # 단어 단위 매핑 — 복합 테마도 처리 (첫 번째 매칭 키워드 사용)
        _word_slug_map = {
            '조용한지능': 'quiet-intelligence', '슬로우라이프': 'slow-life',
            '여백': 'negative-space', '기록': 'record',
            '감각': 'sensory', '침묵': 'silence',
            '물성': 'materiality', '일상': 'daily',
            '미용': 'beauty', '정신건강': 'mental-health', '정신': 'mental',
            '돌봄': 'care', '관계': 'relationship', '시간': 'time',
            '공간': 'space', '음식': 'food', '음악': 'music',
            '독서': 'reading', '글쓰기': 'writing', '사진': 'photo',
            '여행': 'travel', '산책': 'walk', '계절': 'season',
            '빛': 'light', '색': 'color', '질감': 'texture',
            '본질': 'essence', '선택': 'choice', '집중': 'focus',
            '느림': 'slow', '비움': 'empty', '채움': 'fill',
            '습관': 'habit', '루틴': 'routine', '의식': 'ritual',
            '몸': 'body', '마음': 'mind', '영성': 'spirit',
        }
        theme_key = theme.replace(' ', '').replace('의', '').replace('과', '').replace('와', '')
        slug = None
        # 정확 매칭 우선
        for k, v in _word_slug_map.items():
            if k in theme_key:
                slug = v
                break
        if not slug:
            # ASCII 테마는 그대로 slug화
            ascii_slug = _re.sub(r'[^\w-]', '', theme.lower().replace(' ', '-'))
            slug = ascii_slug if ascii_slug and ascii_slug[0].isascii() else 'essay'
        slug = slug[:30]

        folder_name = f"essay-{essay_num_str}-{slug}"
        essay_dir = archive_dir / folder_name
        essay_dir.mkdir(parents=True, exist_ok=True)

        essay_title = result.get('essay_title', theme)
        pull_quote = result.get('pull_quote', '')
        archive_essay = result.get('archive_essay', '')
        today = _dt.now().strftime('%Y.%m.%d')
        content_type_label = result.get('content_type', 'essay').capitalize()
        content_category_label = result.get('content_category', '')

        # ── 에세이 본문 → HTML 단락 변환 ──
        paragraphs = [p.strip() for p in archive_essay.split('\n\n') if p.strip()]
        para_html = ''
        for i, p in enumerate(paragraphs):
            # 중간에 <hr> 한 번 삽입 (절반 지점)
            if i == len(paragraphs) // 2:
                para_html += '            <hr class="article-divider">\n\n'
            if i == len(paragraphs) - 1:
                # 마지막 단락은 closing 스타일
                para_html += f'            <p class="article-closing fade-in">\n                {p}\n            </p>\n\n'
            else:
                para_html += f'            <p class="fade-in">\n                {p}\n            </p>\n\n'

        html = f"""<!DOCTYPE html>
<html lang="ko">

<head>
  <meta charset="UTF-8">
  <link rel="apple-touch-icon" href="../../assets/img/symbol.jpg">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="Issue {essay_num_str} — {essay_title}. {pull_quote}">
  <title>Issue {essay_num_str}: {essay_title} — WOOHWAHAE Archive</title>
  <link rel="manifest" href="/manifest.webmanifest">
  <meta name="theme-color" content="#FAFAF7">
  <link rel="icon" href="../../assets/img/symbol.jpg" type="image/jpeg">
  <link rel="stylesheet" href="../../assets/css/style.css">
</head>

<body>

    <nav>
        <a href="/" class="nav-brand" aria-label="WOOHWAHAE">
            <img src="../../assets/media/brand/symbol.png" alt="WOOHWAHAE" class="nav-symbol">
        </a>
        <ul class="nav-links" id="nav-links">
            <li><a href="../../archive/" class="active">Archive</a></li>
            <li><a href="../../practice/">Practice</a></li>
            <li><a href="../../about/">About</a></li>
        </ul>
        <button class="nav-toggle" id="nav-toggle" aria-label="메뉴" aria-expanded="false">
            <span></span><span></span>
        </button>
    </nav>

    <div class="article-container">

        <header class="article-header">
            <p class="article-meta fade-in">Issue {essay_num_str} &nbsp;·&nbsp; {content_category_label + ' · ' if content_category_label else ''}{content_type_label} &nbsp;·&nbsp; {today}</p>
            <h1 class="article-title fade-in">{essay_title}</h1>
            <p class="article-subtitle fade-in">{pull_quote}</p>
        </header>

        <div class="article-body">

{para_html}
        </div>

    </div>

    <footer class="site-footer">
        <div class="footer-inner">
            <p class="footer-brand">WOOHWAHAE</p>
            <nav class="footer-nav">
                <a href="../../archive/">Archive</a>
                <a href="../../practice/">Practice</a>
                <a href="../../about/">About</a>
                <a href="{instagram_url}" target="_blank" rel="noopener">Instagram</a>
            </nav>
            </div>
        </div>
        <p class="footer-copy">&copy; 2026 WOOHWAHAE. All rights reserved.</p>
    </footer>

    <script src="../../assets/js/main.js"></script>
</body>
</html>"""

        html_path = essay_dir / 'index.html'
        html_path.write_text(html, encoding='utf-8')
        logger.info("CE: HTML 저장 완료 -- %s", html_path.relative_to(PROJECT_ROOT))
        logger.info("CE: Issue %s '%s' -> archive/%s/", essay_num_str, essay_title, folder_name)

        # AgentLogger: 작업 완료
        self.logger.done("Issue %s: %s" % (essay_num_str, essay_title))

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)
        nlm_status = "연결됨" if self.nlm else "fallback"
        logger.info("CE: 큐 감시 시작. LLM=Gemini 2.5 Pro | Brand Voice=NotebookLM RAG (%s) | Tasks=write_content", nlm_status)
        watcher.watch(callback=self.process_task, interval=interval)


if __name__ == '__main__':
    import argparse
    from core.system.env_validator import validate_env
    validate_env("ce_agent")

    parser = argparse.ArgumentParser(description='LAYER OS Chief Editor Agent')
    parser.add_argument('--agent-id', default='ce-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()

    agent = ChiefEditor(agent_id=args.agent_id)

    if args.test:
        logger.info("[TEST] Test Mode: Content Writing")
        test_analysis = {
            'signal_id': 'test_001',
            'themes': ['AI와 창작', '느린 삶'],
            'key_insights': ['AI는 반복을 제거하고 창작에 집중하게 한다', '속도보다 깊이가 더 오래 남는다'],
            'summary': 'AI는 슬로우 라이프를 가능하게 하는 도구다',
        }
        test_visual = {
            'concept_title': '디지털 정원',
            'visual_mood': 'contemplative',
            'brand_alignment': '여백과 느림의 미학',
        }

        result = agent.write_content(test_analysis, test_visual)
        logger.info("[TEST] 콘텐츠 초안 -- 헤드라인: %s | 캡션: %s | 브랜드 보이스 출처: %s",
                    result.get('headline', 'N/A'), result.get('social_caption', 'N/A'), result.get('brand_voice_source', 'N/A'))
        logger.info("[TEST] 테스트 완료")
    else:
        agent.start_watching(interval=args.interval)
