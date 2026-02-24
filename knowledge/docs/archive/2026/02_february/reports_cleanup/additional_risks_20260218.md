# 97LAYER 추가 보고사항: Hidden Risks & Technical Debt

**작성일**: 2026.02.18
**작성자**: 97LAYER 수석 오케스트레이터

---

## 1. 유령 파이프라인 (Phantom Pipeline)

`core/system/woohwahae_pipeline.py`는 인스타그램을 크롤링하여 웹사이트 콘텐츠로 변환하는 정교한 로직을 담고 있습니다. 그러나 현재 배포 구조(`deploy_light.sh`)상 이 코드는 **"전시용"**에 불과합니다.

- **실행 환경 부재**: `deploy_light.sh`는 Docker 컨테이너를 빌드하지만, 해당 컨테이너(`docker-compose.yml`)가 크롤링을 위한 `Playwright`/`Selenium` 등의 의존성을 포함하고 있는지 불분명합니다.
- **데이터 지속성**: 파이프라인이 생성한 데이터(`data/pipeline/`)가 컨테이너 재시작 시 사라지는 구조입니다. Volume 마운트가 명시적으로 관리되지 않는다면 매번 초기화됩니다.
- **스케줄링**: 스크립트 내부에는 `schedule` 라이브러리가 있으나, 이를 구동하는 진입점(Entrypoint)이나 프로세스 매니저(Supervisor/PM2) 설정이 없습니다.

## 2. 자산 관리의 실종 (Asset Management Void)

`knowledge/system/asset_registry.json`은 2026-02-16 이후 업데이트 기록이 전무합니다 (`"total": 0`).
이는 시스템이 생성하거나 관리하는 이미지, 텍스트 자산이 **추적되지 않고 있음**을 의미합니다. "아카이브"를 지향하는 서비스에서 자산 레지스트리가 비어있다는 것은 치명적입니다.

## 3. 코드베이스 파편화 (Codebase Fragmentation)

- **혼재된 언어**: Shell Script, Python, HTML/Jinja2가 뒤섞여 있으며, 각자의 역할 경계가 모호합니다. (예: 배포 스크립트가 파이썬 캐시를 청소하고, 파이썬 스크립트가 HTML 생성)
- **하드코딩된 설정**: `deploy_light.sh` 내 IP(`136.109.201.201`), 사용자명 등이 하드코딩되어 있습니다. 환경변수(.env) 분리가 시급합니다.

---

**제언 (Action Items)**

1. **파이프라인 폐기 또는 소생**: `woohwahae_pipeline.py`를 실제로 쓸 것이라면 `GitHub Actions`로 이관하여 로컬 의존성을 제거하십시오. 쓰지 않을 것이라면 과감히 `deprecated/`로 격리하십시오.
2. **자산 레지스트리 재구축**: 파일 시스템을 스캔하여 `asset_registry.json`을 현행화하는 스크립트를 작성, 주기적으로 실행해야 합니다.
3. **환경변수 전면 분리**: 모든 IP, 경로, 계정 정보를 `.env`로 추출하여 보안 사고를 예방하십시오.

이상입니다.
