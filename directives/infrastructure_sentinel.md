# 97LAYER Infrastructure Sentinel SOP

## Identity

This system is a self-healing infrastructure watchdog designed to maintain 97layerOS in a state of "Pure Intelligence".

## Core Mandate: The Sentinel Logic

The `snapshot_daemon.py` acts as the Sentinel by executing a **Continuous Sanitization** sequence every hour.

### 1. Auto-Purge Protocol

Before every snapshot, the Sentinel automatically removes:

- **Redundant Venvs**: `.venv_runtime`, `venv_new`, etc.
- **Dependency Leaks**: Any folder named `node_modules` at any depth.
- **System Noise**: `.local_node`, `.mcp-source`, `.tmp` files.

### 2. Physical Blockade

- Venv folders (`venv`, `venv_clean`, etc.) are under kernel-level protection (`uchg` flag) or aggressive deletion logic.
- Do NOT attempt to create a local virtual environment within the root directory; it will be purged.

### 3. Shadow Copy Mechanism

To bypass macOS/Cloud sync permission locks, security files like `token.json` are archived via a **Shadow Copy** in `/tmp/97layer_shadow_copy/`.

## Maintenance

If a required file is accidentally purged:

1. Stop the daemon: `pkill -f snapshot_daemon.py`
2. Update the exclusion list in `execution/create_snapshot.py` and `snapshot_daemon.py`.
3. Restart the Sentinel.

> **Status**: ACTIVE & VIGILANT
