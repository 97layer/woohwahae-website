---
id: skill-001
name: Unified Input Protocol (UIP)
description: 유튜브 링크 등 외부 신호를 분석하여 97layerOS 온톨로지 규격으로 변환 및 지식 자산화하는 범용 스킬
version: 1.0.0
target_agents: ["Gemini_Web", "Cursor_Agent", "The_Curator"]
---

# 📝 Unified Input Protocol (UIP)

## 1. 목적 (Goal)

사용자가 어떤 채널(Gemini Web, Cursor, Mobile 등)을 통해 정보를 입력하더라도, 동일한 논리적 필터를 거쳐 `knowledge/` 계층에 정제된 데이터로 안착시키는 것을 목적으로 한다.

## 2. 외부 신호 처리 프로세스 (Workflow)

### Step 1: 신호 수집 (Capture)

- 유튜브 링크가 감지되면 즉시 해당 영상의 트랜스크립트(Transcript) 또는 메타데이터를 추출한다.
- **핵심 필터**: '우화해(WOOHWAHAE)'의 브랜드 철학(미니멀리즘, 본질적 간결함, 슬로우 라이프)과 연결되는 지점만 선별한다.

### Step 2: 온톨로지 변환 (Ontology Transformation)

추출된 인사이트를 아래 규격에 맞춰 구조화한다.

- **ID**: `rs-` (Raw Signal) 접두어를 사용한 고유 번호 부여 (예: rs-005)
- **Context**: 해당 정보가 시스템 내에서 갖는 전략적 위치 정의
- **Relationship**: 기존 `minimal_life_*` 시리즈나 `knowledge/` 내 문서와의 사슬 관계(Link) 명시
- **Tags**: #Essentialism, #Minimalism, #97layer_Logic 중 해당 키워드 포함

### Step 3: 지식 안착 (Commit)

- 생성된 마크다운 파일을 `knowledge/raw_signals/` 경로에 저장한다.
- **파일 명명 규칙**: `rs-[ID]_[핵심키워드].md`

## 3. 환경별 실행 지침 (Execution by Environment)

### A. 제미나이 웹 (Cloud)

- 유튜브 분석 도구를 활용하여 고밀도 인사이트를 즉각 생성한다.
- 생성된 내용을 사용자가 복사할 필요 없이, 구글 드라이브 도구를 사용하여 지정된 경로에 직접 파일을 생성한다.

### B. 안티그래비티 에이전트 (Local)

- 사용자가 링크를 제시하면, 이미 클라우드에서 처리되었는지 `status.json` 혹은 드라이브를 먼저 스캔한다.
- 미처리 상태라면 로컬의 유튜브 파싱 스크립트를 가동하거나 제미나이 웹에 분석을 요청하도록 유도한다.

## 4. 제약 사항 (Constraints)

- 단순 요약은 지양한다. 반드시 97layerOS의 '전략적 자산'으로서의 가치를 중심으로 연산한다.
- 감정적 수식어나 가식적인 공감 표현은 배제하고 고지능적인 논리 전개만 허용한다.
- 관련 없는 시스템 파일(venv 등) 경로는 절대 건드리지 않는다.
