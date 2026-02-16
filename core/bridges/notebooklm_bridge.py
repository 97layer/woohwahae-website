#!/usr/bin/env python3
"""
NotebookLM MCP Bridge - 97layerOS Wrapper

Provides Python interface to NotebookLM MCP CLI tools.
Handles authentication, command execution, and error handling.

Key Features:
- 28 NotebookLM tools via CLI wrapper
- Cookie-based authentication
- Automatic fallback to DIY on failure
- Integration with 97layerOS knowledge base

Author: 97layerOS Technical Director
Created: 2026-02-16
Priority: P0 (Critical for conversational AI)
"""

import subprocess
import json
import os
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime
import logging

logger = logging.getLogger(__name__)

# Project root
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class NotebookLMBridge:
    """
    NotebookLM MCP CLI Wrapper

    Provides Python interface to NotebookLM's 28 tools:
    - notebook_create/list (Foundation)
    - source_add_url/text/file (Source Grounding)
    - notebook_query (RAG)
    - audio_create (Multi-modal)
    - mindmap_create (Multi-modal)
    """

    def __init__(self, cli_command: str = "nlm"):
        """
        Initialize NotebookLM Bridge

        Args:
            cli_command: CLI command name (default: 'nlm')
        """
        self.cli_command = cli_command
        self.authenticated = False

        # Check authentication
        if self._check_auth():
            self.authenticated = True
            logger.info("✅ NotebookLM authenticated")
        else:
            logger.warning("⚠️  NotebookLM not authenticated - run 'nlm login'")

    def _check_auth(self) -> bool:
        """Check if NotebookLM is authenticated"""
        try:
            result = subprocess.run(
                [self.cli_command, "notebook_list"],
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.returncode == 0
        except Exception as e:
            logger.debug(f"Auth check failed: {e}")
            return False

    def _run_command(self, args: List[str], timeout: int = 60) -> Dict[str, Any]:
        """
        Execute CLI command and parse JSON response

        Args:
            args: Command arguments
            timeout: Command timeout in seconds

        Returns:
            Parsed JSON response or text output

        Raises:
            RuntimeError: On command failure
        """
        try:
            result = subprocess.run(
                [self.cli_command] + args,
                capture_output=True,
                text=True,
                timeout=timeout
            )

            if result.returncode != 0:
                error_msg = result.stderr or result.stdout
                raise RuntimeError(f"CLI command failed: {error_msg}")

            # Try to parse JSON
            try:
                return json.loads(result.stdout)
            except json.JSONDecodeError:
                # Return as text if not JSON
                return {"output": result.stdout.strip(), "type": "text"}

        except subprocess.TimeoutExpired:
            raise RuntimeError(f"Command timed out after {timeout}s")
        except Exception as e:
            raise RuntimeError(f"NotebookLM Bridge error: {e}")

    # ========================
    # Foundation Tools
    # ========================

    def create_notebook(self, title: str) -> str:
        """
        Create new notebook

        Args:
            title: Notebook title

        Returns:
            notebook_id
        """
        logger.info(f"Creating notebook: {title}")
        result = self._run_command(["notebook_create", "--title", title])
        notebook_id = result.get("notebook_id", result.get("id"))
        logger.info(f"✅ Notebook created: {notebook_id}")
        return notebook_id

    def list_notebooks(self) -> List[Dict]:
        """
        List all notebooks

        Returns:
            List of notebook dicts with id, title, created_at
        """
        result = self._run_command(["notebook_list"])
        notebooks = result.get("notebooks", [])
        logger.info(f"Found {len(notebooks)} notebooks")
        return notebooks

    def get_notebook(self, notebook_id: str) -> Dict:
        """
        Get notebook details

        Args:
            notebook_id: Notebook ID

        Returns:
            Notebook details dict
        """
        result = self._run_command(["notebook_get", "--notebook-id", notebook_id])
        return result

    # ========================
    # Source Grounding Tools
    # ========================

    def add_source_url(self, notebook_id: str, url: str, title: Optional[str] = None) -> str:
        """
        Add URL source to notebook (YouTube, Web)

        Args:
            notebook_id: Target notebook ID
            url: Source URL
            title: Optional source title

        Returns:
            source_id
        """
        logger.info(f"Adding URL source: {url}")

        args = ["source_add_url", "--notebook-id", notebook_id, "--url", url]
        if title:
            args.extend(["--title", title])

        result = self._run_command(args)
        source_id = result.get("source_id", result.get("id"))
        logger.info(f"✅ Source added: {source_id}")
        return source_id

    def add_source_text(self, notebook_id: str, text: str, title: str) -> str:
        """
        Add text source to notebook

        Args:
            notebook_id: Target notebook ID
            text: Source text content
            title: Source title

        Returns:
            source_id
        """
        logger.info(f"Adding text source: {title}")

        result = self._run_command([
            "source_add_text",
            "--notebook-id", notebook_id,
            "--title", title,
            "--text", text
        ])

        source_id = result.get("source_id", result.get("id"))
        logger.info(f"✅ Source added: {source_id}")
        return source_id

    def add_source_file(self, notebook_id: str, file_path: Path) -> str:
        """
        Add file source to notebook (PDF, DOCX, etc.)

        Args:
            notebook_id: Target notebook ID
            file_path: Path to file

        Returns:
            source_id
        """
        logger.info(f"Adding file source: {file_path.name}")

        if not file_path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")

        result = self._run_command([
            "source_add_file",
            "--notebook-id", notebook_id,
            "--file", str(file_path)
        ])

        source_id = result.get("source_id", result.get("id"))
        logger.info(f"✅ Source added: {source_id}")
        return source_id

    def list_sources(self, notebook_id: str) -> List[Dict]:
        """
        List sources in notebook

        Args:
            notebook_id: Notebook ID

        Returns:
            List of source dicts
        """
        result = self._run_command(["source_list", "--notebook-id", notebook_id])
        sources = result.get("sources", [])
        logger.info(f"Found {len(sources)} sources in notebook {notebook_id}")
        return sources

    # ========================
    # RAG Query Tool
    # ========================

    def query_notebook(self, notebook_id: str, query: str, max_tokens: int = 2048) -> str:
        """
        Query notebook with RAG (Retrieval-Augmented Generation)

        Args:
            notebook_id: Notebook ID to query
            query: User question
            max_tokens: Max response tokens

        Returns:
            Answer text
        """
        logger.info(f"Querying notebook {notebook_id}: {query[:50]}...")

        result = self._run_command([
            "notebook_query",
            "--notebook-id", notebook_id,
            "--query", query,
            "--max-tokens", str(max_tokens)
        ], timeout=120)  # Longer timeout for queries

        answer = result.get("answer", result.get("output", ""))
        logger.info(f"✅ Query answered ({len(answer)} chars)")
        return answer

    # ========================
    # Multi-modal Synthesis Tools
    # ========================

    def create_audio(self, notebook_id: str, output_path: Optional[Path] = None) -> Path:
        """
        Create Audio Overview (Podcast) from notebook

        Args:
            notebook_id: Notebook ID
            output_path: Optional output file path

        Returns:
            Path to generated audio file
        """
        logger.info(f"Creating audio overview for notebook {notebook_id}")

        args = ["audio_create", "--notebook-id", notebook_id]

        if output_path:
            args.extend(["--output", str(output_path)])

        result = self._run_command(args, timeout=300)  # 5 min timeout for audio

        # Get audio file path
        audio_file = result.get("audio_file", result.get("output_file", result.get("output")))
        if not audio_file:
            raise RuntimeError("Audio creation failed - no output file")

        audio_path = Path(audio_file)
        logger.info(f"✅ Audio created: {audio_path}")
        return audio_path

    def create_mindmap(self, notebook_id: str) -> str:
        """
        Create Mind Map (Mermaid.js) from notebook

        Args:
            notebook_id: Notebook ID

        Returns:
            Mermaid.js code
        """
        logger.info(f"Creating mind map for notebook {notebook_id}")

        result = self._run_command([
            "mindmap_create",
            "--notebook-id", notebook_id
        ], timeout=120)

        mermaid_code = result.get("mermaid_code", result.get("output", ""))
        logger.info(f"✅ Mind map created ({len(mermaid_code)} chars)")
        return mermaid_code

    # ========================
    # High-Level Workflows
    # ========================

    def anti_gravity_youtube(self, url: str) -> Dict[str, Any]:
        """
        Anti-Gravity YouTube Analysis Workflow

        Steps:
        1. Create notebook
        2. Add YouTube URL source
        3. Run 3 RAG queries (summary, insights, brand connection)
        4. Create audio overview
        5. Create mind map

        Args:
            url: YouTube URL

        Returns:
            Complete analysis results
        """
        logger.info(f"🛸 Anti-Gravity YouTube Analysis: {url}")

        # Step 1: Create notebook
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        notebook_id = self.create_notebook(f"YouTube Analysis {timestamp}")

        # Step 2: Add source
        source_id = self.add_source_url(notebook_id, url, title="YouTube Video")

        # Step 3: RAG queries
        summary = self.query_notebook(
            notebook_id,
            "이 영상의 핵심 메시지를 3줄로 요약해주세요."
        )

        insights = self.query_notebook(
            notebook_id,
            "이 영상에서 가장 독창적인 인사이트는 무엇인가요? "
            "Aesop 스타일로 절제되고 본질적인 언어로 답해주세요."
        )

        brand_connection = self.query_notebook(
            notebook_id,
            "이 내용이 다음 5가지 브랜드 철학 중 어디에 연결되나요? "
            "1) Authenticity 2) Practicality 3) Elegance 4) Precision 5) Innovation"
        )

        # Step 4: Multi-modal synthesis
        try:
            audio_path = self.create_audio(notebook_id)
        except Exception as e:
            logger.warning(f"Audio creation failed: {e}")
            audio_path = None

        try:
            mindmap_mermaid = self.create_mindmap(notebook_id)
        except Exception as e:
            logger.warning(f"Mind map creation failed: {e}")
            mindmap_mermaid = ""

        result = {
            "notebook_id": notebook_id,
            "source_id": source_id,
            "url": url,
            "summary": summary,
            "insights": insights,
            "brand_connection": brand_connection,
            "audio_file": audio_path,
            "mindmap": mindmap_mermaid,
            "timestamp": datetime.now().isoformat()
        }

        logger.info("✅ Anti-Gravity YouTube Analysis complete")
        return result

    def query_knowledge_base(self, question: str) -> str:
        """
        Query 97layerOS knowledge base

        Creates temporary notebook with knowledge docs and queries it.

        Args:
            question: User question

        Returns:
            Answer based on knowledge base
        """
        logger.info(f"Querying knowledge base: {question[:50]}...")

        # Find or create knowledge base notebook
        notebooks = self.list_notebooks()
        kb_notebook = None

        for nb in notebooks:
            if "97layerOS Knowledge Base" in nb.get("title", ""):
                kb_notebook = nb
                break

        # Create if doesn't exist
        if not kb_notebook:
            logger.info("Creating knowledge base notebook...")
            notebook_id = self.create_notebook("97layerOS Knowledge Base")

            # Add key knowledge documents
            docs_dir = PROJECT_ROOT / 'knowledge' / 'docs'
            if docs_dir.exists():
                for doc_file in docs_dir.glob('*.md'):
                    try:
                        self.add_source_file(notebook_id, doc_file)
                        logger.info(f"  Added: {doc_file.name}")
                    except Exception as e:
                        logger.warning(f"  Failed to add {doc_file.name}: {e}")
        else:
            notebook_id = kb_notebook.get("id", kb_notebook.get("notebook_id"))

        # Query the notebook
        answer = self.query_notebook(notebook_id, question)
        return answer


# ========================
# Convenience Functions
# ========================

def get_bridge() -> NotebookLMBridge:
    """Get NotebookLM Bridge instance (singleton pattern)"""
    if not hasattr(get_bridge, '_instance'):
        get_bridge._instance = NotebookLMBridge()
    return get_bridge._instance


def is_available() -> bool:
    """Check if NotebookLM MCP is available and authenticated"""
    try:
        bridge = NotebookLMBridge()
        return bridge.authenticated
    except Exception:
        return False


# ========================
# CLI Testing
# ========================

def main():
    """CLI test interface"""
    import argparse

    parser = argparse.ArgumentParser(description='NotebookLM Bridge CLI Test')
    parser.add_argument('command', choices=['test', 'list', 'query', 'youtube'])
    parser.add_argument('--notebook-id', help='Notebook ID')
    parser.add_argument('--query', help='Query text')
    parser.add_argument('--url', help='YouTube URL')

    args = parser.parse_args()

    # Configure logging
    logging.basicConfig(level=logging.INFO)

    bridge = NotebookLMBridge()

    if args.command == 'test':
        print(f"Authenticated: {bridge.authenticated}")

    elif args.command == 'list':
        notebooks = bridge.list_notebooks()
        print(json.dumps(notebooks, indent=2))

    elif args.command == 'query':
        if not args.notebook_id or not args.query:
            print("Error: --notebook-id and --query required")
            return
        answer = bridge.query_notebook(args.notebook_id, args.query)
        print(f"\nAnswer:\n{answer}")

    elif args.command == 'youtube':
        if not args.url:
            print("Error: --url required")
            return
        result = bridge.anti_gravity_youtube(args.url)
        print(json.dumps(result, indent=2, default=str))


if __name__ == "__main__":
    main()
