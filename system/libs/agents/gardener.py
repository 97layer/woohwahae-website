# Filename: libs/gardener.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered & Enhanced)

import os
import sys
import json
import logging
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.ai_engine import AIEngine
from libs.memory_manager import MemoryManager
from execution.system.manage_directive import DirectiveManager
from execution.system.log_error import ErrorLogger
from libs.core_config import SYSTEM_CONFIG, AGENT_CREW, MERCENARY_STANDARD, INITIAL_TASK_STATUS

# Silence configuration
from libs.core_config import LOG_LEVEL, MERCENARY_STANDARD
logging.basicConfig(level=getattr(logging, LOG_LEVEL))
logger = logging.getLogger(__name__)

class Gardener:
    """The Gardener: Self-evolution engine for 97LAYER OS."""

    # 브랜드 헌법은 자동 수정 금지 (사령부 지침)
    READ_ONLY_DIRECTIVES = [
        "woohwahae_identity.md",
        "brand_constitution.md",
        "97layer_identity.md"
    ]

    def __init__(self, ai_engine: AIEngine, memory_manager: MemoryManager, workspace_root: str):
        self.ai = ai_engine
        self.memory = memory_manager
        self.workspace = Path(workspace_root)
        self.status_file = self.workspace / "knowledge" / "system" / "task_status.json"
        
        try:
            self.directive_manager = DirectiveManager(workspace_root)
            self.error_logger = ErrorLogger()
            logger.debug("Gardener initialized with meta-tools")
        except Exception as e:
            logger.error(f"Gardener init error: {e}")
            self.directive_manager = None
            self.error_logger = None

    def _read_status(self) -> Dict:
        """Reads current system entropy and task status."""
        if not self.status_file.exists():
            return INITIAL_TASK_STATUS
        try:
            with open(self.status_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception:
            return INITIAL_TASK_STATUS

    def run_cycle(self, days: int = 7) -> str:
        """Runs the evolution cycle (Reflection -> Insight -> Evolve)."""
        logger.debug("Starting evolution cycle...")
        context = self.memory.get_recent_context(hours=days*24)
        status = self._read_status()
        
        prompt = f"""
        당신은 97LAYER OS의 정원사입니다.
        현재 상태: {json.dumps(status, indent=2, ensure_ascii=False)}
        
        최근 대화 맥락:
        {context}
        
        임무:
        1. 시스템의 엔트로피를 분석하십시오.
        2. 어떤 에이전트의 지침(Directive)이 업데이트 되어야 하는지 판단하십시오.
        3. 변경이 필요하다면 구체적인 섹션과 내용을 JSON 형식으로 제안하십시오.
        
        형식 예시:
        {{
            "insight": "전략 분석가의 리포팅 포맷이 너무 복잡함",
            "target_agent": "Strategy_Analyst",
            "section": "Reporting Format",
            "content": "- 간결함을 위해 Insight와 Relevance를 병합한다."
        }}
        
        **모든 분석과 제안은 한국어로 작성되어야 합니다.**
        """
        response = self.ai.generate_response(prompt)
        
        try:
            # Simple heuristic to extract JSON if AI wraps it in markdown
            if "```json" in response:
                json_str = response.split("```json")[1].split("```")[0].strip()
            elif "{" in response:
                json_str = response[response.find("{"):response.rfind("}")+1]
            else:
                return f"Evolution Analysis: {response[:100]}... (No actionable JSON)"
            
            plan = json.loads(json_str)
            return self._evolve_system(plan)
            
        except Exception as e:
            logger.error(f"Evolution failed: {e}")
            return f"Evolution Error: {str(e)}"

    def _evolve_system(self, plan: Dict) -> str:
        """Applies evolutionary changes to the system."""
        target_agent = plan.get('target_agent')
        section = plan.get('section')
        content = plan.get('content')
        insight = plan.get('insight', 'No insight provided')
        
        if not all([target_agent, section, content]):
            return "Evolution Skipped: Missing required fields in plan."
            
        if target_agent not in AGENT_CREW:
            return f"Evolution Skipped: Unknown agent '{target_agent}'"
            
        agent_config = AGENT_CREW[target_agent]
        directive_path = agent_config['directive_path']
        
        # Extract filename from path (e.g., directives/agents/strategy_analyst.md -> strategy_analyst.md)
        filename = Path(directive_path).name

        # 🔒 브랜드 헌법 보호 (사령부 지침)
        if filename in self.READ_ONLY_DIRECTIVES:
            logger.warning(f"🔒 Evolution BLOCKED: {filename} is Read-Only (Brand Constitution)")
            return f"Evolution Denied: {filename} is protected by Brand Constitution. Only 97layer can modify."

        if MERCENARY_STANDARD.get("SILENT_MODE"):
            logger.info(f"Evolving {target_agent} ({filename}) - Silent Update")
        else:
            logger.info(f"Evolving {target_agent} ({filename})...")
        success = self.directive_manager.update_directive(filename, section, content)
        
        if success:
            return f"Evolution Success: Updated {target_agent}'s {section}. Insight: {insight}"
        else:
            return f"Evolution Failed: Could not update {filename}."

    def analyze_and_heal(self):
        """Analyzes error patterns and heals directives."""
        if not self.error_logger: return
        
        patterns = self.error_logger.analyze_patterns()
        for pattern_key, count in patterns.items():
            logger.debug(f"Self-healing pattern detected: {pattern_key} ({count}x)")
            
            # Pattern 1: Lint Error - MD036
            if "MD036" in pattern_key:
                logger.debug("Auto-Healing: Fixing MD036 (Emphasis used as heading)...")
                # In a real scenario, this would parse the file and replace **text** with ### text
                # For now, we simulate the fix by notifying via Synapse
                prompt = f"Detected repeated lint error MD036. Suggest a modification to the Mercenary Standard to prevent this."
                insight = self.ai.generate_response(prompt)
                logger.debug(f"Healer Insight: {insight}")
                
            # Pattern 2: API Connection Failures
            elif "HttpError" in pattern_key:
                 logger.warning("Network instability detected. Suggesting increase in retry intervals.")

