# Parallel Dispatch — 5-Agent Packets

Historical batch note only.
This document is not active runtime authority.
Use `AGENTS.md`, `docs/architecture.md`, `docs/agent-integration.md`, and `docs/agent-role-seeds.md` for current contracts.

## Intent

현재 페이즈는 **프론트 표면 정렬 + 데이터 해자 입력면 확장**이다.
목표는 속도를 올리되, 지식/상태/계약이 파편화되지 않게 하는 것이다.

이번 배치는 **5개 에이전트 내외**로 고정한다.
현재 기준으로는 특정 공급자 이름에 최종 검증/통합/승격을 고정하지 않는다.
그 시점의 control-tower owner가 최종 검증/통합/승격을 맡고, 외부 워커는 **bounded lane**만 맡는다.

워커를 실제로 띄울 때는 `docs/agent-quickstart.md`와 `docs/agent-role-seeds.md`의 역할별 시드를 먼저 사용한다.
이 문서는 lane ownership과 merge order를 설명하는 배치 문서다.

## 관련 문서

- `docs/agent-quickstart.md`에서 공용 worker 규칙, self-start boot 순서, lane selection을 확인할 수 있다.
- `docs/agent-role-seeds.md`에는 역할별 starter prompt 예시와 hot seam rule이 정리되어 있다.
- `docs/operator.md`와 `docs/examples/`의 내용들은 각 lane이 호출할 Wrapper/ingest 명령 샘플을 보여준다.

## Global Guardrails

모든 워커가 공통으로 따라야 하는 규칙:

- 새 루트 폴더를 만들지 않는다.
- `.layer-os/*.json` 직접 수정 금지.
- `review_room.json` 직접 수정 금지.
- `node_modules/`, `.next/`, generated artifact 수정 금지.
- ad-hoc markdown scratchpad 생성 금지.
- 데이터는 항상 **canonical ingest/report path**로만 연결한다.
- 기존 authority/security 방향을 되돌리는 변경 금지.
- 계약/타입/테스트가 필요한 변경이면 같은 작업 안에서 같이 처리한다.
- 파일 소유권이 겹치지 않게 작업한다. 다른 워커의 파일은 건드리지 않는다.
- 너는 혼자 작업하는 것이 아니다. 다른 워커의 변경을 되돌리거나 덮어쓰지 않는다.
- 결과 보고는 `docs/agent-integration.md`의 공식 7개 키 계약을 그대로 사용한다:
  - `summary`
  - `artifacts`
  - `verification`
  - `open_risks`
  - `follow_on`
  - `touched_paths`
  - `blocked_paths`
- `summary`와 `open_risks` 값은 비개발자도 이해할 수 있는 쉬운 말로 먼저 쓴다.
- 배치 문서가 따로 커스텀 보고 포맷을 만들지 않는다.

## Control-Tower-Owned Vertical Lanes

아래 항목은 병렬 위임 금지. 당시 control-tower owner가 계속 직접 잡는다.

- ontology / scoring / dedupe policy
- authority / security / audit semantics
- review-room final transition
- retrieval / ranking / eval core
- multi-lane merge seam / final validation

## Phase Shape

이번 배치의 우선순위는 아래 순서다.

1. **Front Surface** — 대외 표면을 current constitution에 맞게 정렬
2. **Data Intake** — Telegram / personal_db / NotebookLM / crawler 입력면 확장
3. **Later** — retrieval/ranking, recommendation, higher-order intelligence

---

## Agent 1 — Front IA / Copy Shell

### Mission

브랜드 홈의 **메시지 구조와 정보 위계**를 current constitution 기준으로 정렬한다.
목표는 화려한 구현보다 **무슨 메시지를 어떤 순서로 보여줄지**를 명확히 만드는 것이다.

### Read First

- `constitution/brand.md`
- `constitution/surface.md`
- `docs/architecture.md`
- `docs/brand-home/README.md`

### Ownership

이 워커는 아래 파일만 수정 가능:

- `docs/brand-home/app/page.js`
- `docs/brand-home/components/home-shell.js`
- `docs/brand-home/README.md`

### Do

- 현재 홈을 **landing / why / runtime proof / contact CTA** 구조로 재정렬한다.
- 카피는 추상적 브랜딩보다 **Layer OS가 무엇을 하는지**가 드러나게 바꾼다.
- `constitution/brand.md`의 어조를 따르되, 과장/마케팅 문구는 줄인다.
- `runtime proof` 블록이 들어갈 위치와 메시지 뼈대를 만든다.
- CTA는 이후 실제 contact route와 연결될 수 있게 구조만 잡는다.

### Do Not

- CSS 전체 리라이트 금지
- API route 추가 금지
- auth / middleware 수정 금지
- `.next`, `node_modules` 수정 금지

### Acceptance

- 홈이 최소 4개 의미 블록으로 재구성되어 있어야 함
- hero-only empty shell 상태를 벗어나야 함
- copy가 “브랜드 분위기”보다 “무엇을 하는 시스템인가”를 더 명확히 설명해야 함
- `npm`/빌드 관련 새 의존성 추가 금지

### Suggested Validation

- `cd docs/brand-home && npm run lint` 가 있으면 실행
- 없으면 최소한 소스 import/JS 문법 오류 없는지 확인

---

## Agent 2 — Front Proof / Runtime Surface Block

### Mission

브랜드 홈에서 **실제 시스템 존재감**을 보여주는 proof block과 contact/public surface를 정리한다.
목표는 “보기 좋은 페이지”가 아니라 **runtime-backed credibility surface**를 만드는 것이다.

### Read First

- `constitution/brand.md`
- `constitution/surface.md`
- `docs/security.md`
- `docs/brand-home/README.md`

### Ownership

이 워커는 아래 파일만 수정 가능:

- `docs/brand-home/app/globals.css`
- `docs/brand-home/app/layout.js`
- `docs/brand-home/lib/http/request.js`
- `docs/brand-home/lib/validation/public-contact.js`

필요 시 아래 파일도 가능하나 범위를 넘기지 말 것:

- `docs/brand-home/app/api/public/contact/route.js`

### Do

- proof block에 필요한 fetch helper / validation surface를 정리한다.
- public contact surface가 현재 보안 원칙을 해치지 않도록 최소 계약을 정리한다.
- frontend가 future runtime data를 붙일 수 있게 request util을 정리한다.
- CSS는 Agent 1의 메시지 구조를 받쳐주는 수준으로만 정리한다.

### Do Not

- protected admin/auth lane 수정 금지
- middleware / RBAC 리디자인 금지
- Layer OS daemon API 계약 변경 금지

### Acceptance

- public surface validation이 명확해야 함
- proof block을 future data fetch와 연결할 seam이 있어야 함
- contact/public path가 security baseline을 깨지 않아야 함

### Suggested Validation

- `cd docs/brand-home && npm run lint` 또는 equivalent
- route/schema 관련 최소 local check

---

## Agent 3 — Telegram Semantics Profile

### Mission

Telegram을 **단순 메시지 로그가 아니라, 양방향 운영 채널**로 쓰기 위한 분류 프로필을 만든다.
이번 작업은 중앙 ingest 경로를 직접 건드리는 것이 아니라, **Codex가 나중에 통합할 helper/profile + fixtures**를 만드는 것이다.

### Read First

- `docs/architecture.md`
- `docs/operator.md`
- `internal/runtime/evidence_ingest.go`
- `internal/runtime/observation_taxonomy.go`

### Ownership

이 워커는 아래 파일만 수정/생성 가능:

- `internal/runtime/telegram_semantics_profile.go`
- `internal/runtime/telegram_semantics_profile_test.go`
- 필요 시 `docs/examples/telegram_ingest_examples.md`

### Do

- Telegram 상호작용 유형을 구조화한 helper/profile을 만든다.
- 최소 아래 케이스를 first-class로 표현한다:
  - inbound conversation
  - outbound publication
  - feedback/reply
  - youtube link injection
  - bookmark/share
- 각 유형에 대해 추천 `topic`, `content_kind`, `interaction_direction`, `tag/ref` 예시를 fixture로 남긴다.
- later ranking을 방해하지 않도록 어떤 ref는 linkable이고 어떤 ref는 metadata-only인지 설명 가능한 구조로 만든다.

### Do Not

- `internal/runtime/evidence_ingest.go` 직접 수정 금지
- `cmd/layer-osctl/ingest.go` 직접 수정 금지
- Telegram bot lifecycle 전체 재설계 금지

### Acceptance

- Telegram 상호작용 유형별 canonical mapping이 코드/테스트로 보인다
- fixtures만 보고도 Codex가 중앙 ingest에 붙일 수 있어야 한다
- 테스트가 있어야 한다

### Suggested Validation

- `go test ./internal/runtime -run TelegramSemanticsProfile`

---

## Agent 4 — Personal DB / NotebookLM Content Profile

### Mission

개인 DB, 메모, NotebookLM류 자료를 later retrieval/ranking에서 쓰기 좋은 **content profile**로 정리한다.
이번 작업도 중앙 ingest를 직접 건드리지 않고, **profile/helper + fixtures**를 만든다.

### Read First

- `docs/architecture.md`
- `internal/runtime/evidence_ingest.go`
- `internal/runtime/observation_taxonomy.go`

### Ownership

이 워커는 아래 파일만 수정/생성 가능:

- `internal/runtime/personal_content_profile.go`
- `internal/runtime/personal_content_profile_test.go`
- 필요 시 `docs/examples/content_ingest_examples.md`

### Do

- personal_db / notebook_lm 자료 유형을 profile로 정리한다.
- 최소 아래 kind를 안정적으로 표현한다:
  - note
  - summary
  - qa
  - transcript
  - bookmark
  - outline
- `doc-id`, `title`, `author`, `source url`, `tags`가 later retrieval 자산으로 남게 fixture를 만든다.
- personal_db와 notebook_lm가 같은 family에 묶일지, 별도 family를 가질지 명시적인 추천안을 코드/테스트로 남긴다.

### Do Not

- `internal/runtime/evidence_ingest.go` 직접 수정 금지
- `cmd/layer-osctl/ingest.go` 직접 수정 금지
- corpus promotion semantics 변경 금지

### Acceptance

- personal_db / notebook_lm canonical mapping이 코드/테스트로 드러나야 함
- 재수집/재정리에 강한 doc identity 전략이 포함돼야 함
- Codex가 나중에 중앙 ingest에 쉽게 붙일 수 있어야 함

### Suggested Validation

- `go test ./internal/runtime -run PersonalContentProfile`

---

## Agent 5 — Crawler / Content Capture Profile

### Mission

외부 콘텐츠 캡처를 later intelligence의 원재료로 만들기 위한 **crawler/content profile**을 만든다.
목표는 실제 크롤러를 만드는 것이 아니라, 문서 identity / host / kind / dedupe 규칙을 helper/test로 정리하는 것이다.

### Read First

- `docs/architecture.md`
- `docs/security.md`
- `internal/runtime/evidence_ingest.go`
- `internal/runtime/observation_taxonomy.go`

### Ownership

이 워커는 아래 파일만 수정/생성 가능:

- `internal/runtime/crawler_content_profile.go`
- `internal/runtime/crawler_content_profile_test.go`
- 필요 시 `docs/examples/content_ingest_examples.md`

### Do

- crawler/content capture용 canonical content profile을 만든다.
- 최소 아래 kind를 안정적으로 표현한다:
  - article
  - video
  - thread
  - post
  - content_capture
- URL variation이 있어도 same doc으로 묶일 수 있는 dedupe strategy를 helper/test로 제안한다.
- host/doc/author/title metadata가 이후 source-quality scoring에 도움이 되게 정리한다.

### Do Not

- 실제 네트워크 fetch loop 구현 금지
- `internal/runtime/evidence_ingest.go` 직접 수정 금지
- `cmd/layer-osctl/ingest.go` 직접 수정 금지

### Acceptance

- crawler/content canonical mapping이 코드/테스트로 드러나야 함
- URL variation / same-doc 문제에 대한 명시적 전략이 있어야 함
- Codex가 나중에 중앙 ingest에 쉽게 붙일 수 있어야 함

### Suggested Validation

- `go test ./internal/runtime -run CrawlerContentProfile`

---

## Merge Order

워크는 아래 순서로 merge candidate를 만들 것.

1. Agent 3, 4, 5 결과 먼저 검토
2. Telegram / personal / crawler profile helper를 `Codex`가 중앙 ingest seam에 통합
3. Agent 1, 2 프론트 결과는 그 다음 통합
4. 최종 전체 검증은 `Codex`가 수행

## Report Template

각 워커는 아래 형식으로만 결과를 보낸다.

```text
[Summary]
- 무엇을 바꿨는지 3줄 이내

[Files]
- path1
- path2

[Validation]
- command1
- command2

[Risks]
- risk1
- risk2

[Need From Codex]
- 최종 통합 전에 확인할 포인트 1개
```

## Founder Note

이번 배치는 **속도용 멀티스레드**지만, 구조 권한은 계속 Go kernel 쪽에 있다.
즉 워커가 만드는 것은 구현 조각이고, 정본 의미/판단/승격은 여전히 central merge를 거친다.
