# 97server Communication Protocol (Bidirectional)

이 문서는 사용자(Sovereign)와 97layerOS(System) 간의 양방향 소통 스케줄 및 프로토콜을 정의합니다.

## 1. 정기 보고 (System -> User)

### 🌿 06:00 AM : Daily Essence Report (The Gardener)

- **내용**: 전날 `knowledge/inbox`에 수집된 모든 원시 신호(Raw Signals)를 정제하여 핵심 통찰(Insight)로 요약.
- **목적**: 사용자가 아침에 가장 맑은 정신으로 영감을 섭취.
- **채널**: Telegram

### 📊 09:00 AM : Morning Briefing (Strategy Analyst)

- **내용**: 밤사이 발생한 미용 업계 트렌드, SNS 알고리즘 변동, 시스템 상태 점검 결과.
- **목적**: 비즈니스 의사결정을 위한 데이터 제공.
- **채널**: Telegram (Technical Daemon 주관)

## 2. 수시 소통 (User -> System)

### 📥 Raw Signal Injection (Anytime)

- **방법**: 텔레그램 채팅방에 텍스트, 링크(YouTube, URL), 메모를 전송.
- **처리**:
  - **Text**: `knowledge/inbox`에 저장 후 다음 날 06:00 리포트에 반영.
  - **YouTube/URL**: 즉시 `UIP`가 가동되어 내용을 분석하고 요약 보고.

### ⚡ Command & Control (Anytime)

- **/council [주제]**: 여러 에이전트(CD, TD, SA 등)를 소집하여 다각도로 토론 후 결론 도출.
- **/status**: 현재 시스템의 작업 대기열 및 데몬 상태 확인.
- **/evolve**: 강제로 정원사(Gardener)를 깨워 현재까지의 인사이트를 정리.

## 3. 작업 완료 알림 (System -> User)

- **Trigger**: 백그라운드에서 수행된 자율 태스크(예: 스킬 생성, 대규모 데이터 분석) 완료 시.
- **내용**: 결과물 요약 및 다운로드/확인 경로 안내.

---
**System Note**: 이 프로토콜은 `telegram_daemon.py` 및 `technical_daemon.py`에 의해 자동 집행됩니다.
