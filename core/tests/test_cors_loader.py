"""CORS origin loader tests."""

from core.system.security import load_cors_origins


def test_load_cors_origins_prefers_single_source(monkeypatch):
    monkeypatch.setenv("CORS_ORIGINS", "https://legacy.example")
    monkeypatch.setenv("FASTAPI_CORS_ORIGINS", "https://fastapi.example")
    monkeypatch.setenv("BACKEND_CORS_ORIGINS", "https://single.example")

    origins = load_cors_origins(default=["https://fallback.example"])

    assert origins == ["https://single.example"]


def test_load_cors_origins_parses_and_dedupes(monkeypatch):
    monkeypatch.delenv("BACKEND_CORS_ORIGINS", raising=False)
    monkeypatch.delenv("FASTAPI_CORS_ORIGINS", raising=False)
    monkeypatch.delenv("CORS_ORIGINS", raising=False)

    parsed = load_cors_origins(
        default=["https://fallback.example", "https://fallback.example"],
    )
    assert parsed == ["https://fallback.example"]

    monkeypatch.setenv(
        "BACKEND_CORS_ORIGINS",
        " https://a.example , https://b.example,https://a.example ",
    )
    parsed = load_cors_origins(default=["https://fallback.example"])
    assert parsed == ["https://a.example", "https://b.example"]
