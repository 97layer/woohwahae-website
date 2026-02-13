import sys
import unittest
import json
from unittest.mock import MagicMock, patch
from pathlib import Path

# Add project root to sys.path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

# Mock AI engine before importing technical_daemon
with patch('libs.ai_engine.AIEngine'):
    from execution.technical_daemon import _process_remote_commands

class TestTelegramParser(unittest.TestCase):
    @patch('urllib.request.urlopen')
    @patch('execution.technical_daemon._get_ai')
    @patch('execution.technical_daemon._load_status')
    @patch('execution.technical_daemon._save_status')
    @patch('execution.technical_daemon._telegram_send')
    def test_parse_natural_command(self, mock_send, mock_save, mock_load, mock_ai, mock_urlopen):
        # 1. Setup Mock Telegram Response
        mock_response = MagicMock()
        mock_response.read.return_value = json.dumps({
            "ok": True,
            "result": [{
                "update_id": 12345,
                "message": {
                    "chat": {"id": 999},
                    "text": "내일 시장 분석 보고서 작성해줘"
                }
            }]
        }).encode()
        mock_response.__enter__.return_value = mock_response
        mock_urlopen.return_value = mock_response

        # 2. Setup Mock Status
        mock_load.return_value = {
            "pending_tasks": [],
            "last_telegram_update_id": 0
        }

        # 3. Setup Mock AI Response (Task JSON)
        mock_ai_instance = MagicMock()
        mock_ai_instance.generate_response.return_value = json.dumps({
            "type": "ANALYSIS",
            "agent": "SA",
            "instruction": "시장 분석 보고서 작성",
            "council": False
        })
        mock_ai.return_value = mock_ai_instance

        # 4. Execute
        _process_remote_commands()

        # 5. Assertions
        mock_save.assert_called_once()
        saved_status = mock_save.call_args[0][0]
        self.assertEqual(len(saved_status["pending_tasks"]), 1)
        self.assertEqual(saved_status["pending_tasks"][0]["agent"], "SA")
        self.assertEqual(saved_status["last_telegram_update_id"], 12345)
        mock_send.assert_called_with(999, unittest.mock.ANY)

if __name__ == '__main__':
    unittest.main()
