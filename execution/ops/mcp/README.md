# 97layerOS MCP-GDrive Server

**Container-First MCP Integration for Google Drive Knowledge Base**

## 개요

이 디렉토리는 Model Context Protocol (MCP) 서버를 Podman 컨테이너로 격리하여 실행하는 인프라를 제공합니다. 맥북 네이티브 환경에 Node.js를 설치하지 않고, 모든 MCP 연산을 컨테이너 내부에서 처리합니다.

## 철학: Container-First Protocol

```
로컬 맥북 (관제실)
  ├─ directives/ (철학, 규칙)
  ├─ .env (환경 변수)
  └─ credentials/ (인증 정보)

컨테이너 (실행 환경)
  ├─ MCP Server (Google Drive 연동)
  └─ Node.js 20 + MCP 패키지
```

## 파일 구조

```
execution/ops/mcp/
├─ Containerfile              # MCP-GDrive 컨테이너 이미지 정의
├─ build_mcp_container.sh     # 컨테이너 빌드 스크립트
├─ run_mcp_server.sh          # MCP 서버 실행 스크립트
├─ claude_desktop_config.json # Claude Desktop 설정 예시
└─ README.md                  # 본 파일
```

## 설치 및 사용법

### 1. Google Drive 인증 설정

```bash
# 1) Google Cloud Console에서 Service Account 생성
# 2) JSON 키 파일 다운로드
# 3) credentials 디렉토리에 저장

mkdir -p credentials
mv ~/Downloads/service-account-key.json credentials/gdrive_auth.json
chmod 600 credentials/gdrive_auth.json
```

### 2. 환경 변수 설정

`.env` 파일에 Google Drive 폴더 ID 추가:

```bash
# Google Drive - NotebookLM Knowledge Base
GOOGLE_DRIVE_FOLDER_ID="your_folder_id_here"
```

**폴더 ID 찾기:**
- Google Drive에서 NotebookLM 소스 폴더 열기
- URL에서 마지막 부분 복사: `https://drive.google.com/drive/folders/{FOLDER_ID}`

### 3. 컨테이너 빌드

```bash
cd execution/ops/mcp
./build_mcp_container.sh
```

### 4. MCP 서버 테스트 실행

```bash
./run_mcp_server.sh
```

### 5. Claude Desktop 설정

`~/Library/Application Support/Claude/claude_desktop_config.json` 편집:

```json
{
  "mcpServers": {
    "97layer-mcp-gdrive": {
      "command": "podman",
      "args": [
        "run", "-i", "--rm",
        "--env-file", "/Users/97layer/97layerOS/.env",
        "-v", "/Users/97layer/97layerOS/credentials:/app/credentials:ro",
        "97layer-mcp-gdrive:latest"
      ]
    }
  }
}
```

**중요:** 절대 경로를 사용하세요. 상대 경로는 Claude Desktop에서 작동하지 않습니다.

### 6. Claude Desktop 재시작

설정 파일 변경 후 Claude Desktop을 재시작하면 MCP 서버가 자동으로 연결됩니다.

## 세션 연속성 프로토콜

### Read-First (출근 루틴)

세션 시작 시 Claude가 자동으로 수행:

1. MCP 서버를 통해 Google Drive 검색
2. `INTELLIGENCE_QUANTA.md` 최신 버전 읽기
3. 마지막 'Handoff' 섹션에서 작업 맥락 복구

### Write-Back (퇴근 루틴)

작업 종료 시 수동 또는 자동 실행:

```bash
# handoff.py 실행 (로컬 저장)
python3 execution/system/handoff.py --handoff "Phase 2.4 MCP 통합 완료"

# Google Drive 업로드 (Phase 2.4.2에서 자동화 예정)
# 현재는 수동으로 업로드하거나 rclone 사용
```

## NotebookLM 통합

### 자동 업로드 대상

1. **일일 브리핑/리포트**
   - `knowledge/reports/daily/morning_YYYYMMDD.json`
   - `knowledge/reports/daily/evening_YYYYMMDD.json`
   - `knowledge/reports/daily/weekly_YYYYWWW.json`

2. **세션 핸드오버**
   - `knowledge/agent_hub/INTELLIGENCE_QUANTA.md`

3. **멀티에이전트 분석 결과**
   - `knowledge/content/` 디렉토리의 최종 콘텐츠

### NotebookLM 소스 관리

Google Drive 폴더 구조:

```
NotebookLM Sources (GOOGLE_DRIVE_FOLDER_ID)
├─ intelligence/
│  └─ INTELLIGENCE_QUANTA.md (세션 연속성)
├─ daily_reports/
│  ├─ morning_20260216.json
│  └─ evening_20260216.json
└─ content/
   └─ final_approved_assets.md
```

## 문제 해결

### "Podman not found" 오류

```bash
brew install podman
podman machine init
podman machine start
```

### "Permission denied" 오류

```bash
# credentials 파일 권한 확인
ls -la credentials/gdrive_auth.json

# 필요 시 권한 수정
chmod 600 credentials/gdrive_auth.json
```

### macOS 샌드박스 오류

스크립트에 이미 포함되어 있지만, 수동 실행 시:

```bash
export TMPDIR=/tmp
podman run ...
```

## 다음 단계

- [ ] Phase 2.4.1: MCP-GDrive 컨테이너 구축 ✅ (현재)
- [ ] Phase 2.4.2: Handoff 클라우드 동기화
- [ ] Phase 2.4.3: NotebookLM 자동 업로드
- [ ] Phase 2.4.4: Telegram 비서 체계 강화 (MCP 연동)

## 참고 자료

- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)
- [MCP Google Drive Server](https://github.com/modelcontextprotocol/servers/tree/main/src/gdrive)
- [Podman Documentation](https://docs.podman.io/)

---

**97layerOS - Slow Life Archive System**
*Container-First. Context-Preserved. Cloud-Synced.*
