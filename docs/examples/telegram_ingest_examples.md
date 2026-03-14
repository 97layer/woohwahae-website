# Telegram Ingest Examples (Phase 1.5)

Use canonical interaction profiles so topics/kinds/tags/dedupe refs stay stable.

## Inbound Conversation

```bash
go run ./cmd/layer-osctl ingest telegram \
  --interaction inbound_conversation \
  --message-id 12345 \
  --chat founder \
  --excerpt "Founder ping about deploy gate" \
  --dry-run
```

Expected refs include `interaction_direction:inbound`, `tag:founder_inbox`, `content_kind:message`, `telegram_message:12345`.

## Outbound Publication

```bash
go run ./cmd/layer-osctl ingest telegram \
  --interaction outbound_publication \
  --title "Weekly release note" \
  --excerpt "Cutover scheduled after security gate" \
  --dry-run
```

Expected refs include `interaction_direction:outbound`, `tag:publication`, `content_kind:publication`.

## Feedback Reply

```bash
go run ./cmd/layer-osctl ingest telegram \
  --interaction feedback_reply \
  --excerpt "Release window is too tight" \
  --username partner_a \
  --message-id 778899 \
  --dry-run
```

Expected refs include `interaction_direction:inbound`, `tag:feedback`, `telegram_user:partner_a`, `telegram_message:778899`.

## YouTube Link Injection

```bash
go run ./cmd/layer-osctl ingest telegram \
  --interaction youtube_link_injection \
  --url "https://youtu.be/example" \
  --excerpt "See minute 2:15 for rollout pattern" \
  --dry-run
```

Expected refs include `content_kind:link`, `tag:youtube`, and normalized `content_host:youtube_com`.

## Bookmark Share

```bash
go run ./cmd/layer-osctl ingest telegram \
  --interaction bookmark_share \
  --url "https://example.com/case-study" \
  --excerpt "Relevant rollout case" \
  --dry-run
```

Expected refs include `interaction_direction:outbound`, `tag:bookmark`, and dedupe ref `content_doc:url_<hash>` when a URL is provided.

## Related Docs

- `docs/operator.md` lays out the wrapper surface where these ingestion samples are invoked.
- `docs/agent-quickstart.md` clarifies that scoped implementer lanes can tidy these docs without touching hot seams.
