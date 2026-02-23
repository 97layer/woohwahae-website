# Environment Segmentation Specification (v9.0)

> **상태**: 활성 (Active)
> **최종 갱신**: 2026-02-21
> **목적**: 로컬 개발과 차세대 외부 배포 환경의 완전한 분리 및 기록

## 1. Environment Topology

| 환경명 | 역할 | 주소 | 관리 주체 |
| :--- | :--- | :--- | :--- |
| **Local Lab** | 원형 개발 및 에이전트 추론 (Dev Base) | `localhost:8081` | 97layer Orchestrator |
| **Remote Lab** | 고정밀 렌더링 및 대외부용 실험실 (Ext Lab) | `136.109.201.201` | GCP VM (97layer-vm) |

## 2. Server Metadata (Next-Gen)

- **Provider**: Google Cloud Platform (GCP)
- **Internal User**: `skyto5339_gmail_com`
- **Root Path**: `/home/skyto5339_gmail_com/97layerOS/website/website/`
- **Web Server**: Nginx (Port 80)

## 3. Deployment Workflow

1. **Stage (Local)**: `website/lab/` 하위에서 프로토타입 개발 및 로컬 검증.
2. **Transfer (Sync)**: 검증된 소스를 SCP를 통해 외부 서버로 전이.
    - 명령: `scp <file> 97layer-vm:<remote_path>`
3. **Release (External)**: 외부 IP를 통해 최종 UX 및 성능 확인.

## 4. Environment Neutrality Rule

- 외부 서버 소스는 로컬 소스와 물리적으로 격리되며, 로컬의 `.ai_rules`나 `lab/index.html` 등의 핵심 인프라를 직접 수정하기 전 외부 서버에서 먼저 실험적 배포를 수행한다.
- 모든 외부 배포물은 GUI 상단에 환경 정보를 명시하여 오인 배포를 방지한다.
