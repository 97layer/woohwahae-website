# GCP 자동 보고 시스템 설정 가이드

## 📋 현재 상황
- ✅ GCP 서버 실행 중 (35.184.30.182)
- ✅ 텔레그램 봇 실행 중 (새 토큰 필요)
- ✅ 자동 보고서 스크립트 준비됨
- ⚠️ Cron 스케줄 설정 필요

## 🔧 설정 단계

### 1. GCP 봇 토큰 업데이트
```bash
# Mac에서 실행
chmod +x UPDATE_GCP_BOT.sh
./UPDATE_GCP_BOT.sh
```

### 2. 자동 보고서 스크립트 업로드
```bash
# auto_reporter.py를 GCP로 복사
scp -i ~/.ssh/id_ed25519_gcp execution/ops/auto_reporter.py skyto5339@35.184.30.182:~/97layerOS/execution/ops/
```

### 3. GCP에서 Cron 설정
```bash
# GCP 접속
ssh -i ~/.ssh/id_ed25519_gcp skyto5339@35.184.30.182

# crontab 편집
crontab -e

# 다음 라인들 추가:
# 매일 오전 9시 일일 보고서
0 9 * * * cd ~/97layerOS && /home/skyto5339/97layerOS/.venv/bin/python execution/ops/auto_reporter.py daily

# 매시간 상태 체크
0 * * * * cd ~/97layerOS && /home/skyto5339/97layerOS/.venv/bin/python execution/ops/auto_reporter.py hourly

# 매주 월요일 오전 10시 주간 보고서
0 10 * * 1 cd ~/97layerOS && /home/skyto5339/97layerOS/.venv/bin/python execution/ops/auto_reporter.py weekly
```

## 📊 보고서 종류

### 1. **일일 보고서** (매일 오전 9시)
- 시스템 상태
- 24시간 메시지 통계
- 작업 완료 현황
- 에이전트 활동

### 2. **시간별 체크** (매시간)
- 시스템 정상 작동 확인
- 최근 1시간 활동
- 간단한 상태 알림

### 3. **주간 보고서** (월요일 오전 10시)
- 주간 통계 요약
- 주요 완료 작업
- 시스템 성능 지표

## ✅ 맥북 꺼도 작동하는 기능들

**GCP에서 독립 실행:**
1. ✅ **텔레그램 봇** - 24/7 메시지 수신/응답
2. ✅ **Technical Daemon** - 기술 작업 자동 처리
3. ✅ **자동 보고서** - 정기적으로 상태 전송
4. ✅ **Memory Sync** - 대화 기록 자동 백업
5. ✅ **에이전트 시스템** - 자율 작업 수행

**Mac 필요 작업:**
- Council Meeting (고성능 필요)
- Nightly Consolidation
- Draft Approval
- 72시간 규칙 체크

## 🚀 테스트 명령어

### 즉시 보고서 테스트
```bash
# GCP에서 실행
cd ~/97layerOS
source .venv/bin/activate
python execution/ops/auto_reporter.py daily  # 일일 보고서 테스트
python execution/ops/auto_reporter.py hourly # 시간별 체크 테스트
```

## 📱 텔레그램 명령어

봇에게 직접 명령:
- `/status` - 현재 상태
- `/report` - 즉시 보고서
- `/agents` - 에이전트 상태
- `/help` - 도움말