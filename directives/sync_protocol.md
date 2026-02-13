# Google Drive Sanitizer Protocol (97LAYER)

## 1. Objective

- 로컬 자산을 구글 드라이브 친화적인 텍스트 기반 형식(.txt)으로 변환하여 동기화 오류를 원천 차단한다.
- 웹 제미나이(Gemini)가 코드 및 설정 파일을 일반 문서로 쉽게 인식하고 분석할 수 있도록 가독성을 상수화(Constant)한다.
- 민감 정보(.env)는 구조만 유지하고 값을 제거하여 보안 사고를 예방한다.

## 2. Transformation Rules

| 원본 형식 | 변환 형식 | 비고 |
| :--- | :--- | :--- |
| `.py` | `[파일명]_py.txt` | 파이썬 실행 파일 -> 텍스트 문서 |
| `.cursorrules` | `cursorrules.txt` | 숨김 설정 -> 명시적 텍스트 |
| `.env` | `env_structure.txt` | **보안 처리**: 키(KEY)만 유지, 값(VALUE) 제거 |
| `.md`, `.txt` | `[파일명]` | 그대로 유지 |
| 기타 코드 | `[파일명]_[확장자].txt` | 필요 시 확장 |

## 3. Storage Architecture

- **Local Source**: `/Users/97layer/97layerOS`
- **Cloud Destination**: `/Users/97layer/내 드라이브/97layerOS_Sync`
- **Exclusion Filters**:
  - `.tmp`
  - `venv`, `venv_old_broken`
  - `.git`
  - `__pycache__`
  - `node_modules`
  - `.DS_Store`

## 4. Execution Logic (Sanitizer)

- 스크립트 실행 시 `sync_manager.py`가 작동하여 변환 및 동기화를 수행한다.
- 파일의 수정 시간(mtime)을 비교하여 변경된 파일만 업데이트한다.
- `.env` 파일은 내용을 파싱하여 `KEY=` 형태로만 저장하고 실제 값은 기록하지 않는다.
