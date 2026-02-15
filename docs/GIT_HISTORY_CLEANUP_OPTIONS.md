# Git 히스토리 정리 옵션 비교

**날짜**: 2026-02-15
**목적**: 노출된 토큰을 Git 히스토리에서 제거

---

## 🎯 목표

GitHub Public Repo에서 이미 노출된 토큰을 히스토리에서 완전히 제거하기.

**노출된 파일들**:
- `test_bot.py`
- `simple_test_bot.py`
- `execution/five_agent_hub_integrated.py`
- `execution/five_agent_multimodal.py`
- `execution/five_agent_async.py`
- 그 외 10여개 파일

---

## ⚠️ 중요 고려사항

### 문제: GitHub에 이미 Push됨
```bash
# 확인
git log --oneline --remotes
# 5f0c5c20 (origin/main) feat: 자가 순환 장기 기억 시스템 구축 완료
# 84a7321c feat: 텔레그램 통신 복구 및 5-Agent Hub 통합
```

**의미**:
1. 토큰이 이미 인터넷에 노출됨
2. 누군가 이미 clone했을 수 있음
3. GitHub Actions, 검색 엔진에 캐시됨

### 결론: 히스토리 정리해도 토큰 재발급은 필수!

---

## 📊 옵션 비교

| 옵션 | 장점 | 단점 | 소요시간 |
|-----|------|------|---------|
| **A. 히스토리 재작성** | 파일만 제거 | 복잡, Force Push 필요 | 10분 |
| **B. Repo 재생성** | 깨끗한 시작 | 모든 이력 손실 | 5분 |
| **C. 아무것도 안함** | 간단 | 히스토리에 토큰 남음 | 0분 |

**권장**: **옵션 B (Repo 재생성)**
- 혼자 작업 중 (협업자 없음)
- 히스토리보다 보안이 중요
- 이미 토큰 재발급 예정

---

## 🔧 옵션 A: Git 히스토리 재작성 (BFG Repo-Cleaner)

### 준비
```bash
# BFG 설치
brew install bfg
```

### 실행 방법

#### 1. 백업
```bash
cd /Users/97layer/97layerOS
git branch backup-$(date +%Y%m%d)
```

#### 2. 삭제할 파일 목록 작성
```bash
cat > /tmp/files_to_remove.txt << 'EOF'
test_bot.py
simple_test_bot.py
execution/five_agent_hub_integrated.py
execution/five_agent_multimodal.py
execution/five_agent_async.py
execution/launchers/WORKING_BOT.py
execution/ops/clear_webhook.py
execution/ops/bots/*.py
.archive/legacy_scripts/send_test_message.py
EOF
```

#### 3. BFG로 파일 제거
```bash
cd /Users/97layer/97layerOS

# 전체 히스토리에서 파일 제거
bfg --delete-files test_bot.py
bfg --delete-files simple_test_bot.py
bfg --delete-files five_agent_hub_integrated.py
bfg --delete-files five_agent_multimodal.py
# ... (각 파일마다 반복)

# Git GC 실행
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

#### 4. Force Push
```bash
git push --force-with-lease
```

### ⚠️ 문제점
1. **현재 커밋(c6a1f987)에서 파일이 다시 추가됨**
   - 히스토리에서 제거했지만 최신 커밋에 존재
   - 모순 발생

2. **GitHub 캐시**
   - GitHub UI에서 이전 커밋 여전히 접근 가능
   - 완전히 사라지려면 시간 필요

3. **복잡성**
   - 10개 이상 파일 개별 처리
   - 실수 가능성 높음

---

## 🚀 옵션 B: Repository 재생성 (권장)

### 장점
- 깨끗한 Git 히스토리
- 토큰이 히스토리에 전혀 없음
- 간단하고 확실함

### 단점
- 기존 이력 손실 (하지만 중요한가?)
- Remote 변경 필요

### 실행 방법

#### 1. 현재 repo 백업
```bash
cd /Users/97layer
cp -r 97layerOS 97layerOS.backup
```

#### 2. Git 이력 삭제 및 재시작
```bash
cd /Users/97layer/97layerOS

# 기존 .git 제거
rm -rf .git

# 새로 초기화
git init
git add .
git commit -m "feat: 97layerOS 보안 강화 완료

## 변경사항
- 모든 하드코딩 토큰 제거
- 환경변수 기반 인증
- Hybrid Zero-Cost 인프라 구축
- 텔레그램 대화 플로우 개선

## 보안
- API 토큰 완전 제거
- .gitignore 보호 강화
- 토큰 재발급 완료

## 아키텍처
- 맥북 (전투기) + GCP VM (정찰기) + Cloud Run (레이더)
- Night Guard 24/7 감시
- Handshake 프로토콜

## 비용
- $0/월 (Google Free Tier)

Initial commit with clean security."
```

#### 3. GitHub에 새 Private Repo 생성
```bash
# GitHub CLI 사용
gh repo create 97layerOS --private --source=. --remote=origin --push
```

**또는 수동**:
1. https://github.com/new 접속
2. Repository name: `97layerOS`
3. **Private** 선택 ✅ (중요!)
4. Create repository

```bash
git remote add origin https://github.com/97layer/97layerOS.git
git branch -M main
git push -u origin main
```

#### 4. 기존 Public Repo 삭제
1. https://github.com/97layer/woohwahae-website 접속
2. Settings → Danger Zone
3. Delete this repository
4. Repo 이름 입력하여 확인

---

## 💡 옵션 C: 아무것도 안함

### 언제 선택?
- 이미 토큰 재발급 완료
- GitHub를 당장 Private로 전환 예정

### 실행 방법
```bash
# GitHub에서 Public → Private 전환
# Settings → Danger Zone → Change visibility → Make private
```

**단점**:
- 이미 인터넷에 노출된 토큰은 회수 불가
- 검색 엔진 캐시에 남아있을 수 있음

---

## 🎯 최종 권장사항

### 상황: 혼자 작업 중, 협업자 없음

**추천**: **옵션 B (Repo 재생성)** + **Private로 전환**

**이유**:
1. ✅ **간단함**: 5분이면 완료
2. ✅ **확실함**: 히스토리에 토큰 완전 제거
3. ✅ **깨끗함**: 새로운 시작
4. ✅ **보안**: Private repo로 전환
5. ❌ 이력 손실: 중요하지 않음 (최신 코드가 중요)

---

## 📋 실행 체크리스트

### 옵션 B 선택 시

- [ ] 1. 현재 디렉토리 백업
- [ ] 2. .git 제거 및 재초기화
- [ ] 3. 깨끗한 첫 커밋
- [ ] 4. GitHub에 새 Private Repo 생성
- [ ] 5. Push
- [ ] 6. 기존 Public Repo 삭제
- [ ] 7. 토큰 재발급 (필수!)
- [ ] 8. .env 업데이트
- [ ] 9. Cloud Run + VM 업데이트
- [ ] 10. 검증

**예상 소요 시간**: 15분 (토큰 재발급 포함)

---

## 🚨 주의사항

### 반드시 기억할 것

1. **히스토리 정리 ≠ 보안 완료**
   - 이미 노출된 토큰은 회수 불가
   - 반드시 토큰 재발급 필요

2. **Force Push 위험**
   - 협업 중이면 팀원과 협의
   - 혼자 작업 중이면 안전

3. **GitHub 캐시**
   - 삭제 후에도 일시적으로 접근 가능
   - 시간이 지나면 사라짐

---

## 🤖 자동화 스크립트

옵션 B를 자동화한 스크립트:

```bash
#!/bin/bash
# Git 히스토리 재생성 스크립트

cd /Users/97layer/97layerOS

# 백업
echo "1️⃣ 백업 중..."
cd ..
cp -r 97layerOS 97layerOS.backup.$(date +%Y%m%d_%H%M%S)
cd 97layerOS

# Git 재초기화
echo "2️⃣ Git 재초기화..."
rm -rf .git
git init
git add .

# 커밋
echo "3️⃣ 첫 커밋 생성..."
git commit -m "feat: 97layerOS 보안 강화 완료

Initial commit with clean security.
All hardcoded tokens removed.
Hybrid Zero-Cost infrastructure deployed."

# GitHub Private Repo 생성
echo "4️⃣ GitHub Private Repo 생성..."
gh repo create 97layerOS --private --source=. --remote=origin --push

echo ""
echo "✅ 완료!"
echo ""
echo "다음 단계:"
echo "1. 기존 Public Repo 삭제: https://github.com/97layer/woohwahae-website"
echo "2. 토큰 재발급 실행: ./execution/system/update_tokens.sh"
```

---

**다음 액션**: 어떤 옵션을 선택하시겠습니까?
