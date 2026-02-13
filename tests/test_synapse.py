import sys
import logging
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.ai_engine import AIEngine
from libs.synapse import Synapse

# Configure logging to see the debate
logging.basicConfig(level=logging.INFO)

def test_council():
    print("Initializing AIEngine & Synapse...")
    ai = AIEngine()
    synapse = Synapse(ai)
    
    topic = "Should 97LAYER adopt a neon-green color scheme for high visibility?"
    print(f"\n[TOPIC]: {topic}\n")
    
    # Run the council
    decision = synapse.council_meeting(topic)
    
    print("\n" + "="*50)
    print("FINAL COUNCIL DECISION")
    print("="*50)
    print(decision)

if __name__ == "__main__":
    test_council()
