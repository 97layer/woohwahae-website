---
type: documentation
status: active
created: 2026-02-12
last_updated: 2026-02-12
---

# Context7 MCP 서버 설정 완료

## 개요

Context7 MCP 서버가 Claude Code에 성공적으로 연동되었습니다. Context7은 최신 라이브러리 문서와 코드 예제를 실시간으로 가져와 LLM 컨텍스트에 제공하는 도구입니다.

## 설치된 구성 요소

### 1. Context7 MCP 서버
- **패키지**: `@upstash/context7-mcp@2.1.1`
- **위치**: `/Users/97layer/97layerOS/.local_node/node_modules/@upstash/context7-mcp/`
- **실행 파일**: `dist/index.js`

### 2. MCP 설정 파일
- **위치**: `/Users/97layer/97layerOS/.claude/mcp_config.json`
- **설정 내용**:
```json
{
  "mcpServers": {
    "context7": {
      "command": "node",
      "args": [
        "/Users/97layer/97layerOS/.local_node/node_modules/@upstash/context7-mcp/dist/index.js"
      ],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

## 사용 방법

### Claude Code에서 Context7 사용하기

프롬프트 끝에 `use context7`를 추가하면 최신 문서를 자동으로 가져옵니다:

```txt
Create a Next.js middleware that checks for a valid JWT in cookies
and redirects unauthenticated users to `/login`. use context7
```

```txt
Configure a Cloudflare Worker script to cache
JSON API responses for five minutes. use context7
```

### 자동 호출 설정 (선택사항)

매번 `use context7`을 입력하지 않으려면, Claude Code의 규칙(Rule)을 설정하여 자동으로 호출되도록 할 수 있습니다.

## API 키 설정 (선택사항)

Context7은 API 키 없이도 작동하지만, 속도 제한을 완화하려면 API 키를 설정할 수 있습니다:

1. `.env` 파일에 추가:
```bash
CONTEXT7_API_KEY=your_api_key_here
```

2. MCP 설정 파일 업데이트:
```json
{
  "mcpServers": {
    "context7": {
      "command": "node",
      "args": [
        "/Users/97layer/97layerOS/.local_node/node_modules/@upstash/context7-mcp/dist/index.js"
      ],
      "env": {
        "NODE_ENV": "production",
        "CONTEXT7_API_KEY": "${CONTEXT7_API_KEY}"
      }
    }
  }
}
```

## Context7이 해결하는 문제

### ❌ Context7 없이:
- 오래된 코드 예제 (1년 전 학습 데이터 기반)
- 존재하지 않는 API 환각(hallucination)
- 구버전 패키지에 대한 일반적인 답변

### ✅ Context7 사용:
- 최신 버전별 문서 자동 검색
- 실제 작동하는 코드 예제
- 소스에서 직접 가져온 정확한 API 정보

## 지원하는 라이브러리

Context7은 다음과 같은 주요 라이브러리를 지원합니다:
- Next.js
- React
- Vue
- Cloudflare Workers
- AWS SDK
- 기타 수백 개의 인기 라이브러리

전체 목록: https://context7.com/docs/adding-libraries

## 문제 해결

### MCP 서버가 시작되지 않는 경우

1. Node.js 버전 확인 (v18 이상 필요):
```bash
node --version
```

2. Context7 패키지 재설치:
```bash
cd /Users/97layer/97layerOS/.local_node
npm install @upstash/context7-mcp@latest
```

3. Claude Code 재시작

### 문서를 가져오지 못하는 경우

- 네트워크 연결 확인
- Context7 API 상태 확인: https://status.context7.com
- Claude Code 로그 확인

## 다음 단계

1. **다른 MCP 서버 추가**:
   - `fetch` 서버: 웹 페이지 크롤링
   - `brave-search` 서버: 웹 검색

2. **에이전트 시스템과 통합**:
   - Technical Director 에이전트가 자동으로 Context7 사용
   - 코드 생성 시 최신 문서 참조

3. **자동화 규칙 설정**:
   - 코드 관련 질문 시 자동으로 Context7 활성화

## 참고 자료

- [Context7 공식 문서](https://context7.com/docs)
- [Context7 GitHub](https://github.com/upstash/context7)
- [MCP 프로토콜](https://modelcontextprotocol.io)
- [Upstash 문서](https://upstash.com/docs)

## 업데이트 이력

- **2026-02-12**: 초기 설정 완료
  - Context7 MCP v2.1.1 설치
  - Claude Code MCP 설정 생성
  - 기본 구성으로 작동 확인
