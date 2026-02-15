#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: libs/skill_engine.py
Author: 97LAYER Mercenary
Date: 2026-02-14
Description: Skill loader, registry, and execution orchestrator for 97layerOS
"""

import logging
import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Any
import yaml
import sys

# Add project root and system directory to path
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))
sys.path.insert(0, str(PROJECT_ROOT / "system"))

# Import core config for container-aware paths
from libs.core_config import KNOWLEDGE_PATHS

logger = logging.getLogger(__name__)

class Skill:
    """Represents a single skill with its metadata and workflow."""

    def __init__(self, skill_id: str, name: str, description: str,
                 version: str, target_agents: List[str],
                 tools: Optional[List[str]] = None,
                 workflow: Optional[str] = None):
        self.id = skill_id
        self.name = name
        self.description = description
        self.version = version
        self.target_agents = target_agents
        self.tools = tools or []
        self.workflow = workflow or ""

    def __repr__(self):
        return f"<Skill {self.id}: {self.name}>"


class SkillEngine:
    """
    Loads skills from the skills/ directory and orchestrates their execution.
    Skills are defined in SKILL.md files with YAML frontmatter.
    """

    def __init__(self, skills_dir: Optional[Path] = None):
        self.skills_dir = skills_dir or Path(__file__).parent.parent / "skills"
        self.registry: Dict[str, Skill] = {}
        self._load_all_skills()

    def _load_all_skills(self) -> None:
        """Scans skills directory and loads all SKILL.md files."""
        if not self.skills_dir.exists():
            logger.warning("Skills directory not found: %s", self.skills_dir)
            return

        for skill_dir in self.skills_dir.iterdir():
            if not skill_dir.is_dir() or skill_dir.name.startswith("."):
                continue

            skill_file = skill_dir / "SKILL.md"
            if skill_file.exists():
                skill = self._parse_skill_file(skill_file)
                if skill:
                    self.registry[skill.id] = skill
                    logger.debug("Loaded skill: %s", skill.id)

    def _parse_skill_file(self, filepath: Path) -> Optional[Skill]:
        """Parses a SKILL.md file with YAML frontmatter."""
        try:
            with open(filepath, "r", encoding="utf-8") as f:
                content = f.read()

            # Extract YAML frontmatter between --- delimiters
            match = re.match(r'^---\s*\n(.*?)\n---\s*\n(.*)$', content, re.DOTALL)
            if not match:
                logger.error("No YAML frontmatter found in: %s", filepath)
                return None

            yaml_content = match.group(1)
            markdown_content = match.group(2)

            # Parse YAML
            metadata = yaml.safe_load(yaml_content)

            # Create Skill object
            skill = Skill(
                skill_id=metadata.get("id", metadata.get("name", filepath.parent.name)),
                name=metadata.get("name", ""),
                description=metadata.get("description", ""),
                version=metadata.get("version", "1.0.0"),
                target_agents=metadata.get("target_agents", []),
                tools=metadata.get("tools", []),
                workflow=markdown_content.strip()
            )

            return skill

        except Exception as e:
            logger.error("Error parsing skill file %s: %s", filepath, e)
            return None

    def get_skill(self, skill_id: str) -> Optional[Skill]:
        """Retrieves a skill by its ID."""
        return self.registry.get(skill_id)

    def list_skills(self) -> List[Skill]:
        """Returns all loaded skills."""
        return list(self.registry.values())

    def detect_skill_from_input(self, text: str) -> Optional[str]:
        """
        Analyzes user input to determine if a skill should be triggered.
        Returns skill_id if a match is found, None otherwise.
        """
        # YouTube URL detection → UIP skill
        youtube_patterns = [
            r'youtube\.com/watch\?v=',
            r'youtu\.be/',
            r'youtube\.com/shorts/'
        ]
        for pattern in youtube_patterns:
            if re.search(pattern, text, re.IGNORECASE):
                return "skill-001"  # UIP skill

        # Add more detection patterns as needed
        # Instagram URL → instagram_content_curator
        if re.search(r'instagram\.com/', text, re.IGNORECASE):
            return "instagram_content_curator"

        return None

    def execute_skill(self, skill_id: str, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Executes a skill based on its ID and context.
        Returns a result dictionary with status and outputs.
        """
        skill = self.get_skill(skill_id)
        if not skill:
            return {
                "status": "error",
                "message": f"Skill not found: {skill_id}"
            }

        logger.info("Executing skill: %s", skill.name)

        # Route to appropriate executor based on skill ID
        if skill_id == "skill-001":  # UIP
            return self._execute_uip(context)
        elif skill_id == "signal_capture":
            return self._execute_signal_capture(context)
        elif skill_id == "data_curation":
            return self._execute_data_curation(context)
        elif skill_id == "instagram_content_curator":
            return self._execute_instagram_curator(context)
        else:
            return {
                "status": "error",
                "message": f"No executor implemented for skill: {skill_id}"
            }

    def _execute_uip(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Executes the Unified Input Protocol (UIP) skill.
        Workflow: Parse URL → Extract content → Transform to ontology → Save to knowledge
        """
        url = context.get("url", "")
        if not url:
            return {"status": "error", "message": "No URL provided"}

        try:
            import subprocess
            import sys
            import json
            from datetime import datetime
            import time

            # Step 1: Parse YouTube video using subprocess
            project_root = Path(__file__).parent.parent
            parser_script = project_root / "execution" / "youtube_parser.py"

            result = subprocess.run(
                [sys.executable, str(parser_script), url],
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0:
                return {
                    "status": "error",
                    "message": "YouTube parsing failed",
                    "details": result.stderr
                }

            parse_data = json.loads(result.stdout)

            if "error" in parse_data:
                return {
                    "status": "error",
                    "message": f"YouTube parsing error: {parse_data['error']}"
                }

            # Step 2: Create raw signal file using container-aware paths
            metadata = parse_data.get("metadata", {})
            raw_signals_dir = KNOWLEDGE_PATHS["signals"]
            raw_signals_dir.mkdir(parents=True, exist_ok=True)

            # Generate signal ID
            signal_id = f"rs-{int(time.time())}"
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

            # Create safe filename from title
            title = metadata.get("title", "youtube_video")
            safe_title = re.sub(r'[^\w\s-]', '', title)[:50]
            safe_title = re.sub(r'[-\s]+', '_', safe_title)

            filename = f"{signal_id}_youtube_{safe_title}.md"
            filepath = raw_signals_dir / filename

            # Generate markdown content
            content = f"""---
id: {signal_id}
source: YouTube
url: {url}
title: {metadata.get("title", "N/A")}
author: {metadata.get("author", "N/A")}
captured_at: {timestamp}
tags: [raw_signal, youtube, uip]
---

# {metadata.get("title", "YouTube Video")}

**Source**: {url}
**Author**: {metadata.get("author", "N/A")}
**Video ID**: {metadata.get("id", "N/A")}

## Metadata

{json.dumps(metadata, indent=2, ensure_ascii=False)}

## Transcript

{parse_data.get("transcript", "[Pending] Transcript extraction requires Gemini Web integration.")}

## Notes

This raw signal was automatically captured by the UIP (Unified Input Protocol) skill.
Next step: Analyze and transform into strategic knowledge assets for 97layerOS.
"""

            # Write to file
            with open(filepath, "w", encoding="utf-8") as f:
                f.write(content)

            logger.info("UIP skill created raw signal: %s", filepath)

            return {
                "status": "success",
                "message": f"UIP executed successfully. Raw signal created.",
                "output_file": str(filepath),
                "signal_id": signal_id,
                "title": metadata.get("title")
            }

        except subprocess.TimeoutExpired:
            return {
                "status": "error",
                "message": "YouTube parsing timeout"
            }
        except Exception as e:
            logger.error("UIP execution failed: %s", e)
            import traceback
            return {
                "status": "error",
                "message": f"Execution error: {e}",
                "traceback": traceback.format_exc()
            }

    def _execute_signal_capture(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """Executes signal capture skill."""
        # Placeholder for signal capture implementation
        return {
            "status": "not_implemented",
            "message": "Signal capture executor not yet implemented"
        }

    def _execute_data_curation(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """Executes data curation skill."""
        return {
            "status": "not_implemented",
            "message": "Data curation executor not yet implemented"
        }

    def _execute_instagram_curator(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """Executes Instagram content curator skill."""
        return {
            "status": "not_implemented",
            "message": "Instagram curator executor not yet implemented"
        }

    def get_skill_summary(self) -> str:
        """Returns a formatted summary of all loaded skills."""
        if not self.registry:
            return "No skills loaded."

        lines = ["Loaded Skills:"]
        for skill in self.registry.values():
            lines.append(f"- {skill.id}: {skill.name} (v{skill.version})")
            lines.append(f"  {skill.description[:80]}...")

        return "\n".join(lines)
