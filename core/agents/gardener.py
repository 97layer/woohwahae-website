#!/usr/bin/env python3
"""
Gardener — 97layerOS 자가진화 에이전트

매일 새벽 3시 실행. 데이터를 분석하고 시스템을 진화시킨다.

수정 권한 3단계:
  FROZEN  — 절대 불가 (IDENTITY.md, CD.md)
  PROPOSE — 순호 승인 후 적용 (SA/AD/CE.md, intent 기준)
  AUTO    — 자동 갱신 (long_term_memory, QUANTA)

Author: 97layerOS
Updated: 2026-02-16
"""

import os
import sys
import json
import logging
import asyncio
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

try:
    from dotenv import load_dotenv
    load_dotenv(PROJECT_ROOT / '.env')
except ImportError:
    pass

import google.genai as genai

logger = logging.getLogger(__name__)

# ── 권한 정의 ─────────────────────────────────────
FROZEN = {
    # 순호의 본질 — 절대 불가
    "IDENTITY.md",
    "CD.md",
}

PROPOSE = {
    # 에이전트 행동 지침 — 순호 승인 필요
    "SA.md",
    "AD.md",
    "CE.md",
}

# AUTO: long_term_memory.json, INTELLIGENCE_QUANTA.md → 기존 SA/CE가 이미 처리
# Gardener는 분석 + 제안만 담당


class Gardener:
    """
    24시간 주기 자가진화 에이전트.
    데이터 분석 → AUTO 갱신 → PROPOSE 텔레그램 전송 → 승인 대기
    """

    def __init__(self):
        api_key = os.getenv('GOOGLE_API_KEY') or os.getenv('GEMINI_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY 또는 GEMINI_API_KEY 필요")

        self.client = genai.Client(api_key=api_key)
        self._model = 'gemini-2.5-flash'

        self.knowledge_dir = PROJECT_ROOT / 'knowledge'
        self.directives_dir = PROJECT_ROOT / 'directives'
        self.pending_file = self.knowledge_dir / 'system' / 'gardener_pending.json'

        # 대기 중인 제안 로드
        self.pending: List[Dict] = self._load_pending()

        logger.info("🌱 Gardener 초기화 완료")

    # ── 데이터 수집 ───────────────────────────────

    def _collect_stats(self, days: int = 7) -> Dict:
        """지난 N일 데이터 통계 수집"""
        cutoff = datetime.now() - timedelta(days=days)
        stats = {
            'period_days': days,
            'signal_count': 0,
            'sa_analyzed': 0,
            'avg_score': 0,
            'top_themes': [],
            'top_concepts': [],
            'low_score_patterns': [],
        }

        # signals/ 카운트
        signals_dir = self.knowledge_dir / 'signals'
        if signals_dir.exists():
            for sf in signals_dir.glob('**/*.json'):
                try:
                    data = json.loads(sf.read_text(encoding='utf-8'))
                    captured = data.get('captured_at', '')
                    if captured:
                        try:
                            dt = datetime.fromisoformat(captured[:19])
                            if dt < cutoff:
                                continue
                        except Exception:
                            pass
                    stats['signal_count'] += 1
                except Exception:
                    pass

        # corpus entries에서 SA 분석 점수/테마 수집 (signal 파일엔 analysis key 없음)
        scores = []
        theme_counter: Dict[str, int] = {}
        corpus_entries_dir = self.knowledge_dir / 'corpus' / 'entries'

        if corpus_entries_dir.exists():
            for ef in corpus_entries_dir.glob('*.json'):
                try:
                    entry = json.loads(ef.read_text(encoding='utf-8'))
                    indexed = entry.get('indexed_at', '')
                    if indexed:
                        try:
                            dt = datetime.fromisoformat(indexed[:19])
                            if dt < cutoff:
                                continue
                        except Exception:
                            pass
                    stats['sa_analyzed'] += 1
                    score = entry.get('strategic_score', 0)
                    if score:
                        scores.append(score)
                    for theme in entry.get('themes', []):
                        theme_counter[theme] = theme_counter.get(theme, 0) + 1
                except Exception:
                    pass

        if scores:
            stats['avg_score'] = round(sum(scores) / len(scores), 1)
            stats['low_score_patterns'] = [s for s in scores if s < 50]

        stats['top_themes'] = sorted(
            theme_counter.items(), key=lambda x: x[1], reverse=True
        )[:8]

        # long_term_memory 개념
        lm_path = self.knowledge_dir / 'long_term_memory.json'
        if lm_path.exists():
            try:
                lm = json.loads(lm_path.read_text(encoding='utf-8'))
                concepts = lm.get('concepts', {})
                stats['top_concepts'] = sorted(
                    concepts.items(), key=lambda x: x[1], reverse=True
                )[:10]
            except Exception:
                pass

        return stats

    def _load_directive(self, filename: str) -> str:
        """에이전트 지시어 로드"""
        path = self.directives_dir / 'agents' / filename
        if path.exists():
            return path.read_text(encoding='utf-8')
        return ""

    # ── 분석 + 제안 생성 ──────────────────────────

    def _analyze_and_propose(self, stats: Dict) -> List[Dict]:
        """
        Gemini로 데이터 분석 → PROPOSE 목록 생성
        각 제안: {target_file, section, current, proposed, reason}
        """
        proposals = []

        # SA.md 분석 — SA 집중 테마 업데이트 제안
        joon_content = self._load_directive('SA.md')
        if joon_content and stats['top_themes']:
            themes_str = ', '.join(f"{t}({c}회)" for t, c in stats['top_themes'][:5])
            prompt = f"""너는 97layerOS Gardener다.

지난 {stats['period_days']}일 데이터:
- 신호 수: {stats['signal_count']}개
- SA 분석: {stats['sa_analyzed']}개
- 평균 점수: {stats['avg_score']}
- 상위 테마: {themes_str}
- 상위 개념: {', '.join(k for k, _ in stats['top_concepts'][:5])}

현재 SA.md 일부:
{joon_content[:800]}

질문: 이 데이터를 보면 SA.md에서 어떤 부분을 미세조정하면 좋을까?
- 집중할 테마/카테고리 업데이트가 필요한가?
- 분석 기준에서 놓치고 있는 패턴이 있는가?

응답 형식 (JSON):
{{
  "needs_update": true/false,
  "section": "업데이트할 섹션명",
  "reason": "왜 필요한지 한 문장",
  "proposed_addition": "추가/수정할 내용 (2-3줄)"
}}

개선이 불필요하면 needs_update: false.
JSON만 출력."""

            try:
                resp = self.client.models.generate_content(
                    model=self._model, contents=[prompt]
                )
                text = resp.text.strip()
                import re
                m = re.search(r'\{.*\}', text, re.DOTALL)
                if m:
                    result = json.loads(m.group())
                    if result.get('needs_update'):
                        proposals.append({
                            'id': f"joon_{datetime.now().strftime('%Y%m%d')}",
                            'target_file': 'SA.md',
                            'section': result.get('section', '분석 집중 영역'),
                            'reason': result.get('reason', ''),
                            'proposed_addition': result.get('proposed_addition', ''),
                            'status': 'pending',
                            'created_at': datetime.now().isoformat(),
                        })
            except Exception as e:
                logger.warning("SA.md 분석 실패: %s", e)

        return proposals

    # ── AUTO 갱신 ─────────────────────────────────

    def _auto_update_quanta(self, stats: Dict):
        """INTELLIGENCE_QUANTA.md 자동 업데이트"""
        quanta_path = self.knowledge_dir / 'agent_hub' / 'INTELLIGENCE_QUANTA.md'
        if not quanta_path.exists():
            return

        try:
            content = quanta_path.read_text(encoding='utf-8')
            now = datetime.now().strftime('%Y-%m-%d %H:%M')
            themes_str = ', '.join(t for t, _ in stats['top_themes'][:5])
            concepts_str = ', '.join(k for k, _ in stats['top_concepts'][:5])

            # Gardener 업데이트 섹션 찾아서 갱신
            marker = "## 🌱 Gardener 자동 업데이트"
            new_section = (
                f"{marker}\n"
                f"최종 실행: {now}\n"
                f"분석 기간: {stats['period_days']}일\n"
                f"신호 수집: {stats['signal_count']}개 / SA 분석: {stats['sa_analyzed']}개\n"
                f"평균 전략점수: {stats['avg_score']}\n"
                f"부상 테마: {themes_str}\n"
                f"핵심 개념: {concepts_str}\n"
            )

            if marker in content:
                # 기존 섹션 교체
                import re
                content = re.sub(
                    rf"{re.escape(marker)}.*?(?=\n##|\Z)",
                    new_section,
                    content,
                    flags=re.DOTALL
                )
            else:
                content += f"\n\n{new_section}"

            quanta_path.write_text(content, encoding='utf-8')
            logger.info("✅ INTELLIGENCE_QUANTA.md 자동 업데이트")
        except Exception as e:
            logger.warning("QUANTA 업데이트 실패: %s", e)

    def _evolve_concept_memory(self, stats: Dict):
        """
        개념 진화 기록 — 대화/신호가 쌓일수록 사고가 깊어지는 구조의 핵심.

        기존 long_term_memory.json의 concepts는 카운트(슬로우라이프: 1)만 존재.
        이 메서드는 Gemini가 corpus entry들을 읽고 각 핵심 개념이 어떻게 심화됐는지
        서술로 기록한다. 모델이 바뀌어도 이 파일을 읽으면 동일한 사고 수준에서 출발 가능.
        """
        lm_path = self.knowledge_dir / 'long_term_memory.json'
        if not lm_path.exists():
            return

        try:
            lm = json.loads(lm_path.read_text(encoding='utf-8'))
        except Exception:
            return

        # corpus entries 로드 (최근 30개 — 사고 흐름 파악용)
        corpus_dir = self.knowledge_dir / 'corpus' / 'entries'
        recent_entries = []
        if corpus_dir.exists():
            entry_files = sorted(corpus_dir.glob('*.json'), reverse=True)[:30]
            for f in entry_files:
                try:
                    recent_entries.append(json.loads(f.read_text(encoding='utf-8')))
                except Exception:
                    pass

        if not recent_entries:
            # corpus 비어있으면 experiences에서 추출
            recent_entries = [
                {"summary": e.get("summary", ""), "themes": [], "key_insights": []}
                for e in lm.get("experiences", [])[-20:]
            ]

        if not recent_entries:
            return

        # 상위 개념 목록
        concepts = lm.get("concepts", {})
        top_concepts = sorted(concepts.items(), key=lambda x: x[1] if isinstance(x[1], (int, float)) else 0, reverse=True)[:6]
        if not top_concepts:
            return

        # 기존 concept_evolution 로드
        concept_evolution = lm.get("concept_evolution", {})

        # 각 상위 개념에 대해 진화 서술 생성
        entries_text = ""
        for e in recent_entries[:15]:
            entries_text += f"- {e.get('summary', '')[:120]}\n"

        concepts_str = ", ".join(k for k, _ in top_concepts)

        prompt = f"""너는 97layerOS의 지식 큐레이터다.

아래는 최근 수집된 신호들의 요약이다:
{entries_text}

이 사람이 반복적으로 다루는 핵심 개념들: {concepts_str}

각 개념에 대해 답하라:
1. 이 개념이 초기에는 어떤 맥락이었는가?
2. 최근 신호들을 통해 어떻게 심화/확장되었는가?
3. 현재 이 사람의 이 개념에 대한 사고 수준을 한 문장으로.

응답 형식 (JSON):
{{
  "concept_evolution": {{
    "개념명": {{
      "current_depth": "현재 사고 깊이를 한 문장으로",
      "trajectory": "초기 → 현재 방향으로 어떻게 변화했는지",
      "last_updated": "{datetime.now().strftime('%Y-%m-%d')}"
    }}
  }}
}}

분석 가능한 개념만 포함. JSON만 출력."""

        try:
            response = self.client.models.generate_content(
                model=self._model,
                contents=[prompt]
            )
            import re as re_module
            text = response.text.strip()
            match = re_module.search(r'\{.*\}', text, re_module.DOTALL)
            if not match:
                return

            result = json.loads(match.group())
            new_evolution = result.get("concept_evolution", {})

            # 기존 기록과 병합 (덮어쓰지 않고 누적)
            for concept, data in new_evolution.items():
                if concept not in concept_evolution:
                    concept_evolution[concept] = data
                else:
                    # 기존 trajectory 보존 + 현재 depth 갱신
                    concept_evolution[concept]["current_depth"] = data.get("current_depth", "")
                    concept_evolution[concept]["last_updated"] = data.get("last_updated", "")
                    prev_traj = concept_evolution[concept].get("trajectory", "")
                    new_traj = data.get("trajectory", "")
                    if new_traj and new_traj != prev_traj:
                        concept_evolution[concept]["trajectory"] = new_traj

            lm["concept_evolution"] = concept_evolution
            lm["metadata"]["last_updated"] = datetime.now().strftime('%Y-%m-%dT%H:%M')

            lm_path.write_text(json.dumps(lm, ensure_ascii=False, indent=2), encoding='utf-8')
            logger.info("🧠 개념 진화 기록 갱신: %d개 개념", len(new_evolution))

        except Exception as e:
            logger.warning("개념 진화 기록 실패: %s", e)

    def _update_quanta_with_growth(self, stats: Dict):
        """
        INTELLIGENCE_QUANTA.md를 상태 스냅샷이 아닌 사고 성장 일지로 갱신.
        어떤 모델이 읽어도 현재 사고 수준을 즉시 파악할 수 있도록.
        """
        quanta_path = self.knowledge_dir / 'agent_hub' / 'INTELLIGENCE_QUANTA.md'
        lm_path = self.knowledge_dir / 'long_term_memory.json'

        if not quanta_path.exists():
            return

        try:
            # concept_evolution 로드
            concept_evolution = {}
            if lm_path.exists():
                lm = json.loads(lm_path.read_text(encoding='utf-8'))
                concept_evolution = lm.get("concept_evolution", {})

            content = quanta_path.read_text(encoding='utf-8')
            now = datetime.now().strftime('%Y-%m-%d %H:%M')
            themes_str = ', '.join(t for t, _ in stats['top_themes'][:5]) or '없음'

            # 개념 진화 요약 텍스트
            evolution_lines = ""
            for concept, data in list(concept_evolution.items())[:4]:
                depth = data.get("current_depth", "")
                if depth:
                    evolution_lines += f"- **{concept}**: {depth}\n"

            if not evolution_lines:
                evolution_lines = "- (아직 충분한 신호 미축적)\n"

            marker = "## 🌱 Gardener 자동 업데이트"
            new_section = (
                f"{marker}\n"
                f"최종 실행: {now}\n\n"
                f"**수집 현황** | 신호: {stats['signal_count']}개 / SA분석: {stats['sa_analyzed']}개 / 평균점수: {stats['avg_score']}\n\n"
                f"**부상 테마** | {themes_str}\n\n"
                f"**개념 사고 수준** (세션 간 연속성 앵커)\n"
                f"{evolution_lines}\n"
            )

            import re as re_module
            if marker in content:
                content = re_module.sub(
                    rf"{re_module.escape(marker)}.*?(?=\n##|\Z)",
                    new_section,
                    content,
                    flags=re_module.DOTALL
                )
            else:
                content += f"\n\n{new_section}"

            quanta_path.write_text(content, encoding='utf-8')
            logger.info("✅ INTELLIGENCE_QUANTA.md 성장 일지 갱신")

        except Exception as e:
            logger.warning("QUANTA 성장 갱신 실패: %s", e)

    # ── 제안 관리 ─────────────────────────────────

    def _load_pending(self) -> List[Dict]:
        if self.pending_file.exists():
            try:
                return json.loads(self.pending_file.read_text(encoding='utf-8'))
            except Exception:
                pass
        return []

    def _save_pending(self):
        self.pending_file.parent.mkdir(parents=True, exist_ok=True)
        self.pending_file.write_text(
            json.dumps(self.pending, ensure_ascii=False, indent=2),
            encoding='utf-8'
        )

    def approve_proposal(self, proposal_id: str) -> Tuple[bool, str]:
        """순호 승인 → 실제 파일 수정"""
        proposal = next((p for p in self.pending if p['id'] == proposal_id), None)
        if not proposal:
            return False, "제안을 찾을 수 없음"

        filename = proposal['target_file']

        # FROZEN 이중 체크
        if filename in FROZEN:
            return False, f"🔒 {filename}은 수정 불가 (FROZEN)"

        if filename not in PROPOSE:
            return False, f"알 수 없는 파일: {filename}"

        # 실제 파일 수정
        path = self.directives_dir / 'agents' / filename
        try:
            content = path.read_text(encoding='utf-8')
            section = proposal['section']
            addition = proposal['proposed_addition']
            now = datetime.now().strftime('%Y-%m-%d')

            # 섹션 찾아서 추가, 없으면 끝에 추가
            if f"## {section}" in content:
                insert_point = content.find(f"## {section}") + len(f"## {section}")
                # 다음 ## 앞에 삽입
                next_section = content.find('\n##', insert_point)
                if next_section > 0:
                    content = (
                        content[:next_section]
                        + f"\n\n<!-- Gardener {now} -->\n{addition}"
                        + content[next_section:]
                    )
                else:
                    content += f"\n\n<!-- Gardener {now} -->\n{addition}"
            else:
                content += f"\n\n## {section}\n<!-- Gardener {now} -->\n{addition}"

            path.write_text(content, encoding='utf-8')

            # pending에서 제거
            self.pending = [p for p in self.pending if p['id'] != proposal_id]
            self._save_pending()

            logger.info("✅ 승인 적용: %s / %s", filename, section)
            return True, f"✅ {filename} — {section} 업데이트 완료"

        except Exception as e:
            return False, f"적용 실패: {e}"

    def reject_proposal(self, proposal_id: str) -> bool:
        """순호 거절 → pending에서 제거"""
        self.pending = [p for p in self.pending if p['id'] != proposal_id]
        self._save_pending()
        return True

    # ── 메인 사이클 ───────────────────────────────

    def _trigger_essay_for_cluster(self, cluster: Dict) -> Optional[str]:
        """
        성숙한 군집 → CE Agent에게 에세이 작성 지시
        Magazine B 방식: 단일 신호가 아닌 군집 전체를 RAG해서 에세이 작성

        Returns: task_id or None
        """
        from core.system.corpus_manager import CorpusManager
        from core.system.queue_manager import QueueManager

        corpus = CorpusManager()
        entries = corpus.get_entries_for_essay(cluster["entry_ids"])

        if not entries:
            return None

        # 에세이 RAG 컨텍스트 구성
        rag_context = []
        for e in entries:
            rag_context.append({
                "summary": e.get("summary", ""),
                "key_insights": e.get("key_insights", []),
                "themes": e.get("themes", []),
                "captured_at": e.get("captured_at", ""),
                "signal_type": e.get("signal_type", ""),
                "preview": e.get("raw_content_preview", ""),
            })

        payload = {
            "mode": "corpus_essay",
            "content_type": cluster.get("content_type", "archive"),
            "theme": cluster["theme"],
            "entry_count": cluster["entry_count"],
            "rag_context": rag_context,
            "avg_strategic_score": cluster["avg_strategic_score"],
            "time_span_hours": cluster["hours_span"],
            "instruction": (
                f"주제 '{cluster['theme']}'에 관한 {cluster['entry_count']}개의 신호를 바탕으로 "
                f"원소스 멀티유즈 콘텐츠를 생성하라. "
                f"archive_essay(롱폼) / instagram_caption(150자) / "
                f"carousel_slides(3~5장) / telegram_summary(3줄) / pull_quote(1문장) "
                f"5개 포맷을 동시에. 모두 같은 본질에서 파생."
            ),
        }

        try:
            queue = QueueManager()
            task_id = queue.create_task(
                agent_type="CE",
                task_type="write_corpus_essay",
                payload=payload,
            )
            logger.info(f"🖊️  에세이 트리거: {cluster['theme']} ({cluster['entry_count']}개 entry) → CE task {task_id}")
            return task_id
        except Exception as e:
            logger.error(f"에세이 트리거 실패: {e}")
            return None

    def _check_corpus_clusters(self) -> Dict:
        """
        Corpus 군집 성숙도 점검 → 익은 군집 에세이 트리거
        """
        try:
            from core.system.corpus_manager import CorpusManager
            corpus = CorpusManager()
            summary = corpus.get_summary()
            ripe = corpus.get_ripe_clusters()

            triggered = []
            for cluster in ripe:
                task_id = self._trigger_essay_for_cluster(cluster)
                if task_id:
                    triggered.append({
                        "theme": cluster["theme"],
                        "entry_count": cluster["entry_count"],
                        "task_id": task_id,
                    })

            logger.info(
                "📚 Corpus 점검: 총 %d개 entry / 군집 %d개 / 성숙 %d개 / 에세이 트리거 %d개",
                summary["total_entries"], summary["total_clusters"],
                summary["ripe_clusters"], len(triggered)
            )

            return {
                "corpus_summary": summary,
                "ripe_clusters": len(ripe),
                "essay_triggered": triggered,
            }
        except Exception as e:
            logger.warning(f"Corpus 점검 실패: {e}")
            return {"corpus_summary": {}, "ripe_clusters": 0, "essay_triggered": []}

    def _check_revisit_due(self) -> None:
        """재방문 시기가 된 고객 → Telegram 알림"""
        try:
            from core.modules.ritual import get_ritual_module
            due_clients = get_ritual_module().get_due_clients()
            if not due_clients:
                return

            admin_id = os.getenv('ADMIN_TELEGRAM_ID')
            bot_token = os.getenv('TELEGRAM_BOT_TOKEN')
            if not (admin_id and bot_token):
                logger.warning("Telegram 환경변수 미설정 — 재방문 알림 생략")
                return

            lines = [f"⏰ <b>재방문 예정 고객 {len(due_clients)}명</b>"]
            for c in due_clients:
                lines.append(f"• {c['name']} ({c.get('rhythm', '보통')} 리듬)")
            msg = "\n".join(lines)

            import httpx
            httpx.post(
                f"https://api.telegram.org/bot{bot_token}/sendMessage",
                json={"chat_id": admin_id, "text": msg, "parse_mode": "HTML"},
                timeout=10,
            )
            logger.info("⏰ 재방문 알림 전송: %d명", len(due_clients))
        except Exception as e:
            logger.warning("재방문 알림 실패: %s", e)

    def _record_growth_snapshot(self) -> None:
        """월별 성장 지표를 Growth Module에 자동 기록"""
        try:
            from core.modules.growth import get_growth_module
            period = datetime.now().strftime('%Y-%m')
            gm = get_growth_module()
            gm.auto_count_content(period)
            gm.auto_count_service(period)
            logger.info("Growth snapshot 저장: %s", period)
        except Exception as e:
            logger.warning("Growth snapshot 실패: %s", e)

    def run_cycle(self, days: int = 7) -> Dict:
        """
        Gardener 메인 사이클
        Returns: {stats, proposals, corpus_check, auto_updates}
        """
        logger.info("🌱 Gardener 사이클 시작 (지난 %d일)", days)

        # 1. 데이터 수집
        stats = self._collect_stats(days)
        logger.info(
            "📊 신호:%d / SA분석:%d / 평균점수:%s",
            stats['signal_count'], stats['sa_analyzed'], stats['avg_score']
        )

        # 2. 개념 진화 기록 (핵심: 대화가 쌓일수록 사고가 깊어지는 구조)
        self._evolve_concept_memory(stats)

        # 3. QUANTA 성장 일지 갱신 (상태 스냅샷 → 사고 수준 앵커로)
        self._update_quanta_with_growth(stats)

        # 4. Corpus 군집 성숙도 점검 → 익은 것 에세이 트리거 (핵심 신규)
        corpus_result = self._check_corpus_clusters()

        # 5. Growth Module 월간 집계 자동 기록
        self._record_growth_snapshot()

        # 6. 재방문 시기 고객 알림
        self._check_revisit_due()

        # 7. PROPOSE 생성 (신호가 10개 이상일 때만)
        new_proposals = []
        if stats['signal_count'] >= 10:
            new_proposals = self._analyze_and_propose(stats)
            if new_proposals:
                self.pending.extend(new_proposals)
                self._save_pending()
                logger.info("📝 새 제안 %d개 생성", len(new_proposals))
        else:
            logger.info("⏭️  신호 부족 (%d개) — 제안 생략", stats['signal_count'])

        return {
            'stats': stats,
            'new_proposals': new_proposals,
            'pending_count': len(self.pending),
            'corpus': corpus_result,
        }

    def format_telegram_report(self, result: Dict) -> str:
        """텔레그램 전송용 리포트 포맷"""
        stats = result['stats']
        proposals = result['new_proposals']

        themes = ', '.join(f"{t}" for t, _ in stats['top_themes'][:4]) or '없음'
        concepts = ', '.join(k for k, _ in stats['top_concepts'][:4]) or '없음'

        lines = [
            f"🌱 <b>Gardener 주간 리포트</b>",
            f"",
            f"<b>지난 {stats['period_days']}일 현황</b>",
            f"신호 수집: {stats['signal_count']}개",
            f"SA 분석: {stats['sa_analyzed']}개",
            f"평균 전략점수: {stats['avg_score']}",
            f"",
            f"<b>부상 테마</b>",
            f"{themes}",
            f"",
            f"<b>핵심 개념</b>",
            f"{concepts}",
        ]

        if proposals:
            lines += ["", f"<b>시스템 개선 제안 {len(proposals)}건</b>"]
            for p in proposals:
                lines.append(f"• {p['target_file']}: {p['reason']}")
            lines.append("")
            lines.append("승인하려면 /approve, 거절하려면 /reject")

        return "\n".join(lines)


# ── 스케줄러 (GCP systemd에서 실행) ──────────────

async def run_scheduled(hour: int = 3):
    """매일 지정 시각에 실행"""
    from core.agents.gardener import Gardener

    gardener = Gardener()

    while True:
        now = datetime.now()
        # 다음 실행 시각 계산
        next_run = now.replace(hour=hour, minute=0, second=0, microsecond=0)
        if next_run <= now:
            next_run += timedelta(days=1)

        wait_seconds = (next_run - now).total_seconds()
        logger.info("🌱 Gardener 대기 중 — 다음 실행: %s (%.0f초 후)",
                    next_run.strftime('%m/%d %H:%M'), wait_seconds)

        await asyncio.sleep(wait_seconds)

        try:
            result = gardener.run_cycle(days=7)

            # 텔레그램 리포트 전송
            admin_id = os.getenv('ADMIN_TELEGRAM_ID')
            bot_token = os.getenv('TELEGRAM_BOT_TOKEN')
            if admin_id and bot_token and result['stats']['signal_count'] > 0:
                import httpx
                msg = gardener.format_telegram_report(result)
                httpx.post(
                    f"https://api.telegram.org/bot{bot_token}/sendMessage",
                    json={
                        'chat_id': admin_id,
                        'text': msg,
                        'parse_mode': 'HTML'
                    },
                    timeout=10
                )
                logger.info("📨 텔레그램 리포트 전송 완료")

        except Exception as e:
            logger.error("Gardener 사이클 실패: %s", e)


if __name__ == '__main__':
    import argparse
    from core.system.env_validator import validate_env
    validate_env("gardener")

    logging.basicConfig(level=logging.INFO, format='%(message)s')

    parser = argparse.ArgumentParser(description='97layerOS Gardener')
    parser.add_argument('--run-now', action='store_true', help='즉시 1회 실행')
    parser.add_argument('--days', type=int, default=7, help='분석 기간 (기본: 7일)')
    parser.add_argument('--schedule', action='store_true', help='24시간 스케줄 모드')
    parser.add_argument('--hour', type=int, default=3, help='실행 시각 (기본: 3시)')
    args = parser.parse_args()

    if args.run_now:
        g = Gardener()
        result = g.run_cycle(days=args.days)
        print(g.format_telegram_report(result))

    elif args.schedule:
        asyncio.run(run_scheduled(hour=args.hour))

    else:
        parser.print_help()
