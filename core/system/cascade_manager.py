#!/usr/bin/env python3
"""
Cascade Manager - CD 알림 전송 (Telegram/Email)
CI/CD 파이프라인 이벤트를 Telegram 및 Email로 알림 전송
"""

import os
import logging
import smtplib
import ssl
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from datetime import datetime
from typing import Optional

import httpx

logger = logging.getLogger(__name__)


def _require_env(key: str) -> str:
    """환경변수 필수 조회 — fallback 없음."""
    value = os.environ.get(key)
    if not value:
        raise RuntimeError("Required environment variable not set: %s" % key)
    return value


# ─── Telegram ─────────────────────────────────────────────────

def _build_cd_message(event: str, detail: str, status: str, timestamp: Optional[str] = None) -> str:
    """CD 알림 메시지 템플릿 생성."""
    ts = timestamp or datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    status_emoji = {"success": "✅", "failure": "❌", "started": "🚀", "cancelled": "⚠️"}.get(status, "ℹ️")
    return (
        f"{status_emoji} *LAYER OS CD 알림*\n"
        f"━━━━━━━━━━━━━━━━━━━━\n"
        f"이벤트: `{event}`\n"
        f"상태: `{status}`\n"
        f"상세: {detail}\n"
        f"시각: `{ts}`"
    )


def send_telegram_cd_notification(
    event: str,
    detail: str,
    status: str,
    timestamp: Optional[str] = None,
    bot_token: Optional[str] = None,
    chat_id: Optional[str] = None,
) -> bool:
    """
    Telegram으로 CD 알림 전송.

    Args:
        event: 이벤트 이름 (예: "deploy", "rollback")
        detail: 상세 설명
        status: "success" | "failure" | "started" | "cancelled"
        timestamp: 타임스탬프 문자열 (None이면 현재 시각)
        bot_token: Telegram Bot Token (None이면 환경변수 TELEGRAM_BOT_TOKEN)
        chat_id: Telegram Chat ID (None이면 환경변수 TELEGRAM_CHAT_ID)

    Returns:
        True if sent successfully, False otherwise.
    """
    try:
        token = bot_token or _require_env("TELEGRAM_BOT_TOKEN")
        cid = chat_id or _require_env("TELEGRAM_CHAT_ID")
    except RuntimeError as exc:
        logger.warning("Telegram CD notification skipped: %s", exc)
        return False

    message = _build_cd_message(event, detail, status, timestamp)
    url = "https://api.telegram.org/bot%s/sendMessage" % token
    payload = {
        "chat_id": cid,
        "text": message,
        "parse_mode": "Markdown",
        "disable_web_page_preview": True,
    }

    try:
        response = httpx.post(url, json=payload, timeout=10)
        response.raise_for_status()
        logger.info("Telegram CD notification sent: event=%s status=%s", event, status)
        return True
    except httpx.HTTPStatusError as exc:
        logger.error(
            "Telegram CD notification HTTP error: status=%s body=%s",
            exc.response.status_code,
            exc.response.text,
        )
        return False
    except httpx.RequestError as exc:
        logger.error("Telegram CD notification request error: %s", exc)
        return False


# ─── Email ────────────────────────────────────────────────────

def send_email_cd_notification(
    event: str,
    detail: str,
    status: str,
    timestamp: Optional[str] = None,
    smtp_host: Optional[str] = None,
    smtp_port: Optional[int] = None,
    smtp_user: Optional[str] = None,
    smtp_password: Optional[str] = None,
    from_addr: Optional[str] = None,
    to_addr: Optional[str] = None,
) -> bool:
    """
    Email로 CD 알림 전송 (TLS).

    환경변수:
        SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD,
        NOTIFY_FROM_EMAIL, NOTIFY_TO_EMAIL

    Returns:
        True if sent successfully, False otherwise.
    """
    try:
        host = smtp_host or _require_env("SMTP_HOST")
        port = smtp_port or int(_require_env("SMTP_PORT"))
        user = smtp_user or _require_env("SMTP_USER")
        password = smtp_password or _require_env("SMTP_PASSWORD")
        sender = from_addr or _require_env("NOTIFY_FROM_EMAIL")
        recipient = to_addr or _require_env("NOTIFY_TO_EMAIL")
    except RuntimeError as exc:
        logger.warning("Email CD notification skipped: %s", exc)
        return False

    ts = timestamp or datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    status_label = {"success": "성공", "failure": "실패", "started": "시작", "cancelled": "취소"}.get(status, status)
    subject = "[LAYER OS CD] %s — %s (%s)" % (event, status_label, ts)

    body = (
        "LAYER OS CD 알림\n"
        "===================\n"
        "이벤트 : %s\n"
        "상태   : %s\n"
        "상세   : %s\n"
        "시각   : %s\n"
    ) % (event, status, detail, ts)

    msg = MIMEMultipart("alternative")
    msg["Subject"] = subject
    msg["From"] = sender
    msg["To"] = recipient
    msg.attach(MIMEText(body, "plain", "utf-8"))

    try:
        context = ssl.create_default_context()
        with smtplib.SMTP(host, port) as server:
            server.ehlo()
            server.starttls(context=context)
            server.login(user, password)
            server.sendmail(sender, [recipient], msg.as_string())
        logger.info("Email CD notification sent: event=%s status=%s to=%s", event, status, recipient)
        return True
    except smtplib.SMTPAuthenticationError as exc:
        logger.error("Email CD notification SMTP auth error: %s", exc)
        return False
    except smtplib.SMTPException as exc:
        logger.error("Email CD notification SMTP error: %s", exc)
        return False


# ─── Unified Entry Point ──────────────────────────────────────

def notify_cd_event(
    event: str,
    detail: str,
    status: str,
    timestamp: Optional[str] = None,
    channels: Optional[list] = None,
) -> dict:
    """
    CD 이벤트 알림 — Telegram 및/또는 Email로 전송.

    Args:
        event: 이벤트 이름
        detail: 상세 설명
        status: "success" | "failure" | "started" | "cancelled"
        timestamp: 타임스탬프 (None이면 현재 시각)
        channels: ["telegram", "email"] 중 선택 (None이면 둘 다 시도)

    Returns:
        {"telegram": bool, "email": bool}
    """
    if channels is None:
        channels = ["telegram", "email"]

    results = {}

    if "telegram" in channels:
        results["telegram"] = send_telegram_cd_notification(
            event=event,
            detail=detail,
            status=status,
            timestamp=timestamp,
        )

    if "email" in channels:
        results["email"] = send_email_cd_notification(
            event=event,
            detail=detail,
            status=status,
            timestamp=timestamp,
        )

    logger.info(
        "CD notification results: event=%s status=%s results=%s",
        event,
        status,
        results,
    )
    return results
