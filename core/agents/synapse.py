# Filename: libs/synapse.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12

import logging
import json
from pathlib import Path
from typing import List, Dict, Optional
from libs.core_config import AGENT_CREW, SYNAPSE_CONFIG
from core.system.manage_directive import DirectiveManager
from libs.notifier import Notifier


# Silence configuration
from libs.core_config import LOG_LEVEL
logging.basicConfig(level=getattr(logging, LOG_LEVEL))
logger = logging.getLogger(__name__)

class Synapse:
    """
    The Neural Network for 97LAYER OS.
    Facilitates inter-agent communication, debate, and consensus building.
    """

    def __init__(self, ai_engine):
        self.ai = ai_engine
        self.dm = DirectiveManager()

    def _get_agent_persona(self, role: str) -> str:
        """Loads the persona/directive for a specific agent."""
        if role not in AGENT_CREW:
            return "Standard AI Assistant"
        
        config = AGENT_CREW[role]
        directive = self.dm.read_directive(config['directive_path'])
        
        # Combine Identity and Mandate
        persona = f"Role: {role} ({config['legacy_name']})\n"
        if "IDENTITY" in directive:
            persona += f"Identity:\n{directive['IDENTITY']}\n"
        if "CORE MANDATE" in directive:
            persona += f"Mandate:\n{directive['CORE MANDATE']}\n"
            
        return persona

    def council_meeting(self, topic: str, participants: Optional[List[str]] = None) -> str:
        """
        Conducts a multi-agent debate/discussion on a topic.
        Returns the synthesized conclusion.
        """
        if not participants:
            participants = SYNAPSE_CONFIG["DEFAULT_COUNCIL"]
            
        if logger.isEnabledFor(logging.INFO):
            logger.info(f"Convening Council: {participants} on '{topic}'")
        
        import time
        from datetime import datetime
        
        # 1. Archive Start
        session_id = datetime.now().strftime("%Y%m%d_%H%M%S")
        log_content = f"# 🏛️ The Council Log ({session_id})\n\nTopic: {topic}\nParticipants: {', '.join(participants)}\n\n---\n\n"
        
        # 2. Initial Rounds - Each agent gives their perspective
        perspectives = {}
        for i, agent in enumerate(participants):
            if i > 0:
                time.sleep(10)  # Reduced wait time slightly for better UX, but keeping it safe
            persona = self._get_agent_persona(agent)
            prompt = f"""
            당신은 97LAYER의 고위 전략가인 {agent}입니다.
            현재 우리는 '{topic}'이라는 중요한 안건에 대해 토론 중입니다.

            [Persona & Mandate]
            {persona}

            [Instruction]
            위 주제에 대해 당신의 전문적 시각(Strategy, Technology, Design, or Narrative)에서 날카로운 통찰을 제시하십시오.
            단순한 동의나 일반론은 지양하며, 브랜드 헌법에 입각한 구체적인 제언을 하십시오.
            
            [Output Logic]
            1. 핵심 주장 (1문장)
            2. 근거 및 우려사항 (2-3문장)
            
            반드시 한국어로 작성하십시오.
            **볼드(**) 표기를 절대 사용하지 마십시오.**
            """
            response = self.ai.generate_response(prompt)
            perspectives[agent] = response
            
            log_entry = f"## 🗣️ {agent}\n{response}\n\n"
            log_content += log_entry
            # Silence: Use debug instead of info
            logger.debug(f"{agent}: {response[:50]}...")

        # 3. Synthesis & Decision (Creative Director has final say)
        time.sleep(10)
        synthesis_prompt = f"""
        당신은 97LAYER의 Creative Director (Sovereign)입니다.
        이사회(The Council) 멤버들의 다양한 의견을 종합하여 최종적인 전략적 결정을 내려야 합니다.

        [Agenda]
        {topic}

        [Council Opinions]
        {json.dumps(perspectives, indent=2, ensure_ascii=False)}

        [Mandate]
        - 각 의견의 본질을 꿰뚫어 보고, 상충되는 부분을 조율하십시오.
        - 브랜드의 장기적 비전과 미니멀리즘 철학에 가장 부합하는 결론을 내리십시오.
        - 실행 가능한 지침(Actionable Guideline)을 포함하십시오.

        [Final Output Logic]
        반드시 한국어로 작성하십시오.
        **볼드(**) 표기를 절대 사용하지 마십시오.**
        형식:
        1. 종합 분석 (Synthesis): 의견들의 공통점과 대립점 분석
        2. 최종 결정 (The Decision): 확정된 방향성
        3. 실행 지침 (Directives): 에이전트별 구체적 행동 강령
        """
        
        final_decision = self.ai.generate_response(synthesis_prompt)
        
        log_content += f"---\n\n## 👑 Final Decision (CD)\n{final_decision}\n"
        
        # 4. Archive Save
        try:
            log_path = self.dm.workspace_root / "knowledge" / "council_log" / f"council_{session_id}.md"
            with open(log_path, "w", encoding="utf-8") as f:
                f.write(log_content)
            logger.info(f"Council Log saved: {log_path}")
            
            return f"{final_decision}\n\n[Log Archived]: council_{session_id}.md"
        except Exception as e:
            logger.error(f"Failed to save council log: {e}")
            return final_decision

    def autonomous_though_loop(self):
        """
        태스크 파일을 스캔해서 council=True인 태스크를 자동 처리.
        technical_daemon의 10분 스케줄러가 호출.
        """
        import json
        from pathlib import Path

        task_file = Path(__file__).resolve().parent.parent / "task_status.json"
        if not task_file.exists():
            return

        try:
            with open(task_file, 'r', encoding='utf-8') as f:
                status = json.load(f)

            pending = status.get("pending_tasks", [])
            council_tasks = [t for t in pending if t.get("council") is True]

            if not council_tasks:
                logger.debug("[Synapse] 자율 처리 대상 태스크 없음.")
                return

            task = council_tasks[0]
            logger.info(f"[Synapse] Council 태스크 감지: {task.get('instruction', '')[:60]}")

            result = self.council_meeting(task["instruction"])

            # 완료 처리
            status["pending_tasks"] = [t for t in pending if t.get("id") != task.get("id")]
            status.setdefault("completed_tasks", []).append(f"{task.get('id', 'unknown')}_council_done")
            status["last_council_result"] = result[:500]

            with open(task_file, 'w', encoding='utf-8') as f:
                json.dump(status, f, indent=4, ensure_ascii=False)

            logger.info(f"[Synapse] Council 완료.")
            
            # Broadcast to Telegram
            try:
                notifier = Notifier()
                notifier.broadcast(f"🏛️ 위원회 결론 수립:\n\n{result[:1500]}")
            except Exception as tg_e:
                logger.error(f"[Synapse] Broadcast error: {tg_e}")
                
            return result

        except Exception as e:
            logger.error(f"[Synapse] autonomous_loop error: {e}")

    def propose_content_action(self, signal_path: str) -> str:
        """
        Analyzes a newly generated Raw Signal and proposes a collaborative task sequence
        if the insight is high-value and requires cross-agent cooperation.
        """
        try:
            with open(signal_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            prompt = f"""
            당신은 97LAYER의 Strategist입니다. 
            새롭게 수집된 지식 신호(Raw Signal)를 분석하여 협업 가치가 있는지 판단하십시오.

            [Raw Signal]
            {content[:3000]}

            [ 협업 설계 원칙 ]
            1. **No Fragmentation**: 정보가 파편화되지 않도록 기존 프로젝트 맥락과 연결하십시오.
            2. **Collaborative Chain**: 필요하다면 여러 에이전트가 순차적으로 협업하는 '태스크 체인'을 설계하십시오.
               예: [SA: 시장성 조사] -> [CE: 초안 작성] -> [AD: 비주얼 설계]
            3. **Valuation**: 0-10점 중 7점 이상일 때만 제안하십시오.

            [Output Format]
            - If PASS: returns "PASS"
            - If GO: returns JSON list of tasks:
            [
                {{
                    "type": "COLLABORATIVE_WORK",
                    "agent": "AgentRole",
                    "instruction": "지시사항",
                    "council": false
                }},
                ...
            ]
            """
            
            response = self.ai.generate_response(prompt)
            
            if "PASS" in response:
                return "PASS"
            
            # JSON parsing (naive but robust)
            import re
            json_match = re.search(r"\[.*\]", response, re.DOTALL)
            if json_match:
                tasks = json.loads(json_match.group(0))
                
                # Load current status
                task_status_path = Path(__file__).resolve().parent.parent / "task_status.json"
                if task_status_path.exists():
                    with open(task_status_path, 'r', encoding='utf-8') as f:
                        status = json.load(f)
                    
                    import time
                    for i, task_data in enumerate(reversed(tasks)):
                        task_data["id"] = f"collab_{int(time.time())}_{i}"
                        # Insert at the top in reverse order so the first task in the chain is executed first
                        status.setdefault("pending_tasks", []).insert(0, task_data)
                    
                    with open(task_status_path, 'w', encoding='utf-8') as f:
                        json.dump(status, f, indent=4, ensure_ascii=False)
                        
                    return f"Collaborative Chain Triggered: {len(tasks)} tasks."
            
            return "PASS (No valid JSON chain)"
            
        except Exception as e:
            logger.error(f"[Synapse] Collaborative Proposal Error: {e}")
            return f"Error: {e}"
