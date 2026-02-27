---
description: LAYER OS 시스템 상태 진단 (QUANTA + work_lock + cache + 서비스 상태)
---

# /doctor — System Health Check

LAYER OS 전체 상태를 진단한다.

## 실행 순서

1. **state.md 상태**
```bash
cat knowledge/agent_hub/state.md | head -40
```

2. **Work Lock 상태**
```bash
cat knowledge/system/work_lock.json
```

3. **Filesystem Cache 상태**
```bash
cat knowledge/system/filesystem_cache.json | python3 -c "import json,sys; d=json.load(sys.stdin); print(f'캐시 항목: {len(d)} 개')" 2>/dev/null || echo "캐시 없음"
```

4. **Handoff 상태**
```bash
python core/system/handoff.py --status 2>/dev/null || echo "handoff.py --status 미지원"
```

5. **VM 서비스 상태** (infrastructure_sentinel 스킬 참조)
```bash
ssh 97layer-vm "systemctl is-active telegram-bot 2>/dev/null; pm2 status 2>/dev/null | head -10" 2>/dev/null || echo "VM 연결 불가"
```

## 진단 결과 판정

- work_lock = `locked: true` → **STOP. 다른 에이전트 작업 중**
- QUANTA 없거나 비어있음 → CRITICAL. 세션 시작 불가
- 캐시 항목 0 → 주의. filesystem_cache 재빌드 고려
