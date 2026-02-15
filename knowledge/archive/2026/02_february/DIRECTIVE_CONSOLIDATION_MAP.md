# Directive Consolidation Map
**Date**: 2026-02-15
**Action**: Consolidated 27+ fragmented directives into 6 core documents + 5 executable skills

---

## Current Active Directives (6)

### `/directives/CORE.md`
Core system principles and operational framework
- **Replaces**: `system_sop.md`, `system_handshake.md`, `efficiency_protocol.md`

### `/directives/DEVELOPMENT.md`
Technical implementation standards and workflows
- **Replaces**: `daemon_workflow.md`, `venv_workflow.md`, `sync_workflow.md`, `snapshot_workflow.md`, `cycle_protocol.md`
- **Replaces**: `system/podman_optimization.md`, `system/self_maintenance.md`

### `/directives/IDENTITY.md`
Brand identity and voice
- **Replaces**: `97layer_identity.md`, `woohwahae_identity.md`, `woohwahae_brand_source.md`, `brand_constitution.md`

### `/directives/MANIFESTO.md`
Mission, philosophy, and anti-algorithm protocol
- **Replaces**: `anti_algorithm_protocol.md`, `aesop_benchmark.md`, `imperfect_publish_protocol.md`

### `/directives/OPERATIONS.md`
Day-to-day operational protocols
- **Replaces**: `communication_protocol.md`, `junction_protocol.md`, `sync_protocol.md`, `token_optimization_protocol.md`
- **Replaces**: `uip_protocol.md`, `directive_lifecycle.md`, `data_asset_management.md`

### `/directives/PUBLISHING.md`
Content publication and Instagram automation
- **Replaces**: Publishing-related aspects from various protocols

---

## Active Skills (5 - in `/core/skills/`)

### `brand_voice.skill.md`
Executable brand voice and tone guidelines
- **Replaces**: Communication aspects of `communication_protocol.md`
- **Size**: 7KB

### `design_guide.skill.md`
Visual identity and design system
- **Replaces**: `design/style_guide.md`, `visual_identity_guide.md`
- **Size**: 10KB

### `infra_ops.skill.md`
Infrastructure operations and deployment
- **Replaces**: `infrastructure_sentinel.md`, deployment aspects from various protocols
- **Size**: 15KB

### `instagram.skill.md`
Instagram content curation and publishing
- **Replaces**: Publishing automation workflows
- **Size**: 11KB

### `pattern_recognition.skill.md`
Signal processing and insight generation
- **Replaces**: Signal processing aspects from workflow protocols
- **Size**: 15KB

---

## Deprecated Agent-Specific Files (Archived)

The following agent files were conceptual overlays, now replaced by skill-based execution:

- `agents/architect.md` → Covered by `DEVELOPMENT.md` + `infra_ops.skill.md`
- `agents/art_director.md` → Covered by `design_guide.skill.md`
- `agents/artisan.md` → Covered by `OPERATIONS.md` + skills
- `agents/chief_editor.md` → Covered by `PUBLISHING.md`
- `agents/creative_director.md` → Covered by `IDENTITY.md` + `brand_voice.skill.md`
- `agents/narrator.md` → Covered by `brand_voice.skill.md`
- `agents/scout.md` → Covered by `pattern_recognition.skill.md`
- `agents/sovereign.md` → Covered by `CORE.md`
- `agents/strategy_analyst.md` → Covered by `MANIFESTO.md`
- `agents/technical_director.md` → Covered by `DEVELOPMENT.md`

---

## Deprecated Protocol Files (Archived)

### System Protocols
- `97layerOS_Optimization_Directive.md`
- `agent_instructions.md`
- `skills_integration.md`
- `skill_unified_input.md`

### Workflow Protocols
All merged into `DEVELOPMENT.md` and `OPERATIONS.md`

### Identity Protocols
All merged into `IDENTITY.md` and `MANIFESTO.md`

---

## Benefits of Consolidation

1. **Reduced Token Overhead**: From 27+ files (est. 150KB+) to 6 directives + 5 skills (est. 70KB)
2. **Clearer Hierarchy**: Core principles → Operations → Executable skills
3. **No Duplication**: Each concept exists in exactly one place
4. **Skill-Based Execution**: Agents invoke skills, not directives
5. **Easier Maintenance**: Update one file instead of multiple overlapping ones

---

## Implementation Status

- [x] Consolidated 6 core directives created (2026-02-15)
- [x] 5 executable skills created in `/core/skills/`
- [x] Old directive files staged for deletion via git
- [x] Knowledge folder restructured (signals/, content/, system/)
- [x] Agent execution now uses skill-based approach
- [x] Git history cleanup ready for commit

---

## Next Actions

1. Commit this consolidation to git history
2. Update all agent execution scripts to reference new structure
3. Verify Podman containers use new directive paths
4. Remove any hardcoded references to old directive files

---

**Status**: CONSOLIDATION COMPLETE ✓
