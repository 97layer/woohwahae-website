# Claude API Engine for 97layerOS
# Sovereign (Creative Director) Decision Engine
# Author: 97LAYER
# Date: 2026-02-14

import os
import json
import hashlib
import logging
from pathlib import Path
from typing import Optional, Dict, Any, List
from datetime import datetime, timedelta
import anthropic

logger = logging.getLogger(__name__)

class ClaudeEngine:
    """
    Claude API wrapper for deep philosophical judgment.
    Used sparingly for Sovereign (CD) final decisions.
    Optimized for minimal token usage - only called 10-20 times per month.
    """

    def __init__(self, config: Optional[Dict[str, Any]] = None):
        self.config = config or {}

        # Load API key from environment
        self.api_key = os.getenv("ANTHROPIC_API_KEY")
        if not self.api_key:
            # Try loading from .env file
            project_root = Path(__file__).resolve().parent.parent
            env_path = project_root / ".env"
            if env_path.exists():
                with open(env_path, "r", encoding="utf-8") as f:
                    for line in f:
                        if "ANTHROPIC_API_KEY" in line and "=" in line:
                            _, value = line.strip().split("=", 1)
                            self.api_key = value.strip('"').strip("'")
                            break

        # Initialize client if API key exists
        self.client = None
        if self.api_key:
            try:
                self.client = anthropic.Anthropic(api_key=self.api_key)
                logger.info("Claude API client initialized successfully")
            except Exception as e:
                logger.error(f"Failed to initialize Claude client: {e}")

        # Cache directory for expensive Claude responses
        project_root = Path(__file__).resolve().parent.parent
        self.cache_dir = project_root / ".tmp" / "claude_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)

        # Monthly usage tracking
        self.usage_file = self.cache_dir / "monthly_usage.json"
        self.load_usage()

    def load_usage(self):
        """Load monthly usage statistics"""
        if self.usage_file.exists():
            with open(self.usage_file, "r", encoding="utf-8") as f:
                self.usage = json.load(f)
        else:
            self.usage = {
                "month": datetime.now().strftime("%Y-%m"),
                "calls": 0,
                "tokens": 0,
                "cost_estimate": 0.0
            }

    def save_usage(self):
        """Save usage statistics"""
        # Reset if new month
        current_month = datetime.now().strftime("%Y-%m")
        if self.usage["month"] != current_month:
            self.usage = {
                "month": current_month,
                "calls": 0,
                "tokens": 0,
                "cost_estimate": 0.0
            }

        with open(self.usage_file, "w", encoding="utf-8") as f:
            json.dump(self.usage, f, indent=2, ensure_ascii=False)

    def check_usage_limit(self) -> bool:
        """Check if monthly usage limit is exceeded"""
        # Limit: 20 calls per month
        if self.usage["calls"] >= 20:
            logger.warning(f"Claude monthly limit reached: {self.usage['calls']} calls")
            return False
        return True

    def get_cache_key(self, prompt: str, context: str = "") -> str:
        """Generate cache key for prompt"""
        content = f"{prompt}:{context}"
        return hashlib.md5(content.encode()).hexdigest()

    def get_cached_response(self, cache_key: str) -> Optional[str]:
        """Retrieve cached response if exists and still valid (24 hours)"""
        cache_file = self.cache_dir / f"{cache_key}.json"
        if cache_file.exists():
            with open(cache_file, "r", encoding="utf-8") as f:
                cached = json.load(f)
                cached_time = datetime.fromisoformat(cached["timestamp"])
                if datetime.now() - cached_time < timedelta(hours=24):
                    logger.info(f"Using cached Claude response: {cache_key}")
                    return cached["response"]
        return None

    def save_cache(self, cache_key: str, response: str):
        """Save response to cache"""
        cache_file = self.cache_dir / f"{cache_key}.json"
        with open(cache_file, "w", encoding="utf-8") as f:
            json.dump({
                "timestamp": datetime.now().isoformat(),
                "response": response
            }, f, ensure_ascii=False, indent=2)

    def sovereign_judgment(self,
                          content: str,
                          criteria: Optional[Dict[str, Any]] = None,
                          use_cache: bool = True) -> Dict[str, Any]:
        """
        Sovereign (Creative Director) makes final judgment.
        This is the most expensive operation - use sparingly.

        Args:
            content: The content to judge (essay, visual, strategy)
            criteria: MBQ criteria or custom judgment framework
            use_cache: Whether to use cached responses

        Returns:
            Judgment result with approval status and feedback
        """

        if not self.client:
            return {
                "approved": False,
                "reason": "Claude API not configured",
                "fallback": True
            }

        if not self.check_usage_limit():
            return {
                "approved": False,
                "reason": "Monthly Claude limit exceeded - falling back to Gemini",
                "fallback": True
            }

        # Default MBQ criteria from brand constitution
        if not criteria:
            criteria = {
                "philosophy_alignment": "Does it embody 97layer's 5 philosophical pillars?",
                "tone_consistency": "Is it Aesop-like (measured, intellectual, no hyperbole)?",
                "structure_completeness": "Does it follow Hook→Manuscript→Afterglow structure?",
                "anti_algorithm": "Does it resist virality in favor of depth?"
            }

        # Build prompt
        prompt = f"""You are the Sovereign (Creative Director) of WOOHWAHAE, reviewing content for final approval.

Content to Review:
{content}

Judgment Criteria:
{json.dumps(criteria, indent=2, ensure_ascii=False)}

Brand Context:
- WOOHWAHAE: A minimalist lifestyle salon that removes noise, reveals essence
- 97layer: Solitary philosopher, time archivist, algorithm resister
- Standard: Aesop-inspired tone, 400-800 chars, monochrome aesthetic

Provide judgment as JSON:
{{
  "approved": true/false,
  "score": 0-100,
  "strengths": ["..."],
  "weaknesses": ["..."],
  "suggestions": ["..."],
  "revised_version": "..." (only if not approved)
}}

Be strict but constructive. Approve only if it truly embodies the brand essence."""

        # Check cache
        if use_cache:
            cache_key = self.get_cache_key(prompt, content)
            cached = self.get_cached_response(cache_key)
            if cached:
                return json.loads(cached)

        try:
            # Call Claude API - Using Opus for sovereign authority
            response = self.client.messages.create(
                model="claude-3-5-opus-20241022",  # Latest Opus for supreme judgment
                max_tokens=1000,
                temperature=0.3,  # Lower temperature for consistent judgment
                messages=[
                    {
                        "role": "user",
                        "content": prompt
                    }
                ]
            )

            # Parse response
            result_text = response.content[0].text

            # Try to parse as JSON
            try:
                # Find JSON block in response
                import re
                json_match = re.search(r'\{.*\}', result_text, re.DOTALL)
                if json_match:
                    result = json.loads(json_match.group())
                else:
                    result = {"approved": False, "reason": "Could not parse response"}
            except json.JSONDecodeError:
                result = {"approved": False, "reason": "Invalid JSON response"}

            # Update usage
            self.usage["calls"] += 1
            self.usage["tokens"] += len(prompt) // 4 + 1000  # Rough estimate
            self.usage["cost_estimate"] += 0.015 * (len(prompt) // 1000) + 0.075  # Opus pricing
            self.save_usage()

            # Cache response
            if use_cache:
                self.save_cache(cache_key, json.dumps(result, ensure_ascii=False))

            logger.info(f"Claude judgment completed. Calls this month: {self.usage['calls']}/20")
            return result

        except Exception as e:
            logger.error(f"Claude API error: {e}")
            return {
                "approved": False,
                "reason": f"API error: {str(e)}",
                "fallback": True
            }

    def philosophical_reflection(self,
                                trigger: str,
                                context: Optional[List[str]] = None) -> str:
        """
        Deep philosophical reflection on a trigger/event.
        Used for monthly reviews or significant moments.

        Args:
            trigger: The event or thought that triggered reflection
            context: Additional context or past reflections

        Returns:
            A philosophical essay/reflection
        """

        if not self.client:
            return "Claude API not configured - cannot generate deep reflection"

        if not self.check_usage_limit():
            return "Monthly Claude limit exceeded - please wait for next month"

        context_str = "\n".join(context) if context else "No additional context"

        prompt = f"""You are 97layer, a solitary philosopher and time archivist. Generate a deep philosophical reflection.

Trigger:
{trigger}

Past Context:
{context_str}

Guidelines:
- Write in first person, as 97layer
- 800-1500 characters
- Reference time, solitude, imperfection, or algorithmic resistance
- No emoji, no bold text, no hyperbole
- End with a question or incomplete thought
- Korean or English based on trigger language

This is not content for Instagram, but a private philosophical journal entry."""

        try:
            response = self.client.messages.create(
                model="claude-3-sonnet-20240229",  # Sonnet for creative writing
                max_tokens=500,
                temperature=0.7,
                messages=[
                    {
                        "role": "user",
                        "content": prompt
                    }
                ]
            )

            reflection = response.content[0].text

            # Update usage
            self.usage["calls"] += 1
            self.save_usage()

            # Save to journal
            journal_dir = Path(__file__).resolve().parent.parent / "knowledge" / "journal"
            journal_dir.mkdir(parents=True, exist_ok=True)
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            journal_file = journal_dir / f"reflection_{timestamp}.md"

            with open(journal_file, "w", encoding="utf-8") as f:
                f.write(f"# Philosophical Reflection\n\n")
                f.write(f"**Date**: {datetime.now().isoformat()}\n")
                f.write(f"**Trigger**: {trigger}\n\n")
                f.write(f"---\n\n{reflection}\n")

            logger.info(f"Philosophical reflection saved to {journal_file}")
            return reflection

        except Exception as e:
            logger.error(f"Claude reflection error: {e}")
            return f"Error generating reflection: {str(e)}"

    def get_usage_summary(self) -> str:
        """Get monthly usage summary"""
        return (
            f"Claude Usage ({self.usage['month']}):\n"
            f"  Calls: {self.usage['calls']}/20\n"
            f"  Estimated cost: ${self.usage['cost_estimate']:.2f}\n"
            f"  Remaining: {20 - self.usage['calls']} calls"
        )


class DualAIEngine:
    """
    Orchestrator for Gemini (primary) and Claude (premium) engines.
    Implements intelligent routing based on task complexity.
    """

    def __init__(self):
        # Load existing Gemini engine
        from libs.ai_engine import AIEngine
        self.gemini = AIEngine()

        # Initialize Claude engine
        self.claude = ClaudeEngine()

        # Decision threshold
        self.complexity_keywords = [
            "sovereign", "final", "approval", "judgment",
            "philosophy", "brand", "essence", "deep",
            "창조", "철학", "본질", "최종"
        ]

    def requires_sovereign(self, task: str, context: str = "") -> bool:
        """Determine if task requires Sovereign (Claude) judgment"""
        task_lower = task.lower()

        # Check for explicit keywords
        for keyword in self.complexity_keywords:
            if keyword in task_lower:
                return True

        # Check for MBQ approval tasks
        if "mbq" in context.lower() or "publish" in task_lower:
            return True

        # Monthly review or reflection
        if "monthly" in task_lower or "review" in task_lower:
            return True

        return False

    def process(self,
               prompt: str,
               context: Optional[str] = None,
               force_claude: bool = False,
               force_gemini: bool = False) -> Dict[str, Any]:
        """
        Process request with appropriate AI engine.

        Args:
            prompt: The task or question
            context: Additional context
            force_claude: Override to use Claude
            force_gemini: Override to use Gemini

        Returns:
            Response with metadata about which engine was used
        """

        # Forced engine selection
        if force_gemini:
            response = self.gemini.generate(prompt, context)
            return {
                "response": response,
                "engine": "gemini",
                "cost": 0
            }

        if force_claude or self.requires_sovereign(prompt, context or ""):
            # Use Claude for deep judgment
            if "judgment" in prompt.lower() or "approval" in prompt.lower():
                result = self.claude.sovereign_judgment(
                    content=context or prompt
                )
                return {
                    "response": result,
                    "engine": "claude",
                    "cost": self.claude.usage["cost_estimate"]
                }
            else:
                # Philosophical reflection
                result = self.claude.philosophical_reflection(
                    trigger=prompt,
                    context=[context] if context else None
                )
                return {
                    "response": result,
                    "engine": "claude",
                    "cost": self.claude.usage["cost_estimate"]
                }

        # Default to Gemini for efficiency
        response = self.gemini.generate(prompt, context)
        return {
            "response": response,
            "engine": "gemini",
            "cost": 0
        }

    def get_status(self) -> str:
        """Get status of both engines"""
        gemini_status = "✓ Online" if self.gemini.api_key else "✗ No API key"
        claude_status = "✓ Online" if self.claude.client else "✗ No API key"

        status = f"""Dual AI Engine Status:

Gemini Flash (Primary):
  Status: {gemini_status}
  Model: {self.gemini.model_name}
  Cache: {len(list(self.gemini.cache_dir.glob('*.json')))} entries

Claude Opus (Sovereign):
  Status: {claude_status}
  {self.claude.get_usage_summary() if self.claude.client else '  Not configured'}

Routing: Gemini handles 99%, Claude for final judgments only"""

        return status