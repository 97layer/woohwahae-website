import logging
import sys
from pathlib import Path
from io import StringIO

# Mock project root
project_root = Path("/Users/97layer/97layerOS")
sys.path.append(str(project_root))

from libs.core_config import MERCENARY_STANDARD, LOG_LEVEL
from libs.synapse import Synapse
from libs.gardener import Gardener

# Capture stdout/stderr
capture = StringIO()
handler = logging.StreamHandler(capture)
handler.setLevel(LOG_LEVEL)

# Setup logger interception
root_logger = logging.getLogger()
root_logger.setLevel(LOG_LEVEL)
# Remove existing handlers to avoid double logging during test
for h in root_logger.handlers[:]:
    root_logger.removeHandler(h)
root_logger.addHandler(handler)

print(f"Testing Silent Mode: {MERCENARY_STANDARD.get('SILENT_MODE')}")
print(f"Log Level: {LOG_LEVEL}")

# 1. Test Synapse Silence
print("--- Initializing Synapse ---")
try:
    # Minimal mock for AI engine
    class MockAI:
        def generate_response(self, prompt):
            return "Silent thought..."
    
    synapse = Synapse(MockAI())
    # This should be DEBUG now, so it shouldn't show up in capture if level is WARNING
    synapse.council_meeting("Test Topic", participants=["Creative_Director"]) 
except Exception as e:
    print(f"Synapse init failed: {e}")

# 2. Test Gardener Silence
print("--- Initializing Gardener ---")
try:
    # Minimal mocks
    class MockMemory:
        def get_recent_context(self, hours): return "context"
    
    gardener = Gardener(MockAI(), MockMemory(), str(project_root))
    # This should be DEBUG now
    gardener.run_cycle(days=1)
except Exception as e:
    print(f"Gardener run failed: {e}")

# Check captured output
captured_logs = capture.getvalue()
print(f"--- Captured Logs (Length: {len(captured_logs)}) ---")
print(captured_logs)
print("--- End Capture ---")

if len(captured_logs) == 0:
    print("SUCCESS: Absolute Silence Achieved.")
elif len(captured_logs) < 100:
    print("WARNING: Some logs leaked, but mostly silent.")
else:
    print("FAILURE: Too much noise.")
