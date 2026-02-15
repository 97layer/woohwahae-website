# System Directory

## Container-First Architecture

This directory is a placeholder for local development. All system state files are stored in:
- Container: `/app/.container_data/knowledge/system/`
- Local: `PROJECT_ROOT/.container_data/knowledge/system/`

## Purpose

System directory stores runtime state, task boards, and agent loop tracking data.

## Files

- `task_board.json` - Current task status and priority queue
- `ralph_loop.json` - Agent execution loop tracking
- `snapshot_status.json` - System snapshot metadata
- `sync_state.json` - Synchronization state tracking

## Access

Use `KNOWLEDGE_PATHS["system"]` from `libs/core_config.py` to access the environment-aware path.
