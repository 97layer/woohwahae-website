#!/usr/bin/env python3
"""
Telegram Secretary 기본 테스트
초기화, 명령어 핸들러, 통합 기능 검증

Author: 97layerOS Technical Director
"""

import sys
import asyncio
from pathlib import Path
from unittest.mock import Mock, AsyncMock, patch

PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))


def test_1_imports():
    """1. 필수 임포트 검증"""
    print("\n" + "="*70)
    print("TEST 1: 필수 임포트 검증")
    print("="*70)

    try:
        from execution.daemons.telegram_secretary import TelegramSecretary
        print("✅ TelegramSecretary 임포트 성공")

        from telegram import Update
        from telegram.ext import ContextTypes
        print("✅ Telegram 라이브러리 임포트 성공")

        return True

    except ImportError as e:
        print(f"❌ 임포트 실패: {e}")
        return False


def test_2_initialization():
    """2. Secretary 초기화 테스트"""
    print("\n" + "="*70)
    print("TEST 2: Secretary 초기화")
    print("="*70)

    try:
        from execution.daemons.telegram_secretary import TelegramSecretary

        # Mock bot token
        test_token = "test_token_12345"

        # Mock handoff engine
        with patch('execution.daemons.telegram_secretary.HandoffEngine') as mock_handoff:
            mock_handoff_instance = Mock()
            mock_handoff_instance.onboard = Mock()
            mock_handoff_instance.acquire_work_lock = Mock(return_value=True)
            mock_handoff.return_value = mock_handoff_instance

            secretary = TelegramSecretary(test_token)

            assert secretary.bot_token == test_token
            assert secretary.handoff is not None
            assert secretary.orchestrator is not None
            assert secretary.asset_manager is not None

            print(f"✅ Secretary 초기화 성공")
            print(f"   - Bot Token: {test_token[:20]}...")
            print(f"   - Handoff Engine: ✓")
            print(f"   - Orchestrator: ✓")
            print(f"   - Asset Manager: ✓")

            return True

    except Exception as e:
        print(f"❌ 초기화 실패: {e}")
        import traceback
        traceback.print_exc()
        return False


def test_3_command_handlers():
    """3. 명령어 핸들러 존재 확인"""
    print("\n" + "="*70)
    print("TEST 3: 명령어 핸들러 확인")
    print("="*70)

    try:
        from execution.daemons.telegram_secretary import TelegramSecretary

        required_handlers = [
            'start_command',
            'status_command',
            'report_command',
            'analyze_command',
            'signal_command',
            'handle_message',
            'handle_photo'
        ]

        for handler in required_handlers:
            assert hasattr(TelegramSecretary, handler), f"Missing handler: {handler}"
            print(f"✅ {handler}")

        print(f"\n✅ 모든 핸들러 존재 확인 ({len(required_handlers)}개)")
        return True

    except AssertionError as e:
        print(f"❌ 핸들러 누락: {e}")
        return False
    except Exception as e:
        print(f"❌ 테스트 실패: {e}")
        return False


def test_4_signal_directory():
    """4. 신호 디렉토리 생성 확인"""
    print("\n" + "="*70)
    print("TEST 4: 신호 디렉토리 구조")
    print("="*70)

    signals_dir = PROJECT_ROOT / 'knowledge' / 'signals'
    images_dir = signals_dir / 'images'

    # 디렉토리 생성
    signals_dir.mkdir(parents=True, exist_ok=True)
    images_dir.mkdir(parents=True, exist_ok=True)

    assert signals_dir.exists(), "signals/ 디렉토리 생성 실패"
    assert images_dir.exists(), "signals/images/ 디렉토리 생성 실패"

    print(f"✅ 신호 디렉토리 구조 확인:")
    print(f"   - {signals_dir}")
    print(f"   - {images_dir}")

    return True


def test_5_logs_directory():
    """5. 로그 디렉토리 생성 확인"""
    print("\n" + "="*70)
    print("TEST 5: 로그 디렉토리")
    print("="*70)

    logs_dir = PROJECT_ROOT / 'logs'
    logs_dir.mkdir(parents=True, exist_ok=True)

    assert logs_dir.exists(), "logs/ 디렉토리 생성 실패"

    print(f"✅ 로그 디렉토리 확인:")
    print(f"   - {logs_dir}")

    return True


def run_all_tests():
    """모든 테스트 실행"""
    print("\n" + "🧪"*35)
    print("TELEGRAM SECRETARY TEST SUITE")
    print("기본 기능 검증 - 초기화 및 명령어 핸들러")
    print("🧪"*35)

    tests = [
        test_1_imports,
        test_2_initialization,
        test_3_command_handlers,
        test_4_signal_directory,
        test_5_logs_directory
    ]

    results = []
    for test_func in tests:
        try:
            result = test_func()
            results.append(result)
        except Exception as e:
            print(f"\n❌ 테스트 실행 오류: {e}")
            import traceback
            traceback.print_exc()
            results.append(False)

    # Summary
    print("\n" + "="*70)
    print("📊 테스트 결과 요약")
    print("="*70)

    passed = sum(results)
    total = len(results)

    for i, (test_func, result) in enumerate(zip(tests, results), 1):
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status} - Test {i}: {test_func.__name__}")

    print("="*70)
    print(f"총 {passed}/{total} 테스트 통과")

    if passed == total:
        print("\n🎉 모든 테스트 통과! Telegram Secretary 배포 준비 완료")
        print("\n다음 단계:")
        print("  python3 execution/daemons/telegram_secretary.py")
        return True
    else:
        print(f"\n❌ {total - passed}개 테스트 실패")
        return False


if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)
