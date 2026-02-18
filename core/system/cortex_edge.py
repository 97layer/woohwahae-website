#!/usr/bin/env python3
"""
97layerOS Cortex Edge
The Central Intelligence Gateway for systemic coordination.

Features:
- Unified Engine: Single interface for text, vision, and RAG.
- Contextual Awareness: Integrated Identity + Memory + Signals.
- Multi-Source RAG: High-performance local search + Cloud fallback.
- API Standardization: Uniform request/response for Telegram and Web.
"""

import os
import json
import logging
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime

# Import Engines
from core.system.gemini_engine import get_gemini_engine
from core.system.conversation_engine import ConversationEngine
from core.system.queue_manager import QueueManager

logger = logging.getLogger(__name__)
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent

class CortexEdge:
    """
    97layerOS의 중앙 사고 엔진.
    시스템 전체의 인텔리전스를 조정하고 통합된 맥락을 제공함.
    """
    def __init__(self):
        self.gemini = get_gemini_engine()
        self.conv_engine = ConversationEngine()  # Existing logic for memory/RAG
        self.queue = QueueManager()
        
        self.knowledge_dir = PROJECT_ROOT / 'knowledge'
        self.signals_dir = self.knowledge_dir / 'signals'
        
        logger.info("🧠 Cortex Edge Engine Active")

    def query(self, user_id: str, text: str, mode: str = "chat") -> Dict[str, Any]:
        """
        통합 쿼리 처리기.
        
        Args:
            user_id: 요청자 식별자
            text: 입력 메시지
            mode: 'chat', 'analyze', 'search' 등
            
        Returns:
            통합 응답 객체
        """
        start_time = datetime.now()
        
        # 1. 의도 분석 및 맥락 구성 (ConversationEngine 로직 활용)
        response_text = self.conv_engine.chat(user_id, text)
        
        # 2. 장기 기억 업데이트 (Graph Extraction) - 비동기 시뮬레이션
        try:
            self._update_long_term_memory(text, response_text)
        except Exception as e:
            logger.error(f"Memory update error: {e}")
        
        # 2. 성능 메트릭 및 메타데이터 추가
        latency = (datetime.now() - start_time).total_seconds()
        
        return {
            "user_id": user_id,
            "response": response_text,
            "latency": f"{latency:.2f}s",
            "timestamp": datetime.now().isoformat(),
            "engine": "Cortex-V1"
        }

    def inject_signal(self, text: str, source: str = "web") -> Dict[str, Any]:
        """
        신호(인사이트) 주입 및 비동기 분석 트리거.
        """
        self.signals_dir.mkdir(parents=True, exist_ok=True)
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        signal_id = f"{source}_{timestamp}"
        
        signal_data = {
            'signal_id': signal_id,
            'type': 'text_insight',
            'content': text,
            'captured_at': datetime.now().isoformat(),
            'from_user': source,
            'status': 'captured'
        }
        
        # 파일 저장
        signal_path = self.signals_dir / f"{signal_id}.json"
        with open(signal_path, 'w', encoding='utf-8') as f:
            json.dump(signal_data, f, ensure_ascii=False, indent=2)
            
        # SA 분석 태스크 생성 (큐 매니저 연동)
        task_id = self.queue.create_task(
            agent_type='SA',
            task_type='analyze',
            payload={
                'signal_id': signal_id,
                'content': text,
                'source': f'cortex_{source}'
            }
        )
        
        return {
            "success": True,
            "signal_id": signal_id,
            "task_id": task_id
        }

    def _update_long_term_memory(self, user_message: str, assistant_answer: str):
        """
        장기 기억을 업데이트하고 개념 및 관계를 추출하여 온톨로지에 저장.
        """
        memory_path = self.knowledge_dir / 'long_term_memory.json'
        data = {"concepts": {}, "experiences": [], "metadata": {}}

        if memory_path.exists():
            try:
                with open(memory_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
            except json.JSONDecodeError:
                logger.warning("long_term_memory.json is corrupted, starting fresh.")
                data = {"concepts": {}, "experiences": [], "metadata": {}}

        # Add new experience
        data['experiences'].append({
            "user": user_message,
            "assistant": assistant_answer,
            "timestamp": datetime.now().isoformat()
        })
        
        # Gemini로 개념 및 관계 추출 (Graph Extraction)
        extract_prompt = f"""다음 대화에서 97layer 개인의 핵심 개념(Nodes)과 개념 간의 관계(Edges)를 추출하라.

사용자: {user_message[:300]}
비서: {assistant_answer[:300]}

JSON으로만 응답:
{{
  "concepts": ["개념1", "개념2"],
  "relations": [
    {{"source": "개념1", "target": "개념2", "relation": "연결어", "strength": 0.5}}
  ],
  "summary": "한 문장 요약",
  "category": "브랜드/개인/기술/비즈니스/라이프스타일 중 하나"
}}"""

        try:
            resp = self.gemini.generate_text(extract_prompt) # Use GeminiEngine
            import re
            json_match = re.search(r'\{.*\}', resp, re.DOTALL)
            if not json_match:
                return
            extracted = json.loads(json_match.group())
        except Exception as e:
            logger.error(f"Failed to extract knowledge: {e}")
            return

        # Prepare Ontology section
        if 'ontology' not in data:
            data['ontology'] = {"nodes": [], "edges": []}

        # concepts 및 nodes 업데이트
        for concept in extracted.get('concepts', []):
            concept = concept.strip()
            if concept:
                # Flat count for backward compatibility
                data['concepts'][concept] = data['concepts'].get(concept, 0) + 1
                
                # Graph Nodes
                node_exists = False
                for node in data['ontology']['nodes']:
                    if node['id'] == concept:
                        node['weight'] = node.get('weight', 0) + 1
                        node_exists = True
                        break
                if not node_exists:
                    data['ontology']['nodes'].append({
                        "id": concept,
                        "type": extracted.get('category', 'unknown'),
                        "weight": 1
                    })

        # Graph Edges 업데이트
        for rel in extracted.get('relations', []):
            source = rel.get('source')
            target = rel.get('target')
            if source and target:
                edge_exists = False
                for edge in data['ontology']['edges']:
                    if edge['source'] == source and edge['target'] == target:
                        edge['strength'] = min(1.0, edge.get('strength', 0.5) + 0.1)
                        edge_exists = True
                        break
                if not edge_exists:
                    data['ontology']['edges'].append({
                        "source": source,
                        "target": target,
                        "relation": rel.get('relation', 'connected'),
                        "strength": rel.get('strength', 0.5)
                    })

        data['metadata']['last_updated'] = datetime.now().isoformat()

        with open(memory_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def get_system_status(self) -> Dict[str, Any]:
        """
        시스템 상태 요약 (Cockpit용)
        """
        memory_path = self.knowledge_dir / 'long_term_memory.json'
        memory = {}
        if memory_path.exists():
            try:
                memory = json.loads(memory_path.read_text(encoding='utf-8'))
            except:
                pass
        
        recent_signals = sorted(
            self.signals_dir.glob('*.json'),
            key=lambda p: p.stat().st_mtime,
            reverse=True
        )[:5]
        
        return {
            "intelligence": {
                "concepts": len(memory.get('concepts', {})),
                "experiences": len(memory.get('experiences', [])),
                "last_update": memory.get('metadata', {}).get('last_updated')
            },
            "signals": {
                "recent": [f.name for f in recent_signals]
            }
        }

# Singleton
_ctx_instance = None

def get_cortex() -> CortexEdge:
    global _ctx_instance
    if _ctx_instance is None:
        _ctx_instance = CortexEdge()
    return _ctx_instance
