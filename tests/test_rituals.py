import sys
import unittest
from datetime import datetime
from unittest.mock import MagicMock, patch

# Add project root to sys.path
import os
from pathlib import Path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from execution.technical_daemon import _check_rituals

class TestRituals(unittest.TestCase):
    def setUp(self):
        self.mock_config = {
            "TEST_RITUAL": {
                "trigger_hour": 9,
                "agent": "Test_Agent",
                "task_type": "TEST",
                "instruction": "Test Instruction",
                "council": False
            }
        }

    @patch('libs.core_config.RITUALS_CONFIG', new_callable=dict)
    @patch('execution.technical_daemon.datetime')
    @patch('execution.technical_daemon._get_chat_ids', return_value=[])
    @patch('execution.technical_daemon._telegram_send')
    def test_ritual_triggers_at_correct_time(self, mock_send, mock_chat_ids, mock_datetime, mock_config_dict):
        # Setup Mock Config
        mock_config_dict.update(self.mock_config)
        
        # Mock Time: 09:00 AM (Should Trigger)
        mock_now = datetime(2026, 2, 13, 9, 0, 0)
        mock_datetime.now.return_value = mock_now
        
        status = {"pending_tasks": [], "rituals_log": {}}
        
        _check_rituals(status)
        
        # Verify Task Added
        self.assertEqual(len(status["pending_tasks"]), 1)
        self.assertEqual(status["pending_tasks"][0]["type"], "TEST")
        # Verify Log Updated
        self.assertEqual(status["rituals_log"]["TEST_RITUAL"], "2026-02-13")

    @patch('libs.core_config.RITUALS_CONFIG', new_callable=dict)
    @patch('execution.technical_daemon.datetime')
    @patch('execution.technical_daemon._get_chat_ids', return_value=[])
    def test_ritual_does_not_trigger_wrong_time(self, mock_chat_ids, mock_datetime, mock_config_dict):
        # Setup Mock Config
        mock_config_dict.update(self.mock_config)
        
        # Mock Time: 10:00 AM (Should NOT Trigger)
        mock_now = datetime(2026, 2, 13, 10, 0, 0)
        mock_datetime.now.return_value = mock_now
        
        status = {"pending_tasks": [], "rituals_log": {}}
        
        _check_rituals(status)
        
        # Verify Task NOT Added
        self.assertEqual(len(status["pending_tasks"]), 0)

    @patch('libs.core_config.RITUALS_CONFIG', new_callable=dict)
    @patch('execution.technical_daemon.datetime')
    @patch('execution.technical_daemon._get_chat_ids', return_value=[])
    def test_ritual_idempotency(self, mock_chat_ids, mock_datetime, mock_config_dict):
        # Setup Mock Config
        mock_config_dict.update(self.mock_config)
        
        # Mock Time: 09:00 AM
        mock_now = datetime(2026, 2, 13, 9, 0, 0)
        mock_datetime.now.return_value = mock_now
        
        # Status indicates already run today
        status = {
            "pending_tasks": [], 
            "rituals_log": {"TEST_RITUAL": "2026-02-13"}
        }
        
        _check_rituals(status)
        
        # Verify Task NOT Added (Idempotency)
        self.assertEqual(len(status["pending_tasks"]), 0)

if __name__ == '__main__':
    unittest.main()
