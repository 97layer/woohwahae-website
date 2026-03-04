"""
Photo upload security regression tests.
"""

from __future__ import annotations

import importlib
import sys

import pytest
from fastapi.testclient import TestClient


@pytest.fixture
def photo_upload_client(monkeypatch, tmp_path):
    monkeypatch.setenv("PHOTO_UPLOAD_ADMIN_TOKEN", "upload-test-token")
    if "core.backend.photo_upload" in sys.modules:
        del sys.modules["core.backend.photo_upload"]

    module = importlib.import_module("core.backend.photo_upload")
    module.UPLOAD_DIR = tmp_path / "uploads"
    module.UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
    for category in module.CATEGORIES:
        (module.UPLOAD_DIR / category).mkdir(parents=True, exist_ok=True)
    module.METADATA_FILE = module.UPLOAD_DIR / "metadata.json"
    module.save_metadata([])

    with TestClient(module.app) as client:
        yield client, module


def _auth_headers() -> dict[str, str]:
    return {"Authorization": "Bearer upload-test-token"}


def test_images_requires_auth(photo_upload_client):
    client, _ = photo_upload_client
    resp = client.get("/api/images")
    assert resp.status_code == 401


def test_upload_rejects_invalid_extension(photo_upload_client):
    client, _ = photo_upload_client
    resp = client.post(
        "/api/upload",
        headers=_auth_headers(),
        data={"category": "hair"},
        files=[("photos", ("bad.txt", b"not-image", "text/plain"))],
    )
    assert resp.status_code == 400


def test_upload_rejects_svg(photo_upload_client):
    client, _ = photo_upload_client
    resp = client.post(
        "/api/upload",
        headers=_auth_headers(),
        data={"category": "hair"},
        files=[("photos", ("vector.svg", b"<svg></svg>", "image/svg+xml"))],
    )
    assert resp.status_code == 400
    assert "Invalid file extension" in resp.json()["detail"]


def test_upload_rejects_mismatched_signature(photo_upload_client):
    client, _ = photo_upload_client
    resp = client.post(
        "/api/upload",
        headers=_auth_headers(),
        data={"category": "hair"},
        files=[("photos", ("fake.png", b"not-a-real-png", "image/png"))],
    )
    assert resp.status_code == 400
    assert "Invalid image signature" in resp.json()["detail"]


def test_upload_and_list_with_auth(photo_upload_client):
    client, _ = photo_upload_client

    upload = client.post(
        "/api/upload",
        headers=_auth_headers(),
        data={"category": "hair"},
        files=[("photos", ("sample.png", b"\x89PNG\r\n\x1a\nPNGDATA", "image/png"))],
    )
    assert upload.status_code == 200
    payload = upload.json()
    assert payload["success"] is True
    assert payload["count"] == 1
    assert payload["files"][0]["category"] == "hair"

    listed = client.get("/api/images", headers=_auth_headers())
    assert listed.status_code == 200
    assert len(listed.json()) == 1


def test_upload_rejects_oversized_file(photo_upload_client):
    client, module = photo_upload_client
    module.MAX_UPLOAD_BYTES = 4
    resp = client.post(
        "/api/upload",
        headers=_auth_headers(),
        data={"category": "hair"},
        files=[("photos", ("large.png", b"12345", "image/png"))],
    )
    assert resp.status_code == 400
    assert "File too large" in resp.json()["detail"]
