# Git Workflow Rules

globs: **

## 커밋 형식
- `{type}: {한국어 설명}` (type: feat/fix/refactor/docs/chore/style/test)
- 예시: `feat: 통합 신호 스키마 적용`, `fix: SA 큐 전달 누락 수정`
- 50자 이내 제목, 본문은 선택

## 커밋 전 체크
- state.md 갱신 여부 확인 (코드 변경 시)
- .env, credentials, API 키 포함 파일 커밋 금지
- knowledge/signals/ 내 JSON 데이터 파일은 커밋 제외 (.gitignore)

## 브랜치
- main: 안정 배포 브랜치
- 기능 작업은 현재 main 직접 커밋 (단일 개발자)
- force push 금지

## 배포 연동
- VM 배포 전 로컬 테스트 필수
- 배포 후 서비스 상태 확인
