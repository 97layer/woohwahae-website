#!/usr/bin/env python3
"""
NotebookLM MCP Bridge - 97layerOS Wrapper
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Python wrapper for notebooklm-mcp-cli (nlm command)

28개 도구 중 Anti-Gravity 핵심 8개 래핑:
1. notebook_create/list (Foundation)
2. source_add_url (Source Grounding - YouTube, Web)
3. notebook_query (RAG)
4. audio_create (Multi-modal)

Container-First:
- CLI 호출: subprocess로 nlm 명령 실행
- 인증: ~/.notebooklm-mcp-cli/profiles/default/cookies.json
- 에러 핸들링: 인증 실패 → DIY fallback 트리거

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import subprocess
import json
import sys
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime


class NotebookLMBridge:
    """
    NotebookLM MCP CLI Wrapper

    Wraps `nlm` CLI commands in Python for 97layerOS integration.
    """

    def __init__(self):
        self.cli_command = "nlm"

        # 인증 확인
        if not self._check_auth():
            raise RuntimeError(
                "NotebookLM 인증 필요.\n"
                "Container: 쿠키 파일이 /root/.notebooklm-mcp-cli/profiles/default/에 있는지 확인\n"
                "Host: nlm login 실행 후 쿠키 복사 필요"
            )

        print("✅ NotebookLM MCP Bridge 초기화 완료")

    def _check_auth(self) -> bool:
        """인증 상태 확인"""
        try:
            result = subprocess.run(
                [self.cli_command, "notebook", "list"],
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.returncode == 0
        except Exception as e:
            print(f"⚠️  인증 확인 실패: {e}")
            return False

    def _run_command(self, args: List[str], timeout: int = 60) -> Dict[str, Any]:
        """
        CLI 명령 실행 및 JSON 파싱

        Args:
            args: CLI 인자 리스트 (예: ["notebook", "list"])
            timeout: 타임아웃 (초)

        Returns:
            파싱된 JSON 또는 텍스트 응답
        """
        try:
            result = subprocess.run(
                [self.cli_command] + args,
                capture_output=True,
                text=True,
                timeout=timeout
            )

            if result.returncode != 0:
                raise RuntimeError(f"CLI Error: {result.stderr}")

            # JSON 파싱 시도
            try:
                return json.loads(result.stdout)
            except json.JSONDecodeError:
                # 텍스트 응답인 경우
                return {"output": result.stdout.strip(), "type": "text"}

        except subprocess.TimeoutExpired:
            raise RuntimeError(f"Command timeout after {timeout}s")

        except Exception as e:
            raise RuntimeError(f"NotebookLM Bridge 오류: {e}")

    # === Foundation Tools ===

    def create_notebook(self, title: str) -> str:
        """
        새 노트북 생성

        Args:
            title: 노트북 제목

        Returns:
            생성된 노트북 ID
        """
        result = self._run_command(["notebook", "create", title])

        # CLI 응답 파싱
        if isinstance(result, dict):
            # JSON 응답
            notebook_id = result.get("id", result.get("notebook_id", ""))
            if notebook_id:
                return notebook_id

            # 텍스트 응답에서 ID 추출
            output = result.get("output", "")
            if "ID:" in output:
                # "ID: ab952c9b-..." 형식
                import re
                match = re.search(r'ID:\s*([a-f0-9\-]+)', output)
                if match:
                    return match.group(1)

        return str(result)

    def list_notebooks(self) -> List[Dict]:
        """
        노트북 목록 조회

        Returns:
            노트북 리스트 [{"id": "...", "title": "...", "source_count": N}, ...]
        """
        result = self._run_command(["notebook", "list"])

        if isinstance(result, list):
            return result
        elif isinstance(result, dict) and "notebooks" in result:
            return result["notebooks"]
        return []

    # === Source Grounding Tools ===

    def add_source_url(self, notebook_id: str, url: str, wait: bool = True) -> str:
        """
        URL 소스 추가 (YouTube, Web)

        Args:
            notebook_id: 대상 노트북 ID
            url: YouTube URL 또는 웹페이지 URL
            wait: 소스 처리 완료까지 대기 (권장)

        Returns:
            생성된 소스 ID
        """
        args = ["source", "add", notebook_id, "--url", url]

        if wait:
            args.append("--wait")

        result = self._run_command(args, timeout=180)  # YouTube 처리 시간 고려

        if isinstance(result, dict):
            source_id = result.get("id", result.get("source_id", ""))
            if source_id:
                return source_id

            # 텍스트 응답에서 ID 추출
            output = result.get("output", "")
            if "ID:" in output or "id:" in output:
                import re
                match = re.search(r'[Ii][Dd]:\s*([a-f0-9\-]+)', output)
                if match:
                    return match.group(1)

        return str(result)

    # === RAG Tool ===

    def query_notebook(self, notebook_id: str, query: str) -> str:
        """
        노트북 소스 기반 질의 (RAG)

        Args:
            notebook_id: 질의할 노트북 ID
            query: 질문 (예: "이 영상의 핵심 메시지를 3줄로 요약해주세요")

        Returns:
            RAG 기반 답변
        """
        result = self._run_command([
            "notebook", "query",
            notebook_id,
            query
        ], timeout=120)  # RAG 처리 시간 고려

        if isinstance(result, dict):
            # 다양한 응답 포맷 처리
            return result.get("answer", result.get("response", result.get("output", "")))
        return str(result)

    # === Multi-modal Synthesis Tools ===

    def create_audio(self, notebook_id: str, format: str = "deep_dive", confirm: bool = True) -> Dict[str, Any]:
        """
        Audio Overview 생성 (Podcast)

        Args:
            notebook_id: 소스 노트북 ID
            format: Overview 형식 (deep_dive, brief, critique, debate)
            confirm: 확인 스킵 (자동 실행)

        Returns:
            {"status": "success", "audio_url": "...", "download_url": "..."}
        """
        args = ["audio", "create", notebook_id, "--format", format]

        if confirm:
            args.append("--confirm")

        result = self._run_command(args, timeout=300)  # Audio 생성 시간 고려 (5분)

        return result


# === Anti-Gravity YouTube Workflow ===

def anti_gravity_youtube(url: str, notebook_title: Optional[str] = None) -> Dict[str, Any]:
    """
    Anti-Gravity YouTube 분석 (NotebookLM 엔진)

    Workflow:
    1. 노트북 생성
    2. YouTube URL 소스 추가
    3. 3가지 RAG 질의 (요약, 인사이트, 브랜드 연결)
    4. Audio Overview 생성
    5. 결과 반환

    Args:
        url: YouTube URL
        notebook_title: 노트북 제목 (기본값: 자동 생성)

    Returns:
        {
            "notebook_id": "...",
            "source_id": "...",
            "summary": "3줄 요약",
            "insights": "핵심 인사이트",
            "brand_connection": "5 Pillars 연결",
            "audio": {"audio_url": "...", ...}
        }
    """

    print("🛸 Anti-Gravity YouTube 분석 시작...")
    print(f"🔗 URL: {url}")

    bridge = NotebookLMBridge()

    # Step 1: 노트북 생성
    if not notebook_title:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        notebook_title = f"YouTube Analysis {timestamp}"

    print(f"📓 노트북 생성: {notebook_title}")
    notebook_id = bridge.create_notebook(notebook_title)
    print(f"✅ Notebook ID: {notebook_id}")

    # Step 2: 소스 추가
    print(f"📥 YouTube 소스 추가 중...")
    source_id = bridge.add_source_url(notebook_id, url)
    print(f"✅ Source ID: {source_id}")

    # Step 3: RAG 질의 (3가지)
    print(f"🤖 RAG 분석 중...")

    print("   [1/4] 3줄 요약...")
    summary = bridge.query_notebook(
        notebook_id,
        "이 영상의 핵심 메시지를 3줄로 요약해주세요."
    )

    print("   [2/4] 인사이트 추출...")
    insights = bridge.query_notebook(
        notebook_id,
        "이 영상에서 가장 독창적인 인사이트는 무엇인가요? "
        "Aesop 스타일로 절제되고 본질적인 언어로 답해주세요."
    )

    print("   [3/4] 브랜드 연결...")
    brand_connection = bridge.query_notebook(
        notebook_id,
        "이 내용이 다음 5가지 브랜드 철학 중 어디에 연결되나요? "
        "1) Authenticity 2) Practicality 3) Elegance 4) Precision 5) Innovation"
    )

    # Step 4: Audio Overview 생성
    print("   [4/4] Audio Overview 생성...")
    audio_result = bridge.create_audio(notebook_id)

    print("✅ Anti-Gravity 분석 완료!")

    return {
        "notebook_id": notebook_id,
        "source_id": source_id,
        "summary": summary,
        "insights": insights,
        "brand_connection": brand_connection,
        "audio": audio_result
    }


# === CLI Entry Point ===

def main():
    """
    CLI 테스트 인터페이스

    Usage:
        python3 notebooklm_bridge.py list
        python3 notebooklm_bridge.py create "Test Notebook"
        python3 notebooklm_bridge.py analyze "https://youtu.be/xxxxx"
    """
    import sys

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python3 notebooklm_bridge.py list")
        print("  python3 notebooklm_bridge.py create <title>")
        print("  python3 notebooklm_bridge.py analyze <youtube_url>")
        sys.exit(1)

    command = sys.argv[1]

    try:
        bridge = NotebookLMBridge()

        if command == "list":
            notebooks = bridge.list_notebooks()
            print(json.dumps(notebooks, indent=2, ensure_ascii=False))

        elif command == "create":
            title = sys.argv[2] if len(sys.argv) > 2 else "Test Notebook"
            notebook_id = bridge.create_notebook(title)
            print(f"Created notebook: {notebook_id}")

        elif command == "analyze":
            url = sys.argv[2]
            result = anti_gravity_youtube(url)
            print("\n" + "="*70)
            print("📊 분석 결과:")
            print("="*70)
            print(f"\n📓 Notebook ID: {result['notebook_id']}")
            print(f"\n📝 요약:\n{result['summary']}")
            print(f"\n💡 인사이트:\n{result['insights']}")
            print(f"\n🎯 브랜드 연결:\n{result['brand_connection']}")
            print(f"\n🎙️  Audio: {result['audio']}")

        else:
            print(f"Unknown command: {command}")
            sys.exit(1)

    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
