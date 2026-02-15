# Filename: libs/context_router.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import logging
from enum import Enum
from typing import Dict, Any

logger = logging.getLogger(__name__)

class ContextType(Enum):
    IDEA = "idea"
    BRANDING = "branding"
    CODING = "coding"
    SYSTEM = "system"
    UNKNOWN = "unknown"

class ContextRouter:
    """Routes messages based on intent and content analysis."""
    
    def __init__(self, ai_engine):
        self.ai = ai_engine

    def route(self, text: str) -> ContextType:
        """Analyzes text and returns the appropriate ContextType."""
        prompt = f"다음 텍스트의 범주를 하나만 선택하십시오: [idea, branding, coding, system]\n텍스트: {text}"
        res = self.ai.generate_response(prompt, model_type="flash").lower().strip()
        
        for ct in ContextType:
            if ct.value in res:
                return ct
        return ContextType.UNKNOWN
