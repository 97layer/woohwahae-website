"""safe_env_export script tests."""

from __future__ import annotations

import subprocess
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[2]
SCRIPT = PROJECT_ROOT / "core" / "scripts" / "safe_env_export.py"


def test_safe_env_export_handles_json_value(tmp_path):
    env_file = tmp_path / ".env"
    env_file.write_text(
        "\n".join(
            [
                "GEMINI_API_KEY=test-key",
                'NOTEBOOKLM_AUTH_JSON={"cookies":[{"name":"sid","value":"abc"}]}',
            ]
        ),
        encoding="utf-8",
    )

    proc = subprocess.run(
        ["python3", str(SCRIPT), "--file", str(env_file), "--keys", "NOTEBOOKLM_AUTH_JSON"],
        check=True,
        capture_output=True,
        text=True,
    )

    out = proc.stdout.strip()
    assert out.startswith("export NOTEBOOKLM_AUTH_JSON=")
    assert '{"cookies"' in out


def test_safe_env_export_selective_keys(tmp_path):
    env_file = tmp_path / ".env"
    env_file.write_text(
        "A=1\nB=2\n",
        encoding="utf-8",
    )

    proc = subprocess.run(
        ["python3", str(SCRIPT), "--file", str(env_file), "--keys", "B"],
        check=True,
        capture_output=True,
        text=True,
    )

    assert proc.stdout.strip() == "export B=2"
