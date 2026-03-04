#!/usr/bin/env python3
"""Read-only backend hardening report (easy Korean summary)."""

from __future__ import annotations

import argparse
import ast
import json
import re
import sys
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Callable

PROJECT_ROOT = Path(__file__).resolve().parents[2]
SEVERITY_ORDER = ("high", "medium", "low")
SEVERITY_LABEL = {
    "high": "높음",
    "medium": "중간",
    "low": "낮음",
}
RALPH_STAGE_MAX = 25.0
RALPH_STAGES: tuple[tuple[str, str, tuple[str, ...]], ...] = (
    (
        "auth_access",
        "1단계 인증/권한",
        ("fastapi-admin-header-auth", "photo-upload-security"),
    ),
    (
        "secrets_seed",
        "2단계 시크릿/시드",
        ("ecommerce-secret-policy", "ecommerce-admin-seed-policy"),
    ),
    (
        "integrity_payment",
        "3단계 데이터무결성/결제",
        ("growth-report-integrity", "ecommerce-payments-router"),
    ),
    (
        "operations_persistence",
        "4단계 운영지속성",
        ("queue-timestamp-parse", "admin-password-persistence"),
    ),
)


@dataclass(frozen=True)
class Finding:
    check_id: str
    title: str
    severity: str
    status: str
    target_file: str
    summary: str
    recommendation: str


CheckFn = Callable[[Path], Finding]


def _read_text(root: Path, rel_path: str) -> str | None:
    file_path = root / rel_path
    if not file_path.exists():
        return None
    try:
        return file_path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return file_path.read_text(encoding="utf-8", errors="ignore")


def _function_contains(text: str, func_name: str, needle: str) -> bool:
    try:
        tree = ast.parse(text)
    except SyntaxError:
        return False

    for node in tree.body:
        if isinstance(node, (ast.AsyncFunctionDef, ast.FunctionDef)) and node.name == func_name:
            segment = ast.get_source_segment(text, node) or ""
            return needle in segment
    return False


def _ok(
    check_id: str,
    title: str,
    severity: str,
    target_file: str,
    summary: str,
    recommendation: str,
) -> Finding:
    return Finding(
        check_id=check_id,
        title=title,
        severity=severity,
        status="pass",
        target_file=target_file,
        summary=summary,
        recommendation=recommendation,
    )


def _fail(
    check_id: str,
    title: str,
    severity: str,
    target_file: str,
    summary: str,
    recommendation: str,
) -> Finding:
    return Finding(
        check_id=check_id,
        title=title,
        severity=severity,
        status="fail",
        target_file=target_file,
        summary=summary,
        recommendation=recommendation,
    )


def check_fastapi_admin_headers(root: Path) -> Finding:
    rel = "core/backend/main.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "fastapi-admin-header-auth",
            "FastAPI 관리자 헤더 인증",
            "high",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "파일 경로가 맞는지 확인하고 관리자 인증 로직을 복구하세요.",
        )

    missing: list[str] = []
    required_snippets = [
        "def _extract_admin_token",
        "request.headers.get(\"authorization\"",
        "request.headers.get(\"x-admin-token\"",
    ]
    for snippet in required_snippets:
        if snippet not in text:
            missing.append(snippet)

    guarded_funcs = ("get_page_contents", "get_content", "get_all_pages")
    for func_name in guarded_funcs:
        if not _function_contains(text, func_name, "check_admin(request)"):
            missing.append(f"{func_name}:check_admin(request)")

    forbidden_snippets = (
        "request.query_params.get('token')",
        'request.query_params.get("token")',
        "?token=",
    )
    forbidden_found = [s for s in forbidden_snippets if s in text]

    if not missing and not forbidden_found:
        return _ok(
            "fastapi-admin-header-auth",
            "FastAPI 관리자 헤더 인증",
            "high",
            rel,
            "관리자 인증이 헤더 기반으로 동작하고, 주요 조회 API에서 인증 검사를 수행합니다.",
            "현재 상태 유지. 인증 누락 회귀만 정기 점검하세요.",
        )

    details: list[str] = []
    if missing:
        details.append(f"누락: {', '.join(missing)}")
    if forbidden_found:
        details.append(f"금지 패턴 존재: {', '.join(forbidden_found)}")
    return _fail(
        "fastapi-admin-header-auth",
        "FastAPI 관리자 헤더 인증",
        "high",
        rel,
        " / ".join(details),
        "쿼리스트링 토큰 사용을 제거하고, 보호 API 함수에 `check_admin(request)`를 넣으세요.",
    )


def check_photo_upload_security(root: Path) -> Finding:
    rel = "core/backend/photo_upload.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "photo-upload-security",
            "사진 업로드 인증/제한",
            "high",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "업로드 서브앱 파일을 복구하고 인증/용량 제한을 다시 점검하세요.",
        )

    required = (
        "PHOTO_UPLOAD_ADMIN_TOKEN",
        "def _require_admin_auth",
        "def _matches_image_signature",
        "MAX_UPLOAD_BYTES",
        "MAX_UPLOAD_FILES",
    )
    missing = [token for token in required if token not in text]
    forbidden_found: list[str] = []
    if "image/svg+xml" in text:
        forbidden_found.append("image/svg+xml")
    ext_match = re.search(r"ALLOWED_EXTENSIONS\\s*=\\s*\\{([^}]*)\\}", text)
    if ext_match and "svg" in ext_match.group(1).lower():
        forbidden_found.append("ALLOWED_EXTENSIONS contains svg")

    if not missing and not forbidden_found:
        return _ok(
            "photo-upload-security",
            "사진 업로드 인증/제한",
            "high",
            rel,
            "업로드 API에 관리자 인증, 파일 제한, 시그니처 검증, SVG 차단이 적용되어 있습니다.",
            "현재 제한값이 운영 정책과 맞는지만 주기 점검하세요.",
        )
    details: list[str] = []
    if missing:
        details.append(f"누락: {', '.join(missing)}")
    if forbidden_found:
        details.append(f"금지 패턴 존재: {', '.join(forbidden_found)}")
    return _fail(
        "photo-upload-security",
        "사진 업로드 인증/제한",
        "high",
        rel,
        " / ".join(details),
        "업로드 API에 관리자 인증, 용량/개수 제한, 시그니처 검증, SVG 차단을 적용하세요.",
    )


def check_ecommerce_secret_policy(root: Path) -> Finding:
    rel = "core/backend/ecommerce/config.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "ecommerce-secret-policy",
            "이커머스 JWT/DB 설정 안전성",
            "high",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "이커머스 설정 파일을 복구하고 JWT/DB 기본값 정책을 다시 검토하세요.",
        )

    required = (
        "def _load_jwt_secret",
        "JWT_SECRET_KEY",
        "RuntimeError(\"JWT_SECRET_KEY environment variable is required\")",
    )
    forbidden = (
        "your-secret-key-here",
        "postgresql://user:password@localhost",
    )

    missing = [token for token in required if token not in text]
    forbidden_found = [token for token in forbidden if token in text]

    if not missing and not forbidden_found:
        return _ok(
            "ecommerce-secret-policy",
            "이커머스 JWT/DB 설정 안전성",
            "high",
            rel,
            "JWT 비밀키 강제 정책과 안전한 DB 기본값 구성이 확인됩니다.",
            "운영 환경에서는 반드시 환경변수 기반 키/DB 주소를 유지하세요.",
        )

    chunks: list[str] = []
    if missing:
        chunks.append(f"누락: {', '.join(missing)}")
    if forbidden_found:
        chunks.append(f"금지 기본값 존재: {', '.join(forbidden_found)}")
    return _fail(
        "ecommerce-secret-policy",
        "이커머스 JWT/DB 설정 안전성",
        "high",
        rel,
        " / ".join(chunks),
        "하드코딩 시크릿/DB 문자열을 제거하고 환경변수 강제 로직을 적용하세요.",
    )


def check_admin_seed_policy(root: Path) -> Finding:
    rel = "core/backend/ecommerce/migrations/init_db.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "ecommerce-admin-seed-policy",
            "이커머스 관리자 시드 정책",
            "high",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "초기 관리자 계정 생성 정책을 환경변수 기반으로 복구하세요.",
        )

    required = (
        "ECOMMERCE_ADMIN_EMAIL",
        "ECOMMERCE_ADMIN_PASSWORD",
        "최소 12자",
    )
    forbidden = ("changeme123", "admin@woohwahae.kr")

    missing = [token for token in required if token not in text]
    forbidden_found = [token for token in forbidden if token in text]

    if not missing and not forbidden_found:
        return _ok(
            "ecommerce-admin-seed-policy",
            "이커머스 관리자 시드 정책",
            "high",
            rel,
            "기본 관리자 계정 하드코딩 없이 환경변수 기반 시드 정책으로 동작합니다.",
            "운영 비밀번호는 12자 이상으로 주기 교체하세요.",
        )

    chunks: list[str] = []
    if missing:
        chunks.append(f"누락: {', '.join(missing)}")
    if forbidden_found:
        chunks.append(f"금지 계정 정보 존재: {', '.join(forbidden_found)}")
    return _fail(
        "ecommerce-admin-seed-policy",
        "이커머스 관리자 시드 정책",
        "high",
        rel,
        " / ".join(chunks),
        "기본 계정 문자열을 제거하고 환경변수 기반 관리자 시드만 허용하세요.",
    )


def check_payments_router(root: Path) -> Finding:
    rel_main = "core/backend/ecommerce/main.py"
    rel_api = "core/backend/ecommerce/api/payments.py"
    rel_service = "core/backend/ecommerce/services/payment.py"
    main_text = _read_text(root, rel_main)
    api_text = _read_text(root, rel_api)
    service_text = _read_text(root, rel_service)
    target = f"{rel_main}, {rel_api}, {rel_service}"

    if main_text is None or api_text is None or service_text is None:
        return _fail(
            "ecommerce-payments-router",
            "결제 MVP 라우팅",
            "medium",
            target,
            "결제 라우터 파일 일부가 없습니다.",
            "결제 API/서비스 파일과 라우터 연결(include_router)을 복구하세요.",
        )

    missing: list[str] = []
    if "include_router(payments_router" not in main_text:
        missing.append("main.py: include_router(payments_router)")
    required_api = (
        "@router.post(\"/intent\")",
        "@router.post(\"/webhook\")",
        "@router.post(\"/orders/{order_id}/mark-paid\")",
        "_WEBHOOK_EVENT_CACHE_FILE",
        "_mark_webhook_event_processed",
        "\"idempotent\"",
    )
    for token in required_api:
        if token not in api_text:
            missing.append(f"payments.py: {token}")
    required_service = (
        "idempotency_key",
        "create_kwargs",
        "stripe.PaymentIntent.create",
    )
    for token in required_service:
        if token not in service_text:
            missing.append(f"services/payment.py: {token}")

    if not missing:
        return _ok(
            "ecommerce-payments-router",
            "결제 MVP 라우팅",
            "medium",
            target,
            "결제 라우트 + 웹훅 멱등성 + PaymentIntent 멱등키가 적용되어 있습니다.",
            "웹훅 서명키/이벤트 캐시 파일 경로를 운영값으로 주기 점검하세요.",
        )

    return _fail(
        "ecommerce-payments-router",
        "결제 MVP 라우팅",
        "medium",
        target,
        f"누락: {', '.join(missing)}",
        "결제 라우트 3종(intent/webhook/mark-paid), 웹훅 멱등성, 서비스 멱등키 적용을 복구하세요.",
    )


def check_growth_report_integrity(root: Path) -> Finding:
    rel = "core/system/growth.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "growth-report-integrity",
            "성장 리포트 JSON 무결성",
            "medium",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "월간 리포트 생성부에서 JSON 구조가 유지되는지 복구하세요.",
        )

    required = (
        "data[\"monthly_report_markdown\"] = report",
        "report_path = self.save_month(data)",
    )
    missing = [token for token in required if token not in text]
    if not missing:
        return _ok(
            "growth-report-integrity",
            "성장 리포트 JSON 무결성",
            "medium",
            rel,
            "리포트 마크다운이 JSON 필드로 저장되어 지표 데이터가 보존됩니다.",
            "필드명을 유지해 하위 소비 모듈과 호환성을 지키세요.",
        )
    return _fail(
        "growth-report-integrity",
        "성장 리포트 JSON 무결성",
        "medium",
        rel,
        f"누락: {', '.join(missing)}",
        "마크다운을 별도 필드로 저장하고 save_month(data) 호출 흐름을 유지하세요.",
    )


def check_queue_timestamp_parsing(root: Path) -> Finding:
    rel = "core/system/queue_manager.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "queue-timestamp-parse",
            "큐 이벤트 타임스탬프 파싱",
            "medium",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "이벤트 타임스탬프 파서와 사용 지점을 복구하세요.",
        )

    required = (
        "def _parse_event_timestamp",
        "datetime.fromisoformat(raw_timestamp)",
        "event_time = self._parse_event_timestamp(event.timestamp)",
    )
    missing = [token for token in required if token not in text]
    if not missing:
        return _ok(
            "queue-timestamp-parse",
            "큐 이벤트 타임스탬프 파싱",
            "medium",
            rel,
            "신규/레거시 타임스탬프를 모두 파싱하는 방어 로직이 있습니다.",
            "이벤트 파일 포맷 변경 시 파서 테스트를 먼저 추가하세요.",
        )
    return _fail(
        "queue-timestamp-parse",
        "큐 이벤트 타임스탬프 파싱",
        "medium",
        rel,
        f"누락: {', '.join(missing)}",
        "_parse_event_timestamp 도입 후 get_events에서 파서 결과를 사용하세요.",
    )


def check_admin_password_persistence(root: Path) -> Finding:
    rel = "core/admin/app.py"
    text = _read_text(root, rel)
    if text is None:
        return _fail(
            "admin-password-persistence",
            "관리자 비밀번호 영속 저장",
            "low",
            rel,
            "핵심 파일을 찾지 못했습니다.",
            "비밀번호 해시 override 저장/로드 로직을 복구하세요.",
        )

    required = (
        "ADMIN_PASSWORD_OVERRIDE_FILE",
        "def _load_admin_password_hash",
        "def _save_admin_password_hash",
    )
    missing = [token for token in required if token not in text]
    if not missing:
        return _ok(
            "admin-password-persistence",
            "관리자 비밀번호 영속 저장",
            "low",
            rel,
            "비밀번호 변경 시 파일 저장 기반 영속화 로직이 확인됩니다.",
            "권한 관리 정책에 맞춰 저장 파일 접근권한만 유지하세요.",
        )
    return _fail(
        "admin-password-persistence",
        "관리자 비밀번호 영속 저장",
        "low",
        rel,
        f"누락: {', '.join(missing)}",
        "override 파일 경로와 load/save 함수 2개를 함께 복구하세요.",
    )


CHECKS: tuple[CheckFn, ...] = (
    check_fastapi_admin_headers,
    check_photo_upload_security,
    check_ecommerce_secret_policy,
    check_admin_seed_policy,
    check_payments_router,
    check_growth_report_integrity,
    check_queue_timestamp_parsing,
    check_admin_password_persistence,
)


def run_assessment(root: Path) -> list[Finding]:
    return [check(root) for check in CHECKS]


def _build_summary(findings: list[Finding]) -> dict[str, object]:
    total = len(findings)
    passed = sum(1 for f in findings if f.status == "pass")
    failed = total - passed

    failed_by_severity = {
        severity: sum(1 for f in findings if f.status == "fail" and f.severity == severity)
        for severity in SEVERITY_ORDER
    }

    if failed_by_severity["high"]:
        overall_risk = "높음"
    elif failed_by_severity["medium"]:
        overall_risk = "중간"
    elif failed:
        overall_risk = "낮음"
    else:
        overall_risk = "안정"

    return {
        "total": total,
        "passed": passed,
        "failed": failed,
        "failed_by_severity": failed_by_severity,
        "overall_risk": overall_risk,
    }


def _build_actions(findings: list[Finding]) -> list[str]:
    failed = [f for f in findings if f.status == "fail"]
    if not failed:
        return [
            "즉시 조치 항목은 없습니다.",
            "배포 전 이 보고서를 다시 실행해 회귀 여부만 확인하세요.",
            "월 1회 기준으로 정책/키/권한 설정 점검을 유지하세요.",
        ]

    ranked = sorted(
        failed,
        key=lambda f: (SEVERITY_ORDER.index(f.severity), f.check_id),
    )
    actions: list[str] = []
    for item in ranked[:5]:
        severity = SEVERITY_LABEL[item.severity]
        actions.append(f"[{severity}] {item.title}: {item.recommendation}")
    return actions


def _build_ralph_loop(findings: list[Finding], target_score: float, max_loops: int) -> dict[str, object]:
    by_id = {item.check_id: item for item in findings}
    stages: list[dict[str, object]] = []
    total_score = 0.0

    for stage_id, stage_label, check_ids in RALPH_STAGES:
        stage_items = [by_id[cid] for cid in check_ids if cid in by_id]
        total_checks = len(stage_items)
        passed_checks = sum(1 for item in stage_items if item.status == "pass")
        if total_checks == 0:
            earned = 0.0
            ratio = 0.0
        else:
            ratio = passed_checks / total_checks
            earned = round(RALPH_STAGE_MAX * ratio, 1)
        total_score += earned
        stages.append(
            {
                "stage_id": stage_id,
                "label": stage_label,
                "max_points": RALPH_STAGE_MAX,
                "earned_points": earned,
                "passed_checks": passed_checks,
                "total_checks": total_checks,
                "pass_ratio": round(ratio, 4),
            }
        )

    target = float(max(0.0, min(100.0, target_score)))
    total_score = round(total_score, 1)
    gap = round(max(0.0, target - total_score), 1)

    failed_ranked = sorted(
        [item for item in findings if item.status == "fail"],
        key=lambda item: (SEVERITY_ORDER.index(item.severity), item.check_id),
    )
    loop_actions: list[dict[str, object]] = []
    for loop_idx, item in enumerate(failed_ranked[:max_loops], start=1):
        loop_actions.append(
            {
                "loop": loop_idx,
                "check_id": item.check_id,
                "title": item.title,
                "severity": item.severity,
                "target_file": item.target_file,
                "action": item.recommendation,
            }
        )

    return {
        "enabled": True,
        "stage_max_points": RALPH_STAGE_MAX,
        "stages": stages,
        "total_score": total_score,
        "target_score": target,
        "gap_to_target": gap,
        "target_met": total_score >= target,
        "max_loops": max_loops,
        "loop_actions": loop_actions,
    }


def build_report_payload(
    findings: list[Finding],
    checked_at: datetime,
    root: Path,
    include_ralph_loop: bool = False,
    target_score: float = 90.0,
    max_loops: int = 3,
) -> dict[str, object]:
    summary = _build_summary(findings)
    actions = _build_actions(findings)
    payload: dict[str, object] = {
        "checked_at": checked_at.isoformat(),
        "root": str(root.resolve()),
        "summary": summary,
        "findings": [asdict(f) for f in findings],
        "actions": actions,
    }
    if include_ralph_loop:
        payload["ralph_loop"] = _build_ralph_loop(
            findings=findings,
            target_score=target_score,
            max_loops=max_loops,
        )
    return payload


def render_text_report(payload: dict[str, object]) -> str:
    summary = payload["summary"]
    findings = payload["findings"]
    actions = payload["actions"]

    failed_items = [item for item in findings if item["status"] == "fail"]

    lines: list[str] = []
    lines.append("[백엔드 안정화 쉬운 보고서]")
    lines.append(f"점검 시각: {payload['checked_at']}")
    lines.append(f"점검 경로: {payload['root']}")
    lines.append(
        "총 %d건 중 통과 %d건, 보완 필요 %d건 | 현재 위험도: %s"
        % (
            summary["total"],
            summary["passed"],
            summary["failed"],
            summary["overall_risk"],
        )
    )
    lines.append("")

    lines.append("1) 위험도별 요약")
    for severity in SEVERITY_ORDER:
        label = SEVERITY_LABEL[severity]
        count = summary["failed_by_severity"][severity]
        lines.append(f"- {label}: {count}건")
    lines.append("")

    lines.append("2) 미완성/보완 필요 항목")
    if not failed_items:
        lines.append("- 없음 (현재 기준 회귀 징후 미탐지)")
    else:
        for item in failed_items:
            label = SEVERITY_LABEL[item["severity"]]
            lines.append(
                "- [%s] %s | 파일: %s | 상태: %s"
                % (label, item["title"], item["target_file"], item["summary"])
            )
    lines.append("")

    lines.append("3) 권장 다음 액션")
    for idx, action in enumerate(actions, start=1):
        lines.append(f"{idx}. {action}")

    ralph_loop = payload.get("ralph_loop")
    if isinstance(ralph_loop, dict):
        lines.append("")
        lines.append("4) 랄프루프 진단")
        lines.append(
            "- 총점 %.1f / 목표 %.1f (갭 %.1f)"
            % (
                ralph_loop.get("total_score", 0.0),
                ralph_loop.get("target_score", 90.0),
                ralph_loop.get("gap_to_target", 0.0),
            )
        )
        lines.append(
            "- 목표 달성: %s"
            % ("달성" if ralph_loop.get("target_met") else "미달")
        )
        stage_rows = ralph_loop.get("stages", [])
        if isinstance(stage_rows, list):
            for stage in stage_rows:
                if not isinstance(stage, dict):
                    continue
                lines.append(
                    "- %s: %.1f/%.1f (%d/%d 통과)"
                    % (
                        stage.get("label", "단계"),
                        stage.get("earned_points", 0.0),
                        stage.get("max_points", RALPH_STAGE_MAX),
                        stage.get("passed_checks", 0),
                        stage.get("total_checks", 0),
                    )
                )

        lines.append("")
        lines.append("5) 루프 실행 플랜")
        loop_actions = ralph_loop.get("loop_actions", [])
        if isinstance(loop_actions, list) and loop_actions:
            for idx, loop in enumerate(loop_actions, start=1):
                if not isinstance(loop, dict):
                    continue
                severity = SEVERITY_LABEL.get(str(loop.get("severity", "low")), "낮음")
                lines.append(
                    "%d. [%s] %s (%s) -> %s"
                    % (
                        idx,
                        severity,
                        loop.get("title", "항목"),
                        loop.get("target_file", "unknown"),
                        loop.get("action", ""),
                    )
                )
        else:
            lines.append("- 현재 실패 항목이 없어 루프 액션이 없습니다.")

    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description="백엔드 하드닝 상태를 읽기 전용으로 보고합니다.")
    parser.add_argument(
        "--root",
        default=str(PROJECT_ROOT),
        help="점검할 프로젝트 루트 경로",
    )
    parser.add_argument(
        "--format",
        choices=("text", "json"),
        default="text",
        help="출력 형식 (기본: text)",
    )
    parser.add_argument(
        "--fail-on-high",
        action="store_true",
        help="높음(high) 위험 실패가 있으면 종료코드 1 반환",
    )
    parser.add_argument(
        "--ralph-loop",
        action="store_true",
        help="4단계 랄프루프 진단/루프 플랜 섹션 추가",
    )
    parser.add_argument(
        "--target-score",
        type=float,
        default=90.0,
        help="랄프루프 목표 점수 (기본: 90)",
    )
    parser.add_argument(
        "--max-loops",
        type=int,
        default=3,
        help="루프 실행 플랜 최대 항목 수 (기본: 3)",
    )
    args = parser.parse_args()
    if args.target_score < 0 or args.target_score > 100:
        parser.error("--target-score는 0~100 범위여야 합니다.")
    if args.max_loops < 1:
        parser.error("--max-loops는 1 이상이어야 합니다.")

    root = Path(args.root).resolve()
    checked_at = datetime.now(timezone.utc)
    findings = run_assessment(root)
    payload = build_report_payload(
        findings=findings,
        checked_at=checked_at,
        root=root,
        include_ralph_loop=args.ralph_loop,
        target_score=args.target_score,
        max_loops=args.max_loops,
    )

    if args.format == "json":
        print(json.dumps(payload, ensure_ascii=False, indent=2))
    else:
        print(render_text_report(payload))

    if args.fail_on_high and any(
        f.status == "fail" and f.severity == "high" for f in findings
    ):
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
