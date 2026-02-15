# 97layerOS 보안 감사 보고서 (검증 완료)

**날짜**: 2026-02-15
**감사자**: Claude Code (Technical Director)
**상태**: 🔴 **CRITICAL - 즉시 조치 필요**

---

## 🚨 검증 결과: 보고서 100% 정확

### 1. GitHub Repository: **PUBLIC** ✅ 확인됨
```json
{
  "repo": "97layer/woohwahae-website",
  "private": false  ← PUBLIC 상태
}
```

**의미**: Git 히스토리에 커밋된 모든 토큰이 인터넷에 공개되어 있음.

---

### 2. 하드코딩된 실제 토큰 확인 ✅

#### Telegram Bot Token
```python
# test_bot.py:12
# simple_test_bot.py:13
# execution/five_agent_hub_integrated.py:32
BOT_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"
```

#### .env 파일 (평문 노출)
```env
TELEGRAM_BOT_TOKEN=8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI
GEMINI_API_KEY=AIzaSyCGgHVPjEEI3OI3tSNW3SSHNbZuYpHrH-g
ANTHROPIC_API_KEY=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
```

---

### 3. Git 히스토리 확인 ✅

```bash
origin	https://github.com/97layer/woohwahae-website.git (public)

# 최근 3개 커밋에 토큰 포함된 파일 변경 이력
58043786 structure: Major cleanup + Constitution enforcement (336MB freed)
5f0c5c20 feat: 자가 순환 장기 기억 시스템 구축 완료
a2cfcbdb Initial commit - exclude large files
```

**확인된 하드코딩 파일**: 33개

---

## 🔴 CRITICAL - 즉시 조치 (1시간 이내)

### 조치 1: 모든 API 토큰 즉시 재발급 ⚠️

| 서비스 | 재발급 방법 | 예상 시간 |
|--------|------------|----------|
| **Telegram** | BotFather → `/revoke` → `/newbot` 또는 `/token` | 2분 |
| **Gemini** | Google Cloud Console → API 키 삭제 → 새 키 생성 | 3분 |
| **Anthropic** | Anthropic Console → Revoke → Create new key | 3분 |

**⚠️ 주의**: 재발급 시 기존 키는 즉시 무효화됨 → 실행 중인 서비스 중단 → 순차 재발급 필요

### 조치 2: 하드코딩된 토큰 제거 (33개 파일)

**즉시 수정 필요 파일** (우선순위 순):
1. `test_bot.py`
2. `simple_test_bot.py`
3. `execution/five_agent_hub_integrated.py`
4. `execution/five_agent_multimodal.py`
5. `execution/five_agent_async.py`
6. `execution/launchers/WORKING_BOT.py`
7. `execution/ops/clear_webhook.py`
8. `execution/ops/gcp_legacy/*.py` (2개)
9. `execution/ops/bots/*.py` (6개)
10. 나머지 19개 파일

**수정 패턴**:
```python
# ❌ Before
BOT_TOKEN = "8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI"

# ✅ After
import os
from dotenv import load_dotenv
load_dotenv()
BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
```

### 조치 3: Git 히스토리 정리

**옵션 A**: Git Filter-Repo (히스토리 재작성)
```bash
git filter-repo --invert-paths --path test_bot.py --path simple_test_bot.py
git push --force-with-lease
```

**옵션 B**: Repository 재생성 (권장)
```bash
# 1. 새 private repo 생성
# 2. 현재 코드만 깨끗하게 커밋
# 3. 기존 repo 삭제
```

**⚠️ 주의**: Force push는 협업자에게 영향 → 혼자 작업 중이므로 안전

---

## 🟡 HIGH - 24시간 이내 조치

### 조치 4: Code Executor 샌드박싱

**파일**: `execution/code_executor.py`

**현재 위험**:
```python
# 임의 Python 코드를 텔레그램에서 받아 실행
subprocess.run([sys.executable, "-c", user_code], timeout=30)
```

**개선 방안**:
1. **허용 모듈 화이트리스트**
   ```python
   ALLOWED_MODULES = ["math", "datetime", "json", "re"]
   ```

2. **Docker 샌드박스** (권장)
   ```bash
   docker run --rm --network=none --cpus=0.5 --memory=256m python:3.10 python -c "$user_code"
   ```

3. **텔레그램 사용자 ID 검증**
   ```python
   ADMIN_USER_IDS = [123456789]  # .env에서 로드
   if chat_id not in ADMIN_USER_IDS:
       return "권한 없음"
   ```

### 조치 5: Pickle → JSON 전환

**파일**: `execution/autonomous_workflow.py`

**현재 위험**:
```python
# pickle deserialization = 임의 코드 실행 가능
with open(checkpoint_file, 'rb') as f:
    workflow = pickle.load(f)
```

**개선**:
```python
# JSON 직렬화 (이미 백업 로직 있음)
with open(checkpoint_file, 'r') as f:
    workflow = json.load(f)
```

---

## 🟢 MEDIUM - 1주일 이내 조치

### 조치 6: .gitignore에 config.json 추가

```bash
echo "config.json" >> .gitignore
git add .gitignore
git commit -m "security: Add config.json to .gitignore"
```

### 조치 7: .env 파일 퍼미션 강화

```bash
chmod 600 .env  # 현재: 644 (그룹/타인 읽기 가능) → 600 (소유자만)
```

---

## 📊 실행 계획 (우선순위 순)

| 순서 | 작업 | 시간 | 다운타임 |
|-----|------|------|---------|
| 1 | Telegram 토큰 재발급 | 2분 | ✅ 봇 중단 2분 |
| 2 | Gemini API 키 재발급 | 3분 | ✅ AI 응답 중단 3분 |
| 3 | Anthropic API 키 재발급 | 3분 | ✅ Claude 응답 중단 3분 |
| 4 | .env 업데이트 (새 키 입력) | 2분 | - |
| 5 | 33개 파일 하드코딩 제거 | 30분 | - |
| 6 | Cloud Run 재배포 | 5분 | ✅ 웹훅 중단 5분 |
| 7 | VM 동기화 | 3분 | - |
| 8 | Git 커밋 (토큰 제거) | 2분 | - |
| 9 | Git 히스토리 정리 (옵션) | 10분 | - |
| **총합** | - | **60분** | **13분** |

---

## 🎯 즉시 실행 체크리스트

- [ ] **1단계**: Telegram BotFather에서 토큰 재발급
- [ ] **2단계**: Gemini Console에서 API 키 재발급
- [ ] **3단계**: Anthropic Console에서 API 키 재발급
- [ ] **4단계**: .env 파일에 새 키 업데이트
- [ ] **5단계**: 하드코딩 파일 10개 수정 (우선순위)
- [ ] **6단계**: Cloud Run 재배포
- [ ] **7단계**: Git 커밋 및 히스토리 정리
- [ ] **8단계**: Code Executor 샌드박싱
- [ ] **9단계**: Pickle → JSON 전환
- [ ] **10단계**: .gitignore에 config.json 추가

---

## 🔒 보안 강화 완료 후 검증

```bash
# 1. 하드코딩 확인
grep -r "8501568801" . --exclude-dir=.git

# 2. .env 퍼미션 확인
ls -la .env  # -rw------- (600) 확인

# 3. .gitignore 확인
cat .gitignore | grep -E "\.env|config\.json"

# 4. Git status 확인
git status  # .env가 Untracked Files에 없어야 함

# 5. Public repo 확인
curl -s https://api.github.com/repos/97layer/woohwahae-website | jq '.private'
# true 또는 repo 삭제 확인
```

---

## 📞 보안 사고 대응

만약 이미 토큰이 악용된 흔적이 있다면:

1. **Telegram Bot 확인**
   ```bash
   curl "https://api.telegram.org/bot$OLD_TOKEN/getUpdates"
   # 알 수 없는 chat_id 확인
   ```

2. **Gemini API 사용량 확인**
   - Google Cloud Console → APIs & Services → Quotas

3. **Anthropic API 사용량 확인**
   - Anthropic Console → Usage

---

**다음 액션**: 토큰 재발급부터 시작할까요?
