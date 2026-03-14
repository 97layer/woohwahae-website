# Layer OS

Layer OS is the next-generation runtime rebuilt from the lived experience of `97layerOS`.

This workspace inherits constitutional direction only.
This workspace keeps only active constitutional documents and a lineage note.
This workspace keeps a CLI-first skill library as a first-class source axis.

The active constitution is:
- `constitution/charter.md`
- `constitution/lineage.md`
- `constitution/origin.md`
- `constitution/brand.md`
- `constitution/voice.md`
- `constitution/surface.md`
- `constitution/orchestra.md`

It does not inherit Python runtime structure.

## Path Map

- Current active constitutional docs live in `constitution/` inside this workspace: `/Users/97layer/layer OS/constitution`.
- Legacy `97layerOS` philosophy/reference docs live in sibling workspace `97layerOS` under `directives/`.
- Legacy keep/port/drop inventory lives in `docs/legacy_inventory.md`.
- Bounded legacy/operator/verifier support docs live in `docs/legacy_session.md`, `docs/operator.md`, and `docs/verifier.md`.
- Agent startup docs live in `docs/agent-quickstart.md`, `docs/agent-role-seeds.md`, and `docs/agent-integration.md`.
- If a shell scanner targets the current workspace with an absolute path, it must quote `/Users/97layer/layer OS/...` because the root folder contains a space.

## Scope

- Go runtime only
- Prompt operating model in `docs/prompting.md`
- Contract-first architecture
- Founder-led approval model
- Agent-friendly naming grammar
- Standard-library-first implementation
- CLI + skill + terminal workflow

## Dev Quick Checks

- `GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test -cover ./...`
- `GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache GOLANGCI_LINT_TEMP_DIR=/tmp/golangci-lint-temp golangci-lint run`
- 또는 편의 스크립트: `scripts/dev_checks.sh` (`--lint-only` | `--test-only` 지원)

## Agent Startup (Self-Start)

- 공용 worker 규칙과 self-start 동작: `docs/agent-quickstart.md`
- 한국어 역할별 복붙 시드: `docs/agent-role-seeds.md`
- 외부 에이전트 패킷/보고 계약: `docs/agent-integration.md`
- 실행 예시: `docs/examples/content_ingest_examples.md`, `docs/examples/telegram_ingest_examples.md`

프롬프트 복붙은 여기서 다시 들고 있지 않는다. canonical prompt text는 `docs/agent-quickstart.md`와 `docs/agent-role-seeds.md`만 본다.

## Compact Scaffold

```text
layer OS/
  constitution/   active constitutional intent
  contracts/      typed JSON contracts
  docs/           operator and architecture reference
  skills/         active/imported skill grammar
  cmd/            layer-osctl + layer-osd entrypoints
  internal/       api + runtime only
  scripts/        local bootstrap and operator harness
  bin/            local build outputs only
  .layer-os/      runtime state only
```

- source scaffold stays inside `constitution`, `contracts`, `docs`, `skills`, `cmd`, `internal`, and `scripts`
- `bin/` is not a source root; it is the only allowed local binary output path
- runtime state lives in `.layer-os/` and is never a source authority
- canonical CLI/API surface inventory lives in `internal/runtime/surface_spec.go`
- `layer-osctl audit surface` is the drift check for README, CLI dispatch, and API routes
- `contracts/` holds both core runtime contracts and harness/transport contracts such as `AgentJob`, `AgentRunPacket`, `SessionBootstrapPacket`, `ObservationRecord`, and `TelegramPacket`

See:
- `docs/architecture.md`
- `docs/grammar.md`
- `docs/operator.md`
- `skills/README.md`

## Run

```bash
./scripts/start.sh --check
./scripts/start.sh
```

```bash
./scripts/control.sh
./scripts/control.sh watch --interval 5
```

```bash
go run ./cmd/layer-osctl status
go run ./cmd/layer-osctl handoff
go run ./cmd/layer-osctl knowledge
```

- `./scripts/start.sh` is the canonical local bootstrap
- `./scripts/control.sh` is the compact operator report and worktree lane view
- full wrapper examples, worker loop usage, and authenticated mutation flows live in `docs/operator.md`

## Validate

```bash
go test ./...
```

```bash
go run ./cmd/layer-osctl audit docs --strict
go run ./cmd/layer-osctl audit contracts --strict
go run ./cmd/layer-osctl audit surface --strict
go run ./cmd/layer-osctl audit structure --strict
```
