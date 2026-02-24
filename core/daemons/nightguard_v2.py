#!/usr/bin/env python3
"""
97layerOS Nightguard V2: Self-Diagnostic Autonomous Daemon
Purpose: Monitor system health, detect failures, auto-recover, and alert admin
Philosophy: Intelligence Autonomy - self-awareness and self-healing

Critical Responsibilities:
1. Authentication Monitoring (Cookie Risk mitigation)
2. API Quota Tracking (Gemini, Anthropic)
3. Service Health Checks (Telegram bot, MCP servers)
4. Auto-Recovery (restart failed services)
5. Admin Alerts (Telegram notifications)

Author: 97layerOS Technical Director
Created: 2026-02-16
Priority: P0 (Critical)
"""

import os
import sys
import json
import time
import asyncio
import subprocess
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import logging

# Add project root to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Environment variables
from dotenv import load_dotenv
load_dotenv(PROJECT_ROOT / '.env')

# Telegram for admin alerts
try:
    from telegram import Bot
    TELEGRAM_AVAILABLE = True
except ImportError:
    TELEGRAM_AVAILABLE = False

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [Nightguard] %(levelname)s: %(message)s',
    handlers=[
        logging.FileHandler(PROJECT_ROOT / 'knowledge' / 'system' / 'nightguard.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)


class NightguardV2:
    """
    Self-diagnostic autonomous daemon

    Monitors:
    - NotebookLM cookie expiration
    - Gemini API quota
    - Anthropic API quota
    - Telegram bot health
    - MCP server health
    - Disk space
    - Memory usage

    Actions:
    - Send Telegram alerts
    - Auto-restart services
    - Log all incidents
    """

    def __init__(self, admin_telegram_id: Optional[str] = None):
        """
        Initialize Nightguard V2

        Args:
            admin_telegram_id: Telegram user ID for admin alerts
        """
        self.project_root = PROJECT_ROOT
        self.admin_id = admin_telegram_id or os.getenv('ADMIN_TELEGRAM_ID')
        self.bot_token = os.getenv('TELEGRAM_BOT_TOKEN')
        self.google_api_key = os.getenv('GOOGLE_API_KEY')
        self.anthropic_api_key = os.getenv('ANTHROPIC_API_KEY')

        # State files
        self.state_dir = PROJECT_ROOT / 'knowledge' / 'system'
        self.state_dir.mkdir(parents=True, exist_ok=True)
        self.state_file = self.state_dir / 'nightguard_state.json'
        self.cookie_file = self.state_dir / 'notebooklm_cookie.json'

        # Initialize Telegram bot for alerts
        self.bot = None
        if TELEGRAM_AVAILABLE and self.bot_token:
            try:
                self.bot = Bot(token=self.bot_token)
            except Exception as e:
                logger.error(f"Failed to initialize Telegram bot: {e}")

        # Load state
        self.state = self._load_state()

        logger.info("🤖 Nightguard V2 initialized")

    def _load_state(self) -> Dict:
        """Load persistent state"""
        if not self.state_file.exists():
            return {
                'last_check': None,
                'alerts_sent': [],
                'incidents': []
            }

        try:
            with open(self.state_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            logger.error(f"Failed to load state: {e}")
            return {
                'last_check': None,
                'alerts_sent': [],
                'incidents': []
            }

    def _save_state(self):
        """Save persistent state"""
        try:
            with open(self.state_file, 'w', encoding='utf-8') as f:
                json.dump(self.state, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to save state: {e}")

    async def send_alert(self, message: str, critical: bool = False):
        """
        Send alert to admin via Telegram

        Args:
            message: Alert message
            critical: If True, prepend 🚨, else ⚠️
        """
        emoji = "🚨" if critical else "⚠️"
        full_message = f"{emoji} 나이트가드 알림\n\n{message}\n\n🕐 {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"

        logger.warning(full_message)

        # Send Telegram alert
        if self.bot and self.admin_id:
            try:
                await self.bot.send_message(
                    chat_id=self.admin_id,
                    text=full_message,
                    parse_mode='Markdown'
                )
                logger.info(f"Alert sent to admin: {self.admin_id}")
            except Exception as e:
                logger.error(f"Failed to send Telegram alert: {e}")
        else:
            logger.warning("Telegram bot not configured - alert not sent")

        # Record alert
        self.state['alerts_sent'].append({
            'timestamp': datetime.now().isoformat(),
            'message': message,
            'critical': critical
        })
        self._save_state()

    def _log_incident(self, component: str, issue: str, action: str):
        """Log incident for audit trail"""
        incident = {
            'timestamp': datetime.now().isoformat(),
            'component': component,
            'issue': issue,
            'action': action
        }
        self.state['incidents'].append(incident)
        self._save_state()
        logger.info(f"Incident logged: {component} - {issue} - {action}")

    # ========================
    # Authentication Monitoring
    # ========================

    def check_notebooklm_cookie(self) -> Dict:
        """
        Check NotebookLM cookie expiration

        Returns:
            {
                'status': 'ok' | 'warning' | 'critical' | 'missing',
                'expires_in_hours': int,
                'message': str
            }
        """
        if not self.cookie_file.exists():
            return {
                'status': 'missing',
                'expires_in_hours': 0,
                'message': 'NotebookLM cookie file not found'
            }

        try:
            with open(self.cookie_file, 'r', encoding='utf-8') as f:
                cookie_data = json.load(f)

            # Check if cookie has expiration info
            if 'updated_at' not in cookie_data:
                return {
                    'status': 'warning',
                    'expires_in_hours': 0,
                    'message': 'Cookie file missing update timestamp'
                }

            # Google cookies typically expire after 14 days
            updated_at = datetime.fromisoformat(cookie_data['updated_at'])
            age_hours = (datetime.now() - updated_at).total_seconds() / 3600
            expires_in_hours = (14 * 24) - age_hours

            if expires_in_hours < 0:
                status = 'critical'
                message = f'Cookie EXPIRED {abs(expires_in_hours):.0f}h ago'
            elif expires_in_hours < 48:
                status = 'critical'
                message = f'Cookie expires in {expires_in_hours:.0f}h (< 48h)'
            elif expires_in_hours < 72:
                status = 'warning'
                message = f'Cookie expires in {expires_in_hours:.0f}h (< 72h)'
            else:
                status = 'ok'
                message = f'Cookie healthy - expires in {expires_in_hours:.0f}h'

            return {
                'status': status,
                'expires_in_hours': expires_in_hours,
                'message': message
            }

        except Exception as e:
            return {
                'status': 'warning',
                'expires_in_hours': 0,
                'message': f'Failed to check cookie: {str(e)}'
            }

    # ========================
    # API Quota Monitoring
    # ========================

    def check_gemini_quota(self) -> Dict:
        """
        Check Gemini API quota (approximate)

        Note: Google doesn't provide direct quota API, this is estimation based on usage logs

        Returns:
            {
                'status': 'ok' | 'warning' | 'critical',
                'estimated_quota_remaining': float,
                'message': str
            }
        """
        # TODO: Implement actual quota tracking based on usage logs
        # For now, just check if API key is present

        if not self.google_api_key:
            return {
                'status': 'critical',
                'estimated_quota_remaining': 0,
                'message': 'GOOGLE_API_KEY not configured'
            }

        # Placeholder: assume OK for now
        return {
            'status': 'ok',
            'estimated_quota_remaining': 100.0,
            'message': 'Gemini API key configured (quota tracking not implemented)'
        }

    def check_anthropic_quota(self) -> Dict:
        """
        Check Anthropic API quota

        Returns:
            {
                'status': 'ok' | 'warning' | 'critical',
                'estimated_quota_remaining': float,
                'message': str
            }
        """
        if not self.anthropic_api_key:
            return {
                'status': 'warning',
                'estimated_quota_remaining': 0,
                'message': 'ANTHROPIC_API_KEY not configured (CD agent will not work)'
            }

        # TODO: Implement actual quota tracking via Anthropic API
        # For now, just check if key is present

        return {
            'status': 'ok',
            'estimated_quota_remaining': 100.0,
            'message': 'Anthropic API key configured (quota tracking not implemented)'
        }

    # ========================
    # Service Health Checks
    # ========================

    def check_telegram_bot(self) -> Dict:
        """
        Check if Telegram bot is running

        Returns:
            {
                'status': 'ok' | 'critical',
                'process_id': int | None,
                'message': str
            }
        """
        try:
            # Check if telegram_secretary*.py is running
            result = subprocess.run(
                ['pgrep', '-f', 'telegram_secretary'],
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                pid = result.stdout.strip().split('\n')[0]
                return {
                    'status': 'ok',
                    'process_id': int(pid),
                    'message': f'Telegram bot running (PID: {pid})'
                }
            else:
                return {
                    'status': 'critical',
                    'process_id': None,
                    'message': 'Telegram bot NOT running'
                }

        except Exception as e:
            return {
                'status': 'warning',
                'process_id': None,
                'message': f'Failed to check Telegram bot: {str(e)}'
            }

    def check_mcp_server(self) -> Dict:
        """
        Check if NotebookLM MCP server is accessible

        Returns:
            {
                'status': 'ok' | 'warning' | 'critical',
                'message': str
            }
        """
        # Check if container is running
        try:
            result = subprocess.run(
                ['podman', 'ps', '--filter', 'name=97layer-os', '--format', '{{.Names}}'],
                capture_output=True,
                text=True,
                cwd=self.project_root
            )

            if '97layer-os' in result.stdout:
                return {
                    'status': 'ok',
                    'message': 'Podman container "97layer-os" running'
                }
            else:
                return {
                    'status': 'critical',
                    'message': 'Podman container "97layer-os" NOT running'
                }

        except Exception as e:
            return {
                'status': 'warning',
                'message': f'Failed to check MCP server: {str(e)}'
            }

    # ========================
    # System Resources
    # ========================

    def check_disk_space(self) -> Dict:
        """
        Check available disk space

        Returns:
            {
                'status': 'ok' | 'warning' | 'critical',
                'available_gb': float,
                'percent_used': float,
                'message': str
            }
        """
        try:
            import shutil
            usage = shutil.disk_usage(self.project_root)

            available_gb = usage.free / (1024**3)
            percent_used = (usage.used / usage.total) * 100

            if percent_used > 95:
                status = 'critical'
                message = f'Disk {percent_used:.1f}% full (< 5% free)'
            elif percent_used > 90:
                status = 'warning'
                message = f'Disk {percent_used:.1f}% full (< 10% free)'
            else:
                status = 'ok'
                message = f'Disk {percent_used:.1f}% full ({available_gb:.1f}GB free)'

            return {
                'status': status,
                'available_gb': available_gb,
                'percent_used': percent_used,
                'message': message
            }

        except Exception as e:
            return {
                'status': 'warning',
                'available_gb': 0,
                'percent_used': 0,
                'message': f'Failed to check disk space: {str(e)}'
            }

    # ========================
    # Auto-Recovery
    # ========================

    async def restart_telegram_bot(self) -> bool:
        """
        Attempt to restart Telegram bot via systemd (on GCP VM)

        Returns:
            True if restart successful, False otherwise
        """
        try:
            logger.info("Attempting to restart Telegram bot...")

            # Try systemd restart
            result = subprocess.run(
                ['sudo', 'systemctl', 'restart', '97layer-telegram'],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode == 0:
                logger.info("✅ Telegram bot restarted successfully")
                self._log_incident('telegram_bot', 'service_down', 'auto_restarted')
                await self.send_alert("✅ 텔레그램 봇이 중단되어 자동으로 재시작했습니다", critical=False)
                return True
            else:
                logger.error(f"Failed to restart Telegram bot: {result.stderr}")
                self._log_incident('telegram_bot', 'service_down', 'restart_failed')
                await self.send_alert(f"❌ 텔레그램 봇 중단 — 자동 재시작 실패\n\n오류: {result.stderr}", critical=True)
                return False

        except Exception as e:
            logger.error(f"Exception during Telegram bot restart: {e}")
            await self.send_alert(f"❌ 텔레그램 봇 중단 — 재시작 중 오류 발생: {str(e)}", critical=True)
            return False

    # ========================
    # Main Diagnostic Loop
    # ========================

    async def run_diagnostics(self) -> Dict:
        """
        Run full system diagnostics

        Returns:
            {
                'timestamp': str,
                'overall_status': 'healthy' | 'degraded' | 'critical',
                'components': {
                    'notebooklm_cookie': {...},
                    'gemini_quota': {...},
                    'anthropic_quota': {...},
                    'telegram_bot': {...},
                    'mcp_server': {...},
                    'disk_space': {...}
                },
                'actions_taken': [...]
            }
        """
        logger.info("🔍 Running system diagnostics...")

        # Run all checks
        components = {
            'notebooklm_cookie': self.check_notebooklm_cookie(),
            'gemini_quota': self.check_gemini_quota(),
            'anthropic_quota': self.check_anthropic_quota(),
            'telegram_bot': self.check_telegram_bot(),
            'mcp_server': self.check_mcp_server(),
            'disk_space': self.check_disk_space()
        }

        # Determine overall status
        critical_count = sum(1 for c in components.values() if c.get('status') == 'critical')
        warning_count = sum(1 for c in components.values() if c.get('status') == 'warning')

        if critical_count > 0:
            overall_status = 'critical'
        elif warning_count > 0:
            overall_status = 'degraded'
        else:
            overall_status = 'healthy'

        actions_taken = []

        # Handle critical issues

        # 1. NotebookLM Cookie
        cookie_status = components['notebooklm_cookie']
        if cookie_status['status'] == 'critical':
            await self.send_alert(
                f"🍪 노트북LM 로그인 만료\n\n{cookie_status['message']}\n\n"
                "조치 방법: 쿠키를 수동으로 갱신해주세요\n"
                "1. Chrome → notebooklm.google.com 접속\n"
                "2. DevTools(F12) → Application → Cookies 복사\n"
                "3. knowledge/system/notebooklm_cookie.json 업데이트",
                critical=True
            )
            actions_taken.append('alerted_admin_cookie_expiry')

        # 2. Telegram Bot
        bot_status = components['telegram_bot']
        if bot_status['status'] == 'critical':
            # Auto-restart
            restart_success = await self.restart_telegram_bot()
            if restart_success:
                actions_taken.append('restarted_telegram_bot')
                # Re-check
                components['telegram_bot'] = self.check_telegram_bot()
            else:
                actions_taken.append('telegram_restart_failed')

        # 3. MCP Server
        mcp_status = components['mcp_server']
        if mcp_status['status'] == 'critical':
            await self.send_alert(
                f"🐳 맥북 연결 끊김\n\n{mcp_status['message']}\n\n"
                "조치 방법: 맥북에서 포드맨을 재시작해주세요\n"
                "`podman start 97layer-os`",
                critical=True
            )
            actions_taken.append('alerted_admin_mcp_down')

        # 4. Disk Space
        disk_status = components['disk_space']
        if disk_status['status'] == 'critical':
            await self.send_alert(
                f"💾 서버 저장공간 부족\n\n{disk_status['message']}\n\n"
                "조치 방법: 불필요한 파일을 정리해주세요\n"
                "확인 위치: knowledge/signals/, logs/",
                critical=True
            )
            actions_taken.append('alerted_admin_disk_space')

        # Update state
        self.state['last_check'] = datetime.now().isoformat()
        self._save_state()

        result = {
            'timestamp': datetime.now().isoformat(),
            'overall_status': overall_status,
            'components': components,
            'actions_taken': actions_taken
        }

        logger.info(f"✅ Diagnostics complete - Status: {overall_status.upper()}")

        return result

    async def run_daemon(self, interval_minutes: int = 30):
        """
        Run Nightguard as a continuous daemon

        Args:
            interval_minutes: Check interval in minutes (default: 30)
        """
        logger.info(f"🛡️  Nightguard V2 daemon starting...")
        logger.info(f"   Check interval: {interval_minutes} minutes")
        logger.info(f"   Admin Telegram ID: {self.admin_id or 'NOT CONFIGURED'}")

        if not self.admin_id:
            logger.warning("⚠️  Admin Telegram ID not configured - alerts will only be logged")

        try:
            while True:
                result = await self.run_diagnostics()

                # Sleep until next check
                await asyncio.sleep(interval_minutes * 60)

        except KeyboardInterrupt:
            logger.info("\n🛑 Nightguard daemon stopped by user")
        except Exception as e:
            logger.error(f"❌ Nightguard daemon crashed: {e}")
            raise


async def main():
    """CLI entry point"""
    import argparse

    parser = argparse.ArgumentParser(description='97layerOS Nightguard V2 - Self-Diagnostic Daemon')
    parser.add_argument(
        '--interval',
        type=int,
        default=30,
        help='Check interval in minutes (default: 30)'
    )
    parser.add_argument(
        '--once',
        action='store_true',
        help='Run diagnostics once and exit (no daemon mode)'
    )
    parser.add_argument(
        '--admin-id',
        type=str,
        help='Telegram user ID for admin alerts (or set ADMIN_TELEGRAM_ID in .env)'
    )

    args = parser.parse_args()

    nightguard = NightguardV2(admin_telegram_id=args.admin_id)

    if args.once:
        result = await nightguard.run_diagnostics()
        print("\n" + "="*80)
        print("NIGHTGUARD DIAGNOSTICS REPORT".center(80))
        print("="*80)
        print(f"\n⏰ Timestamp: {result['timestamp']}")
        print(f"📊 Overall Status: {result['overall_status'].upper()}")
        print(f"\n🔍 Component Status:")
        for component, status in result['components'].items():
            emoji = {'ok': '✅', 'warning': '⚠️', 'critical': '🚨', 'missing': '❓'}.get(status['status'], '•')
            print(f"  {emoji} {component}: {status['message']}")
        if result['actions_taken']:
            print(f"\n🛠️  Actions Taken: {', '.join(result['actions_taken'])}")
        print("\n" + "="*80)
    else:
        await nightguard.run_daemon(interval_minutes=args.interval)


if __name__ == "__main__":
    asyncio.run(main())
