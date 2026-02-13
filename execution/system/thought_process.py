# Filename: execution/system/thought_process.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12

import sys
import time
import json
import logging
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.ai_engine import AIEngine
from libs.synapse import Synapse
from libs.gardener import Gardener
from libs.memory_manager import MemoryManager
from libs.core_config import SYNAPSE_CONFIG, INITIAL_TASK_STATUS

# Setup Logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [THOUGHT] - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("thought_process.log"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class ThoughtLoop:
    def __init__(self):
        self.ai = AIEngine()
        self.synapse = Synapse(self.ai)
        self.memory = MemoryManager(str(project_root))
        self.gardener = Gardener(self.ai, self.memory, str(project_root))
        self.status_file = project_root / "task_status.json"

    def _read_status(self):
        if not self.status_file.exists():
            return INITIAL_TASK_STATUS
        with open(self.status_file, 'r', encoding='utf-8') as f:
            return json.load(f)

    def run(self):
        logger.info("Starting Autonomous Thought Loop...")
        
        while True:
            try:
                status = self._read_status()
                pending = status.get('pending_tasks', [])
                
                if pending:
                    task = pending[0] # Take highest priority
                    logger.info(f"Detected Pending Task: {task}")
                    
                    # Initiate Council for the task
                    logger.info(f"Convening Council for: {task}")
                    decision = self.synapse.council_meeting(f"How should we execute: {task}?")
                    
                    logger.info(f"Council Decision: {decision}")
                    
                    # TODO: Execute decision via Technical Director (future implementation)
                    # For now, we log it and respect the loop interval

                # Self-Healing Check
                logger.info("Running System Health Check...")
                self.gardener.analyze_and_heal()
                    
                time.sleep(SYNAPSE_CONFIG["AUTONOMY_INTERVAL"])
                
            except KeyboardInterrupt:
                logger.info("Stopping Thought Loop...")
                break
            except Exception as e:
                logger.error(f"Loop Error: {e}")
                time.sleep(60)

if __name__ == "__main__":
    loop = ThoughtLoop()
    loop.run()
