# 97layerOS 구조 명확화 및 중복 제거 제안

**작성일:** 2026-02-15
**분석 범위:** 전체 프로젝트 (194 Python 파일, 334 Markdown 파일)
**목표:** 중복 제거, 명확성 향상, 헌법(PROJECT_STRUCTURE.md) 준수

---

## 요약

97layerOS는 **3-Layer 아키텍처**를 잘 따르고 있지만, 다음 문제가 발견되었습니다:

1. **2개의 Archive 폴더** (`.archive/`, `archive/`) - 용도 불명확
2. **루트에 테스트/설정 스크립트 6개** - 레거시 파일들
3. **282MB의 .tmp.driveupload** - 정리 필요
4. **152MB의 스냅샷 폴더** - 클라우드로 이전 필요
5. **execution/ops/gcp_legacy/** - 레거시 스크립트 8개
6. **execution/ 루트에 37개 Python 파일** - 분류 필요
7. **3개의 동일한 MD 파일** (CLAUDE.md, AGENTS.md, GEMINI.md) - 중복

---

## 문제점 상세 분석

### 1. 중복된 Archive 폴더 ⚠️

**현재 상황:**
```
97layerOS/
├── .archive/           (192KB)
│   ├── legacy_scripts/
│   ├── reports/
│   └── setup_guides/
└── archive/            (24KB)
    ├── dashboard/
    └── tests/
```

**문제:**
- 용도가 불명확하고 중복됨
- `.archive/`는 숨김 폴더, `archive/`는 공개 폴더
- 헌법(PROJECT_STRUCTURE.md)에 archive 언급 없음

**제안:**
```
97layerOS/
└── .archive/           (통합)
    ├── legacy_scripts/
    ├── reports/
    ├── setup_guides/
    ├── dashboard/      ← archive/dashboard 이동
    └── tests/          ← archive/tests 이동
```

**이유:**
- 숨김 폴더(`.archive/`)로 통일하여 일반 작업 시 가시성 감소
- 모든 레거시 항목을 단일 위치에 보관
- `archive/` 폴더 제거로 루트 정리

---

### 2. 루트 레거시 스크립트 ⚠️

**현재 상황:**
```
97layerOS/ (루트)
├── test_bot.py                      (2.7KB) - 테스트용
├── simple_test_bot.py               (3.3KB) - 테스트용
├── start_system.sh                  (941B)  - 레거시 런처
├── quick_start.sh                   (331B)  - 레거시 런처
├── install_systemd_services.sh      (1.9KB) - GCP VM용 (레거시)
└── systemd_install_commands.sh      (1.6KB) - GCP VM용 (레거시)
```

**문제:**
- Git 로그 확인 결과 **최근 30일 내 2회만 커밋** (거의 사용 안 함)
- 헌법 위반: "루트에 새 폴더 생성 금지"는 있지만 스크립트 정리 규칙 없음
- 테스트 스크립트는 `tests/` 폴더에 있어야 함
- 런처 스크립트는 `execution/launchers/`에 있어야 함

**제안:**
```
이동 계획:
test_bot.py              → tests/legacy/test_bot.py
simple_test_bot.py       → tests/legacy/simple_test_bot.py
start_system.sh          → .archive/legacy_scripts/
quick_start.sh           → .archive/legacy_scripts/
install_systemd_services.sh → deployment/legacy/
systemd_install_commands.sh → deployment/legacy/
```

**이유:**
- 루트 정리 (6개 파일 제거)
- 역할별 명확한 위치
- 레거시 스크립트는 archive로 격리
- 테스트 파일은 tests/ 하위로

---

### 3. 임시 파일 폴더 정리 ⚠️

**현재 상황:**
```bash
$ du -sh .tmp*
  0B    .tmp/
  0B    .tmp.drivedownload/
282MB   .tmp.driveupload/
```

**문제:**
- `.tmp.driveupload/`에 **282MB**의 데이터 적재
- `.tmp.drivedownload/`는 비어 있지만 존재
- 헌법은 `.tmp/` 하나만 언급

**제안:**
```
97layerOS/
└── .tmp/               (통합)
    ├── cache/          (AI 캐시)
    ├── token_cache/    (기존)
    ├── ai_cache/       (기존)
    ├── drive/          ← .tmp.driveupload, .tmp.drivedownload 통합
    └── snapshots/      ← 임시 스냅샷 저장
```

**액션:**
1. `.tmp.driveupload/` 내용 정리 (282MB 검토 후 삭제 또는 이동)
2. `.tmp.drivedownload/` 제거 (비어 있음)
3. `.tmp/drive/` 하위로 통합

**이유:**
- 임시 파일은 `.tmp/` 단일 위치
- 282MB는 Git 추적에서 제외되어야 함 (.gitignore 확인 필요)
- 명확한 하위 구조 (`drive/` 서브폴더)

---

### 4. 스냅샷 폴더 클라우드 이전 ⚠️

**현재 상황:**
```bash
$ du -sh 97layerOS_Snapshots/
152MB   97layerOS_Snapshots/
```

**문제:**
- **152MB**가 로컬에 저장됨 (Git 저장소 무거워짐)
- 스냅샷은 Google Drive로 업로드되는데 로컬에 중복 보관

**제안:**
```
97layerOS/
└── 97layerOS_Snapshots/  ← 제거 또는 .gitignore에 추가
```

**액션:**
1. `.gitignore`에 `97layerOS_Snapshots/` 추가 확인
2. 기존 스냅샷은 Google Drive에만 보관
3. 로컬에서 삭제 또는 `.tmp/snapshots/`로 이동

**이유:**
- 스냅샷은 백업용이므로 클라우드에만 보관
- 로컬 저장소 경량화 (152MB 절약)
- snapshot_daemon.py가 자동으로 Drive 업로드 중

---

### 5. GCP Legacy Scripts 격리 ✅

**현재 상황:**
```
execution/ops/gcp_legacy/  (8개 스크립트)
├── UPDATE_GCP_BOT.sh
├── auto_deploy_full_sync.sh
├── deploy_gcp_command.sh
├── deploy_sync_to_gcp.md
├── download_gcp_memory.sh
├── fetch_gcp_data.py
├── gcp_sync_install.sh
└── stop_gcp_bot.py
```

**상태:** ✅ **이미 잘 정리되어 있음**

**이유:**
- `gcp_legacy/` 폴더로 명확하게 분리됨
- 사용하지 않는 GCP 스크립트들이 격리되어 있음
- 추가 조치 불필요

---

### 6. execution/ 루트 정리 제안 ⚠️

**현재 상황:**
```
execution/ (루트에 37개 Python 파일)
├── onboard_agent.py
├── gardener_daily_digest.py
├── instagram_generator.py
├── five_agent_multimodal.py
├── archive_daemon.py
├── async_telegram_daemon.py
├── integrated_orchestrator.py
├── sovereign_judgment.py
├── async_five_agent_multimodal.py
├── auto_publisher.py
├── instagram_publisher.py
├── single_telegram_bot.py
├── five_agent_async.py
├── self_healing_monitor.py
├── progress_analyzer.py
├── telegram_daemon.py
├── integrated_bot.py
├── autonomous_developer.py
├── ontology_transform.py
├── telegram_webhook.py
├── youtube_parser.py
├── snapshot_daemon.py
├── create_snapshot.py
├── switch_to_webhook.py
├── unified_system.py
├── gdocs_reader.py
├── technical_daemon.py
├── gdrive_import.py
├── junction_executor.py
├── ingest_gatekeeper.py
├── cycle_manager.py
├── dashboard_server.py
├── web_parser.py
├── five_agent_hub_integrated.py
├── clipboard_sentinel.py
└── command_parser.py
```

**문제:**
- 37개 파일이 한 폴더에 나열되어 있어 가독성 낮음
- 역할별 분류가 불명확 (데몬, 봇, 파서, 도구 등 혼재)

**제안:**
```
execution/
├── daemons/           ← 백그라운드 데몬 (NEW)
│   ├── snapshot_daemon.py
│   ├── archive_daemon.py
│   ├── technical_daemon.py
│   ├── self_healing_monitor.py
│   └── clipboard_sentinel.py
│
├── bots/              ← 텔레그램 봇 (기존 경로 유지)
│   ├── telegram_daemon.py
│   ├── async_telegram_daemon.py
│   ├── single_telegram_bot.py
│   ├── integrated_bot.py
│   └── telegram_webhook.py
│
├── parsers/           ← 데이터 파서 (NEW)
│   ├── youtube_parser.py
│   ├── web_parser.py
│   ├── command_parser.py
│   └── gdocs_reader.py
│
├── orchestrators/     ← 멀티에이전트 오케스트레이터 (NEW)
│   ├── five_agent_multimodal.py
│   ├── async_five_agent_multimodal.py
│   ├── five_agent_async.py
│   ├── five_agent_hub_integrated.py
│   ├── integrated_orchestrator.py
│   └── unified_system.py
│
├── publishers/        ← 콘텐츠 발행 (NEW)
│   ├── instagram_generator.py
│   ├── instagram_publisher.py
│   ├── auto_publisher.py
│   └── dashboard_server.py
│
├── ops/               ← 운영 스크립트 (기존 유지)
│   ├── gcp_management_server.py
│   ├── mac_realtime_receiver.py
│   └── ...
│
├── system/            ← 시스템 유틸리티 (기존 유지)
│   ├── check_environment.py
│   ├── auto_update.py
│   └── ...
│
├── recovery/          ← 복구 도구 (기존 유지)
│
├── utils/             ← 유틸리티 (기존 유지)
│
└── launchers/         ← 런처 스크립트 (기존 유지)
    ├── onboard_agent.py
    ├── cycle_manager.py
    ├── junction_executor.py
    └── ingest_gatekeeper.py
```

**마이그레이션 필요 파일:**
- `execution/*.py` → 역할별 서브폴더로 이동 (37개 중 약 20개)
- 기존 서브폴더(`ops/`, `system/`, `recovery/`, `utils/`, `launchers/`)는 유지

**이유:**
- 역할별 명확한 분류
- 파일 탐색 용이성 향상
- 새로운 스크립트 추가 시 명확한 위치
- 헌법 준수 (execution/ 하위 구조화)

---

### 7. 중복된 Markdown 파일 (AI 인스트럭션) ⚠️

**현재 상황:**
```
97layerOS/ (루트)
├── CLAUDE.md   (8KB)
├── AGENTS.md   (8KB)
├── GEMINI.md   (8KB)
```

**내용 확인:**
> "This file is mirrored across CLAUDE.md, AGENTS.md, and GEMINI.md so the same instructions load in any AI environment."

**문제:**
- **완전히 동일한 내용** 3개 복제
- 업데이트 시 3곳 모두 수정 필요 (관리 부담)
- DRY(Don't Repeat Yourself) 원칙 위반

**제안:**

**옵션 1: 심볼릭 링크 (추천)**
```bash
# 마스터 파일만 유지
mv CLAUDE.md AGENT_INSTRUCTIONS.md

# 심볼릭 링크 생성
ln -s AGENT_INSTRUCTIONS.md CLAUDE.md
ln -s AGENT_INSTRUCTIONS.md AGENTS.md
ln -s AGENT_INSTRUCTIONS.md GEMINI.md
```

**옵션 2: 자동 동기화 스크립트**
```python
# execution/system/sync_agent_instructions.py
# AGENT_INSTRUCTIONS.md를 읽어서 CLAUDE.md, AGENTS.md, GEMINI.md 자동 생성
```

**옵션 3: 단일 파일로 통합 (비추천)**
```
모든 AI가 AGENT_INSTRUCTIONS.md만 읽도록 강제
→ 하지만 각 AI 환경에서 특정 파일명을 기대할 수 있음
```

**추천:** **옵션 1 (심볼릭 링크)**
- macOS/Linux 환경에서 네이티브 지원
- Git도 심볼릭 링크 추적 가능
- 수정 시 자동으로 3곳 모두 반영
- 원본 파일 1개만 관리

---

### 8. Assets 폴더 크기 문제 ⚠️

**현재 상황:**
```bash
$ du -sh assets/
95MB   assets/
```

**내용:**
```
assets/
└── WOOHWAHAE/
    ├── Google AI Studio/
    ├── Inbox/
    ├── Mood/
    ├── 참고 이미지, 무드 , 브랜드 시그니쳐/
    └── 프로젝트/
```

**문제:**
- **95MB**의 이미지/디자인 파일이 Git 저장소에 포함됨
- Git은 바이너리 파일 관리에 비효율적 (버전별 전체 복사)
- 프로젝트 clone 시 95MB 다운로드 필요

**제안:**

**옵션 1: Git LFS (Large File Storage)**
```bash
# Git LFS 설치 및 초기화
brew install git-lfs
git lfs install

# 이미지 파일을 LFS로 추적
git lfs track "assets/**/*.jpg"
git lfs track "assets/**/*.png"
git lfs track "assets/**/*.jpeg"

# .gitattributes 업데이트
git add .gitattributes
git commit -m "structure: Move assets to Git LFS"
```

**옵션 2: Google Drive로 이전**
```
assets/ → 삭제
Google Drive: "97layerOS Assets" 폴더 생성
README.md에 Drive 링크 추가
```

**옵션 3: .gitignore 추가 (로컬만 보관)**
```
# .gitignore
assets/WOOHWAHAE/
```

**추천:** **옵션 1 (Git LFS)**
- 버전 관리 유지
- Git 저장소 경량화
- 필요한 사람만 LFS로 다운로드

---

## 중복/불필요 파일 종합

### 제거 대상 (총 440MB 절약 예상)

| 항목 | 크기 | 조치 |
|------|------|------|
| `.tmp.driveupload/` | 282MB | 정리 후 `.tmp/drive/`로 통합 |
| `.tmp.drivedownload/` | 0B | 삭제 |
| `97layerOS_Snapshots/` | 152MB | .gitignore 추가 또는 삭제 |
| `archive/` | 24KB | `.archive/`로 통합 후 삭제 |
| 루트 레거시 스크립트 6개 | 10KB | archive 또는 적절한 폴더로 이동 |

---

## 제안된 구조 (최종)

```
97layerOS/
├── .archive/               ← 모든 레거시 통합
│   ├── legacy_scripts/
│   ├── reports/
│   ├── setup_guides/
│   ├── dashboard/
│   └── tests/
│
├── .tmp/                   ← 임시 파일만 (통합)
│   ├── cache/
│   ├── token_cache/
│   ├── ai_cache/
│   ├── drive/              ← .tmp.drive* 통합
│   └── snapshots/
│
├── deployment/             ← 배포 스크립트
│   ├── legacy/             ← systemd 스크립트 이동
│   ├── podman-compose.macbook.yml
│   └── podman-compose.nightguard.yml
│
├── directives/             ← 규범 (변경 없음)
│
├── docs/                   ← 문서
│   ├── milestones/
│   ├── dashboard/
│   └── *.md
│
├── execution/              ← Python 도구들 (재구조화)
│   ├── daemons/            ← NEW: 백그라운드 데몬
│   ├── bots/               ← 텔레그램 봇 (기존)
│   ├── parsers/            ← NEW: 파서들
│   ├── orchestrators/      ← NEW: 멀티에이전트
│   ├── publishers/         ← NEW: 콘텐츠 발행
│   ├── ops/                ← 운영 스크립트 (기존)
│   ├── system/             ← 시스템 유틸리티 (기존)
│   ├── recovery/           ← 복구 도구 (기존)
│   ├── utils/              ← 유틸리티 (기존)
│   └── launchers/          ← 런처 스크립트 (기존)
│
├── knowledge/              ← 기록 (변경 없음)
│
├── libs/                   ← 공유 라이브러리 (변경 없음)
│
├── skills/                 ← 재사용 스킬 (변경 없음)
│
├── tests/                  ← 테스트
│   ├── legacy/             ← NEW: test_bot.py 등 이동
│   └── ...
│
├── website/                ← 웹사이트 (변경 없음)
│
├── worker/                 ← 워커 (변경 없음)
│
├── AGENT_INSTRUCTIONS.md   ← 마스터 파일 (NEW)
├── CLAUDE.md → symlink     ← 심볼릭 링크
├── AGENTS.md → symlink     ← 심볼릭 링크
├── GEMINI.md → symlink     ← 심볼릭 링크
├── PROJECT_STRUCTURE.md    ← 헌법 (업데이트 필요)
├── README.md
└── VISION.md
```

---

## 실행 계획

### Phase 1: 안전한 정리 (즉시 실행 가능)

```bash
# 1. 중복 archive 통합
mv archive/dashboard .archive/
mv archive/tests .archive/
rmdir archive

# 2. 임시 파일 정리
mkdir -p .tmp/drive
# .tmp.driveupload 내용 검토 후 이동 또는 삭제
rm -rf .tmp.drivedownload

# 3. 루트 레거시 스크립트 이동
mkdir -p .archive/legacy_scripts/systemd
mv install_systemd_services.sh .archive/legacy_scripts/systemd/
mv systemd_install_commands.sh .archive/legacy_scripts/systemd/
mv start_system.sh .archive/legacy_scripts/
mv quick_start.sh .archive/legacy_scripts/

mkdir -p tests/legacy
mv test_bot.py tests/legacy/
mv simple_test_bot.py tests/legacy/

# 4. 스냅샷 폴더 .gitignore 추가
echo "97layerOS_Snapshots/" >> .gitignore

# 5. Markdown 통합 (심볼릭 링크)
mv CLAUDE.md AGENT_INSTRUCTIONS.md
ln -s AGENT_INSTRUCTIONS.md CLAUDE.md
ln -s AGENT_INSTRUCTIONS.md AGENTS.md
ln -s AGENT_INSTRUCTIONS.md GEMINI.md

# 6. Git 커밋
git add -A
git commit -m "structure: Consolidate archives, cleanup root, symlink agent instructions"
```

**예상 효과:**
- 루트 정리: 6개 스크립트 제거
- Archive 통합: 단일 위치 관리
- Markdown 중복 제거: 유지보수 부담 감소

---

### Phase 2: execution/ 재구조화 (신중하게 진행)

```bash
# 1. 새 폴더 생성
cd execution
mkdir -p daemons parsers orchestrators publishers

# 2. 데몬 이동
mv snapshot_daemon.py daemons/
mv archive_daemon.py daemons/
mv technical_daemon.py daemons/
mv self_healing_monitor.py daemons/
mv clipboard_sentinel.py daemons/

# 3. 파서 이동
mv youtube_parser.py parsers/
mv web_parser.py parsers/
mv command_parser.py parsers/
mv gdocs_reader.py parsers/

# 4. 오케스트레이터 이동
mv five_agent_multimodal.py orchestrators/
mv async_five_agent_multimodal.py orchestrators/
mv five_agent_async.py orchestrators/
mv five_agent_hub_integrated.py orchestrators/
mv integrated_orchestrator.py orchestrators/
mv unified_system.py orchestrators/

# 5. 발행 도구 이동
mv instagram_generator.py publishers/
mv instagram_publisher.py publishers/
mv auto_publisher.py publishers/
mv dashboard_server.py publishers/

# 6. 봇 이동 (execution/bots/ 이미 존재하므로 루트의 봇 파일만 이동)
mv telegram_daemon.py bots/
mv async_telegram_daemon.py bots/
mv single_telegram_bot.py bots/
mv integrated_bot.py bots/
mv telegram_webhook.py bots/

# 7. 런처 이동
mv onboard_agent.py launchers/
mv cycle_manager.py launchers/
mv junction_executor.py launchers/
mv ingest_gatekeeper.py launchers/

# 8. 나머지 검토 후 분류
# sovereign_judgment.py → orchestrators/
# autonomous_developer.py → launchers/
# ontology_transform.py → utils/
# gdrive_import.py → ops/
# progress_analyzer.py → system/
# create_snapshot.py → daemons/
# switch_to_webhook.py → ops/
# gardener_daily_digest.py → system/

# 9. import 경로 수정 스크립트 실행 필요
# 예: from execution.snapshot_daemon → from execution.daemons.snapshot_daemon

# 10. Git 커밋
git add -A
git commit -m "structure: Reorganize execution/ into role-based subdirectories"
```

**주의사항:**
- **import 경로 업데이트 필수**: `execution/daemons/snapshot_daemon.py`로 이동하면 모든 import문 수정 필요
- **Podman 컨테이너 재시작**: 경로 변경 후 컨테이너 command 수정 필요
- **테스트 필수**: 이동 후 전체 시스템 테스트

---

### Phase 3: Assets Git LFS 적용 (선택 사항)

```bash
# 1. Git LFS 설치
brew install git-lfs
git lfs install

# 2. 이미지 파일 추적
git lfs track "assets/**/*.jpg"
git lfs track "assets/**/*.png"
git lfs track "assets/**/*.jpeg"
git lfs track "assets/**/*.JPG"

# 3. 커밋
git add .gitattributes
git add assets/
git commit -m "structure: Move assets to Git LFS"

# 4. 강제 푸시 (기존 히스토리 재작성)
git lfs migrate import --include="assets/**/*.jpg,assets/**/*.png,assets/**/*.jpeg,assets/**/*.JPG"
```

**효과:**
- Git 저장소 크기 감소 (95MB → ~1MB)
- Clone 속도 향상

---

## 헌법 업데이트 필요

**PROJECT_STRUCTURE.md** 수정 사항:

1. **Section 3 (표준 폴더 구조)** 업데이트:
   ```markdown
   ├── execution/          ← Python 도구들
   │   ├── daemons/        ← 백그라운드 데몬
   │   ├── bots/           ← 텔레그램 봇
   │   ├── parsers/        ← 데이터 파서
   │   ├── orchestrators/  ← 멀티에이전트 오케스트레이터
   │   ├── publishers/     ← 콘텐츠 발행
   │   ├── launchers/      ← 런처 스크립트
   │   ├── ops/            ← 운영 스크립트
   │   ├── system/         ← 시스템 유틸리티
   │   ├── recovery/       ← 복구 도구
   │   └── utils/          ← 유틸리티
   ```

2. **Section 10 (파일 생성 규칙)** 추가:
   ```markdown
   | 데몬 스크립트 | `execution/daemons/` |
   | 봇 스크립트 | `execution/bots/` |
   | 파서 | `execution/parsers/` |
   | 오케스트레이터 | `execution/orchestrators/` |
   ```

3. **Section 2 (금지 사항)** 추가:
   ```markdown
   8. 루트에 스크립트 파일 생성 금지 (execution/ 하위 분류)
   9. archive/, .archive/ 중복 폴더 금지
   ```

---

## 예상 효과

### 정량적 효과

| 항목 | 개선 |
|------|------|
| 루트 파일 개수 | 6개 감소 |
| 중복 폴더 제거 | 1개 (archive/) |
| 임시 파일 정리 | 282MB 절약 |
| 스냅샷 격리 | 152MB 절약 |
| Git 저장소 크기 | ~440MB 감소 (LFS 적용 시) |
| execution/ 가독성 | 37개 → 역할별 5-7개 |

### 정성적 효과

1. **명확성 향상**
   - 파일 위치를 역할로 쉽게 추론 가능
   - 새 스크립트 추가 시 명확한 위치

2. **유지보수성 향상**
   - Markdown 중복 제거로 업데이트 부담 감소
   - 레거시 항목 격리로 현행 코드와 분리

3. **성능 향상**
   - Git clone 속도 향상 (저장소 경량화)
   - 파일 탐색 속도 향상 (구조화)

4. **헌법 준수**
   - PROJECT_STRUCTURE.md와 실제 구조 일치
   - 명확한 규칙으로 파편화 방지

---

## 리스크 및 대응

### Phase 1 (안전)

| 리스크 | 대응 |
|--------|------|
| 심볼릭 링크 미지원 환경 | Windows Git은 심볼릭 링크 지원 (core.symlinks=true) |
| .gitignore 누락 | 커밋 전 `git status` 확인 |

### Phase 2 (주의 필요)

| 리스크 | 대응 |
|--------|------|
| import 경로 오류 | 자동 검색/치환 스크립트 사용 |
| Podman 컨테이너 실패 | 경로 업데이트 후 재시작 테스트 |
| 기존 스크립트 호출 실패 | 전체 grep 검색으로 호출 위치 확인 |

### Phase 3 (선택 사항)

| 리스크 | 대응 |
|--------|------|
| Git LFS 설정 복잡도 | 문서화 및 팀 교육 |
| 히스토리 재작성 | 백업 후 진행 |

---

## 체크리스트

### Phase 1 (안전한 정리)

- [ ] `.archive/` 통합 완료
- [ ] `.tmp/drive/` 통합 완료
- [ ] 루트 레거시 스크립트 이동
- [ ] `97layerOS_Snapshots/` .gitignore 추가
- [ ] Markdown 심볼릭 링크 생성
- [ ] Git 커밋 및 푸시
- [ ] 전체 시스템 테스트 (Podman 컨테이너 정상 작동 확인)

### Phase 2 (execution/ 재구조화)

- [ ] 새 폴더 생성 (daemons, parsers, orchestrators, publishers)
- [ ] 파일 이동 (37개 분류)
- [ ] import 경로 업데이트 스크립트 작성 및 실행
- [ ] Podman compose 파일 경로 업데이트
- [ ] 전체 import 검색 (`grep -r "from execution\."`)
- [ ] 테스트 실행 (`pytest tests/`)
- [ ] Podman 컨테이너 재시작 및 검증
- [ ] Git 커밋 및 푸시

### Phase 3 (Assets Git LFS)

- [ ] Git LFS 설치 및 초기화
- [ ] .gitattributes 설정
- [ ] 마이그레이션 실행
- [ ] 푸시 및 검증

### 문서 업데이트

- [ ] PROJECT_STRUCTURE.md 업데이트
- [ ] README.md 업데이트 (구조 변경 반영)
- [ ] directives/README.md 업데이트 (필요 시)

---

## 결론

97layerOS는 **3-Layer 아키텍처를 잘 구현**하고 있지만, 성장 과정에서 발생한 중복과 레거시 항목 정리가 필요합니다.

**권장 순서:**
1. **Phase 1 즉시 실행** - 안전하고 즉각적인 효과
2. **Phase 2 신중하게 준비** - import 경로 업데이트 자동화 후 진행
3. **Phase 3 선택적 적용** - Assets 관리 전략에 따라 결정

**예상 소요 시간:**
- Phase 1: 30분
- Phase 2: 2-3시간 (import 경로 업데이트 포함)
- Phase 3: 1시간

**최종 효과:**
- 저장소 440MB 경량화
- 파일 탐색 및 유지보수성 대폭 향상
- 헌법(PROJECT_STRUCTURE.md) 완전 준수
- 향후 파편화 방지 기반 확립

---

**작성자:** Claude Code (97layer Technical Director)
**검토 필요:** 97layer (사용자 승인 후 진행)
**우선순위:** Phase 1 > Phase 2 > Phase 3
