"""
LAYER OS System Module  THE ORIGIN Enforcement

Monkey Patch: All Path.write_text() calls are intercepted and validated
against MANIFEST.md rules before execution.

This ensures that even legacy code without explicit safe_write() calls
cannot violate filesystem structure.

Author: THE ORIGIN Agent
Created: 2026-02-26
"""

from pathlib import Path as _Path
import sys

# Import validator
try:
    from core.system.filesystem_validator import validate_write

    # Store original method
    _original_write_text = _Path.write_text

    def _guarded_write_text(self, data, encoding=None, errors=None, newline=None):
        """
        Monkey-patched Path.write_text() with MANIFEST validation.

        Raises:
            PermissionError: If file path violates MANIFEST rules
        """
        ok, reason = validate_write(self)
        if not ok:
            raise PermissionError(
                f"\n{'='*60}\n"
                f"[Filesystem Guard] Write blocked by MANIFEST\n"
                f"{'='*60}\n"
                f"Path: {self}\n"
                f"Reason: {reason}\n"
                f"\nSee: directives/MANIFEST.md\n"
                f"{'='*60}"
            )

        # Call original method
        return _original_write_text(self, data, encoding=encoding, errors=errors, newline=newline)

    # Replace globally
    _Path.write_text = _guarded_write_text

    print("[Filesystem Guard] Monkey patch active  all Path.write_text() calls validated")

except ImportError as e:
    # If validator not available, warn but don't crash
    print(f"[Warning] Filesystem validator not loaded: {e}", file=sys.stderr)
