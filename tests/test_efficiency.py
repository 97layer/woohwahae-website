import sys
from pathlib import Path

# Add project root to sys.path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from libs.memory_manager import MemoryManager
from libs.ai_engine import AIEngine
from libs.core_config import AI_MODEL_CONFIG

def test_efficiency():
    print("=== [Efficiency Infrastructure Test] ===")
    
    mm = MemoryManager(str(PROJECT_ROOT))
    ai = AIEngine(AI_MODEL_CONFIG)
    
    # 1. Test RAG-lite (Snippet Retrieval)
    print("\n1. Testing RAG-lite (Snippet Retrieval)...")
    results = mm.get_relevant_knowledge("WOOHWAHAE brand")
    if results:
        print(f"✓ Found {len(results)} relevant files.")
        for res in results:
            print(f"  - File: {res['file']}, Snippets: {len(res['snippets'])}")
    else:
        print("✗ No relevant knowledge found.")

    # 2. Test Response Caching
    print("\n2. Testing Response Caching...")
    test_prompt = "Efficiency Protocol의 핵심 원칙 3가지를 짧게 말해줘."
    
    print("  [Call 1] Generating response (should hit API)...")
    start_time = sys.time() if hasattr(sys, 'time') else 0 # Dummy for timing
    import time
    t1 = time.time()
    resp1 = ai.generate_response(test_prompt)
    t2 = time.time()
    print(f"  Response 1 received in {t2-t1:.2f}s")
    
    print("\n  [Call 2] Generating same response (should hit CACHE)...")
    t3 = time.time()
    resp2 = ai.generate_response(test_prompt)
    t4 = time.time()
    print(f"  Response 2 received in {t4-t3:.2f}s")
    
    if (t4-t3) < 0.1: # Cache should be near-instant
        print("✓ Cache Hit successful!")
    else:
        print("✗ Cache Hit failed or was slow.")
        
    if resp1 == resp2:
        print("✓ Content consistency verified.")
    else:
        print("✗ Content mismatch between calls.")

if __name__ == "__main__":
    test_efficiency()
