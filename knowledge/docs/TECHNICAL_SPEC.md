# 🛠️ TECHNICAL SPECIFICATION - 통합 기술 명세 v1.0

> **통합**: API Reference + Scripts Guide + Infrastructure Setup
> **상태**: Consolidated (SSOT)
> **갱신**: 2026-02-15

---

## 📖 목차

1. [PWA Backend API](#1-pwa-backend-api): FastAPI, WebSocket, Endpoints.
2. [Operations Scripts](#2-operations-scripts): Consolidation, Management, Quality Gate.
3. [Infrastructure & GCP](#3-infrastructure--gcp): GCP Setup, systemd, Deployment.
4. [Architecture Detail](#4-architecture-detail): Data Flow, Components, Testing.

---

## 1. PWA Backend API

FastAPI 기반의 실시간 에이전트 오케스트레이션 및 하이브리드 시스템 모니터링 백엔드입니다.

### 🚀 서버 실행

- **위치**: `execution/api/`
- **명령**: `uvicorn main:app --reload --host 0.0.0.0 --port 8080`

### 📡 주요 엔드포인트

- `GET /api/health`: 시스템 및 컨테이너 상태.
- `GET /api/status`: 하이브리드 노드 및 에이전트 실시간 상태.
- `GET /api/agents`: 활성 에이전트 목록 및 작업 현황.
- `WS /ws`: 실시간 상태 브로드캐스트 (WebSocket).

---

## 2. Operations Scripts

97layerOS 인프라 유지보수 및 자율 운영을 위한 스크립트 모음입니다.

### 🔄 구조 최적화 (Consolidation)

- **스크립트**: `execution/ops/consolidate_structure.py`
- **용도**: 마크다운 중복 제거, 아카이브 정리, 폴더 계층화.

### ✅ 품질 검증 (Quality Gate)

- **스크립트**: `execution/system/quality_gate.py`
- **용도**: 작업 전후 정합성 체크, 린트 검사, 배포 승인.

---

## 3. Infrastructure & GCP

하이브리드 시스템(Macbook + GCP VM) 환경 구축 및 배포 지침입니다.

### ☁️ GCP & Container Setup

- **Containerization**: 모든 에어전트 데몬은 Podman 기반 컨테이너(`97layer-guardian`) 내에서 격리 실행됨.
- **실행 명령**: `podman run -d --name 97layer-guardian -v /Users/97layer/97layerOS:/app 97layer-os-image`
- **GCP 접속**: `ssh username@gcp-instance-ip`
- **서비스 관리**: `systemd`를 이용한 `97layer-master`, `97layer-cycle` 상시 기동.
- **로그 확인**: `sudo journalctl -u 97layer-master -f`

### 📦 배포 프로토콜

- **동기화**: `rsync -avz`를 이용한 로컬-GCP 코드 배포.
- **자동 복구**: `emergency_recovery.py --full`을 통한 즉각적 시스템 롤백.

---

## 4. Architecture Detail

### 🧩 데이터 흐름

`User Input` → `API/Gateway` → `Agent Orchestration` → `Container Execution` → `Knowledge Layer`

### 🧪 테스트 가이드

- **Integration**: `execution/api/tests/integration_test.py` 실행을 통해 전 계층 통신 검증.
- **GCP Reboot Test**: 인스턴스 자가 재부팅 후 `systemctl` 서비스 자동 기동 여부 확인.

---

**Last Updated**: 2026-02-15
**Authority**: [SYSTEM.md](file:///Users/97layer/97layerOS/directives/system/SYSTEM.md)의 기술적 세부 구현체임.
