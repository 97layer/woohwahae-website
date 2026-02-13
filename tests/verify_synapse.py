
import sys
import os
from pathlib import Path

# Add project root to sys.path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Mock AI Engine to avoid using tokens for test if possible, or use real one?
# Let's use real one to verify prompts work, but maybe short topic.
from libs.ai_engine import AIEngine
from libs.core_config import AI_MODEL_CONFIG
from libs.synapse import Synapse

def test_council():
    print("Initializing AIEngine...")
    ai = AIEngine(AI_MODEL_CONFIG)
    
    print("Initializing Synapse...")
    synapse = Synapse(ai)
    
    topic = "Test Agenda: Should 97LAYER adopt a 'Dark Mode Only' policy for all interfaces?"
    print(f"Starting Council Meeting on: {topic}")
    
    # Run council with a subset of agents to save time/tokens
    # Note: synapse.council_meeting signature is (topic, participants=None)
    participants = ["Creative_Director", "Technical_Director"]
    result = synapse.council_meeting(topic, participants=participants)
    
    print("\n--- Council Result ---")
    print(result)
    print("----------------------")

if __name__ == "__main__":
    test_council()
