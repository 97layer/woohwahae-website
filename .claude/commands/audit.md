---
description: 폴더 구조 감사 — 중복/orphan/금지 파일 탐지
---

# /audit — Filesystem Audit

구조 위반을 탐지한다.

## 실행 순서

1. **루트 .md 파일 탐지** (CLAUDE.md, README.md 외)
```bash
ls *.md 2>/dev/null | grep -v CLAUDE.md | grep -v README.md
```

2. **금지 파일명 패턴 탐지**
```bash
find . -maxdepth 1 -name "SESSION_SUMMARY_*" -o -name "WAKEUP_REPORT*" -o -name "DEEP_WORK_PROGRESS*" -o -name "DEPLOY_*" -o -name "NEXT_STEPS*" 2>/dev/null
```

3. **빈 디렉토리 탐지**
```bash
find knowledge/ directives/ core/ -type d -empty 2>/dev/null
```

4. **임시 파일명 탐지**
```bash
find . -name "temp_*" -o -name "untitled_*" -o -name "무제*" 2>/dev/null
```

5. **중복 구조 탐지**
```bash
find . -path "*/archive/archive/*" 2>/dev/null
find . -path "*/.agent/*" 2>/dev/null
```

6. **SYSTEM.md §10 Filesystem Placement 기준 검증**
```bash
cat directives/SYSTEM.md | grep -A 30 '§10'
```

## 판정

- 위반 0건 → PASS
- 위반 있으면 → 파일 목록 + 권장 조치 출력
