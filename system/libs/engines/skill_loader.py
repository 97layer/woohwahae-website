#!/usr/bin/env python3
"""
Skill Loader for 97layerOS

Loads and manages skill modules from core/skills/
Provides validation, injection, and execution capabilities.
"""

import re
from pathlib import Path
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
import yaml


@dataclass
class Skill:
    """Skill data structure"""
    id: str
    name: str
    path: Path
    purpose: str
    rules: List[str]
    validation_criteria: Dict[str, List[str]]
    examples: Dict[str, str]
    version: str
    content: str


class SkillLoader:
    """
    Load and manage skills from core/skills/

    Usage:
        skill = SkillLoader.load("brand_voice_v1")
        result = skill.validate(content)
    """

    SKILLS_DIR = Path(__file__).parent.parent / "core" / "skills"

    @classmethod
    def load(cls, skill_id: str) -> Skill:
        """
        Load a skill by ID

        Args:
            skill_id: Skill identifier (e.g., "brand_voice_v1")

        Returns:
            Skill object

        Raises:
            FileNotFoundError: If skill file not found
            ValueError: If skill format invalid
        """
        # Find skill file
        skill_file = cls._find_skill_file(skill_id)

        if not skill_file:
            raise FileNotFoundError(f"Skill not found: {skill_id}")

        # Parse skill document
        return cls._parse_skill(skill_file)

    @classmethod
    def list_skills(cls) -> List[Dict[str, str]]:
        """
        List all available skills

        Returns:
            List of skill metadata dicts
        """
        skills = []

        if not cls.SKILLS_DIR.exists():
            return skills

        for skill_file in cls.SKILLS_DIR.glob("*.skill.md"):
            try:
                skill = cls._parse_skill(skill_file)
                skills.append({
                    "id": skill.id,
                    "name": skill.name,
                    "purpose": skill.purpose,
                    "version": skill.version,
                    "path": str(skill.path)
                })
            except Exception as e:
                print(f"Warning: Failed to load {skill_file.name}: {e}")
                continue

        return skills

    @classmethod
    def inject_to_prompt(cls, skill_ids: List[str]) -> str:
        """
        Inject skills into AI prompt

        Args:
            skill_ids: List of skill IDs to inject

        Returns:
            Formatted prompt text with skill rules
        """
        prompt_parts = ["# Active Skills\n"]

        for skill_id in skill_ids:
            try:
                skill = cls.load(skill_id)
                prompt_parts.append(f"\n## {skill.name}\n")
                prompt_parts.append(f"**Purpose**: {skill.purpose}\n\n")
                prompt_parts.append("**Rules**:\n")

                for i, rule in enumerate(skill.rules[:5], 1):  # Top 5 rules
                    prompt_parts.append(f"{i}. {rule}\n")

            except Exception as e:
                print(f"Warning: Failed to inject skill {skill_id}: {e}")
                continue

        return "".join(prompt_parts)

    @classmethod
    def _find_skill_file(cls, skill_id: str) -> Optional[Path]:
        """Find skill file by ID or name"""

        if not cls.SKILLS_DIR.exists():
            return None

        # Try exact ID match
        for skill_file in cls.SKILLS_DIR.glob("*.skill.md"):
            content = skill_file.read_text(encoding="utf-8")
            if f"skill_id: `{skill_id}`" in content or f"Skill ID\n`{skill_id}`" in content:
                return skill_file

        # Try name match
        skill_name = skill_id.replace("_v1", "").replace("_", " ")
        for skill_file in cls.SKILLS_DIR.glob("*.skill.md"):
            if skill_name in skill_file.stem.lower():
                return skill_file

        return None

    @classmethod
    def _parse_skill(cls, skill_file: Path) -> Skill:
        """Parse skill markdown file"""

        content = skill_file.read_text(encoding="utf-8")

        # Extract metadata
        skill_id = cls._extract_field(content, "Skill ID")
        name = cls._extract_title(content)
        purpose = cls._extract_field(content, "Purpose")
        rules = cls._extract_rules(content)
        validation = cls._extract_validation(content)
        examples = cls._extract_examples(content)
        version = cls._extract_version(content)

        return Skill(
            id=skill_id,
            name=name,
            path=skill_file,
            purpose=purpose,
            rules=rules,
            validation_criteria=validation,
            examples=examples,
            version=version,
            content=content
        )

    @staticmethod
    def _extract_title(content: str) -> str:
        """Extract skill title from markdown"""
        match = re.search(r'^# (.+)$', content, re.MULTILINE)
        return match.group(1) if match else "Unknown Skill"

    @staticmethod
    def _extract_field(content: str, field_name: str) -> str:
        """Extract single field value"""

        # Try ## Field format
        pattern = f"## {field_name}\n(.+?)(?=\n##|\Z)"
        match = re.search(pattern, content, re.DOTALL)

        if match:
            value = match.group(1).strip()
            # Remove markdown code blocks
            value = re.sub(r'^`+|`+$', '', value)
            return value

        return ""

    @staticmethod
    def _extract_rules(content: str) -> List[str]:
        """Extract rules from ## Rules section"""

        # Find ## Rules section
        pattern = r"## Rules\n(.+?)(?=\n## |\Z)"
        match = re.search(pattern, content, re.DOTALL)

        if not match:
            return []

        rules_text = match.group(1)

        # Extract numbered/bulleted rules
        rules = []

        # Look for ### subsections as rules
        subsections = re.findall(r'### (\d+\..+?)\n', rules_text)
        if subsections:
            rules.extend(subsections)

        # Look for numbered lists
        numbered = re.findall(r'^\d+\.\s+(.+?)$', rules_text, re.MULTILINE)
        if numbered:
            rules.extend(numbered)

        return rules[:20]  # Limit to top 20 rules

    @staticmethod
    def _extract_validation(content: str) -> Dict[str, List[str]]:
        """Extract validation criteria"""

        validation = {
            "automated": [],
            "manual": []
        }

        # Find ## Validation Criteria section
        pattern = r"## Validation Criteria\n(.+?)(?=\n## |\Z)"
        match = re.search(pattern, content, re.DOTALL)

        if not match:
            return validation

        criteria_text = match.group(1)

        # Extract checkboxes
        checkboxes = re.findall(r'- \[ \] (.+?)$', criteria_text, re.MULTILINE)

        # Categorize
        for checkbox in checkboxes:
            if any(keyword in checkbox.lower() for keyword in ["automated", "score", "check", "ratio", "%"]):
                validation["automated"].append(checkbox)
            else:
                validation["manual"].append(checkbox)

        return validation

    @staticmethod
    def _extract_examples(content: str) -> Dict[str, str]:
        """Extract good/bad examples"""

        examples = {
            "bad": [],
            "good": []
        }

        # Find examples section
        pattern = r"## Examples\n(.+?)(?=\n## |\Z)"
        match = re.search(pattern, content, re.DOTALL)

        if not match:
            return examples

        examples_text = match.group(1)

        # Extract BAD examples
        bad_pattern = r"### ❌ BAD:(.+?)(?=###|\Z)"
        bad_matches = re.findall(bad_pattern, examples_text, re.DOTALL)
        examples["bad"] = [m.strip() for m in bad_matches]

        # Extract GOOD examples
        good_pattern = r"### ✅ GOOD:(.+?)(?=###|\Z)"
        good_matches = re.findall(good_pattern, examples_text, re.DOTALL)
        examples["good"] = [m.strip() for m in good_matches]

        return examples

    @staticmethod
    def _extract_version(content: str) -> str:
        """Extract version from content"""
        match = re.search(r'v(\d+\.\d+)', content)
        return match.group(0) if match else "v1.0"


class SkillValidator:
    """
    Validate content against skill rules

    Usage:
        validator = SkillValidator(skill)
        result = validator.validate(content)
    """

    def __init__(self, skill: Skill):
        self.skill = skill

    def validate(self, content: str) -> Dict[str, Any]:
        """
        Validate content against skill criteria

        Args:
            content: Content to validate

        Returns:
            Validation result dict
        """
        results = {
            "passed": False,
            "skill_id": self.skill.id,
            "checks": {},
            "issues": [],
            "suggestions": []
        }

        # Run automated checks based on skill type
        if "brand_voice" in self.skill.id:
            results = self._validate_brand_voice(content)
        elif "design_guide" in self.skill.id:
            results = self._validate_design(content)
        elif "instagram" in self.skill.id:
            results = self._validate_instagram(content)

        # Determine overall pass/fail
        if results["checks"]:
            passed_count = sum(1 for v in results["checks"].values() if v)
            total_count = len(results["checks"])
            results["passed"] = passed_count / total_count >= 0.7  # 70% threshold

        return results

    def _validate_brand_voice(self, content: str) -> Dict[str, Any]:
        """Validate brand voice criteria"""

        checks = {}
        issues = []
        suggestions = []

        # Check 1: Length (not too short)
        word_count = len(content.split())
        checks["length"] = word_count >= 100
        if not checks["length"]:
            issues.append(f"Content too short: {word_count} words (min 100)")

        # Check 2: No banned words
        banned = ["팔로우", "좋아요", "구독", "알림", "대박", "완전", "여러분"]
        found_banned = [w for w in banned if w in content]
        checks["no_banned_words"] = len(found_banned) == 0
        if found_banned:
            issues.append(f"Banned words found: {', '.join(found_banned)}")
            suggestions.append("Remove algorithm-focused language")

        # Check 3: Has depth (philosophical keywords)
        depth_keywords = ["본질", "진정성", "덜어냄", "여백", "고독", "철학", "의미"]
        has_depth = any(kw in content for kw in depth_keywords)
        checks["has_depth"] = has_depth
        if not has_depth:
            issues.append("Lacks philosophical depth")
            suggestions.append("Add deeper meaning or reflection")

        # Check 4: Structure (paragraphs)
        paragraphs = [p for p in content.split('\n\n') if p.strip()]
        checks["structure"] = len(paragraphs) >= 3
        if not checks["structure"]:
            issues.append("Insufficient structure (need 3+ paragraphs)")

        return {
            "passed": all(checks.values()),
            "skill_id": self.skill.id,
            "checks": checks,
            "issues": issues,
            "suggestions": suggestions
        }

    def _validate_design(self, content: str) -> Dict[str, Any]:
        """Validate design guidelines (placeholder)"""
        # For design validation, this would analyze image files
        # For now, return basic structure check
        return {
            "passed": True,
            "skill_id": self.skill.id,
            "checks": {"structure": True},
            "issues": [],
            "suggestions": ["Use image analysis tools for full validation"]
        }

    def _validate_instagram(self, content: str) -> Dict[str, Any]:
        """Validate Instagram content"""

        checks = {}
        issues = []
        suggestions = []

        # Check 1: Length (Instagram optimal)
        char_count = len(content)
        checks["optimal_length"] = 800 <= char_count <= 1500
        if not checks["optimal_length"]:
            issues.append(f"Length {char_count} chars (optimal: 800-1500)")

        # Check 2: Hashtag count (0-3)
        hashtag_count = content.count('#')
        checks["hashtag_limit"] = hashtag_count <= 3
        if not checks["hashtag_limit"]:
            issues.append(f"Too many hashtags: {hashtag_count} (max 3)")
            suggestions.append("Remove excessive hashtags")

        # Check 3: No aggressive CTA
        aggressive_cta = ["지금 바로", "꼭", "반드시", "무조건"]
        has_aggressive = any(cta in content for cta in aggressive_cta)
        checks["no_aggressive_cta"] = not has_aggressive
        if has_aggressive:
            issues.append("Contains aggressive CTA")
            suggestions.append("Use softer invitation language")

        # Check 4: Has reflective question
        has_question = "?" in content[-200:]  # Question in last 200 chars
        checks["reflective_question"] = has_question
        if not has_question:
            issues.append("Missing reflective question at end")
            suggestions.append("Add open-ended question for afterglow")

        return {
            "passed": all(checks.values()),
            "skill_id": self.skill.id,
            "checks": checks,
            "issues": issues,
            "suggestions": suggestions
        }


# Convenience functions
def load_skill(skill_id: str) -> Skill:
    """Load a skill by ID"""
    return SkillLoader.load(skill_id)


def list_skills() -> List[Dict[str, str]]:
    """List all available skills"""
    return SkillLoader.list_skills()


def validate_content(skill_id: str, content: str) -> Dict[str, Any]:
    """Validate content against a skill"""
    skill = SkillLoader.load(skill_id)
    validator = SkillValidator(skill)
    return validator.validate(content)


def inject_skills_to_prompt(skill_ids: List[str]) -> str:
    """Inject multiple skills into prompt"""
    return SkillLoader.inject_to_prompt(skill_ids)


if __name__ == "__main__":
    # CLI interface
    import sys
    import json

    def print_usage():
        print("""
Usage:
  python skill_loader.py list
  python skill_loader.py load <skill_id>
  python skill_loader.py validate <skill_id> <content_file>
  python skill_loader.py inject <skill_id1> <skill_id2> ...
        """)

    if len(sys.argv) < 2:
        print_usage()
        sys.exit(1)

    command = sys.argv[1]

    if command == "list":
        skills = list_skills()
        print(json.dumps(skills, indent=2, ensure_ascii=False))

    elif command == "load" and len(sys.argv) >= 3:
        skill_id = sys.argv[2]
        skill = load_skill(skill_id)
        print(f"Skill: {skill.name}")
        print(f"ID: {skill.id}")
        print(f"Purpose: {skill.purpose}")
        print(f"Rules: {len(skill.rules)}")
        print(f"Version: {skill.version}")

    elif command == "validate" and len(sys.argv) >= 4:
        skill_id = sys.argv[2]
        content_file = sys.argv[3]

        with open(content_file, 'r', encoding='utf-8') as f:
            content = f.read()

        result = validate_content(skill_id, content)
        print(json.dumps(result, indent=2, ensure_ascii=False))

    elif command == "inject":
        skill_ids = sys.argv[2:]
        prompt = inject_skills_to_prompt(skill_ids)
        print(prompt)

    else:
        print_usage()
        sys.exit(1)
