# Security Rules

globs: core/**/*.py,scripts/**/*.sh,.claude/hooks/**

## 비밀 정보
- 하드코딩 금지: API 키, 토큰, 비밀번호를 코드에 직접 넣지 않음
- 환경변수 또는 .env → os.getenv() 패턴만 허용
- 로그/출력에 비밀 노출 금지

## 입력 검증
- 외부 입력(URL, 파일 경로, 사용자 텍스트)은 반드시 검증 후 사용
- Path traversal (`../`) 방어
- SQL/명령 인젝션 패턴 경계

## 데이터베이스
- WHERE 없는 DELETE/UPDATE 금지
- 사용자 입력을 쿼리에 직접 삽입 금지 (parameterized query 사용)

## 파일 시스템
- 사용자 입력으로 파일 경로 구성 시 정규화(resolve) 필수
- 프로젝트 루트 외부 쓰기 금지
- 임시 파일은 tempfile 모듈 사용
