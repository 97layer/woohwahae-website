#!/usr/bin/env python3
"""
LAYER OS Real-Time Monitoring Dashboard
Purpose: Display current system status, active tasks, and work progress
Philosophy: Transparency in slow life - see the process, not just results

Author: LAYER OS Technical Director
Created: 2026-02-16
"""

import os
import sys
import json
import time
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

try:
    from core.system.progress_graph import build_progress_payload
except Exception:  # noqa: BLE001
    build_progress_payload = None


class MonitorDashboard:
    """
    Real-time system monitoring dashboard

    Displays:
    - Current work lock status (who's working on what)
    - Recent file changes
    - Asset manager statistics
    - Ralph Loop validation stats
    - Daily routine progress
    - Git status
    """

    def __init__(self):
        """Initialize dashboard"""
        self.project_root = PROJECT_ROOT
        self.handoff_file = PROJECT_ROOT / 'knowledge' / 'system' / 'handoff.json'
        self.asset_registry = PROJECT_ROOT / 'knowledge' / 'system' / 'asset_registry.json'
        self.ralph_log = PROJECT_ROOT / 'knowledge' / 'system' / 'ralph_validations.jsonl'
        self.reports_dir = PROJECT_ROOT / 'knowledge' / 'reports' / 'daily'
        self.plan_dispatch_daily_dir = PROJECT_ROOT / 'knowledge' / 'system' / 'plan_dispatch_daily'

    def clear_screen(self):
        """Clear terminal screen"""
        os.system('clear' if os.name != 'nt' else 'cls')

    @staticmethod
    def _display_path(path: Path) -> str:
        try:
            return str(path.relative_to(PROJECT_ROOT))
        except ValueError:
            return str(path)

    def get_work_lock_status(self) -> Dict:
        """Get current work lock status from handoff.json"""
        if not self.handoff_file.exists():
            return {'locked': False}

        try:
            with open(self.handoff_file, 'r', encoding='utf-8') as f:
                handoff = json.load(f)
                return handoff.get('work_lock', {'locked': False})
        except Exception:
            return {'locked': False}

    def get_asset_stats(self) -> Dict:
        """Get asset manager statistics"""
        if not self.asset_registry.exists():
            return {'total': 0, 'by_status': {}}

        try:
            with open(self.asset_registry, 'r', encoding='utf-8') as f:
                registry = json.load(f)
                return registry.get('stats', {'total': 0, 'by_status': {}})
        except Exception:
            return {'total': 0, 'by_status': {}}

    def get_ralph_stats(self) -> Dict:
        """Get Ralph Loop validation statistics"""
        if not self.ralph_log.exists():
            return {'total': 0, 'passed': 0, 'failed': 0}

        try:
            total = 0
            passed = 0
            failed = 0

            with open(self.ralph_log, 'r', encoding='utf-8') as f:
                for line in f:
                    if line.strip():
                        total += 1
                        entry = json.loads(line)
                        if entry.get('decision') == 'pass':
                            passed += 1
                        elif entry.get('decision') == 'archive':
                            failed += 1

            return {
                'total': total,
                'passed': passed,
                'failed': failed,
                'pass_rate': round(passed / total * 100, 1) if total > 0 else 0
            }
        except Exception:
            return {'total': 0, 'passed': 0, 'failed': 0, 'pass_rate': 0}

    def get_daily_reports(self) -> Dict:
        """Get today's daily reports status"""
        if not self.reports_dir.exists():
            return {'morning': False, 'evening': False}

        today = datetime.now().strftime('%Y%m%d')
        morning_file = self.reports_dir / f'morning_{today}.json'
        evening_file = self.reports_dir / f'evening_{today}.json'

        return {
            'morning': morning_file.exists(),
            'evening': evening_file.exists()
        }

    def get_plan_dispatch_daily_status(self) -> Dict:
        """Get latest plan_dispatch daily report status."""
        if not self.plan_dispatch_daily_dir.exists():
            return {'available': False}

        candidates = sorted(self.plan_dispatch_daily_dir.glob('plan_dispatch_*.json'))
        if not candidates:
            return {'available': False}

        latest = candidates[-1]
        try:
            payload = json.loads(latest.read_text(encoding='utf-8'))
        except Exception:
            return {
                'available': True,
                'status': 'invalid',
                'file': self._display_path(latest),
            }

        health = payload.get('health', {}) if isinstance(payload, dict) else {}
        metrics = payload.get('metrics', {}) if isinstance(payload, dict) else {}
        replay = payload.get('replay', {}) if isinstance(payload, dict) else {}
        status = str(health.get('status', 'unknown')).lower()
        generated_at = str(payload.get('generated_at', ''))[:19] if isinstance(payload, dict) else ''
        return {
            'available': True,
            'status': status,
            'generated_at': generated_at,
            'fallback_rate': float(metrics.get('fallback_rate', 0.0)),
            'allowed_rate': float(replay.get('allowed_rate', 0.0)),
            'file': self._display_path(latest),
        }

    def get_progress_trend(self) -> Dict:
        """Get compact trend graph payload."""
        if build_progress_payload is None:
            return {"available": False}
        try:
            payload = build_progress_payload(limit=30, graph_width=28)
        except Exception:
            return {"available": False}
        if not isinstance(payload, dict):
            return {"available": False}
        payload["available"] = True
        return payload

    def get_recent_files(self, limit: int = 5) -> List[Dict]:
        """Get recently modified files"""
        try:
            # Get recent files from execution/ and knowledge/
            files = []
            for pattern in ['execution/**/*.py', 'knowledge/**/*.md', 'knowledge/**/*.json']:
                for file in PROJECT_ROOT.glob(pattern):
                    if file.is_file() and '.git' not in str(file):
                        files.append({
                            'path': str(file.relative_to(PROJECT_ROOT)),
                            'modified': file.stat().st_mtime
                        })

            # Sort by modification time
            files.sort(key=lambda x: x['modified'], reverse=True)

            # Format timestamps
            for file in files[:limit]:
                mtime = datetime.fromtimestamp(file['modified'])
                file['modified'] = mtime.strftime('%Y-%m-%d %H:%M:%S')

            return files[:limit]
        except Exception:
            return []

    def get_git_status(self) -> Dict:
        """Get git repository status"""
        try:
            import subprocess

            # Get current branch
            branch_result = subprocess.run(
                ['git', 'branch', '--show-current'],
                capture_output=True,
                text=True,
                cwd=PROJECT_ROOT
            )
            branch = branch_result.stdout.strip()

            # Get uncommitted changes count
            status_result = subprocess.run(
                ['git', 'status', '--porcelain'],
                capture_output=True,
                text=True,
                cwd=PROJECT_ROOT
            )
            changes = len([l for l in status_result.stdout.strip().split('\n') if l])

            # Get recent commit
            log_result = subprocess.run(
                ['git', 'log', '-1', '--oneline'],
                capture_output=True,
                text=True,
                cwd=PROJECT_ROOT
            )
            recent_commit = log_result.stdout.strip()

            return {
                'branch': branch,
                'uncommitted_changes': changes,
                'recent_commit': recent_commit
            }
        except Exception:
            return {
                'branch': 'unknown',
                'uncommitted_changes': 0,
                'recent_commit': 'N/A'
            }

    def render_dashboard(self):
        """Render complete dashboard"""
        self.clear_screen()

        # Header
        now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        print("=" * 80)
        print("  LAYER OS Real-Time Monitoring Dashboard".center(80))
        print(f"  {now}".center(80))
        print("=" * 80)
        print()

        # Work Lock Status
        lock_status = self.get_work_lock_status()
        print("━━━ 🔒 Work Lock Status ━━━")
        if lock_status.get('locked'):
            print(f"  🟢 Active: {lock_status.get('agent', 'Unknown')}")
            print(f"     Task: {lock_status.get('task', 'Unknown')}")
            print(f"     Since: {lock_status.get('acquired_at', 'Unknown')[:19]}")
            resources = lock_status.get('resources', [])
            if resources:
                print(f"     Resources: {', '.join(resources[:2])}")
        else:
            print("  ⚪ Idle (No active work)")
        print()

        # Asset Manager Statistics
        asset_stats = self.get_asset_stats()
        print("━━━ 📦 Asset Manager ━━━")
        print(f"  Total Assets: {asset_stats.get('total', 0)}")
        by_status = asset_stats.get('by_status', {})
        if by_status:
            for status, count in by_status.items():
                emoji = {
                    'captured': '📥',
                    'analyzed': '🔍',
                    'refined': '✏️',
                    'validated': '✅',
                    'approved': '🎯',
                    'published': '🚀',
                    'archived': '📁'
                }.get(status, '•')
                print(f"  {emoji} {status.capitalize()}: {count}")
        print()

        # Ralph Loop Statistics
        ralph_stats = self.get_ralph_stats()
        print("━━━ 🔄 Ralph Loop (Quality Validation) ━━━")
        print(f"  Total Validations: {ralph_stats.get('total', 0)}")
        print(f"  ✅ Passed: {ralph_stats.get('passed', 0)}")
        print(f"  ❌ Failed: {ralph_stats.get('failed', 0)}")
        print(f"  📊 Pass Rate: {ralph_stats.get('pass_rate', 0)}%")
        print()

        # Daily Reports
        reports = self.get_daily_reports()
        print("━━━ 📅 Daily Routine ━━━")
        morning_status = "✅" if reports['morning'] else "⏳"
        evening_status = "✅" if reports['evening'] else "⏳"
        print(f"  {morning_status} Morning Briefing (09:00)")
        print(f"  {evening_status} Evening Report (21:00)")
        print()

        # plan_dispatch Daily Health
        pd_status = self.get_plan_dispatch_daily_status()
        print("━━━ 🩺 Plan Dispatch Daily ━━━")
        if not pd_status.get('available'):
            print("  ⏳ Daily health report not generated")
        else:
            status = str(pd_status.get('status', 'unknown')).lower()
            icon = {'pass': '✅', 'warn': '⚠️', 'fail': '❌', 'invalid': '❌'}.get(status, '•')
            print(f"  {icon} Health: {status}")
            print(f"  Fallback Rate: {pd_status.get('fallback_rate', 0.0):.3f}")
            print(f"  Allowed Rate: {pd_status.get('allowed_rate', 0.0):.3f}")
            if pd_status.get('generated_at'):
                print(f"  Generated: {pd_status.get('generated_at')}")
            print(f"  File: {pd_status.get('file')}")
        print()

        # Progress Trend
        trend = self.get_progress_trend()
        print("━━━ 📈 Progress Trend ━━━")
        if not trend.get('available'):
            print("  ⏳ Trend graph unavailable")
        else:
            graphs = trend.get('graphs', {})
            metrics = trend.get('metrics', {})
            score = metrics.get('score', {})
            fallback = metrics.get('fallback_rate', {})
            blocked = metrics.get('blocked_rate', {})
            print(f"  Score    {graphs.get('score', '·')}  {score.get('latest', 0):.1f}")
            print(
                f"  Fallback {graphs.get('fallback_rate', '·')}  "
                f"{float(fallback.get('latest', 0.0))*100:.1f}%"
            )
            print(
                f"  Blocked  {graphs.get('blocked_rate', '·')}  "
                f"{float(blocked.get('latest', 0.0))*100:.1f}%"
            )
        print()

        # Recent File Changes
        recent_files = self.get_recent_files(limit=5)
        print("━━━ 📝 Recent File Changes ━━━")
        if recent_files:
            for file in recent_files:
                path = file['path'][:60]  # Truncate long paths
                print(f"  • {path}")
                print(f"    {file['modified']}")
        else:
            print("  (No recent changes)")
        print()

        # Git Status
        git_status = self.get_git_status()
        print("━━━ 🔀 Git Repository ━━━")
        print(f"  Branch: {git_status['branch']}")
        print(f"  Uncommitted Changes: {git_status['uncommitted_changes']}")
        print(f"  Recent Commit: {git_status['recent_commit']}")
        print()

        # Footer
        print("=" * 80)
        print("  💡 Slow Life Reminder: 속도보다 방향, 효율보다 본질".center(80))
        print("  Press Ctrl+C to exit".center(80))
        print("=" * 80)

    def run(self, refresh_interval: int = 5):
        """
        Run dashboard in continuous mode

        Args:
            refresh_interval: Refresh every N seconds
        """
        print(f"🖥️  Starting LAYER OS Monitoring Dashboard...")
        print(f"   Refreshing every {refresh_interval} seconds")
        print()
        time.sleep(2)

        try:
            while True:
                self.render_dashboard()
                time.sleep(refresh_interval)
        except KeyboardInterrupt:
            print("\n\n✅ Dashboard stopped.")
            sys.exit(0)


def main():
    """CLI entry point"""
    import argparse

    parser = argparse.ArgumentParser(description='LAYER OS Real-Time Monitoring Dashboard')
    parser.add_argument(
        '--refresh',
        type=int,
        default=5,
        help='Refresh interval in seconds (default: 5)'
    )
    parser.add_argument(
        '--once',
        action='store_true',
        help='Display once and exit (no continuous refresh)'
    )

    args = parser.parse_args()

    dashboard = MonitorDashboard()

    if args.once:
        dashboard.render_dashboard()
    else:
        dashboard.run(refresh_interval=args.refresh)


if __name__ == "__main__":
    main()
