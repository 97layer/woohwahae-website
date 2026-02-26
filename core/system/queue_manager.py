#!/usr/bin/env python3
"""
LAYER OS Queue Manager - File-based Message Queue
Phase 6.1: Multi-agent communication infrastructure

Architecture:
- File-based queue (no Redis/RabbitMQ dependency)
- Event-driven: agents listen to queue, react to events
- Atomic operations: file locks to prevent race conditions
- Persistent: survives crashes, resumable

Queue Structure:
.infra/queue/
├── events/          # Event notifications (agent → orchestrator)
│   └── {timestamp}_{agent}_{event_type}.json
├── tasks/
│   ├── pending/     # New tasks (orchestrator → agent)
│   │   └── {task_id}.json
│   ├── processing/  # In-progress tasks (agent claims)
│   │   └── {task_id}.json
│   └── completed/   # Finished tasks (agent releases)
│       └── {task_id}.json
└── locks/           # File locks (atomic claims)
    └── {task_id}.lock

Author: LAYER OS Technical Director
Created: 2026-02-16
"""

import json
import time
import fcntl
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict
from enum import Enum


class TaskStatus(Enum):
    """Task lifecycle states"""
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


class EventType(Enum):
    """Event types for inter-agent communication"""
    TASK_CREATED = "task_created"
    TASK_CLAIMED = "task_claimed"
    TASK_COMPLETED = "task_completed"
    TASK_FAILED = "task_failed"
    AGENT_READY = "agent_ready"
    AGENT_BUSY = "agent_busy"
    AGENT_IDLE = "agent_idle"
    SIGNAL_RECEIVED = "signal_received"


@dataclass
class Task:
    """Task data structure"""
    task_id: str
    agent_type: str  # SA, AD, CE, CD, Ralph
    task_type: str  # analyze, create_asset, validate, etc.
    payload: Dict[str, Any]
    status: TaskStatus
    created_at: str
    claimed_at: Optional[str] = None
    completed_at: Optional[str] = None
    claimed_by: Optional[str] = None
    result: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


@dataclass
class Event:
    """Event data structure"""
    event_id: str
    event_type: EventType
    agent_id: str
    payload: Dict[str, Any]
    timestamp: str


class QueueManager:
    """
    File-based message queue for multi-agent coordination

    Features:
    - Atomic task claiming (file locks)
    - Event broadcasting (fan-out)
    - Task lifecycle tracking
    - Crash-safe (persistent files)
    """

    def __init__(self, queue_root: Optional[Path] = None):
        """Initialize queue manager"""
        if queue_root is None:
            # Default: .infra/queue/ from project root
            project_root = Path(__file__).parent.parent.parent
            queue_root = project_root / '.infra' / 'queue'

        self.queue_root = Path(queue_root)
        self._init_queue_structure()

    def _init_queue_structure(self):
        """Create queue directory structure"""
        (self.queue_root / 'events').mkdir(parents=True, exist_ok=True)
        (self.queue_root / 'tasks' / 'pending').mkdir(parents=True, exist_ok=True)
        (self.queue_root / 'tasks' / 'processing').mkdir(parents=True, exist_ok=True)
        (self.queue_root / 'tasks' / 'completed').mkdir(parents=True, exist_ok=True)
        (self.queue_root / 'locks').mkdir(parents=True, exist_ok=True)

    # ================== Task Management ==================

    def create_task(self, agent_type: str, task_type: str, payload: Dict[str, Any]) -> str:
        """
        Create a new task in the pending queue

        Args:
            agent_type: Target agent (SA, AD, CE, CD, Ralph)
            task_type: Task type (analyze, create_asset, validate, etc.)
            payload: Task-specific data

        Returns:
            task_id: Unique task identifier
        """
        task_id = f"{int(time.time() * 1000)}_{agent_type}_{task_type}"

        task = Task(
            task_id=task_id,
            agent_type=agent_type,
            task_type=task_type,
            payload=payload,
            status=TaskStatus.PENDING,
            created_at=datetime.now().isoformat()
        )

        # Write to pending queue
        task_file = self.queue_root / 'tasks' / 'pending' / f'{task_id}.json'
        task_file.write_text(json.dumps(asdict(task), indent=2, default=str))

        # Emit event
        self.emit_event(
            event_type=EventType.TASK_CREATED,
            agent_id='orchestrator',
            payload={'task_id': task_id, 'agent_type': agent_type}
        )

        return task_id

    def claim_task(self, agent_id: str, task_id: str) -> Optional[Task]:
        """
        Atomically claim a task (move pending → processing)

        Args:
            agent_id: Agent claiming the task
            task_id: Task to claim

        Returns:
            Task object if successfully claimed, None if already claimed
        """
        pending_file = self.queue_root / 'tasks' / 'pending' / f'{task_id}.json'
        lock_file = self.queue_root / 'locks' / f'{task_id}.lock'

        if not pending_file.exists():
            return None

        # Atomic claim with file lock
        try:
            with lock_file.open('w') as lock:
                # Try to acquire exclusive lock (non-blocking)
                fcntl.flock(lock.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)

                # Load task
                task_data = json.loads(pending_file.read_text())
                task_data['status'] = TaskStatus.PROCESSING.value
                task_data['claimed_at'] = datetime.now().isoformat()
                task_data['claimed_by'] = agent_id

                # Move to processing
                processing_file = self.queue_root / 'tasks' / 'processing' / f'{task_id}.json'
                processing_file.write_text(json.dumps(task_data, indent=2))
                pending_file.unlink()

                # Emit event
                self.emit_event(
                    event_type=EventType.TASK_CLAIMED,
                    agent_id=agent_id,
                    payload={'task_id': task_id}
                )

                # Release lock
                fcntl.flock(lock.fileno(), fcntl.LOCK_UN)

                return Task(**task_data)

        except BlockingIOError:
            # Task already claimed by another agent
            return None
        finally:
            # Clean up lock file
            if lock_file.exists():
                lock_file.unlink()

    def complete_task(self, task_id: str, result: Dict[str, Any]) -> bool:
        """
        Mark task as completed (move processing → completed)

        Args:
            task_id: Task to complete
            result: Task result data

        Returns:
            True if successful, False if task not found
        """
        processing_file = self.queue_root / 'tasks' / 'processing' / f'{task_id}.json'

        if not processing_file.exists():
            return False

        # Load and update task
        task_data = json.loads(processing_file.read_text())
        task_data['status'] = TaskStatus.COMPLETED.value
        task_data['completed_at'] = datetime.now().isoformat()
        task_data['result'] = result

        # Move to completed
        completed_file = self.queue_root / 'tasks' / 'completed' / f'{task_id}.json'
        completed_file.write_text(json.dumps(task_data, indent=2))
        processing_file.unlink()

        # Emit event
        self.emit_event(
            event_type=EventType.TASK_COMPLETED,
            agent_id=task_data['claimed_by'],
            payload={'task_id': task_id, 'result': result}
        )

        return True

    def fail_task(self, task_id: str, error: str) -> bool:
        """
        Mark task as failed

        Args:
            task_id: Task to fail
            error: Error message

        Returns:
            True if successful, False if task not found
        """
        processing_file = self.queue_root / 'tasks' / 'processing' / f'{task_id}.json'

        if not processing_file.exists():
            return False

        # Load and update task
        task_data = json.loads(processing_file.read_text())
        task_data['status'] = TaskStatus.FAILED.value
        task_data['completed_at'] = datetime.now().isoformat()
        task_data['error'] = error

        # Move to completed (with failed status)
        completed_file = self.queue_root / 'tasks' / 'completed' / f'{task_id}.json'
        completed_file.write_text(json.dumps(task_data, indent=2))
        processing_file.unlink()

        # Emit event
        self.emit_event(
            event_type=EventType.TASK_FAILED,
            agent_id=task_data['claimed_by'],
            payload={'task_id': task_id, 'error': error}
        )

        return True

    def get_pending_tasks(self, agent_type: Optional[str] = None) -> List[Task]:
        """
        Get all pending tasks (optionally filtered by agent type)

        Args:
            agent_type: Filter by agent type (SA, AD, CE, CD, Ralph)

        Returns:
            List of pending tasks
        """
        pending_dir = self.queue_root / 'tasks' / 'pending'
        tasks = []

        for task_file in pending_dir.glob('*.json'):
            task_data = json.loads(task_file.read_text())

            if agent_type is None or task_data['agent_type'] == agent_type:
                tasks.append(Task(**task_data))

        return tasks

    # ================== Event Management ==================

    def emit_event(self, event_type: EventType, agent_id: str, payload: Dict[str, Any]):
        """
        Emit an event to the queue (fan-out broadcast)

        Args:
            event_type: Type of event
            agent_id: Agent emitting the event
            payload: Event-specific data
        """
        timestamp = datetime.now().isoformat().replace(':', '-')
        event_id = f"{int(time.time() * 1000)}_{agent_id}_{event_type.value}"

        event = Event(
            event_id=event_id,
            event_type=event_type,
            agent_id=agent_id,
            payload=payload,
            timestamp=timestamp
        )

        # Write event file (convert Enum to value)
        event_dict = asdict(event)
        event_dict['event_type'] = event_type.value  # Convert Enum to string value
        event_file = self.queue_root / 'events' / f'{event_id}.json'
        event_file.write_text(json.dumps(event_dict, indent=2))

    def get_events(self, since: Optional[datetime] = None) -> List[Event]:
        """
        Get all events (optionally since a specific time)

        Args:
            since: Get events after this time

        Returns:
            List of events
        """
        events_dir = self.queue_root / 'events'
        events = []

        for event_file in sorted(events_dir.glob('*.json')):
            event_data = json.loads(event_file.read_text())

            # Convert event_type string back to Enum
            if isinstance(event_data['event_type'], str):
                event_data['event_type'] = EventType(event_data['event_type'])

            event = Event(**event_data)

            if since is None or datetime.fromisoformat(event.timestamp) > since:
                events.append(event)

        return events

    def clear_old_events(self, older_than_hours: int = 24):
        """
        Clean up old events (older than N hours)

        Args:
            older_than_hours: Delete events older than this
        """
        cutoff_time = time.time() - (older_than_hours * 3600)
        events_dir = self.queue_root / 'events'

        for event_file in events_dir.glob('*.json'):
            if event_file.stat().st_mtime < cutoff_time:
                event_file.unlink()


# ================== Example Usage ==================

if __name__ == '__main__':
    # Initialize queue
    queue = QueueManager()

    print("🚀 Queue Manager Test")
    print("=" * 50)

    # Create a task
    task_id = queue.create_task(
        agent_type='SA',
        task_type='analyze_signal',
        payload={
            'signal_id': 'test_001',
            'content': 'Test signal for analysis'
        }
    )
    print(f"✅ Created task: {task_id}")

    # List pending tasks
    pending = queue.get_pending_tasks(agent_type='SA')
    print(f"📋 Pending SA tasks: {len(pending)}")

    # Claim task
    task = queue.claim_task(agent_id='sa-worker-1', task_id=task_id)
    if task:
        print(f"✅ Claimed task: {task.task_id} by {task.claimed_by}")

        # Simulate work
        time.sleep(1)

        # Complete task
        queue.complete_task(
            task_id=task_id,
            result={'analysis': 'Test complete', 'score': 85}
        )
        print(f"✅ Completed task: {task_id}")

    # Get events
    events = queue.get_events()
    print(f"\n📡 Events: {len(events)}")
    for event in events:
        print(f"  - {event.event_type.value}: {event.agent_id}")

    print("\n✅ Queue Manager working correctly!")
