#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: agent_router.py
Author: 97LAYER Mercenary
Date: 2026-02-12
Description: 에이전트 페르소나 로딩 및 메시지 기반 라우팅
"""

import logging
import os
from typing import Optional, Dict

logger = logging.getLogger(__name__)

AGENT_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "directives", "agents")
DIRECTIVES_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "directives")

AGENT_REGISTRY: Dict[str, Dict[str, str]] = {
    "CD": {"file": "creative_director.md", "name": "Creative Director", "label": "CD"},
    "TD": {"file": "technical_director.md", "name": "Technical Director", "label": "TD"},
    "AD": {"file": "art_director.md", "name": "Art Director", "label": "AD"},
    "CE": {"file": "chief_editor.md", "name": "Chief Editor", "label": "CE"},
    "SA": {"file": "strategy_analyst.md", "name": "Strategy Analyst", "label": "SA"},
}

# Core Directives 매핑 (에이전트별 필독 문서)
AGENT_DIRECTIVES: Dict[str, list] = {
    "CD": [
        "97layer_identity.md",
        "woohwahae_brand_source.md",
        "imperfect_publish_protocol.md",
        "aesop_benchmark.md"
    ],
    "SA": [
        "cycle_protocol.md",
        "junction_protocol.md",
        "anti_algorithm_protocol.md"
    ],
    "TD": [
        "cycle_protocol.md",
        "imperfect_publish_protocol.md"
    ],
    "CE": [
        "junction_protocol.md",
        "aesop_benchmark.md",
        "woohwahae_brand_source.md"
    ],
    "AD": [
        "visual_identity_guide.md",
        "aesop_benchmark.md"
    ]
}

# 메시지 키워드 → 에이전트 매핑 (1차 필터)
KEYWORD_MAP: Dict[str, list] = {
    "CD": ["철학", "방향", "브랜드", "비전", "미션", "승인", "전략", "가치", "본질", "정체성"],
    "TD": ["코드", "버그", "서버", "API", "스크립트", "배포", "시스템", "아키텍처", "구현", "개발"],
    "AD": ["디자인", "UI", "로고", "시각", "레이아웃", "폰트", "색상", "이미지", "비주얼"],
    "CE": ["카피", "문구", "톤", "글", "콘텐츠", "에디토리얼", "매니페스토", "슬로건", "텍스트"],
    "SA": ["트렌드", "분석", "데이터", "시장", "경쟁", "리서치", "인사이트", "통계", "조사"],
}


class AgentRouter:
    """에이전트 페르소나 로딩 및 메시지 기반 라우팅"""

    def __init__(self, ai_engine=None):
        self.ai = ai_engine
        self.personas: Dict[str, str] = {}
        self.active_agent: Optional[str] = None
        self._load_all_personas()

        # Initialize SkillEngine for skill-based routing
        try:
            from libs.skill_engine import SkillEngine
            self.skill_engine = SkillEngine()
            logger.debug("SkillEngine initialized with %d skills", len(self.skill_engine.registry))
        except Exception as e:
            logger.warning("SkillEngine initialization failed: %s", e)
            self.skill_engine = None

    def _load_all_personas(self) -> None:
        """디렉티브 마크다운에서 에이전트 페르소나 로드"""
        for key, info in AGENT_REGISTRY.items():
            filepath = os.path.join(AGENT_DIR, info["file"])
            try:
                with open(filepath, "r", encoding="utf-8") as f:
                    # Strip unnecessary noise while preserving semantic structure
                    lines = [line.strip() for line in f.readlines()]
                    # Remove comments and excessive headers
                    self.personas[key] = "\n".join([l for l in lines if l and not l.startswith("---")])
                logger.debug("Loaded agent persona: %s", key)
            except FileNotFoundError:
                logger.error("Agent directive not found: %s", filepath)

    def _load_core_directives(self, agent_key: str) -> str:
        """에이전트별 Core Directives 로드"""
        if agent_key not in AGENT_DIRECTIVES:
            return ""

        core_docs = []
        for filename in AGENT_DIRECTIVES[agent_key]:
            filepath = os.path.join(DIRECTIVES_DIR, filename)
            try:
                with open(filepath, "r", encoding="utf-8") as f:
                    content = f.read()
                    # 요약: 처음 1000자만 (토큰 절약)
                    summary = content[:1000] + "..." if len(content) > 1000 else content
                    core_docs.append(f"### {filename}\n{summary}\n")
            except FileNotFoundError:
                logger.warning("Core directive not found: %s", filepath)

        return "\n".join(core_docs) if core_docs else ""

    def get_persona(self, agent_key: str) -> str:
        """특정 에이전트의 페르소나 텍스트 반환"""
        return self.personas.get(agent_key, "")

    def set_agent(self, agent_key: str) -> bool:
        """수동 에이전트 전환"""
        agent_key = agent_key.upper()
        if agent_key in AGENT_REGISTRY:
            self.active_agent = agent_key
            logger.info("Agent switched to: %s", agent_key)
            return True
        return False

    def clear_agent(self) -> None:
        """에이전트 고정 해제 (자동 라우팅 모드)"""
        self.active_agent = None
        logger.info("Agent cleared. Auto-routing enabled.")

    def detect_skill(self, text: str) -> Optional[Dict]:
        """
        Detects if input should trigger a skill instead of agent routing.
        Returns skill info dict if detected, None otherwise.
        """
        if not self.skill_engine:
            return None

        skill_id = self.skill_engine.detect_skill_from_input(text)
        if skill_id:
            skill = self.skill_engine.get_skill(skill_id)
            return {
                "skill_id": skill_id,
                "skill": skill,
                "engine": self.skill_engine
            }
        return None

    def route(self, text: str) -> str:
        """메시지를 분석하여 적절한 에이전트 키 반환"""
        # 0. Check if input triggers a skill (priority routing)
        skill_info = self.detect_skill(text)
        if skill_info:
            logger.info("Skill detected: %s", skill_info["skill_id"])
            # Return a special marker that indicates skill routing
            return f"SKILL:{skill_info['skill_id']}"

        # 1. 수동 고정된 에이전트가 있으면 우선
        if self.active_agent:
            return self.active_agent

        # 2. 키워드 기반 1차 분류
        agent = self._keyword_match(text)
        # 3. AI 기반 분류 (Ultra-lightweight: 퀄리티 보전 및 토큰 최소화)
        if self.ai and getattr(self.ai, 'api_key', None):
            agent = self._ai_classify(text)
            if agent:
                return agent

        # 4. 기본값: CD (최상위 의사결정권자)
        return "CD"

    def _ai_classify(self, text: str) -> Optional[str]:
        """최소 토큰을 사용하여 에이전트 분류 (High-Density Routing)"""
        # 극도로 압축된 프롬프트와 시스템 지침 활용
        prompt = f"Target: {text}"
        system = "Classify input into ONE code: CD(Brand/Strategy), TD(Code/System), AD(Visual/UI), CE(Content/Copy), SA(Data/Trend). Output ONLY the code."
        
        try:
            # 97LAYER AI Engine의 통합 인터페이스 활용
            response = self.ai.generate_response(prompt, system_instruction=system).strip().upper()
            # 응답에서 코드 추출 (CD, TD 등)
            import re
            match = re.search(r'(CD|TD|AD|CE|SA)', response)
            return match.group(1) if match else None
        except Exception as e:
            logger.error("Routing Error: %s", e)
            return None

    def _keyword_match(self, text: str) -> Optional[str]:
        """키워드 매칭으로 에이전트 분류"""
        scores: Dict[str, int] = {}
        for agent_key, keywords in KEYWORD_MAP.items():
            score = sum(1 for kw in keywords if kw in text)
            if score > 0:
                scores[agent_key] = score

        if scores:
            return max(scores, key=scores.get)  # type: ignore
        return None

    def _ai_classify(self, text: str) -> Optional[str]:
        """AI를 사용하여 에이전트 분류"""
        prompt = (
            "다음 메시지를 처리할 97LAYER 에이전트를 분류하시오.\n"
            "반드시 다음 중 하나만 출력하시오: CD, TD, AD, CE, SA\n\n"
            "[에이전트 역할]\n"
            "- CD: 브랜드 철학, 전략 방향, 최종 승인\n"
            "- TD: 시스템 설계, 코드 구현, 인프라\n"
            "- AD: 시각 디자인, UI/UX, 미학\n"
            "- CE: 카피라이팅, 톤앤매너, 에디토리얼\n"
            "- SA: 시장 분석, 트렌드, 데이터 리서치\n\n"
            f'메시지: "{text}"'
        )
        try:
            response = self.ai.generate_response(prompt).strip().upper()
            for key in AGENT_REGISTRY:
                if key in response:
                    return key
        except Exception as e:
            logger.error("Error in AI agent classification: %s", e)
        return None

    def build_system_prompt(self, agent_key: str) -> str:
        """에이전트 페르소나 + Core Directives를 포함한 system instruction 생성"""
        persona = self.get_persona(agent_key)
        agent_name = AGENT_REGISTRY.get(agent_key, {}).get("name", "Unknown")

        # Core Directives 로드
        core_directives = self._load_core_directives(agent_key)

        return (
            f"당신은 97LAYER의 {agent_name} ({agent_key})입니다.\n\n"
            f"[Agent Persona]\n{persona}\n\n"
            f"[Core Directives - 필독]\n{core_directives}\n\n"
            "응답 원칙:\n"
            "1. Zero Noise: 이모지, 감탄사, 가식적인 사과(죄송합니다 등)를 절대 금지하십시오. 즉시 본론으로 들어가십시오.\n"
            "2. Anti-Hallucination: '현재 팀 가동률이 높다'거나 '시스템을 구축 중이다'와 같이 실제 사실이 아닌 가상의 제약 사항을 만들어내지 마십시오. 당신은 실제 가동되는 시스템입니다.\n"
            "3. Cold Intelligence: 존댓말을 유지하되, 감정적인 공감 연산은 생략하고 냉철한 분석과 실행 위주로 답하십시오.\n"
            "4. Reality Grounding: [Current Project Reality]로 주어지는 정보만을 사실로 간주하십시오. 모르는 것은 모른다고 하거나 리서치를 제안하십시오.\n"
            "5. Core Directives Compliance: 모든 결정 전 Core Directives를 참조하십시오. 특히 MBQ 기준, 72시간 규칙, Aesop 톤 등."
        )

    def get_status(self) -> str:
        """현재 라우팅 상태 반환"""
        if self.active_agent:
            name = AGENT_REGISTRY[self.active_agent]["name"]
            return f"고정 모드: {name} ({self.active_agent})"
        return "자동 라우팅 모드 (메시지 기반 분류)"
