import sys
import logging
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.synapse import Synapse

# Mock AI Engine for verification without API Key
class MockAIEngine:
    def __init__(self, api_key=None):
        pass
    
    def generate_response(self, prompt, system_instruction=None, model_type=None):
        # Simulate agent responses based on prompt content
        if "Identity" in prompt:
            return "This matches my philosophy. (Mocked Opinion)"
        if "Creative Director" in prompt:
            return "Final Decision: Proceed with caution. (Mocked Decision)"
        return "I have no strong feelings one way or the other."

def test_council_mock():
    print("Initializing Mock AIEngine & Synapse...")
    ai = MockAIEngine()
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
    test_council_mock()
