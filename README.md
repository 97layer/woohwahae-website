# LAYER OS — WOOHWAHAE

> 슬로우라이프를 기록하고 실천하는 브랜드 운영 체제.

---

## 구조

```
directives/          뇌 — SAGE_ARCHITECT(인격) + SYSTEM(운영) + THE_ORIGIN(철학)
knowledge/           기억 — 신호, 상태, 리포트
core/                엔진 — 에이전트, 파이프라인, 스킬
website/             얼굴 — woohwahae.kr (Cloudflare Pages)
```

## 문서

| 파일 | 역할 |
|------|------|
| [SAGE_ARCHITECT.md](directives/SAGE_ARCHITECT.md) | 인격 SSOT. 모든 에이전트의 뿌리 |
| [THE_ORIGIN.md](directives/THE_ORIGIN.md) | 브랜드 철학 경전 |
| [SYSTEM.md](directives/SYSTEM.md) | 운영 매뉴얼. 아키텍처 + 배치 + 거버넌스 |

## 실행

```bash
# 빌드
python3 core/scripts/build.py

# 배포 (Cloudflare Pages)
git push origin main
```

---

> "소음이 걷힌 진공에 다다라서야 명징한 본질이 나선다." — THE ORIGIN
