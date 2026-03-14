# Content Ingest Examples (Phase 1.5)

Use `--profile` to stamp canonical kinds/topics/tags/dedupe hints.

## Crawler Article

```bash
go run ./cmd/layer-osctl ingest content \
  --channel crawler \
  --profile article \
  --url "https://example.com/story?utm_source=twitter&fbclid=abc" \
  --title "OS adoption case study" \
  --excerpt "How a single kernel scaled." \
  --dry-run
```

Expected refs include `tag:crawler`, `content_kind:article`, `content_doc:url_<hash>` with tracking params stripped.

## Personal Archive Outline

```bash
go run ./cmd/layer-osctl ingest content \
  --channel personal_db \
  --profile outline \
  --title "Q2 release outline" \
  --excerpt "Steps and owners" \
  --doc-id outline_q2_release \
  --dry-run
```

Expected refs include `source_family:founder_archive`, `tag:outline`, `content_doc:outline_q2_release`, `content_kind:outline`.

## NotebookLM QA

```bash
go run ./cmd/layer-osctl ingest content \
  --channel notebook_lm \
  --profile qa \
  --title "LLM evaluation QA" \
  --excerpt "Why single-kernel matters?" \
  --dry-run
```

Expected refs include `source_family:founder_archive`, `tag:notebook_lm`, `content_kind:qa`, and a `content_doc:notebooklm_auto` hint if no doc-id is provided.

## Crawler Video

```bash
go run ./cmd/layer-osctl ingest content \
  --channel crawler \
  --profile video \
  --url "https://www.youtube.com/watch?v=dQw4w9WgXcQ" \
  --title "Release demo" \
  --dry-run
```

Expected refs include `content_kind:video`, `content_host:youtube_com`, and dedupe ref `content_doc:url_<hash>`.

## Related Docs

- `docs/operator.md` documents the wrapper commands that surface these ingest samples.
- `docs/agent-quickstart.md` explains how self-start agents should treat docs cleanup lanes before touching these examples.
