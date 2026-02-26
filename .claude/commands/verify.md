# /verify — LAYER OS 코드 품질 전체 검증

Haiku 서브에이전트로 fresh context에서 실행합니다.

다음 항목을 순서대로 검증하세요:

## 1. Python 구문 검증
```bash
find core/ -name "*.py" -exec python3 -m py_compile {} + 2>&1
```

## 2. f-string 로깅 위반
```bash
grep -rn 'logger\.\(info\|debug\|warning\|error\|critical\)(f["\x27]' core/ --include="*.py"
```
위반 발견 시 파일:라인 목록을 출력하세요.

## 3. 빈 except 블록
```bash
grep -rn '^\s*except:\s*$' core/ --include="*.py"
```

## 4. 하드코딩 비밀
```bash
grep -rn '(api_key|secret|token|password)\s*=\s*["\x27][a-zA-Z0-9_-]\{16,\}["\x27]' core/ --include="*.py"
```

## 5. 금지 파일 존재
프로젝트 루트에 금지 패턴 파일이 없는지 확인:
- SESSION_SUMMARY_*.md
- WAKEUP_REPORT.md
- DEEP_WORK_PROGRESS.md
- DEPLOY_*.md
- NEXT_STEPS.md
- temp_*, untitled_*, 무제*

## 6. Import 순서 (표본 검사)
`core/system/signal_router.py`, `core/agents/brand_scout.py` 에서 import 순서가 stdlib → third-party → local 순서인지 확인.

## 7. QUANTA 상태
`knowledge/agent_hub/INTELLIGENCE_QUANTA.md`의 마지막 수정 시각을 확인하고, 2시간 이상 미갱신이면 경고.

## 출력 형식
```
━━━ LAYER OS Verify Report ━━━
✅ Python 구문: PASS (N files)
⚠️  f-string 위반: N건 (파일 목록)
✅ 빈 except: PASS
✅ 하드코딩 비밀: PASS
✅ 금지 파일: PASS
✅ Import 순서: PASS
✅ QUANTA: N분 전 갱신
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```
