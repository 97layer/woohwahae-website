#!/usr/bin/env python3
"""
Model Consistency Manager - AI 모델 간 일관성 유지
Gemini, Claude, GPT 등 어떤 모델을 사용해도 동일한 동작 보장

Features:
- 통합 인터페이스
- 컨텍스트 영속성
- 페르소나 일관성
- 응답 정규화
"""

import json
import os
import hashlib
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional, List
from abc import ABC, abstractmethod
import logging

logger = logging.getLogger(__name__)


class BaseModelAdapter(ABC):
    """모든 AI 모델의 기본 어댑터"""

    @abstractmethod
    def generate_response(self, prompt: str, system_instruction: str = None) -> str:
        """응답 생성 (모든 모델이 구현해야 함)"""
        pass

    @abstractmethod
    def get_model_info(self) -> Dict[str, Any]:
        """모델 정보 반환"""
        pass


class GeminiAdapter(BaseModelAdapter):
    """Gemini 모델 어댑터"""

    def __init__(self, api_key: str, model_name: str = "gemini-1.5-pro-latest"):
        self.api_key = api_key
        self.model_name = model_name
        self.base_url = "https://generativelanguage.googleapis.com/v1beta/models"

    def generate_response(self, prompt: str, system_instruction: str = None) -> str:
        """Gemini API 호출"""
        import urllib.request
        import urllib.error

        url = f"{self.base_url}/{self.model_name}:generateContent?key={self.api_key}"

        # Gemini 형식으로 변환
        request_data = {
            "contents": [{"parts": [{"text": prompt}]}]
        }

        if system_instruction:
            request_data["system_instruction"] = {
                "parts": [{"text": system_instruction}]
            }

        try:
            req = urllib.request.Request(
                url,
                data=json.dumps(request_data).encode('utf-8'),
                headers={'Content-Type': 'application/json'}
            )

            with urllib.request.urlopen(req, timeout=30) as response:
                result = json.loads(response.read().decode('utf-8'))
                return result['candidates'][0]['content']['parts'][0]['text']

        except Exception as e:
            logger.error(f"Gemini API error: {e}")
            return f"Error: {e}"

    def get_model_info(self) -> Dict[str, Any]:
        return {
            "provider": "Google",
            "model": self.model_name,
            "type": "Gemini"
        }


class ClaudeAdapter(BaseModelAdapter):
    """Claude 모델 어댑터"""

    def __init__(self, api_key: str, model_name: str = "claude-3-opus-20240229"):
        self.api_key = api_key
        self.model_name = model_name
        self.base_url = "https://api.anthropic.com/v1"

    def generate_response(self, prompt: str, system_instruction: str = None) -> str:
        """Claude API 호출"""
        import urllib.request
        import urllib.error

        url = f"{self.base_url}/messages"

        # Claude 형식으로 변환
        messages = [{"role": "user", "content": prompt}]

        request_data = {
            "model": self.model_name,
            "messages": messages,
            "max_tokens": 4096
        }

        if system_instruction:
            request_data["system"] = system_instruction

        try:
            req = urllib.request.Request(
                url,
                data=json.dumps(request_data).encode('utf-8'),
                headers={
                    'Content-Type': 'application/json',
                    'x-api-key': self.api_key,
                    'anthropic-version': '2023-06-01'
                }
            )

            with urllib.request.urlopen(req, timeout=30) as response:
                result = json.loads(response.read().decode('utf-8'))
                return result['content'][0]['text']

        except Exception as e:
            logger.error(f"Claude API error: {e}")
            return f"Error: {e}"

    def get_model_info(self) -> Dict[str, Any]:
        return {
            "provider": "Anthropic",
            "model": self.model_name,
            "type": "Claude"
        }


class GPTAdapter(BaseModelAdapter):
    """OpenAI GPT 모델 어댑터"""

    def __init__(self, api_key: str, model_name: str = "gpt-4-turbo-preview"):
        self.api_key = api_key
        self.model_name = model_name
        self.base_url = "https://api.openai.com/v1"

    def generate_response(self, prompt: str, system_instruction: str = None) -> str:
        """OpenAI API 호출"""
        import urllib.request
        import urllib.error

        url = f"{self.base_url}/chat/completions"

        # GPT 형식으로 변환
        messages = []

        if system_instruction:
            messages.append({"role": "system", "content": system_instruction})

        messages.append({"role": "user", "content": prompt})

        request_data = {
            "model": self.model_name,
            "messages": messages,
            "temperature": 0.7
        }

        try:
            req = urllib.request.Request(
                url,
                data=json.dumps(request_data).encode('utf-8'),
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {self.api_key}'
                }
            )

            with urllib.request.urlopen(req, timeout=30) as response:
                result = json.loads(response.read().decode('utf-8'))
                return result['choices'][0]['message']['content']

        except Exception as e:
            logger.error(f"OpenAI API error: {e}")
            return f"Error: {e}"

    def get_model_info(self) -> Dict[str, Any]:
        return {
            "provider": "OpenAI",
            "model": self.model_name,
            "type": "GPT"
        }


class ModelConsistencyManager:
    """모델 일관성 관리자"""

    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.config_file = self.project_root / ".claude" / "model_config.json"
        self.context_dir = self.project_root / "knowledge" / "model_context"
        self.context_dir.mkdir(parents=True, exist_ok=True)

        # 현재 활성 모델
        self.active_model: Optional[BaseModelAdapter] = None
        self.model_type: Optional[str] = None

        # 컨텍스트 캐시
        self.context_cache: Dict[str, Any] = {}

        # 설정 로드
        self.config = self._load_config()

        # 모델 초기화
        self._initialize_model()

    def _load_config(self) -> Dict[str, Any]:
        """설정 파일 로드"""
        default_config = {
            "default_model": "gemini",
            "models": {
                "gemini": {
                    "adapter": "GeminiAdapter",
                    "model_name": "gemini-1.5-pro-latest",
                    "api_key_env": "GEMINI_API_KEY"
                },
                "claude": {
                    "adapter": "ClaudeAdapter",
                    "model_name": "claude-3-opus-20240229",
                    "api_key_env": "CLAUDE_API_KEY"
                },
                "gpt": {
                    "adapter": "GPTAdapter",
                    "model_name": "gpt-4-turbo-preview",
                    "api_key_env": "OPENAI_API_KEY"
                }
            },
            "consistency_rules": {
                "always_korean": True,
                "no_emojis": True,
                "concise": True,
                "technical_focus": True
            }
        }

        if self.config_file.exists():
            try:
                with open(self.config_file) as f:
                    loaded_config = json.load(f)
                    # 기본 설정과 병합
                    default_config.update(loaded_config)
            except Exception as e:
                logger.error(f"Config load error: {e}")

        return default_config

    def _initialize_model(self):
        """모델 초기화"""
        model_type = self.config.get("default_model", "gemini")
        self.switch_model(model_type)

    def switch_model(self, model_type: str) -> bool:
        """모델 전환"""
        if model_type not in self.config["models"]:
            logger.error(f"Unknown model type: {model_type}")
            return False

        model_config = self.config["models"][model_type]

        # API 키 로드
        api_key_env = model_config.get("api_key_env")
        api_key = os.getenv(api_key_env)

        if not api_key:
            # .env 파일에서 로드 시도
            env_file = self.project_root / ".env"
            if env_file.exists():
                with open(env_file) as f:
                    for line in f:
                        if api_key_env in line:
                            api_key = line.split("=")[1].strip()
                            break

        if not api_key:
            logger.error(f"API key not found for {model_type}")
            return False

        # 어댑터 생성
        adapter_class = model_config["adapter"]
        model_name = model_config.get("model_name")

        if adapter_class == "GeminiAdapter":
            self.active_model = GeminiAdapter(api_key, model_name)
        elif adapter_class == "ClaudeAdapter":
            self.active_model = ClaudeAdapter(api_key, model_name)
        elif adapter_class == "GPTAdapter":
            self.active_model = GPTAdapter(api_key, model_name)
        else:
            logger.error(f"Unknown adapter: {adapter_class}")
            return False

        self.model_type = model_type
        logger.info(f"Switched to {model_type} model")

        # 컨텍스트 로드
        self._load_model_context(model_type)

        return True

    def generate_response(self, prompt: str, system_instruction: str = None,
                         agent_key: str = None) -> str:
        """
        통합 응답 생성 (모델 무관)

        Args:
            prompt: 사용자 프롬프트
            system_instruction: 시스템 지시문
            agent_key: 에이전트 식별자

        Returns:
            정규화된 응답
        """
        if not self.active_model:
            return "Error: No active model"

        # 컨텍스트 강화
        enhanced_prompt = self._enhance_prompt(prompt, agent_key)

        # 시스템 지시문 강화
        enhanced_system = self._enhance_system_instruction(system_instruction)

        # 모델 호출
        raw_response = self.active_model.generate_response(
            enhanced_prompt,
            enhanced_system
        )

        # 응답 정규화
        normalized_response = self._normalize_response(raw_response)

        # 컨텍스트 저장
        self._save_interaction(prompt, normalized_response, agent_key)

        return normalized_response

    def _enhance_prompt(self, prompt: str, agent_key: Optional[str]) -> str:
        """프롬프트 강화 (컨텍스트 추가)"""
        enhanced = prompt

        # 에이전트별 컨텍스트 추가
        if agent_key and agent_key in self.context_cache:
            context = self.context_cache[agent_key]
            enhanced = f"[Previous Context]\n{context['summary']}\n\n[Current Input]\n{prompt}"

        # 모델별 특수 처리
        if self.model_type == "gemini":
            # Gemini는 명확한 지시 선호
            enhanced = f"Task: {enhanced}"
        elif self.model_type == "claude":
            # Claude는 대화형 선호
            enhanced = f"Please help with: {enhanced}"

        return enhanced

    def _enhance_system_instruction(self, system_instruction: Optional[str]) -> str:
        """시스템 지시문 강화 (일관성 규칙 적용)"""
        base_instruction = system_instruction or ""

        # 일관성 규칙 추가
        consistency_rules = []

        rules = self.config.get("consistency_rules", {})

        if rules.get("always_korean"):
            consistency_rules.append("Always respond in Korean")

        if rules.get("no_emojis"):
            consistency_rules.append("Do not use emojis unless explicitly requested")

        if rules.get("concise"):
            consistency_rules.append("Be concise and to the point")

        if rules.get("technical_focus"):
            consistency_rules.append("Focus on technical accuracy")

        if consistency_rules:
            rules_text = "\n".join(f"- {rule}" for rule in consistency_rules)
            enhanced = f"{base_instruction}\n\nConsistency Rules:\n{rules_text}"
        else:
            enhanced = base_instruction

        return enhanced

    def _normalize_response(self, response: str) -> str:
        """응답 정규화 (모델 간 차이 제거)"""
        normalized = response

        # 이모지 제거 (설정된 경우)
        if self.config["consistency_rules"].get("no_emojis"):
            import re
            # 이모지 패턴
            emoji_pattern = re.compile("["
                u"\U0001F600-\U0001F64F"  # emoticons
                u"\U0001F300-\U0001F5FF"  # symbols & pictographs
                u"\U0001F680-\U0001F6FF"  # transport & map symbols
                u"\U0001F1E0-\U0001F1FF"  # flags (iOS)
                u"\U00002702-\U000027B0"
                u"\U000024C2-\U0001F251"
                "]+", flags=re.UNICODE)
            normalized = emoji_pattern.sub('', normalized)

        # 과도한 줄바꿈 제거
        normalized = re.sub(r'\n{3,}', '\n\n', normalized)

        # 앞뒤 공백 제거
        normalized = normalized.strip()

        return normalized

    def _save_interaction(self, prompt: str, response: str, agent_key: Optional[str]):
        """상호작용 저장 (컨텍스트 유지)"""
        interaction = {
            "timestamp": datetime.now().isoformat(),
            "model": self.model_type,
            "agent": agent_key,
            "prompt": prompt[:500],  # 요약
            "response": response[:500]  # 요약
        }

        # 에이전트별 컨텍스트 업데이트
        if agent_key:
            if agent_key not in self.context_cache:
                self.context_cache[agent_key] = {
                    "history": [],
                    "summary": ""
                }

            self.context_cache[agent_key]["history"].append(interaction)

            # 최근 5개만 유지
            self.context_cache[agent_key]["history"] = \
                self.context_cache[agent_key]["history"][-5:]

            # 요약 생성
            history = self.context_cache[agent_key]["history"]
            summary_parts = [f"Q: {h['prompt'][:100]}" for h in history[-3:]]
            self.context_cache[agent_key]["summary"] = "\n".join(summary_parts)

        # 파일로 저장
        self._persist_context()

    def _persist_context(self):
        """컨텍스트 영속화"""
        try:
            context_file = self.context_dir / f"context_{self.model_type}.json"

            with open(context_file, 'w') as f:
                json.dump(self.context_cache, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"Context save error: {e}")

    def _load_model_context(self, model_type: str):
        """모델 컨텍스트 로드"""
        try:
            context_file = self.context_dir / f"context_{model_type}.json"

            if context_file.exists():
                with open(context_file) as f:
                    self.context_cache = json.load(f)
                    logger.info(f"Loaded context for {model_type}")

        except Exception as e:
            logger.error(f"Context load error: {e}")
            self.context_cache = {}

    def get_model_info(self) -> Dict[str, Any]:
        """현재 모델 정보"""
        if self.active_model:
            return self.active_model.get_model_info()
        return {"error": "No active model"}

    def list_available_models(self) -> List[str]:
        """사용 가능한 모델 목록"""
        return list(self.config["models"].keys())


# 싱글톤 인스턴스
_consistency_manager: Optional[ModelConsistencyManager] = None


def get_consistency_manager(project_root: str = None) -> ModelConsistencyManager:
    """싱글톤 매니저 반환"""
    global _consistency_manager

    if _consistency_manager is None:
        if project_root is None:
            project_root = str(Path(__file__).resolve().parent.parent)
        _consistency_manager = ModelConsistencyManager(project_root)

    return _consistency_manager