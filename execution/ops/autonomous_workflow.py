#!/usr/bin/env python3
"""
Autonomous Workflow System - 자율 실행 워크플로우
맥북 종료 후에도 GCP에서 계속 실행 가능

Features:
- 워크플로우 상태 저장 및 복원
- GCP 자동 마이그레이션
- 체크포인트 시스템
- 원격 실행 지원
"""

import json
import pickle
import time
import subprocess
import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any, Optional
from dataclasses import dataclass, asdict
import logging

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.append(str(PROJECT_ROOT))


@dataclass
class WorkflowState:
    """워크플로우 상태"""
    workflow_id: str
    name: str
    status: str  # pending, running, paused, completed, failed
    current_step: int
    total_steps: int
    checkpoint_data: Dict[str, Any]
    started_at: str
    last_update: str
    execution_location: str  # mac, gcp
    error_log: List[str]


class AutonomousWorkflow:
    """자율 실행 워크플로우"""

    def __init__(self):
        self.project_root = PROJECT_ROOT
        self.state_dir = self.project_root / "knowledge" / "workflow_state"
        self.state_dir.mkdir(parents=True, exist_ok=True)

        # Google Drive 동기화 경로 (맥북 종료 시에도 접근 가능)
        self.cloud_state_dir = self.project_root / "gdrive_sync" / "workflow_state"
        self.cloud_state_dir.mkdir(parents=True, exist_ok=True)

        self.current_workflow: Optional[WorkflowState] = None
        self.workflows: Dict[str, WorkflowState] = {}

        # 상태 복원
        self._restore_state()

    def create_workflow(self, name: str, steps: List[Dict[str, Any]]) -> str:
        """
        워크플로우 생성

        Args:
            name: 워크플로우 이름
            steps: 실행 단계 리스트

        Returns:
            워크플로우 ID
        """
        workflow_id = f"wf-{datetime.now().timestamp()}"

        workflow = WorkflowState(
            workflow_id=workflow_id,
            name=name,
            status="pending",
            current_step=0,
            total_steps=len(steps),
            checkpoint_data={"steps": steps},
            started_at=datetime.now().isoformat(),
            last_update=datetime.now().isoformat(),
            execution_location="mac",
            error_log=[]
        )

        self.workflows[workflow_id] = workflow
        self._save_state()

        logger.info(f"Workflow created: {workflow_id} - {name}")
        return workflow_id

    def execute_workflow(self, workflow_id: str, resume: bool = False):
        """
        워크플로우 실행

        Args:
            workflow_id: 워크플로우 ID
            resume: 이어서 실행 여부
        """
        if workflow_id not in self.workflows:
            logger.error(f"Workflow not found: {workflow_id}")
            return

        workflow = self.workflows[workflow_id]
        self.current_workflow = workflow

        if not resume:
            workflow.current_step = 0

        workflow.status = "running"
        workflow.execution_location = self._get_execution_location()

        steps = workflow.checkpoint_data.get("steps", [])

        logger.info(f"Executing workflow: {workflow.name} from step {workflow.current_step}")

        try:
            while workflow.current_step < workflow.total_steps:
                # 체크포인트 저장
                self._save_checkpoint(workflow)

                step = steps[workflow.current_step]
                logger.info(f"Step {workflow.current_step + 1}/{workflow.total_steps}: {step.get('name')}")

                # 단계 실행
                success = self._execute_step(step, workflow)

                if not success:
                    workflow.status = "failed"
                    workflow.error_log.append(f"Step {workflow.current_step} failed")
                    break

                workflow.current_step += 1
                workflow.last_update = datetime.now().isoformat()

                # 상태 저장
                self._save_state()

                # 클라우드 동기화
                self._sync_to_cloud()

            if workflow.current_step >= workflow.total_steps:
                workflow.status = "completed"
                logger.info(f"Workflow completed: {workflow.name}")

        except KeyboardInterrupt:
            workflow.status = "paused"
            logger.info(f"Workflow paused: {workflow.name}")
            self._save_state()

        except Exception as e:
            workflow.status = "failed"
            workflow.error_log.append(str(e))
            logger.error(f"Workflow error: {e}")
            self._save_state()

    def _execute_step(self, step: Dict[str, Any], workflow: WorkflowState) -> bool:
        """
        단계 실행

        Args:
            step: 실행할 단계
            workflow: 워크플로우 상태

        Returns:
            성공 여부
        """
        step_type = step.get("type")

        try:
            if step_type == "script":
                # Python 스크립트 실행
                script_path = self.project_root / step["path"]
                result = subprocess.run(
                    [sys.executable, str(script_path)],
                    capture_output=True,
                    text=True,
                    timeout=step.get("timeout", 300)
                )
                return result.returncode == 0

            elif step_type == "command":
                # 쉘 명령 실행
                result = subprocess.run(
                    step["command"],
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=step.get("timeout", 300)
                )
                return result.returncode == 0

            elif step_type == "function":
                # 함수 실행
                func_name = step["function"]
                # 동적 함수 호출 (보안 주의)
                if hasattr(self, func_name):
                    func = getattr(self, func_name)
                    return func(step.get("args", {}))

            elif step_type == "wait":
                # 대기
                time.sleep(step.get("seconds", 1))
                return True

            else:
                logger.warning(f"Unknown step type: {step_type}")
                return False

        except subprocess.TimeoutExpired:
            logger.error(f"Step timeout: {step.get('name')}")
            return False

        except Exception as e:
            logger.error(f"Step execution error: {e}")
            return False

    def migrate_to_gcp(self, workflow_id: str):
        """
        워크플로우를 GCP로 마이그레이션

        Args:
            workflow_id: 워크플로우 ID
        """
        if workflow_id not in self.workflows:
            return

        workflow = self.workflows[workflow_id]

        logger.info(f"Migrating workflow to GCP: {workflow.name}")

        # 상태 파일을 GCP로 전송
        state_file = self.state_dir / f"{workflow_id}.json"

        # rsync 또는 scp로 전송
        gcp_host = "skyto5339@35.184.30.182"
        gcp_path = "~/97layerOS/knowledge/workflow_state/"

        try:
            # SSH 키 경로
            ssh_key = Path.home() / ".ssh" / "id_ed25519_gcp"

            if ssh_key.exists():
                # 상태 파일 전송
                subprocess.run([
                    "scp", "-i", str(ssh_key),
                    str(state_file),
                    f"{gcp_host}:{gcp_path}"
                ], check=True)

                # GCP에서 워크플로우 실행 명령
                subprocess.run([
                    "ssh", "-i", str(ssh_key), gcp_host,
                    f"cd ~/97layerOS && python3 execution/ops/autonomous_workflow.py resume {workflow_id}"
                ])

                workflow.execution_location = "gcp"
                self._save_state()

                logger.info(f"Workflow migrated to GCP: {workflow.name}")

            else:
                logger.error("GCP SSH key not found")

        except Exception as e:
            logger.error(f"Migration error: {e}")

    def resume_from_checkpoint(self, workflow_id: str):
        """
        체크포인트에서 재개

        Args:
            workflow_id: 워크플로우 ID
        """
        # 클라우드에서 상태 복원
        self._restore_from_cloud()

        if workflow_id in self.workflows:
            workflow = self.workflows[workflow_id]
            logger.info(f"Resuming workflow: {workflow.name} from step {workflow.current_step}")
            self.execute_workflow(workflow_id, resume=True)
        else:
            logger.error(f"Workflow not found: {workflow_id}")

    def _save_checkpoint(self, workflow: WorkflowState):
        """체크포인트 저장"""
        checkpoint_file = self.state_dir / f"{workflow.workflow_id}_checkpoint.pkl"

        with open(checkpoint_file, 'wb') as f:
            pickle.dump(workflow, f)

        # JSON 백업도 저장
        json_file = self.state_dir / f"{workflow.workflow_id}.json"
        with open(json_file, 'w') as f:
            json.dump(asdict(workflow), f, indent=2)

    def _save_state(self):
        """전체 상태 저장"""
        state_file = self.state_dir / "workflows.json"

        state_data = {
            "workflows": {
                wf_id: asdict(wf) for wf_id, wf in self.workflows.items()
            },
            "current_workflow": self.current_workflow.workflow_id if self.current_workflow else None,
            "last_save": datetime.now().isoformat()
        }

        with open(state_file, 'w') as f:
            json.dump(state_data, f, indent=2)

    def _restore_state(self):
        """상태 복원"""
        state_file = self.state_dir / "workflows.json"

        if state_file.exists():
            try:
                with open(state_file) as f:
                    state_data = json.load(f)

                for wf_id, wf_data in state_data.get("workflows", {}).items():
                    self.workflows[wf_id] = WorkflowState(**wf_data)

                logger.info(f"Restored {len(self.workflows)} workflows")

            except Exception as e:
                logger.error(f"State restore error: {e}")

    def _sync_to_cloud(self):
        """클라우드 동기화"""
        try:
            # Google Drive 동기화 (rclone 사용)
            subprocess.run([
                "rclone", "sync",
                str(self.state_dir),
                f"gdrive:97layerOS/workflow_state",
                "--update"
            ], capture_output=True)

        except FileNotFoundError:
            # rclone이 없으면 로컬 백업만
            import shutil
            for file in self.state_dir.glob("*.json"):
                shutil.copy2(file, self.cloud_state_dir / file.name)

    def _restore_from_cloud(self):
        """클라우드에서 복원"""
        try:
            # Google Drive에서 동기화
            subprocess.run([
                "rclone", "sync",
                f"gdrive:97layerOS/workflow_state",
                str(self.state_dir),
                "--update"
            ], capture_output=True)

            self._restore_state()

        except Exception as e:
            logger.error(f"Cloud restore error: {e}")

    def _get_execution_location(self) -> str:
        """실행 위치 판단"""
        import socket
        hostname = socket.gethostname()

        if "skyto" in hostname.lower() or "gcp" in hostname.lower():
            return "gcp"
        return "mac"

    def list_workflows(self) -> List[Dict[str, Any]]:
        """워크플로우 목록"""
        return [
            {
                "id": wf.workflow_id,
                "name": wf.name,
                "status": wf.status,
                "progress": f"{wf.current_step}/{wf.total_steps}",
                "location": wf.execution_location
            }
            for wf in self.workflows.values()
        ]


def main():
    """메인 함수 - CLI 인터페이스"""
    workflow = AutonomousWorkflow()

    if len(sys.argv) < 2:
        print("Usage: autonomous_workflow.py [create|execute|resume|migrate|list]")
        return

    command = sys.argv[1]

    if command == "create":
        # 예시 워크플로우 생성
        steps = [
            {"name": "Initialize", "type": "wait", "seconds": 2},
            {"name": "Process data", "type": "script", "path": "execution/data_processor.py"},
            {"name": "Generate report", "type": "command", "command": "echo 'Report generated'"},
            {"name": "Finalize", "type": "wait", "seconds": 1}
        ]

        wf_id = workflow.create_workflow("Example Workflow", steps)
        print(f"Workflow created: {wf_id}")

    elif command == "execute" and len(sys.argv) > 2:
        workflow.execute_workflow(sys.argv[2])

    elif command == "resume" and len(sys.argv) > 2:
        workflow.resume_from_checkpoint(sys.argv[2])

    elif command == "migrate" and len(sys.argv) > 2:
        workflow.migrate_to_gcp(sys.argv[2])

    elif command == "list":
        workflows = workflow.list_workflows()
        for wf in workflows:
            print(f"{wf['id']}: {wf['name']} - {wf['status']} ({wf['progress']}) @ {wf['location']}")


if __name__ == "__main__":
    main()