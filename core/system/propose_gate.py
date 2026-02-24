#!/usr/bin/env python3
"""
propose_gate.py — 텔레그램 diff 확인 게이트

Code Agent가 생성한 diff를 텔레그램으로 전송하고,
순호의 ok/no 응답을 기다린 후 실행/폐기 결정.

State: knowledge/system/propose_queue.json
Timeout: 24시간 무응답 시 자동 폐기
"""
import json
import logging
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
QUEUE_PATH = PROJECT_ROOT / "knowledge" / "system" / "propose_queue.json"

logger = logging.getLogger(__name__)

CONFIRM_WORDS = {"ok", "yes", "응", "ㅇㅋ", "좋아", "해", "승인", "확인"}
REJECT_WORDS = {"no", "취소", "cancel", "아니", "ㄴ", "중지", "폐기", "거절"}
TIMEOUT_HOURS = 24


def _load_queue() -> dict:
    if QUEUE_PATH.exists():
        try:
            return json.loads(QUEUE_PATH.read_text(encoding="utf-8"))
        except Exception:
            return {}
    return {}


def _save_queue(data: dict) -> None:
    QUEUE_PATH.parent.mkdir(parents=True, exist_ok=True)
    # 7일 지난 비-pending 항목 정리
    cutoff = datetime.now(timezone.utc).timestamp() - 7 * 86400
    cleaned = {
        k: v for k, v in data.items()
        if v.get("status") == "pending"
        or datetime.fromisoformat(v.get("created_at", "2000-01-01T00:00:00+00:00")).timestamp() > cutoff
    }
    QUEUE_PATH.write_text(json.dumps(cleaned, ensure_ascii=False, indent=2), encoding="utf-8")


class ProposeGate:
    """
    텔레그램 메시지 ID 기반 confirm/cancel 루프.

    사용 흐름:
        gate = ProposeGate()
        task_id = gate.propose(diff_text, callback_fn)
        # 텔레그램에서 "ok" → gate.confirm("ok", reply_msg_id) → callback_fn 실행
    """

    def propose(
        self,
        task_id: str,
        diff_text: str,
        chat_id: int,
        callback_data: Optional[dict] = None,
    ) -> dict:
        """
        diff를 propose_queue에 등록하고 텔레그램 전송용 메시지를 반환.

        Returns:
            {"task_id": str, "message": str}
        """
        queue = _load_queue()
        now = datetime.now(timezone.utc).isoformat()

        queue[task_id] = {
            "task_id": task_id,
            "diff_text": diff_text,
            "chat_id": chat_id,
            "callback_data": callback_data or {},
            "status": "pending",
            "created_at": now,
            "message_ids": [],  # 텔레그램 메시지 ID 목록 (confirm 매칭용)
        }
        _save_queue(queue)
        logger.info("PROPOSE 등록: %s", task_id)

        short_diff = diff_text[:3000] if len(diff_text) > 3000 else diff_text
        # HTML 이스케이프 — <pre> 안에 <, >, & 있으면 400
        escaped = (short_diff
                   .replace("&", "&amp;")
                   .replace("<", "&lt;")
                   .replace(">", "&gt;"))
        message = (
            f"[PROPOSE] <code>{task_id}</code>\n\n"
            f"<pre>{escaped}</pre>\n\n"
            f"ok = 적용 / no = 폐기 (24시간 타임아웃)"
        )
        return {"task_id": task_id, "message": message}

    def register_message_id(self, task_id: str, telegram_message_id: int) -> None:
        """텔레그램이 실제 메시지를 전송한 후 message_id를 등록."""
        queue = _load_queue()
        if task_id in queue:
            queue[task_id]["message_ids"].append(telegram_message_id)
            _save_queue(queue)

    def find_pending_by_reply(self, reply_to_message_id: int) -> Optional[dict]:
        """reply_to_message_id로 pending 태스크 찾기."""
        queue = _load_queue()
        for task in queue.values():
            if (
                task.get("status") == "pending"
                and reply_to_message_id in task.get("message_ids", [])
            ):
                return task
        return None

    def find_latest_pending(self, chat_id: int) -> Optional[dict]:
        """해당 chat_id의 가장 최근 pending 태스크 반환 (reply 없을 때 fallback)."""
        queue = _load_queue()
        candidates = [
            t for t in queue.values()
            if t.get("status") == "pending" and t.get("chat_id") == chat_id
        ]
        if not candidates:
            return None
        return max(candidates, key=lambda t: t.get("created_at", ""))

    def confirm(self, message_text: str, task_id: str) -> str:
        """
        응답 텍스트로 태스크를 approve/reject.

        Returns:
            "approved" | "rejected" | "unknown"
        """
        normalized = message_text.strip().lower()
        queue = _load_queue()

        if task_id not in queue:
            return "unknown"

        task = queue[task_id]
        if task["status"] != "pending":
            return "unknown"

        if normalized in CONFIRM_WORDS:
            task["status"] = "approved"
            task["resolved_at"] = datetime.now(timezone.utc).isoformat()
            _save_queue(queue)
            logger.info("PROPOSE 승인: %s", task_id)
            return "approved"

        if normalized in REJECT_WORDS:
            task["status"] = "rejected"
            task["resolved_at"] = datetime.now(timezone.utc).isoformat()
            _save_queue(queue)
            logger.info("PROPOSE 폐기: %s", task_id)
            return "rejected"

        return "unknown"

    def get_task(self, task_id: str) -> Optional[dict]:
        """태스크 상태 조회."""
        return _load_queue().get(task_id)

    def is_pending(self, message_id: int) -> bool:
        """주어진 message_id에 pending propose가 있으면 True."""
        queue = _load_queue()
        for task in queue.values():
            if (
                task.get("status") == "pending"
                and message_id in task.get("message_ids", [])
            ):
                return True
        return False

    def expire_old(self) -> list:
        """24시간 초과 pending 태스크 자동 폐기. 만료된 태스크 목록 반환."""
        queue = _load_queue()
        now = datetime.now(timezone.utc)
        expired = []

        for task in queue.values():
            if task.get("status") != "pending":
                continue
            created = datetime.fromisoformat(task["created_at"])
            if (now - created).total_seconds() > TIMEOUT_HOURS * 3600:
                task["status"] = "expired"
                task["resolved_at"] = now.isoformat()
                expired.append({
                    "chat_id": task.get("chat_id"),
                    "task_id": task["task_id"],
                    "summary": task.get("diff_text", "")[:200],
                })
                logger.info("PROPOSE 만료: %s", task["task_id"])

        if expired:
            _save_queue(queue)
        return expired
