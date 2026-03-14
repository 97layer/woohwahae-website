#!/usr/bin/env python3
"""Minimal Layer OS external agent example.

Flow:
1. Fetch `AgentRunPacket` from Layer OS.
2. Execute external work (LLM/tool call placeholder).
3. Report terminal result back to Layer OS.
"""

from __future__ import annotations

import argparse
import json
from typing import Any, Dict

import requests


class LayerOSAgentError(RuntimeError):
    """Raised when packet fetch or report fails."""


def fetch_job_packet(job_id: str, base_url: str) -> Dict[str, Any]:
    response = requests.get(
        f"{base_url.rstrip('/')}/api/layer-os/jobs/packet",
        params={"job_id": job_id},
        timeout=30,
    )
    response.raise_for_status()
    return response.json()


def report_job(
    job_id: str,
    status: str,
    result: Dict[str, Any],
    base_url: str,
    token: str,
) -> Dict[str, Any]:
    payload = {
        "job_id": job_id,
        "status": status,
        "notes": result.get("notes", []),
        "result": result,
    }
    headers = {"Content-Type": "application/json"}
    if token.strip():
        headers["Authorization"] = f"Bearer {token}"
    response = requests.post(
        f"{base_url.rstrip('/')}/api/layer-os/jobs/report",
        headers=headers,
        json=payload,
        timeout=30,
    )
    response.raise_for_status()
    return response.json()


def main() -> None:
    parser = argparse.ArgumentParser(description="Layer OS external agent runner example")
    parser.add_argument("--job-id", required=True, help="Layer OS job id")
    parser.add_argument(
        "--base-url",
        required=True,
        help="Layer OS daemon base URL, e.g. http://127.0.0.1:17808",
    )
    parser.add_argument(
        "--token",
        default="",
        help="Write token value for Authorization: Bearer <token>",
    )
    args = parser.parse_args()

    try:
        packet = fetch_job_packet(args.job_id, args.base_url)

        job = packet.get("job", {})
        runtime = packet.get("runtime", {})

        # TODO: Replace this placeholder with a real LLM/tool execution.
        # Example shape:
        # - read `job["summary"]`, `job.get("payload")`
        # - read `packet["knowledge"]` and `packet["handoff"]`
        # - call Claude Code / Codex / Python agent logic
        # - collect structured output for `result`
        result = {
            "summary": f"Stub external run completed for {job.get('job_id', args.job_id)}",
            "agent": "python-example",
            "dispatch_transport": runtime.get("dispatch_transport", "job_packet"),
            "notes": ["stub_execution", "replace_with_real_llm_call"],
        }
        report = report_job(args.job_id, "succeeded", result, args.base_url, args.token)
        print(json.dumps(report, ensure_ascii=False, indent=2))
    except requests.HTTPError as exc:
        error_text = exc.response.text if exc.response is not None else str(exc)
        failed_result = {
            "error": "http_error",
            "details": error_text,
            "notes": ["http_error"],
        }
        try:
            report = report_job(args.job_id, "failed", failed_result, args.base_url, args.token)
            print(json.dumps(report, ensure_ascii=False, indent=2))
        except Exception as report_exc:  # pragma: no cover - best-effort failure path
            raise LayerOSAgentError(f"failed to report HTTP error: {report_exc}") from exc
    except Exception as exc:
        failed_result = {
            "error": exc.__class__.__name__,
            "details": str(exc),
            "notes": ["agent_exception"],
        }
        try:
            report = report_job(args.job_id, "failed", failed_result, args.base_url, args.token)
            print(json.dumps(report, ensure_ascii=False, indent=2))
        except Exception as report_exc:  # pragma: no cover - best-effort failure path
            raise LayerOSAgentError(f"failed to report agent exception: {report_exc}") from exc


if __name__ == "__main__":
    main()
