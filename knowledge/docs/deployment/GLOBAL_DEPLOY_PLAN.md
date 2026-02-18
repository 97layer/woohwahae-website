# [Deployment] Cortex Global System 배포 계획

시스템을 로컬/VM 환경에서 격리하지 않고, 어디서든 접근 가능하며 소스 코드 검토와 인사이트 주입이 실시간으로 이루어지는 **Cortex Global System** 구축 계획입니다.

## User Review Required

> [!IMPORTANT]
>
> - **Cloudflare Tunnel**: 외부에서 SSH 없이 안전하게 접속하기 위해 Cloudflare 무료 계정 및 도메인 연동이 권장됩니다.
> - **code-server**: 브라우저에서 소스 코드를 직접 검토/수정하므로 보안 설정(Password/SSL)이 필수적입니다.
> - **Docker Compose**: 로컬과 동일한 환경을 유지하기 위해 컨테이너 기반으로 서비스를 이관합니다.

## Proposed Changes

### 1. External Access Layer (Cloudflare Tunnel)

- **Cockpit (Port 5001)**: `https://cockpit.97layer.com` (또는 개별 주소)로 연결.
- **Dashboard (Port 8000)**: `https://signal.97layer.com`으로 연결하여 실시간 SSE Push 수신.

### 2. Remote Management (code-server)

- **Source Review**: 브라우저 기반 VS Code를 설치하여 언제 어디서든 소스 코드 라인 단위 검토 및 수정 가능.
- **Git Sync**: 로컬 MacBook과 VM 간의 코드 정합성을 위해 Git 기반 순환 구조 확립.

### 3. Orchestration (Docker Compose)

- 모든 구성 요소를 단일 설정으로 관리:
  - `cortex-admin`: Flask (5001)
  - `cortex-dashboard`: Dashboard (8000)
  - `cortex-engine`: Signal Processor + Telegram Bot
  - `cortex-editor`: code-server (8080)

## Verification Plan

### Automated Tests

- **Endpoint Health Check**: 배포 후 각 도메인의 HTTP 200 응답 확인.
- **SSE Connection Test**: 온라인 상에서 SSE 스트림이 끊기지 않고 유지되는지 확인.

### Manual Verification

- 외부 네트워크(모바일 등)에서 Cockpit 접속 및 인사이트 주입 테스트.
- 브라우저를 통한 소스 코드 수정 및 시스템 반영 여부 확인.
