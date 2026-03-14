# Brand Home

Next.js App Router scaffold for the extracted WOOHWAHAE home surface, now serving as the Layer OS public shell + founder admin surface for a live B2C service.

## Structure

- `app/` — app router entry
- `components/home-shell.js` — public home shell composition
- `components/admin-*.js` — protected admin shell and founder action widgets
- `public/assets/js/bg-field.js` — copied legacy backfield driver
- `public/assets/media/brand/*` — curated brand assets synced for the rebuilt shell
- `content/absorbed-brand-source.js` — repo-local absorbed brand notes and media manifest
- `content/social-style-corpus.js` — channel style profiles composed from imported legacy captions
- `content/social-style-examples.generated.js` — generated caption examples rebuilt from the repo-local social style snapshot
- `middleware.js` — authentication on protected page/API paths
- `lib/auth/*` — token parsing and RBAC revalidation
- `lib/runtime/*` — server-only Layer OS BFF and view-model helpers
- `lib/env/server.js` — server-only secret access
- `lib/validation/*` — Zod DTO schemas
- `app/api/admin/session/login/route.js` — token login
- `app/api/admin/session/bootstrap/route.js` — localhost one-click dev login
- `app/api/admin/runtime/*` — protected admin BFF to Layer OS daemon
- `app/api/public/proof/route.js` — sanitized runtime-backed proof DTO
- `app/api/public/contact/route.js` — public DTO validation example
- `scripts/dev-session-token.mjs` — optional signed-token helper

## 가장 쉬운 로컬 실행

```bash
cd docs/brand-home
npm install
npm run dev
```

기본 포트는 `8081` 이다.

- public: `http://localhost:8081`
- admin login: `http://localhost:8081/admin/login`
- 비밀번호 로그인도 지원한다. 아래 env 중 하나를 넣고 서버를 다시 띄우면 된다.
- 로컬 개발에서는 로그인 페이지의 **"로컬 founder 세션 바로 열기"** 버튼으로 바로 들어갈 수 있다.
- admin overview에는 터미널 없이 바로 작업을 여는 **Quick work** 카드가 있다.
- admin overview에는 Telegram founder alert preview와 send를 여는 **Telegram** 카드도 있다.
  이 카드는 단순 preview뿐 아니라 inbound mode, founder delivery readiness,
  설정 누락 사유까지 같이 보여준다.
  특히 `command only` 상태는 이제 `noop`처럼 보이지 않고, founder delivery가
  막힌 부분 준비 상태로 분명하게 보인다.
- admin overview에는 **Brand publish** 카드도 있다.
  이 카드는 현재 브랜드 소스팩에서 text draft를 seed하고,
  proposal -> work -> approval -> flow까지 한 번에 열어준다.
  즉 지금은 “바로 게시”가 아니라 founder review corridor를 먼저 닫는
  첫 social 준비면이다.
- admin overview에는 **Source intake** 카드도 있다.
  이 카드는 founder가 던지는 링크나 텍스트를 먼저 정규화된 source unit으로
  저장한다. 지금은 크롤러보다 intake inbox를 먼저 두는 방향이고,
  나중에 Telegram에서 `97layer / 우순호 / 우화해 / hold` route를
  고르는 입구 역할을 한다.
- admin overview에는 daemon, VM target, continuity host evidence, release/deploy/rollback 가시화를 위한 **Deploy lane** 섹션도 있다.
- `review-room` 페이지는 이제 최근 6시간 안의 `Current blockers` 와 예전 누적 `Older unresolved`
  를 분리해서 보여준다. 지금 막는 안건과 역사적 실패 누적을 섞지 않기 위한 정리다.

브랜드 자산의 canonical copy는 이제 현재 repo 안에 있다.
아래 스크립트는 기본값으로 현재 자산 묶음을 검증하고, 필요할 때만 다른 로컬 source directory를 받아 재동기화한다.

```bash
./scripts/sync_brand_assets.sh
```

다른 로컬 source에서 다시 가져와야 할 때만 `BRAND_HOME_ASSET_SOURCE_DIR=...` 를 함께 넘긴다.

소셜 말투 source는 이제 repo-local snapshot이 canonical이다.
평소에는 흡수 스크립트가 필요 없고, 현재 snapshot에서 예시 캡션만 다시 생성하면 된다.

```bash
python3 scripts/import_legacy_social_style.py
```

외부 보관본이나 옛 source에서 다시 가져와야 할 때만:

```bash
LEGACY_SOCIAL_STYLE_SOURCE=/absolute/path/to/social-style-source.json \
python3 scripts/absorb_legacy_social_style_source.py
python3 scripts/import_legacy_social_style.py
```

이 스크립트는 기본값으로 `docs/brand-home/content/social-style-source.json` 을 읽어
`docs/brand-home/content/social-style-examples.generated.js` 로 만든다.
현재 Threads/Instagram style profile은 이 generated corpus를 읽어,
브랜드 spine과 별도로 채널 말투만 교체 가능하게 유지한다.

열리지 않으면 먼저 아래를 확인한다.

```bash
cd docs/brand-home
npm install
npm run dev
```

그리고 브라우저는 `http://localhost:8081/admin/login` 으로 연다.

로컬에서 `port already in use`, Telegram 충돌, 느린 dev seat 같은 이상 징후가
보이면 먼저 도구 노이즈부터 정리한다.

```bash
./scripts/trim_local_operator_noise.sh --check
./scripts/trim_local_operator_noise.sh --apply
```

기준은 단순하다.

- `layer-osd`
- 필요할 때만 `8081/admin`
- 빌더 에이전트 세션 1개
- 필요하면 수동 셸 1개

그 외 반복적으로 쌓인 tmux, MCP, Playwright, NotebookLM helper는 제품 자체가
아니라 로컬 운영 노이즈로 본다.

## 런타임 연결

기본값으로 Layer OS daemon은 `http://127.0.0.1:17808` 를 본다.
필요하면 아래 env로 바꾼다.

```bash
LAYER_OS_BASE_URL=http://127.0.0.1:17808
```

Founder write action까지 웹에서 쓰려면 아래 env를 추가한다.

```bash
LAYER_OS_WRITE_TOKEN=...
```

비밀번호 로그인을 쓰려면 아래 둘 중 하나를 함께 설정한다.

```bash
LAYER_OS_ADMIN_PASSWORD=...
# 또는
LAYER_OS_ADMIN_PASSWORD_SHA256=<sha256 hex>
```

## 토큰 로그인(선택)

원하면 기존처럼 토큰을 직접 만들어 붙여넣을 수도 있다.

```bash
cd docs/brand-home
SESSION_HMAC_SECRET=replace-me npm run token:dev -- founder founder,admin
```

이 토큰은 `/admin/login` 에 붙여넣으면 된다.

## VM production 배포

동일 VM continuity host에 founder/admin 웹을 올릴 때는 standalone 배포를 쓴다.

```bash
./scripts/deploy_brand_home_vm.sh --host 97layer-vm
```

이 경로는 `.next/standalone` + `.next/static` + `public` 만 옮기고, VM에서는
`layer-os-web.service` 가 `127.0.0.1:3081` 에서 웹을 띄운다. 자세한 내용은
`docs/linux_vm_web_bootstrap.md` 를 따른다.

배포 후 현재 VM에서 실제로 뭐가 떠 있는지 보려면:

```bash
./scripts/vm_runtime_inventory.sh --host 97layer-vm
```

이 인벤토리는 서비스 상태만 보는 게 아니라 Telegram 운영 준비도도 같이 보여준다.
예를 들어 `TELEGRAM_BOT_TOKEN`, `TELEGRAM_FOUNDER_CHAT_ID`,
`TELEGRAM_OPS_CHAT_ID`, `TELEGRAM_BRAND_CHAT_ID`, legacy `TELEGRAM_CHAT_ID`,
`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GOOGLE_API_KEY`, `inbound_mode`, `founder_delivery`, `ops_delivery`,
`brand_delivery`, `provider_env_path`를 비밀값 노출 없이 확인할 수 있다.

실제 secret 반영도 수동 SSH 편집 없이 할 수 있다.

```bash
./scripts/seed_vm_providers.sh --host 97layer-vm
./scripts/seed_vm_providers.sh --host 97layer-vm --apply
```

로컬 env나 macOS Keychain(`layer-os` account)에 값이 있으면, 그 값을 VM의
provider env로 안전하게 옮기고 daemon을 재시작한다.

VM을 장기 운영석으로 쓰려면 tmux cockpit도 같이 깔 수 있다.

```bash
./scripts/install_vm_tmux.sh --host 97layer-vm
./scripts/vm_tmux_attach.sh --host 97layer-vm
```

이렇게 하면 `layeros` 세션이 부팅 후에도 다시 만들어지고, `ops`, `status`,
`daemon-log`, `web-log` 창 구성이 표준으로 유지된다.

이제 Telegram은 `1 bot + 여러 chat` 기준으로 잡고 있다.

- founder room: `TELEGRAM_FOUNDER_CHAT_ID`
- founder DM: `TELEGRAM_FOUNDER_DM_CHAT_ID`
- ops: `TELEGRAM_OPS_CHAT_ID`
- brand: `TELEGRAM_BRAND_CHAT_ID`
- 어느 방이든 `/whoami` 를 보내면 현재 `chat_id` 를 바로 확인할 수 있다.

예전 `TELEGRAM_CHAT_ID`만 가지고 있어도 seed 스크립트가 founder route로
재사용할 수 있지만, 새 기준은 founder 전용 키다.

현재 live VM 상태는 여기까지 올라와 있다.

`Source intake` 카드는 이제 intake 저장만 하는 자리가 아니라,
Telegram route로 열린 `draft seed`를 바로 보고 `Threads prep`까지 넘길 수
있는 작은 브리지 역할도 한다. 즉 founder는 raw source를 먼저 쌓고,
route를 정한 뒤, seed가 괜찮으면 admin에서 바로 canonical publish 준비 흐름으로
올릴 수 있다.

- `OPENAI_API_KEY`: present
- `ANTHROPIC_API_KEY`: present
- `GOOGLE_API_KEY`: present
- `TELEGRAM_BOT_TOKEN`: present
- `TELEGRAM_FOUNDER_CHAT_ID`: present (founder room chat id)
- `TELEGRAM_FOUNDER_DM_CHAT_ID`: present (founder 1:1 DM chat id, must differ from founder room)
- `TELEGRAM_OPS_CHAT_ID`: present (`-5060692298`)
- `TELEGRAM_BRAND_CHAT_ID`: present (`-5293295684`)
- `inbound_mode`: `assistant`
- `founder_delivery`: `ready`

중요: `TELEGRAM_FOUNDER_CHAT_ID`와 `TELEGRAM_FOUNDER_DM_CHAT_ID`를 같은 값으로 두면
founder alert가 1:1 DM으로 새기 때문에 split이 성립하지 않는다. 두 값은 반드시 달라야 한다.

즉 Telegram은 founder room, founder 1:1 DM, ops, brand 네 표면을 구분해서 붙이고,
founder alert packet의 canonical send와 대화형 비서 경로를 둘 다 바로 쓸 수 있다.

이제 founder/admin 웹 안에서 첫 social 준비 흐름도 열 수 있다.

- `Brand publish` 카드는 `docs/brand-home/content/public-home-source.js`의
  정규화된 브랜드 소스팩을 기준으로 기본 초안을 채운다.
- 브랜드 정체성 spine과 소셜 말투는 의도적으로 분리되어 있다.
  웹/브랜드 문장은 `public-home-source.js`, 채널 말투는
  `social-style-corpus.js`에서 가져온다.
- 현재 Threads/Instagram profile은 repo-local snapshot example corpus를
  같이 싣고 있어서, founder/admin에서 example count와 sample excerpts까지
  보면서 초안 톤을 맞출 수 있다.
- Brand publish 흐름은 이제 어떤 style example을 참고했는지도 draft prep 안에
  같이 남긴다. 즉 초안을 열 때 참고한 example ids가 observation과 Threads
  candidate/receipt까지 이어져서, 나중에 “왜 이 톤이 나왔는지”를 추적할 수 있다.
- submit 시 `/api/layer-os/proposals`, `/api/layer-os/proposals/promote`,
  `/api/layer-os/approval-inbox`, `/api/layer-os/flows/sync`,
  `/api/layer-os/observations`를 순서대로 사용한다.
- 즉 지금 단계의 의미는 “live publish”가 아니라
  “brand draft -> founder approval corridor”를 웹에서 여는 것이다.

Threads 첫 live publish 경로도 이제 같은 흐름 위에 얹혀 있다.

- founder/admin에서 approval이 끝난 Threads 초안을 바로 publish할 수 있다.
- 실제 publish 실행은 web env가 아니라 daemon의
  `/api/layer-os/social/threads` 경로가 담당한다.
- 이 경로는 `THREADS_ACCESS_TOKEN` 하나만 있으면 켜지고,
  결과는 `brand_publish_threads` observation receipt로 남는다.
- 즉 구조는 `absorbed brand snapshot -> normalized pack -> brand publish prep -> approved Threads publish`
  순서다. 외부 donor 없이도 실게시까지 갈 수 있게 만든 셈이다.

현재 live 배포는 끝났고, 남은 건 토큰만이다.

- `97layer-vm` daemon에는 `/api/layer-os/social/threads` 가 이미 올라가 있다.
- admin web에도 Threads publish surface가 이미 보이게 배포됐다.
- 현재 live 상태는 `publish_configured=false` 이고, 이유는
  `THREADS_ACCESS_TOKEN` 이 아직 VM provider env에 없기 때문이다.

이전 레거시 edge/runtime 잔여물을 같이 정리하려면:

```bash
./scripts/vm_trim_legacy_runtime.sh --host 97layer-vm --check
```

`woohwahae.kr` 를 새 브랜드 홈 origin으로 받을 준비를 VM에 먼저 깔아두려면:

```bash
./scripts/install_public_edge_vm.sh --host 97layer-vm
```

실제 HTTPS와 live DNS cutover까지 닫으려면:

```bash
./scripts/issue_public_tls_vm.sh --host 97layer-vm
./scripts/switch_woohwahae_dns.sh --delete-edgecheck
```

현재 production brand domain은 `https://woohwahae.kr` 이고, 같은 배포에서
public home을 낸다. 보호된 운영면은 이제 `https://admin.woohwahae.kr` 로
분리되어 있고, public apex에서 `/admin*` 나 `/api/admin*` 로 들어오면
admin subdomain으로 리다이렉트된다.

운영자 습관도 이제 이쪽으로 맞추면 된다.

- public brand face: `https://woohwahae.kr`
- protected admin login: `https://admin.woohwahae.kr/admin/login`

이 배포는 사이트가 실제로 쓰는 정제된 이미지/텍스트/브랜드 팩은 함께 올리지만,
원본 Figma/Notion/Drive export 같은 raw 브랜드 소스까지 VM에 싣는 경로는 아니다.
