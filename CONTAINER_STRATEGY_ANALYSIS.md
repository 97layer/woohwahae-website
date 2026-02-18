# 97layerOS Container 전략 분석 - Podman vs Docker

> **작성일**: 2026-02-18
> **발견**: Podman 설치되어 있으나 Docker로 배포 중 (모순)
> **목적**: 컨테이너 전략 정리 및 권장사항

---

## 🔍 현재 상황 분석

### **발견된 사실**

#### **로컬 Mac (개발 환경)**
```bash
✅ Podman 설치됨: version 5.7.1
❌ Docker 미설치: not found
✅ Podman Machine 자동 시작: launchd 설정됨
✅ Container 이름: 97layer-os
```

#### **GCP VM (프로덕션)**
```bash
✅ Docker 설치됨: version 29.2.1
✅ Podman 설치됨: (둘 다 있음)
📄 배포 스크립트: docker-compose.yml 사용 중
```

#### **문서 vs 실제**

**SYSTEM.md 주장**:
```markdown
- Podman 컨테이너: Python 실행, Telegram Bot, MCP CLI
- Container-First: 모든 Python 실행은 Podman 내부에서
```

**실제 배포**:
```yaml
# docker-compose.yml
version: '3.8'
services:
  cortex-admin: ...
  cortex-engine: ...
  # ↑ Docker Compose 사용 중
```

**판정**: 🚨 **문서와 실제 불일치** (Documentation Drift)

---

## ⚖️ Podman vs Docker 비교

### **기술적 차이**

| 항목 | Podman | Docker | 평가 |
|---|---|---|---|
| **Daemon** | ❌ Daemon-less (Systemd 통합) | ✅ Daemon 필요 | Podman 우위 |
| **Root 권한** | ❌ Rootless 기본 (보안 우수) | ⚠️ Root 필요 (보안 위험) | Podman 우위 |
| **Kubernetes 호환** | ✅ Pod YAML 네이티브 지원 | △ 별도 변환 필요 | Podman 우위 |
| **Docker Hub** | ✅ 호환 (docker.io 접근 가능) | ✅ 네이티브 | 동일 |
| **Compose 지원** | ✅ podman-compose (호환) | ✅ docker-compose (네이티브) | Docker 우위 |
| **성숙도** | △ 최근 버전 (5.x) | ✅ 오래됨 (안정) | Docker 우위 |
| **생태계** | △ 제한적 | ✅ 매우 광범위 | Docker 우위 |
| **macOS 지원** | △ VM 필요 (Podman Machine) | ✅ Docker Desktop | Docker 우위 |

---

## 🎯 왜 Docker를 쓰게 되었나? (추정)

### **1. docker-compose.yml 기존 작성됨**

**파일 히스토리**:
```yaml
# docker-compose.yml 존재
# → 누군가 Docker 기반으로 설계
# → 배포 시 Docker 사용
```

### **2. GCP VM에 Docker 기본 설치**

**GCP Compute Engine**:
- Docker 이미지에서 VM 생성 가능
- Docker가 기본 설치된 이미지 사용 추정
- Podman은 나중에 추가 설치한 듯

### **3. Cloudflare Tunnel 이미지가 Docker Hub 기준**

```yaml
cortex-tunnel:
  image: cloudflare/cloudflared:latest
  # ↑ Docker Hub 이미지
```

Podman도 호환되지만 Docker가 더 자연스러움.

### **4. 문서만 "Podman-First"로 작성됨**

**SYSTEM.md 작성 시점**:
- 아키텍처 이상: Podman 사용 계획
- 실제 구현: Docker로 먼저 작동시킴
- 문서 업데이트 안됨 → **Drift 발생**

---

## 🤔 Podman으로 전환해야 하나?

### **전환 시 이점**

#### **✅ 보안 강화**
```bash
# Docker (Root 필요)
sudo docker run ...

# Podman (Rootless)
podman run ...  # sudo 불필요
```

**영향**:
- 컨테이너 탈출 공격 위험 감소
- 호스트 시스템 격리 강화

#### **✅ Systemd 네이티브 통합**

```bash
# Podman은 systemd와 자연스럽게 통합
podman generate systemd --name 97layer-telegram \
  > /etc/systemd/system/97layer-telegram.service

systemctl enable 97layer-telegram
systemctl start 97layer-telegram
```

**현재 문제**:
- Docker Compose로 서비스 관리
- Systemd와 이중 관리 구조

#### **✅ Kubernetes 마이그레이션 용이**

```bash
# Podman은 Pod YAML 직접 사용
podman play kube pod.yaml

# Docker는 변환 필요
kompose convert
```

---

### **전환 시 단점**

#### **❌ 재작업 필요**

**변경 필요 항목**:
1. `docker-compose.yml` → `pod.yaml` 변환
2. `deploy_v2.sh` Docker 명령 → Podman 명령
3. GCP VM systemd 서비스 재구성
4. Cloudflare Tunnel 설정 수정

**소요 시간**: 4-6시간

#### **❌ macOS에서 VM 오버헤드**

```bash
# Podman Machine 필요 (macOS는 Linux 아님)
podman machine start 97layerOS
# ↑ QEMU VM 실행 (메모리/CPU 소모)
```

**Docker Desktop**:
- 동일하게 VM 사용하지만 최적화 더 잘됨

#### **❌ 생태계 차이**

**Docker**:
- Stack Overflow 답변 많음
- 튜토리얼 풍부
- CI/CD 통합 쉬움

**Podman**:
- 문서 적음
- 특정 이미지 호환성 이슈 가능

---

## 📊 권장사항

### **시나리오별 선택**

#### **시나리오 1: 현재 상태 유지 (권장) ✅**

**조건**:
- Docker Compose로 이미 작동 중
- 프로덕션 안정성 최우선
- 배포 경험 축적 필요

**액션**:
1. ✅ **Docker 계속 사용**
2. ✅ **문서 수정** (SYSTEM.md에서 Podman 언급 제거 또는 "계획" 표시)
3. ✅ **Podman은 로컬 개발용으로만 사용** (선택적)

**장점**:
- 추가 작업 없음
- 안정성 유지
- 생태계 활용

**단점**:
- Root 권한 필요 (보안 이슈)
- Systemd 이중 관리

---

#### **시나리오 2: 점진적 Podman 전환 (중기 계획)**

**Phase 1: 로컬 개발만 Podman**
```bash
# Mac에서 개발 시
podman run --rm -it \
  -v $(pwd):/app \
  python:3.11 \
  python /app/core/agents/brand_scout.py
```

**Phase 2: GCP VM에 Podman Compose 도입**
```bash
# docker-compose.yml 유지하되 podman-compose로 실행
pip install podman-compose
podman-compose up -d
```

**Phase 3: Kubernetes 마이그레이션 (미래)**
```bash
# 추후 K8s 클러스터로 이동 시
podman generate kube > deployment.yaml
kubectl apply -f deployment.yaml
```

**소요 기간**: 3-6개월

---

#### **시나리오 3: 즉시 Podman 전환 (비권장)**

**조건**:
- 보안이 최우선 (정부/금융 등)
- Rootless 필수
- Kubernetes 확정

**위험**:
- 프로덕션 다운타임 가능
- 디버깅 시간 증가
- 예상 못한 호환성 이슈

**현 시점 평가**: ❌ **과도한 위험**

---

## 🎯 최종 권장사항

### **단기 (현재~1개월): Docker 유지 ✅**

**이유**:
1. 이미 작동 중 (안정성)
2. docker-compose.yml 기반 구조 완성됨
3. Cloudflare Tunnel 통합 검증됨
4. 학습 곡선 낮음

**액션**:
- ✅ SYSTEM.md 수정 (Podman 언급 → 실제 사용 반영)
- ✅ Docker 기반 배포 문서화
- ✅ Podman은 로컬 테스트용으로만 활용

---

### **중기 (3-6개월): 혼합 전략**

**로컬 (Mac)**:
- Podman: 개발/테스트 (Rootless 장점)
- Docker Desktop: 필요 시만 (호환성)

**프로덕션 (GCP VM)**:
- Docker Compose: 계속 사용
- Podman: 백그라운드 작업용 (선택적)

---

### **장기 (1년+): Kubernetes 고려**

**트래픽 증가 시**:
```
Docker Compose (현재)
  ↓
Podman + Pod YAML (중간 단계)
  ↓
Kubernetes (GKE/EKS) (최종)
```

Podman은 K8s 마이그레이션 시 중간 다리 역할 가능.

---

## 📋 문서 수정 제안

### **SYSTEM.md 수정**

**Before**:
```markdown
**3환경 분리:**
- Podman 컨테이너: Python 실행, Telegram Bot, MCP CLI
```

**After (Option 1: 솔직하게)**:
```markdown
**3환경 분리:**
- Docker 컨테이너 (프로덕션): Python 실행, Telegram Bot, MCP CLI
- Podman (로컬 개발): 선택적 사용 (Rootless 테스트)
```

**After (Option 2: 통합)**:
```markdown
**3환경 분리:**
- 컨테이너 (Docker/Podman): Python 실행, Telegram Bot, MCP CLI
  - 프로덕션: Docker Compose
  - 로컬: Podman (Rootless)
```

---

## 🔧 즉시 적용 가능한 수정

### **1. docker-compose.yml 정리**

**현재**:
```yaml
# 5개 서비스 정의되었으나 일부 미사용
cortex-admin, cortex-dashboard, cortex-engine, cortex-editor, cortex-tunnel
```

**실제 사용 중**:
- cortex-tunnel (Cloudflare)
- 나머지는 직접 실행 중?

**권장**:
```bash
# GCP VM에서 확인
ssh vm "cd ~/97layerOS && docker compose ps"
# 실제 실행 중인 서비스만 남기고 정리
```

---

### **2. launchd Podman 설정 업데이트**

**현재 문제**:
```bash
# com.97layer.podman.plist
podman machine start 97layerOS  # ← Machine 이름
podman start 97layer-os          # ← Container 이름
```

**권장**:
```bash
# 실제 사용하지 않으면 비활성화
launchctl unload ~/Library/LaunchAgents/com.97layer.podman.plist

# 또는 명확한 역할 부여
# "로컬 개발 테스트용 Podman"으로만 사용
```

---

## 🏆 결론

### **질문: "포드맨도 있는데 왜 도커를 쓰는지?"**

### **답변**:

**현재 상황**:
- 로컬: Podman 설치됨 (문서 기준)
- 프로덕션: Docker 사용 중 (실제)
- **문서 ≠ 실제** (Documentation Drift)

**이유 (추정)**:
1. docker-compose.yml 먼저 작성됨
2. GCP VM에 Docker 기본 설치됨
3. 문서만 "Podman-First"로 작성
4. 실제 구현은 Docker로 진행

**권장 전략**:

✅ **Docker 계속 사용** (단기)
- 안정성 우선
- 학습 곡선 낮음
- 생태계 활용

△ **Podman 병행** (중기)
- 로컬 개발: Podman (Rootless)
- 프로덕션: Docker (안정성)

🎯 **Kubernetes 고려** (장기)
- Podman이 K8s 마이그레이션 다리 역할

**즉시 조치**:
1. SYSTEM.md 수정 (Docker 사용 반영)
2. docker-compose.yml 실제 사용 서비스만 정리
3. Podman launchd 역할 명확화 또는 비활성화

---

**Podman이 기술적으로 우수하나, 현재는 Docker가 실용적입니다.**

---

**작성**: 2026-02-18
**작성자**: Container Strategy Analyst
**다음 리뷰**: 2026-05 (3개월 후)
