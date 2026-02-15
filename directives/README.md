# 97layerOS Directives - Agent Constitution (Ver 4.0)

> **상태**: 통합 완료 (Zero Redundancy)
> **최종 갱신**: 2026-02-15

## ⚠️ 핵심 지능 (Single Source of Truth)

신규 에이전트는 아래 두 문서를 최우선으로 숙지하며, 파편화된 과거의 지시서가 아닌 본 통합 문서를 기준으로 행동한다.

1. **[IDENTITY.md](IDENTITY.md)** ⭐ **철학적 북극성**
   - 개인(97layer), 브랜드(WOOHWAHAE), 시스템(97layerOS)의 통합 정체성.
   - 72시간 규칙, 비전 로드맵, 금지 사항 명세.

2. **[system/SYSTEM.md](system/SYSTEM.md)** ⭐ **운영 매뉴얼**
   - Sanctuary 4개 기둥 아키텍처.
   - 5-Agent 프레임워크 및 역할 (CD, CE, SA, AD, TD).
   - 운영 프로토콜 (No Plan No Run, Quality Gate).

---

## 📜 규칙 및 프로토콜 (상세)

통합 문서 외의 보조적 규칙들은 아래 파일을 참조한다.

- `directive_lifecycle.md`: 지시서의 생성, 수정, 소멸 프로토콜.
- `system_handshake.md`: 에이전트 간 교대 및 협업 핸드셰이크.
- `visual_identity_guide.md`: 시각적 일관성 유지를 위한 아트 가이드.

---

## 🌱 Gardener (정원사) 시스템

**실행**:

```bash
python3 -c "from system.libs.gardener import Gardener; from system.libs.engines.ai_engine import AIEngine; g = Gardener(AIEngine(), '.'); print(g.run_cycle(7))"
```

## ⚡ 주요 인터페이스

- `execution/interfaces/telegram/bot.py`: 통합 텔레그램 데몬.
- `system/libs/engines/`: 핵심 AI 엔진 및 메모리 매니저.

---

> "소음을 제거하고 본질을 드러내라." — 97layerOS
