# 97layerOS 프로젝트 구조 - 강제 규칙 (헌법)

## ⚠️ AI 에이전트 필수 준수 사항

이 문서는 97layerOS의 **구조적 헌법**입니다.
위반 시 시스템 무결성이 손상되며, 파편화가 재발합니다.

### 1. 진입점 (Entry Point)

신규 에이전트는 **반드시** 이 순서로 읽어야 함:

1. **CLAUDE.md** (루트) - 3-Layer 아키텍처
2. **directives/README.md** - 전체 매뉴얼 인덱스 (헌법 시작점)
3. **directives/directive_lifecycle.md** ⭐ **핵심 헌법**
4. **directives/agents/README.md** - 역할 선택
5. 해당 역할 MD 파일 + 필수 directive 읽기

온보딩 자동화:
```bash
python3 execution/onboard_agent.py --role SA
```

### 2. 금지 사항 (❌ 헌법 위반)

1. 루트에 새 폴더 생성 금지
2. `memory`, `Memory`, `dashboard` 등 중복 폴더 금지
3. 임시 파일은 반드시 `.tmp/` 안에만
4. Directive 3회 반복 규칙 무시 금지
5. Knowledge와 Directive 혼동 금지
6. 멀티모달 시스템 파일 수정 금지
7. 하드코딩 경로 사용 금지 (포드맨 호환성)

### 3. 표준 폴더 구조

```
97layerOS/
├── directives/          ← 규범 (Normative) - 어떻게 해야 하는가
│   ├── README.md       ← 신규 에이전트 시작점
│   ├── agents/         ← 역할별 매뉴얼
│   └── *.md            ← 27개 프로토콜
├── knowledge/          ← 기록 (Descriptive) - 무엇이 일어났는가
│   ├── system/         ← 시스템 상태
│   ├── agent_hub/      ← 멀티모달 협업
│   └── sessions/       ← 작업 기록
├── execution/          ← Python 도구들
│   ├── launchers/      ← 런처 스크립트
│   ├── ops/            ← 운영 스크립트
│   └── system/         ← 시스템 유틸리티
├── libs/               ← 공유 라이브러리 (멀티모달 포함)
├── deployment/         ← 배포 스크립트
├── docs/               ← 문서
│   ├── milestones/     ← 완료 보고서
│   └── dashboard/      ← 대시보드
├── skills/             ← 재사용 가능 스킬
└── .tmp/               ← 임시 파일만
    ├── cache/          ← AI 캐시
    └── drive/          ← Drive 동기화
```

### 4. Directive vs Knowledge 구분

| 항목 | Directive (규범) | Knowledge (기록) |
|------|-----------------|-----------------|
| 위치 | `directives/` | `knowledge/` |
| 목적 | 어떻게 해야 하는가 | 무엇이 일어났는가 |
| 안정성 | 높음 (검증 후 변경) | 낮음 (자유 기록) |
| 대상 | 반복 작업, Critical Path | 일회성 작업 |

### 5. 브랜드 헌법 보호 (사령부 지침)

다음 파일은 AI 수정 절대 금지:
- `woohwahae_identity.md` 🔒
- `brand_constitution.md` 🔒
- `97layer_identity.md` 🔒

**이유**: 브랜드 정체성은 인간(97layer)의 영역.
Gardener가 자동 수정 시도 시 차단됨.

### 6. 경로 추상화 (포드맨 호환)

모든 Python 스크립트는 상대 경로 사용:
```python
from pathlib import Path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
```

하드코딩 금지:
❌ `/Users/97layer/97layerOS/`
✅ `PROJECT_ROOT`

### 7. Gardener 시스템

**위치**: `libs/gardener.py`
**역할**: 패턴 감지, Directive 승격, 파편화 방지

실행:
```python
from libs.gardener import Gardener
gardener.run_cycle(days=7)
```

**3회 규칙**: 동일 작업 3회 반복 시 Knowledge→Directive 자동 승격

### 8. 멀티모달 시스템 (절대 수정 금지)

- `libs/async_agent_hub.py` - 병렬 처리 허브
- `libs/claude_engine.py` - Claude Opus CD
- `libs/gemini_engine.py` - Gemini Flash SA/AD/CE
- `execution/async_five_agent_multimodal.py` - 5-Agent 시스템
- `execution/async_telegram_daemon.py` - 텔레그램 통합
- `knowledge/agent_hub/synapse_bridge.json` - 협업 상태

**성능**: 2.5x 생산성 (11초 병렬 처리)

### 9. 상태 파일 위치 (표준)

- `knowledge/system_state.json` - 에이전트 실시간 상태
- `knowledge/system/task_status.json` - 작업 진행 (루트에 symlink)
- `knowledge/agent_hub/synapse_bridge.json` - 멀티모달 협업

### 10. 파일 생성 규칙

| 파일 종류 | 위치 |
|----------|------|
| Python 스크립트 | `execution/` |
| 문서/보고서 | `docs/` |
| Directive | `directives/` (3회 규칙 준수) |
| Knowledge | `knowledge/` |
| 설정 파일 | 루트 또는 `config/` |
| 임시 파일 | `.tmp/` |
| 상태/데이터 | `knowledge/` |

### 11. Git 커밋 규칙

```bash
# Directive 변경
git commit -m "directive: [변경 이유]"

# 구조 변경
git commit -m "structure: [변경 내용]"

# 시스템 수정
git commit -m "system: [수정 사항]"
```

### 12. 동기화 핸드셰이크

Drive 업로드 완료 후 다음 작업 진행.
`sync_status.py` 핸드셰이크 로직 준수.

---

**위반 시 시스템 무결성 손상!**

**Last Updated**: 2026-02-15
**Authority**: directives/directive_lifecycle.md
**Protected by**: Gardener System (libs/gardener.py)