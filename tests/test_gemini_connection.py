#!/usr/bin/env python3
"""
Test Gemini API connection
"""

import os
import sys
from pathlib import Path

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

def test_gemini_direct():
    """Direct Gemini API test"""
    import google.generativeai as genai

    # Load API key
    api_key = os.getenv("GEMINI_API_KEY") or "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"

    print("=== Gemini API Direct Test ===")
    print(f"API Key: {api_key[:10]}...")

    try:
        genai.configure(api_key=api_key)
        model = genai.GenerativeModel('gemini-1.5-flash-latest')

        response = model.generate_content("Say 'OK' if you are working")
        print(f"✅ Direct API Response: {response.text}")
        return True
    except Exception as e:
        print(f"❌ Direct API Error: {e}")
        return False

def test_ai_engine():
    """Test our AIEngine wrapper"""
    from libs.ai_engine import AIEngine

    print("\n=== AI Engine Test ===")

    try:
        ai = AIEngine()

        # Check if API key loaded
        if not ai.api_key:
            print("❌ AI Engine: API key not loaded")
            return False

        print(f"✅ AI Engine: API key loaded ({ai.api_key[:10]}...)")

        # Test generation
        response = ai.generate_response("Say 'System operational' if working")
        print(f"✅ AI Engine Response: {response}")
        return True

    except Exception as e:
        print(f"❌ AI Engine Error: {e}")
        return False

def test_multimodal():
    """Test image analysis capability"""
    print("\n=== Multimodal Test ===")

    try:
        from libs.ai_engine import AIEngine
        ai = AIEngine()

        # Create a simple test image (1x1 white pixel)
        import base64
        white_pixel = base64.b64decode("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==")

        response = ai.generate_multimodal(
            "What do you see in this image?",
            white_pixel
        )
        print(f"✅ Vision Response: {response[:100]}...")
        return True

    except Exception as e:
        print(f"❌ Multimodal Error: {e}")
        return False

def main():
    print("97LAYER OS - Gemini Integration Test")
    print("=" * 50)

    # Set environment variable if not set
    if not os.getenv("GEMINI_API_KEY"):
        os.environ["GEMINI_API_KEY"] = "AIzaSyBHpQRFjdZRzzkYGR6eqBezyPteaHX_uMQ"

    tests = [
        ("Direct API", test_gemini_direct),
        ("AI Engine", test_ai_engine),
        ("Multimodal", test_multimodal)
    ]

    results = []
    for name, test_func in tests:
        try:
            success = test_func()
            results.append((name, success))
        except Exception as e:
            print(f"Test {name} failed with exception: {e}")
            results.append((name, False))

    print("\n" + "=" * 50)
    print("Test Results:")
    for name, success in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"  {name}: {status}")

    all_passed = all(r[1] for r in results)
    print(f"\nOverall: {'✅ ALL TESTS PASSED' if all_passed else '❌ SOME TESTS FAILED'}")

    return 0 if all_passed else 1

if __name__ == "__main__":
    sys.exit(main())