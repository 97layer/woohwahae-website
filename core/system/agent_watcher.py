#!/usr/bin/env python3
"""
97layerOS Agent Watcher - Queue Monitoring & Task Dispatch
Phase 6.1: Autonomous agent task claiming

Architecture:
- Agents run this watcher in a loop
- Watch for pending tasks matching their type
- Claim and execute tasks independently
- No central orchestrator dependency (decentralized)

Usage:
  # In agent container/process
  from core.system.agent_watcher import AgentWatcher

  watcher = AgentWatcher(agent_type='SA', agent_id='sa-worker-1')
  watcher.watch(callback=process_task, interval=5)

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import time
import signal
import sys
from pathlib import Path
from typing import Callable, Optional
from datetime import datetime

# Add project root to path
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.queue_manager import QueueManager, Task, EventType


class AgentWatcher:
    """
    Agent-side queue watcher for autonomous task claiming

    Features:
    - Poll pending queue for matching tasks
    - Atomic task claiming (race-condition safe)
    - Callback-based task execution
    - Graceful shutdown (SIGINT/SIGTERM)
    """

    def __init__(self, agent_type: str, agent_id: str, queue_root: Optional[Path] = None):
        """
        Initialize agent watcher

        Args:
            agent_type: Agent type (SA, AD, CE, CD, Ralph)
            agent_id: Unique agent instance ID (e.g., sa-worker-1)
            queue_root: Queue directory (default: .infra/queue)
        """
        self.agent_type = agent_type
        self.agent_id = agent_id
        self.queue = QueueManager(queue_root)
        self.running = False

        # Register signal handlers for graceful shutdown
        signal.signal(signal.SIGINT, self._shutdown_handler)
        signal.signal(signal.SIGTERM, self._shutdown_handler)

    def _shutdown_handler(self, signum, frame):
        """Handle shutdown signals gracefully"""
        print(f"\n🛑 {self.agent_id}: Received shutdown signal, stopping...")
        self.running = False

    def watch(self, callback: Callable[[Task], dict], interval: int = 5, max_iterations: Optional[int] = None):
        """
        Start watching for tasks (blocking loop)

        Args:
            callback: Function to execute when task claimed
                      Must return dict with result (or raise exception for failure)
            interval: Poll interval in seconds (default: 5)
            max_iterations: Stop after N iterations (None = infinite)

        Example callback:
            def process_task(task: Task) -> dict:
                # Do work
                return {'result': 'success', 'data': {...}}
        """
        self.running = True
        iterations = 0

        # Emit agent ready event
        self.queue.emit_event(
            event_type=EventType.AGENT_READY,
            agent_id=self.agent_id,
            payload={'agent_type': self.agent_type}
        )

        print(f"👁️  {self.agent_id}: Watching for {self.agent_type} tasks (interval: {interval}s)...")

        while self.running:
            if max_iterations and iterations >= max_iterations:
                print(f"✅ {self.agent_id}: Max iterations reached, stopping")
                break

            try:
                # Get pending tasks for this agent type
                pending_tasks = self.queue.get_pending_tasks(agent_type=self.agent_type)

                if pending_tasks:
                    print(f"📋 {self.agent_id}: Found {len(pending_tasks)} pending task(s)")

                    # Try to claim the first available task
                    for task in pending_tasks:
                        claimed_task = self.queue.claim_task(
                            agent_id=self.agent_id,
                            task_id=task.task_id
                        )

                        if claimed_task:
                            print(f"✅ {self.agent_id}: Claimed task {claimed_task.task_id}")
                            self._execute_task(claimed_task, callback)
                            break  # Process one task per iteration
                        else:
                            print(f"⏭️  {self.agent_id}: Task {task.task_id} already claimed by another agent")

                # Emit idle event if no tasks processed
                if not pending_tasks or not any(
                    self.queue.claim_task(self.agent_id, t.task_id) for t in pending_tasks
                ):
                    self.queue.emit_event(
                        event_type=EventType.AGENT_IDLE,
                        agent_id=self.agent_id,
                        payload={'agent_type': self.agent_type}
                    )

            except Exception as e:
                print(f"❌ {self.agent_id}: Error in watch loop: {e}")

            # Wait before next poll
            time.sleep(interval)
            iterations += 1

        print(f"👋 {self.agent_id}: Stopped watching")

    def _execute_task(self, task: Task, callback: Callable[[Task], dict]):
        """
        Execute task via callback and update queue

        Args:
            task: Claimed task
            callback: Task execution function
        """
        print(f"🚀 {self.agent_id}: Executing task {task.task_id} ({task.task_type})")

        # Emit busy event
        self.queue.emit_event(
            event_type=EventType.AGENT_BUSY,
            agent_id=self.agent_id,
            payload={'task_id': task.task_id, 'task_type': task.task_type}
        )

        start_time = time.time()

        try:
            # Execute callback
            result = callback(task)

            # Mark task as completed
            self.queue.complete_task(task.task_id, result)

            elapsed = time.time() - start_time
            print(f"✅ {self.agent_id}: Task {task.task_id} completed in {elapsed:.2f}s")

        except Exception as e:
            # Mark task as failed
            error_msg = f"{type(e).__name__}: {str(e)}"
            self.queue.fail_task(task.task_id, error_msg)

            elapsed = time.time() - start_time
            print(f"❌ {self.agent_id}: Task {task.task_id} failed after {elapsed:.2f}s: {error_msg}")


# ================== Example Usage ==================

def example_callback(task: Task) -> dict:
    """Example task processor"""
    print(f"  📝 Processing: {task.payload}")

    # Simulate work
    time.sleep(2)

    # Return result
    return {
        'status': 'analyzed',
        'insights': ['Insight 1', 'Insight 2'],
        'score': 85
    }


if __name__ == '__main__':
    print("🚀 Agent Watcher Test")
    print("=" * 50)

    # Create test queue
    queue = QueueManager()

    # Create some test tasks
    print("\n📝 Creating test tasks...")
    task1_id = queue.create_task(
        agent_type='SA',
        task_type='analyze_signal',
        payload={'signal_id': 'test_001', 'content': 'Test signal 1'}
    )
    task2_id = queue.create_task(
        agent_type='SA',
        task_type='analyze_signal',
        payload={'signal_id': 'test_002', 'content': 'Test signal 2'}
    )
    print(f"✅ Created 2 tasks: {task1_id}, {task2_id}")

    # Start watcher (will process both tasks then stop)
    watcher = AgentWatcher(agent_type='SA', agent_id='sa-test-worker')

    print("\n👁️  Starting watcher...")
    watcher.watch(callback=example_callback, interval=1, max_iterations=3)

    # Check events
    print("\n📡 Events:")
    events = queue.get_events()
    for event in events[-5:]:  # Last 5 events
        print(f"  - {event.event_type.value}: {event.agent_id}")

    print("\n✅ Agent Watcher test complete!")
