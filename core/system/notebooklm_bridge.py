#!/usr/bin/env python3
"""
NotebookLM Bridge - 97layerOS
notebooklm-py 라이브러리 기반 (HTTP API 직접 호출, 브라우저 불필요)

인증: ~/.notebooklm/storage_state.json (1회 로그인 후 영구 재사용)
      또는 NOTEBOOKLM_AUTH_JSON 환경변수 (GCP/컨테이너 배포용)

Author: 97layerOS Technical Director
Updated: 2026-02-16 (notebooklm-py 마이그레이션)
"""

import asyncio
import json
import os
import logging
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent

# 97layer 전용 노트북 타이틀
NB_SIGNAL_ARCHIVE = "97layerOS: Signal Archive"
NB_ESSAY_ARCHIVE  = "97layerOS: Essay Archive"
NB_BRAND_GUIDE    = "97layerOS: Identity Framework and System Implementation Guide"


def _get_storage_path() -> Optional[Path]:
    """인증 파일 경로 반환. 없으면 None."""
    p = Path.home() / ".notebooklm" / "storage_state.json"
    return p if p.exists() else None


def _write_auth_from_env():
    """
    NOTEBOOKLM_AUTH_JSON 환경변수 → ~/.notebooklm/storage_state.json 기록
    GCP VM / Podman 컨테이너 배포 시 사용
    """
    auth_json = os.getenv("NOTEBOOKLM_AUTH_JSON", "").strip()
    if not auth_json:
        return False
    storage_dir = Path.home() / ".notebooklm"
    storage_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
    storage_path = storage_dir / "storage_state.json"
    storage_path.write_text(auth_json, encoding="utf-8")
    storage_path.chmod(0o600)
    logger.info("✅ NOTEBOOKLM_AUTH_JSON → storage_state.json 기록 완료")
    return True


async def _get_client():
    """
    NotebookLMClient 비동기 컨텍스트 반환.
    환경변수 우선, 없으면 로컬 파일 사용.
    """
    # 환경변수 → 파일로 먼저 쓰기
    if os.getenv("NOTEBOOKLM_AUTH_JSON"):
        _write_auth_from_env()

    from notebooklm import NotebookLMClient
    storage = _get_storage_path()
    if not storage:
        raise RuntimeError(
            "NotebookLM 인증 없음. "
            "Mac: notebooklm login 실행. "
            "GCP/컨테이너: NOTEBOOKLM_AUTH_JSON 환경변수 설정."
        )
    return await NotebookLMClient.from_storage(storage)


# ──────────────────────────────────────────────
# 동기 래퍼 (기존 코드와 호환 — sync 인터페이스)
# ──────────────────────────────────────────────

def _run(coro):
    """비동기 코루틴을 동기로 실행"""
    try:
        loop = asyncio.get_event_loop()
        if loop.is_running():
            import concurrent.futures
            with concurrent.futures.ThreadPoolExecutor() as pool:
                future = pool.submit(asyncio.run, coro)
                return future.result(timeout=120)
        return loop.run_until_complete(coro)
    except RuntimeError:
        return asyncio.run(coro)


class NotebookLMBridge:
    """
    97layerOS NotebookLM 브릿지

    notebooklm-py 기반. 브라우저 없이 HTTP API 직접 호출.
    동기 인터페이스 제공 (기존 에이전트 코드와 호환).
    """

    def __init__(self):
        self.authenticated = False
        self._nb_cache: Dict[str, str] = {}  # title → id 캐시

        # 인증 확인
        if os.getenv("NOTEBOOKLM_AUTH_JSON"):
            _write_auth_from_env()

        if _get_storage_path():
            self.authenticated = True
            logger.info("✅ NotebookLM 인증 확인")
        else:
            logger.warning("⚠️  NotebookLM 미인증 — fallback 모드")

    # ── 노트북 관리 ──────────────────────────────

    def list_notebooks(self) -> List[Dict]:
        """노트북 목록 반환"""
        return _run(self._async_list_notebooks())

    async def _async_list_notebooks(self) -> List[Dict]:
        client = await _get_client()
        async with client:
            nbs = await client.notebooks.list()
            result = []
            for nb in nbs:
                result.append({
                    "id": nb.id,
                    "title": nb.title,
                })
            return result

    def get_or_create_notebook(self, title: str) -> str:
        """타이틀로 노트북 찾기. 없으면 생성. notebook_id 반환."""
        return _run(self._async_get_or_create(title))

    async def _async_get_or_create(self, title: str) -> str:
        if title in self._nb_cache:
            return self._nb_cache[title]

        client = await _get_client()
        async with client:
            nbs = await client.notebooks.list()
            for nb in nbs:
                if nb.title == title:
                    self._nb_cache[title] = nb.id
                    logger.info("📖 기존 노트북 사용: %s (%s...)", title, nb.id[:20])
                    return nb.id

            # 없으면 생성
            nb = await client.notebooks.create(title)
            self._nb_cache[title] = nb.id
            logger.info("✅ 노트북 생성: %s (%s...)", title, nb.id[:20])
            return nb.id

    # ── 소스 추가 ────────────────────────────────

    def add_source_url(self, notebook_id: str, url: str, title: Optional[str] = None) -> str:
        """URL 소스 추가 (YouTube, 웹페이지)"""
        return _run(self._async_add_url(notebook_id, url, title))

    async def _async_add_url(self, notebook_id: str, url: str, title: Optional[str]) -> str:
        client = await _get_client()
        async with client:
            kwargs = {"wait": True}
            if title:
                kwargs["title"] = title
            source = await client.sources.add_url(notebook_id, url, **kwargs)
            source_id = getattr(source, "id", str(source))
            logger.info("✅ URL 소스 추가: %s → %s...", url[:60], source_id[:20])
            return source_id

    def add_source_text(self, notebook_id: str, text: str, title: str) -> str:
        """텍스트 소스 추가"""
        return _run(self._async_add_text(notebook_id, text, title))

    async def _async_add_text(self, notebook_id: str, text: str, title: str) -> str:
        client = await _get_client()
        async with client:
            source = await client.sources.add_text(notebook_id, title, text, wait=True)
            source_id = getattr(source, "id", str(source))
            logger.info("✅ 텍스트 소스 추가: %s → %s...", title, source_id[:20])
            return source_id

    # ── 쿼리 (RAG) ──────────────────────────────

    def query_notebook(self, notebook_id: str, query: str) -> str:
        """노트북 RAG 쿼리"""
        return _run(self._async_query(notebook_id, query))

    async def _async_query(self, notebook_id: str, query: str) -> str:
        client = await _get_client()
        async with client:
            result = await client.chat.ask(notebook_id, query)
            answer = getattr(result, "answer", str(result))
            logger.info("✅ 쿼리 완료 (%s자)", len(answer))
            return answer

    # ── 고수준 워크플로우 ────────────────────────

    def add_signal_to_archive(self, signal_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        신호 → Signal Archive 노트북에 텍스트 소스로 추가
        텔레그램 /analyze 명령 후 호출
        """
        return _run(self._async_add_signal(signal_data))

    async def _async_add_signal(self, signal_data: Dict[str, Any]) -> Dict[str, Any]:
        signal_id = signal_data.get("signal_id", f"signal_{datetime.now().strftime('%Y%m%d_%H%M%S')}")
        content   = signal_data.get("content", "")
        source    = signal_data.get("source", "unknown")
        analysis  = signal_data.get("analysis", {})

        # 노트북 ID 확보
        nb_id = await self._async_get_or_create(NB_SIGNAL_ARCHIVE)

        # 소스 텍스트 구성
        score    = analysis.get("strategic_score", "?")
        category = analysis.get("category", "?")
        summary  = analysis.get("summary", "")
        themes   = ", ".join(analysis.get("themes", []))

        text = f"""# Signal: {signal_id}
날짜: {datetime.now().strftime('%Y-%m-%d %H:%M')}
출처: {source}
SA 점수: {score}
카테고리: {category}
테마: {themes}
요약: {summary}

---

{content}
"""
        title = f"[{score}] {summary[:50] or signal_id}"

        client = await _get_client()
        async with client:
            source_obj = await client.sources.add_text(nb_id, title, text, wait=True)
            source_id = getattr(source_obj, "id", str(source_obj))

        logger.info("📚 Signal Archive 추가: %s", title)
        return {
            "notebook_id": nb_id,
            "source_id": source_id,
            "title": title,
            "signal_id": signal_id,
        }

    def query_brand_guide(self, question: str) -> str:
        """
        브랜드 가이드 노트북 RAG 쿼리
        AD/CE 에이전트가 브랜드 컨텍스트 참조 시 사용
        """
        return _run(self._async_query_brand(question))

    async def _async_query_brand(self, question: str) -> str:
        # 브랜드 가이드 노트북 (기존 것 사용)
        nb_id = await self._async_get_or_create(NB_BRAND_GUIDE)
        client = await _get_client()
        async with client:
            result = await client.chat.ask(nb_id, question)
            return getattr(result, "answer", str(result))

    def query_knowledge_base(self, question: str) -> str:
        """
        브랜드/아이덴티티 컨텍스트 쿼리 (기존 코드 호환용)
        """
        return self.query_brand_guide(question)

    def add_essay_to_archive(self, essay_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        CE 에세이 → Essay Archive 노트북에 텍스트 소스로 추가.
        /draft 승인 후 CE write_corpus_essay 완료 시 호출.
        """
        return _run(self._async_add_essay(essay_data))

    async def _async_add_essay(self, essay_data: Dict[str, Any]) -> Dict[str, Any]:
        essay_title  = essay_data.get("essay_title", "제목 없음")
        theme        = essay_data.get("theme", "unknown")
        archive_essay = essay_data.get("archive_essay", "")
        pull_quote   = essay_data.get("pull_quote", "")
        instagram_caption = essay_data.get("instagram_caption", "")
        issue_num    = essay_data.get("issue_num", "")
        today        = datetime.now().strftime("%Y-%m-%d")

        # Essay Archive 노트북 확보
        nb_id = await self._async_get_or_create(NB_ESSAY_ARCHIVE)

        # 소스 텍스트 구성 — 에세이 전문 + 멀티포맷 메타
        text = f"""# {essay_title}
날짜: {today}
테마: {theme}
이슈: {issue_num}

---

## 풀쿼트
{pull_quote}

---

## 본문
{archive_essay}

---

## 인스타그램 캡션
{instagram_caption if isinstance(instagram_caption, str) else chr(10).join(instagram_caption)}
"""
        source_title = f"[{issue_num}] {essay_title} — {theme}"

        client = await _get_client()
        async with client:
            source_obj = await client.sources.add_text(nb_id, source_title, text, wait=True)
            source_id = getattr(source_obj, "id", str(source_obj))

        logger.info("📚 Essay Archive 추가: %s", source_title)
        return {
            "notebook_id": nb_id,
            "source_id": source_id,
            "title": source_title,
            "essay_title": essay_title,
        }


# ── 싱글턴 / 편의 함수 ────────────────────────

_instance: Optional[NotebookLMBridge] = None


def get_bridge() -> NotebookLMBridge:
    """싱글턴 브릿지 인스턴스"""
    global _instance
    if _instance is None:
        _instance = NotebookLMBridge()
    return _instance


def is_available() -> bool:
    """NotebookLM 인증 여부"""
    try:
        return get_bridge().authenticated
    except Exception:
        return False


# ── CLI 테스트 ────────────────────────────────

def main():
    import argparse
    logging.basicConfig(level=logging.INFO, format="%(message)s")

    parser = argparse.ArgumentParser(description="97layerOS NotebookLM Bridge CLI")
    parser.add_argument("command", choices=["status", "list", "query", "add-signal"])
    parser.add_argument("--query",   help="쿼리 텍스트")
    parser.add_argument("--content", help="신호 내용")
    parser.add_argument("--source",  default="cli-test")
    args = parser.parse_args()

    bridge = get_bridge()

    if args.command == "status":
        print(f"인증: {'✅' if bridge.authenticated else '❌'}")

    elif args.command == "list":
        nbs = bridge.list_notebooks()
        for nb in nbs:
            print(f"  {nb['title'][:50]:50s} | {nb['id'][:20]}...")

    elif args.command == "query":
        if not args.query:
            print("--query 필요")
            return
        answer = bridge.query_brand_guide(args.query)
        print(f"\n답변:\n{answer}")

    elif args.command == "add-signal":
        if not args.content:
            print("--content 필요")
            return
        result = bridge.add_signal_to_archive({
            "signal_id": f"test_{datetime.now().strftime('%H%M%S')}",
            "content": args.content,
            "source": args.source,
            "analysis": {"strategic_score": 75, "category": "test", "summary": "CLI 테스트 신호", "themes": ["테스트"]},
        })
        print(f"✅ 추가 완료: {result}")


if __name__ == "__main__":
    main()
