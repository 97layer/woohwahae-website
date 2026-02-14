# GCP SSH 접속 설정 가이드

**생성일**: 2026-02-14
**서버 IP**: 35.184.30.182

---

## Step 1: SSH 공개키 등록

### 방법 A: GCP Console (웹 UI)

1. [GCP Console](https://console.cloud.google.com) 접속
2. **Compute Engine** → **메타데이터** 클릭
3. **SSH Keys** 탭 선택
4. **수정** 버튼 클릭
5. **항목 추가** 클릭
6. 아래 공개키를 붙여넣기:

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJd0G87SFvzDq4dJmSw8O6Jj0cxx8dPWSRANgoEz0NDp 97layer@gcp-production
```

7. **저장** 클릭

### 방법 B: gcloud CLI (터미널)

```bash
# gcloud CLI가 설치되어 있다면
gcloud compute project-info add-metadata \
  --metadata-from-file ssh-keys=~/.ssh/id_ed25519_gcp.pub
```

---

## Step 2: SSH 접속 테스트

```bash
# 맥북 터미널에서 실행
ssh -i ~/.ssh/id_ed25519_gcp 97layer@35.184.30.182

# 또는 SSH config 설정 후
ssh gcp-97layer
```

---

## Step 3: SSH Config 설정 (선택사항)

`~/.ssh/config` 파일에 추가:

```
Host gcp-97layer
    HostName 35.184.30.182
    User 97layer
    IdentityFile ~/.ssh/id_ed25519_gcp
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

설정 후 간단하게 접속:
```bash
ssh gcp-97layer
```

---

## Troubleshooting

### 문제 1: Permission denied (publickey)
- 공개키가 GCP에 등록되었는지 확인
- 사용자 이름이 `97layer`인지 확인

### 문제 2: Connection timeout
- 방화벽 규칙 확인 (GCP Console → VPC 네트워크 → 방화벽)
- SSH(22번 포트) 허용 필요

### 문제 3: 키 권한 오류
```bash
chmod 600 ~/.ssh/id_ed25519_gcp
```

---

## 다음 단계

SSH 접속이 성공하면:
1. 서버 상태 확인: `ls -la ~/97layerOS`
2. Daemon 상태 확인: `ps aux | grep daemon`
3. 구글 드라이브 동기화 설정 시작
