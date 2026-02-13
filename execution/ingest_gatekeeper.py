import os
import time
import shutil
import logging
from pathlib import Path
from datetime import datetime

# Setup Logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

BASE_DIR = Path(__file__).resolve().parent.parent
INBOX_DIR = BASE_DIR / "knowledge" / "inbox"
PROCESSING_DIR = BASE_DIR / "knowledge" / "processing"
RAW_SIGNALS_DIR = BASE_DIR / "knowledge" / "raw_signals"

# Ensure directories exist
INBOX_DIR.mkdir(parents=True, exist_ok=True)
PROCESSING_DIR.mkdir(parents=True, exist_ok=True)
RAW_SIGNALS_DIR.mkdir(parents=True, exist_ok=True)

ALLOWED_EXTENSIONS = {".txt", ".md", ".json", ".csv"}

def process_inbox():
    """Scans inbox for new files and processes them."""
    logger.info("Scanning Inbox for new intelligence...")
    
    files = [f for f in INBOX_DIR.iterdir() if f.is_file() and f.suffix in ALLOWED_EXTENSIONS]
    
    if not files:
        logger.info("Inbox is empty.")
        return

    from execution.ontology_transform import OntologyEngine
    engine = OntologyEngine()

    for file_path in files:
        try:
            logger.info(f"Processing: {file_path.name}")
            
            # 1. Move to Processing (Safety)
            processing_path = PROCESSING_DIR / file_path.name
            shutil.move(str(file_path), str(processing_path))
            
            # 2. Extract & Classify
            with open(processing_path, "r", encoding="utf-8", errors="ignore") as f:
                content = f.read()
            
            # Call Ontology Engine (LLM based classification)
            result = engine.process_content(content, source=file_path.name)
            
            if result:
                logger.info(f"Successfully processed: {result}")
                
                # 3. Trigger Synapse Action Proposal
                try:
                    # Lazy import to avoid circular dependency if any (though here it's fine)
                    from libs.synapse import Synapse
                    from libs.ai_engine import AIEngine
                    from libs.core_config import AI_MODEL_CONFIG
                    
                    # We need AI instance for Synapse
                    ai_engine = AIEngine(AI_MODEL_CONFIG)
                    synapse = Synapse(ai_engine)
                    
                    action_result = synapse.propose_content_action(result)
                    logger.info(f"Synapse Action: {action_result}")
                    
                    # Proactive Notify: Tell user that a new signal was processed and an action was proposed
                    from libs.notifier import Notifier
                    notifier = Notifier()
                    
                    msg = f"🧩 [Proactive Pulse] 새로운 신호가 자산화되었습니다.\nID: {result.get('id', 'N/A')}\n분류: {result.get('type', 'N/A')}\n\n추천 액션:\n{action_result[:300]}..."
                    notifier.broadcast(msg)
                    
                except Exception as se:
                    logger.error(f"Synapse trigger failed: {se}")

                # Optional: Archive the original raw file or delete?
                # For now, let's keep it in processing or move to archive
                archive_dir = BASE_DIR / "knowledge" / "archive" / datetime.now().strftime("%Y%m")
                archive_dir.mkdir(parents=True, exist_ok=True)
                shutil.move(str(processing_path), str(archive_dir / file_path.name))
            else:
                logger.error(f"Failed to process {file_path.name}")

        except Exception as e:
            logger.error(f"Error processing {file_path.name}: {e}")

if __name__ == "__main__":
    # Add project root to sys.path for imports
    import sys
    sys.path.insert(0, str(BASE_DIR))
    
    process_inbox()
