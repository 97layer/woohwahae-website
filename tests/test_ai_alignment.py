import sys
from pathlib import Path
import os
from unittest.mock import MagicMock

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.ai_engine import AIEngine
from libs.agent_router import AgentRouter

def test_ai_alignment():
    print("Testing AI Alignment and System Instruction Support...")
    
    # Mocking API key to bypass real API calls if key is missing
    os.environ["GEMINI_API_KEY"] = "mock_key"
    
    # 1. AIEngine initialization test
    engine = AIEngine(system_instruction="You are a Stoic philosopher.")
    print("[1] AIEngine Initialization: OK")
    
    # 2. Interface Alignment test
    if hasattr(engine, 'generate_response'):
        print("[2] AIEngine Interface Alignment (generate_response): OK")
    else:
        print("[2] AIEngine Interface Alignment (generate_response): FAILED")
        sys.exit(1)
        
    # 3. AgentRouter integration test
    router = AgentRouter(ai_engine=engine)
    if router.ai.api_key == "mock_key":
        print("[3] AgentRouter AI Engine Integration: OK")
    else:
        print("[3] AgentRouter AI Engine Integration: FAILED")
        sys.exit(1)

    print("AI Alignment Test Completed Successfully.")

if __name__ == "__main__":
    test_ai_alignment()
