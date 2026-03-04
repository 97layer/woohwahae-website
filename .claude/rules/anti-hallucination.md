# 환각 방지 실행 규칙 (Evidence-First)

globs: **
source_of_truth: AGENTS.md

## Core Rule

1. 주장보다 근거가 먼저다.
2. 근거가 없으면 추측하지 않는다.
3. 검증 불가 사실은 `모름/검증 필요`로 처리한다.

## Evidence Priority

1. 로컬 1차 근거: 파일/코드/로그/실행 결과
2. 공식 1차 근거: 공식 문서/스펙/원문 링크
3. 시간 민감 정보: 반드시 최신 검증 후 답변

## Mandatory Response Contract

1. 사실 단정에는 최소 1개 근거를 남긴다.
2. 상대 날짜 표현(오늘/어제/내일)은 절대 날짜를 함께 표기한다.
3. 최신성 관련 질문은 확인 시점을 같이 명시한다.

## Evidence Logging

1. 비자명한 사실 응답/구현 판단 시 `knowledge/system/evidence_log.jsonl`에 근거를 남긴다.
2. JSONL 1행 1증거 원칙: `timestamp`, `claim`, `evidence_type`, `source` 필수.

## Hard Bans

1. "아마도", "대충", "기억상" 기반 단정 금지
2. 근거 없는 수치/버전/상태 보고 금지
3. 출처 없는 인용/요약 금지
