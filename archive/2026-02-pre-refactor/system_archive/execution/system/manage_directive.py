# Filename: execution/system/manage_directive.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import os
import re
import logging
from pathlib import Path
from typing import Dict, List, Optional, Tuple

logger = logging.getLogger(__name__)

class DirectiveManager:
    """Manages system directives and agent protocols."""
    
    def __init__(self, workspace_root: Optional[str] = None):
        if workspace_root is None:
            workspace_root = Path(__file__).resolve().parent.parent.parent
        self.workspace_root = Path(workspace_root)
        self.directives_dir = self.workspace_root / 'directives'
        self.directives_dir.mkdir(parents=True, exist_ok=True)

    def parse_markdown(self, content: str) -> Dict[str, str]:
        """Parses markdown into sections based on H2 headers."""
        sections = {}
        current_section = "PREAMBLE"
        lines = []
        
        for line in content.split('\n'):
            if line.startswith('## '):
                sections[current_section] = '\n'.join(lines).strip()
                current_section = line.replace('## ', '').strip().upper()
                lines = []
            else:
                lines.append(line)
        
        sections[current_section] = '\n'.join(lines).strip()
        return sections

    def read_directive(self, topic: str) -> Dict[str, str]:
        """Reads and parses a directive file."""
        # Check if topic is a full path or just a name. 
        # If it's a name like 'strategy_analyst.md', check agents folder too.
        
        potential_paths = [
            self.directives_dir / topic,
            self.directives_dir / 'agents' / topic
        ]
        
        if not topic.endswith('.md'):
             potential_paths.append(self.directives_dir / f"{topic}.md")
             potential_paths.append(self.directives_dir / 'agents' / f"{topic}.md")

        path = None
        for p in potential_paths:
            if p.exists():
                path = p
                break
        
        if not path:
            return {}
        
        try:
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read()
            return self.parse_markdown(content)
        except Exception as e:
            logger.error(f"Failed to read directive {topic}: {e}")
            return {}

    def update_directive(self, topic: str, section: str, content: str, mode: str = 'replace') -> bool:
        """Updates a specific section of a directive."""
        sections = self.read_directive(topic)
        section = section.upper()
        
        if mode == 'append' and section in sections:
            sections[section] = sections[section] + "\n" + content
        else:
            sections[section] = content
            
        # Reconstruct markdown
        md_content = ""
        for sec_name, sec_body in sections.items():
            if sec_name == "PREAMBLE":
                md_content += sec_body + "\n\n"
            else:
                md_content += f"## {sec_name}\n{sec_body}\n\n"
        
        
        try:
            # Re-resolve path logic (similar to read_directive)
            potential_paths = [
                self.directives_dir / topic,
                self.directives_dir / 'agents' / topic
            ]
            if not topic.endswith('.md'):
                 potential_paths.append(self.directives_dir / f"{topic}.md")
                 potential_paths.append(self.directives_dir / 'agents' / f"{topic}.md")

            path = None
            for p in potential_paths:
                if p.exists():
                    path = p
                    break
            
            # If not found, default to direct path for creation
            if not path:
                path = self.directives_dir / topic

            path.parent.mkdir(parents=True, exist_ok=True)
            with open(path, 'w', encoding='utf-8') as f:
                f.write(md_content.strip())
            return True
        except Exception as e:
            logger.error(f"Failed to update directive {topic}: {e}")
            return False
