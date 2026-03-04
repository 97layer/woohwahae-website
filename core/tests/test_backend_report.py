"""Backend report-only logic tests."""

from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path

from core.system import backend_report


def _write(root: Path, rel_path: str, content: str) -> None:
    target = root / rel_path
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(content, encoding="utf-8")


def _seed_minimal_project(root: Path) -> None:
    _write(
        root,
        "core/backend/main.py",
        """
def _extract_admin_token(request):
    auth = request.headers.get(\"authorization\", \"\").strip()
    return request.headers.get(\"x-admin-token\", \"\").strip()

async def get_page_contents(request):
    check_admin(request)

async def get_content(request):
    check_admin(request)

async def get_all_pages(request):
    check_admin(request)
""",
    )
    _write(
        root,
        "core/backend/photo_upload.py",
        """
PHOTO_UPLOAD_ADMIN_TOKEN = ""
MAX_UPLOAD_BYTES = 100
MAX_UPLOAD_FILES = 10

def _require_admin_auth(request):
    return None
""",
    )
    _write(
        root,
        "core/backend/ecommerce/config.py",
        """
def _load_jwt_secret():
    raise RuntimeError(\"JWT_SECRET_KEY environment variable is required\")

JWT_SECRET_KEY = _load_jwt_secret()
""",
    )
    _write(
        root,
        "core/backend/ecommerce/migrations/init_db.py",
        """
admin_email = os.getenv("ECOMMERCE_ADMIN_EMAIL")
admin_password = os.getenv("ECOMMERCE_ADMIN_PASSWORD")
if len(admin_password) < 12:
    raise RuntimeError("최소 12자")
""",
    )
    _write(
        root,
        "core/backend/ecommerce/main.py",
        """
app.include_router(payments_router, prefix=API_V1_PREFIX)
""",
    )
    _write(
        root,
        "core/backend/ecommerce/api/payments.py",
        """
@router.post(\"/intent\")
def intent():
    pass

@router.post(\"/webhook\")
def webhook():
    pass

@router.post(\"/orders/{order_id}/mark-paid\")
def mark_paid():
    pass
""",
    )
    _write(
        root,
        "core/system/growth.py",
        """
data[\"monthly_report_markdown\"] = report
report_path = self.save_month(data)
""",
    )
    _write(
        root,
        "core/system/queue_manager.py",
        """
def _parse_event_timestamp(raw_timestamp):
    return datetime.fromisoformat(raw_timestamp)

def get_events(self):
    event_time = self._parse_event_timestamp(event.timestamp)
""",
    )
    _write(
        root,
        "core/admin/app.py",
        """
ADMIN_PASSWORD_OVERRIDE_FILE = BASE_DIR / 'knowledge' / 'system' / 'admin_password_hash.json'

def _load_admin_password_hash(default_hash):
    return default_hash

def _save_admin_password_hash(password_hash):
    return None
""",
    )


def test_backend_report_all_pass(tmp_path):
    _seed_minimal_project(tmp_path)

    findings = backend_report.run_assessment(tmp_path)

    assert findings
    assert all(item.status == "pass" for item in findings)


def test_backend_report_detects_failure_and_action(tmp_path):
    _seed_minimal_project(tmp_path)
    _write(
        tmp_path,
        "core/backend/main.py",
        """
def _extract_admin_token(request):
    token = request.query_params.get('token')
    return token

async def get_page_contents(request):
    pass

async def get_content(request):
    pass

async def get_all_pages(request):
    pass
""",
    )

    findings = backend_report.run_assessment(tmp_path)
    payload = backend_report.build_report_payload(
        findings=findings,
        checked_at=datetime.now(timezone.utc),
        root=tmp_path,
    )

    failed = [item for item in findings if item.status == "fail"]
    assert failed
    assert any(item.check_id == "fastapi-admin-header-auth" for item in failed)
    assert payload["summary"]["failed"] >= 1
    assert payload["actions"]


def test_backend_report_text_sections(tmp_path):
    _seed_minimal_project(tmp_path)
    findings = backend_report.run_assessment(tmp_path)
    payload = backend_report.build_report_payload(
        findings=findings,
        checked_at=datetime.now(timezone.utc),
        root=tmp_path,
    )

    text = backend_report.render_text_report(payload)

    assert "1) 위험도별 요약" in text
    assert "2) 미완성/보완 필요 항목" in text
    assert "3) 권장 다음 액션" in text


def test_backend_report_ralph_loop_extension(tmp_path):
    _seed_minimal_project(tmp_path)
    _write(
        tmp_path,
        "core/backend/main.py",
        """
def _extract_admin_token(request):
    return request.headers.get("x-admin-token", "")

async def get_page_contents(request):
    pass

async def get_content(request):
    check_admin(request)

async def get_all_pages(request):
    check_admin(request)
""",
    )
    _write(
        tmp_path,
        "core/backend/ecommerce/api/payments.py",
        """
@router.post("/intent")
def intent():
    pass

@router.post("/webhook")
def webhook():
    pass
""",
    )

    findings = backend_report.run_assessment(tmp_path)
    payload = backend_report.build_report_payload(
        findings=findings,
        checked_at=datetime.now(timezone.utc),
        root=tmp_path,
        include_ralph_loop=True,
        target_score=95.0,
        max_loops=1,
    )
    text = backend_report.render_text_report(payload)

    ralph = payload["ralph_loop"]
    assert len(ralph["stages"]) == 4
    assert ralph["target_score"] == 95.0
    assert ralph["total_score"] < 100.0
    assert len(ralph["loop_actions"]) == 1
    assert "4) 랄프루프 진단" in text
    assert "5) 루프 실행 플랜" in text
