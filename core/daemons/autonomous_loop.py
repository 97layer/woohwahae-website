import os
import time
import json
import logging
from pathlib import Path
from datetime import datetime
# Legacy imports - autonomous_loop not actively used
# from system.libs.engines.ai_engine import AIEngine
# from system.libs.core_config import KNOWLEDGE_PATHS

PROJECT_ROOT = Path(__file__).parent.parent.parent
KNOWLEDGE_PATHS = {
    'signals': PROJECT_ROOT / 'knowledge' / 'signals',
    'insights': PROJECT_ROOT / 'knowledge' / 'insights',
    'content': PROJECT_ROOT / 'knowledge' / 'content',
    'system': PROJECT_ROOT / 'knowledge' / 'system',
    'archive': PROJECT_ROOT / 'knowledge' / 'archive',
}

# Logging setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler()
    ]
)
logger = logging.getLogger("AutonomousLoop")

class AutonomousLoop:
    """97layerOS 자율 운영 루프.
    신호 포착 -> 에이전트 협업 -> 결과물 생성 사이클을 자동화함.
    """
    def __init__(self):
        # self.ai = AIEngine()  # Legacy - not used
        self.signal_inbox = Path(KNOWLEDGE_PATHS["signals"])
        self.bridge_path = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"
        self.signal_inbox.mkdir(parents=True, exist_ok=True)

    def _update_bridge(self, agent: str, status: str, task: str = None):
        """Synapse Bridge 상태 업데이트"""
        if not self.bridge_path.exists():
            return
        
        try:
            with open(self.bridge_path, 'r', encoding='utf-8') as f:
                bridge = json.load(f)
            
            bridge["last_updated"] = datetime.now().isoformat()
            if agent in bridge["agents"]:
                bridge["agents"][agent]["status"] = status
                if task:
                    bridge["agents"][agent]["current_task"] = task
            
            with open(self.bridge_path, 'w', encoding='utf-8') as f:
                json.dump(bridge, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Bridge update failed: {e}")

    async def run_cycle(self):
        """단일 운영 사이클 실행"""
        logger.info("Checking for new signals...")
        signals = list(self.signal_inbox.glob("*"))
        
        if not signals:
            logger.info("No new signals found. Sleeping.")
            return

        for signal in signals:
            if signal.suffix not in [".md", ".txt"]: continue
            logger.info(f"Processing signal: {signal.name}")
            
            # Check for associated image file
            image_data = None
            for ext in [".jpg", ".jpeg", ".png"]:
                img_path = signal.with_suffix(ext)
                if img_path.exists():
                    logger.info(f"Found visual asset: {img_path.name}")
                    image_data = img_path.read_bytes()
                    break

            # 1. Strategy Analyst - 신호 분석
            self._update_bridge("Strategy Analyst", "active", f"Analyzing {signal.name}")
            analysis = await self.ai.generate_with_role(
                prompt=f"아래 신호(파일명: {signal.name})를 분석하여 브랜드 전략 인사이트를 도출하십시오.",
                role="SA",
                image_data=image_data
            )
            # Actually SA, CE, AD all might want images. Let's pass if relevant.
            
            logger.info("Strategy Analyst Analysis complete.")

            # 2. Chief Editor - 분석 데이터 기반 콘텐츠 초안 생성
            self._update_bridge("Chief Editor", "active", "Drafting content based on SA analysis")
            draft = await self.ai.generate_with_role(
                prompt=f"SA의 분석 결과를 바탕으로 브랜드 블로그/포스트 초안을 작성하십시오.\n\nAnalysis: {analysis}",
                role="CE"
            )
            logger.info("Chief Editor Drafting complete.")

            # 3. Art Director - 시각적 가이드 및 미학 검수 (Vision call)
            self._update_bridge("Art Director", "active", "Designing visual layout and aesthetic guide")
            visual_guide = await self.ai.generate_with_role(
                prompt=f"CE가 작성한 초안에 어울리는 시각적 레이아웃 가이드와 이미지 생성 프롬프트를 제안하십시오.\n\nDraft: {draft}",
                role="AD",
                image_data=image_data
            )
            logger.info("Art Director Visual Guide complete.")

            # 4. Creative Director - 최종 검토 및 승인
            self._update_bridge("Creative Director", "active", "Reviewing draft and visual guide for identity compliance")
            approval_report = await self.ai.generate_with_role(
                prompt=f"CE의 초안과 AD의 시각 가이드가 우리 브랜드 정체성(IDENTITY.md)에 부합하는지 최종 검토하십시오. 완벽하다면 [APPROVED]를 포함하십시오.\n\nDraft: {draft}\nVisual Guide: {visual_guide}",
                role="CD"
            )
            
            # 결과 저장
            output_dir = Path(KNOWLEDGE_PATHS["content"]) / datetime.now().strftime("%Y-%m")
            output_dir.mkdir(parents=True, exist_ok=True)
            
            report_file = output_dir / f"automation_report_{int(time.time())}.md"
            report_content = f"# Intelligence Report: {signal.name}\n\n## SA Analysis\n{analysis}\n\n## CE Draft\n{draft}\n\n## AD Visual Guide\n{visual_guide}\n\n## CD Approval Report\n{approval_report}"
            report_file.write_text(report_content, encoding="utf-8")
            
            # 처리 완료된 신호 아카이브
            archive_dir = Path(KNOWLEDGE_PATHS["archive"]) / "signals"
            archive_dir.mkdir(parents=True, exist_ok=True)
            signal.replace(archive_dir / signal.name)
            
            logger.info(f"Cycle complete for {signal.name}. Report saved to {report_file}")
            self._update_bridge("Strategy Analyst", "idle")
            self._update_bridge("Chief Editor", "idle")
            self._update_bridge("Art Director", "idle")
            self._update_bridge("Creative Director", "idle")

    async def run_evolution_phase(self):
        """시스템 자가 진단 및 개선(Spiral Evolution) 단계"""
        logger.info("Starting Evolution Phase (Self-Improvement)...")
        self._update_bridge("Technical Director", "active", "System Health Check")
        
        # 1. Technical Director가 시스템 병목 진단
        issue = await self.ai.generate_with_role(
            prompt="현재 97layerOS의 시스템 구조나 코드에서 개선할 수 있는 기술적 병목 또는 고도화 안건을 하나 제안하십시오.",
            role="TD"
        )
        
        # 2. Agent Council 토론
        self._update_bridge("Creative Director", "active", "Council Debate on Evolution")
        council_report = await self.ai.council_debate(topic=issue)
        
        # 3. 결과 기록 (Council Room)
        council_room = Path("/Users/97layer/97layerOS/knowledge/agent_hub/council_room.md")
        if council_room.exists():
            log_entry = f"\n### [{datetime.now().strftime('%Y-%m-%d %H:%M')}] Evolution Proposal: {issue[:50]}...\n{council_report}\n"
            current_content = council_room.read_text(encoding="utf-8")
            council_room.write_text(current_content + log_entry, encoding="utf-8")
        
        logger.info("Evolution Phase complete. Decision recorded in Council Room.")
        self._update_bridge("Creative Director", "idle")
        self._update_bridge("Technical Director", "idle")

    def start(self, interval: int = 60):
        """루프 시작"""
        logger.info("97layerOS Autonomous Loop started.")
        import asyncio
        cycle_count = 0
        while True:
            asyncio.run(self.run_cycle())
            
            # 5회 사이클마다 또는 신호가 없을 때 진화 단계 실행
            cycle_count += 1
            if cycle_count % 5 == 0:
                asyncio.run(self.run_evolution_phase())
                
            time.sleep(interval)

if __name__ == "__main__":
    loop = AutonomousLoop()
    # Test execution: Run one cycle and one evolution phase
    import asyncio
    logger.info("Starting Single Test Cycle...")
    asyncio.run(loop.run_cycle())
    logger.info("Starting Single Evolution Phase...")
    asyncio.run(loop.run_evolution_phase())
    logger.info("Single Test Run Complete.")
