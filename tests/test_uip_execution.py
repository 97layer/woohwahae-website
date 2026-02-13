#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Manual test for UIP skill execution
"""

import sys
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(BASE_DIR))

def main():
    from libs.skill_engine import SkillEngine

    # Test URL (Rick Astley - Never Gonna Give You Up, a stable classic)
    test_url = "https://youtube.com/watch?v=dQw4w9WgXcQ"

    print(f"Testing UIP skill with URL: {test_url}")
    print("-" * 60)

    engine = SkillEngine()
    result = engine.execute_skill("skill-001", {"url": test_url})

    print(f"Status: {result.get('status')}")
    print(f"Message: {result.get('message')}")

    if result.get("status") == "success":
        print(f"Signal ID: {result.get('signal_id')}")
        print(f"Title: {result.get('title')}")
        print(f"Output file: {result.get('output_file')}")
        print()
        print("Raw signal created successfully!")
        print()
        print("To view the file:")
        print(f"  cat {result.get('output_file')}")
    else:
        print(f"Error: {result.get('message')}")
        if "details" in result:
            print(f"Details: {result['details']}")
        if "traceback" in result:
            print("\nTraceback:")
            print(result["traceback"])

if __name__ == "__main__":
    main()
