import os
import json
import time
import urllib.request
import urllib.error
import logging
from pathlib import Path
from typing import Optional, Dict, Any, List
import anthropic  # Added Claude support

logger = logging.getLogger(__name__)

# Token Optimization Integration
try:
    from execution.system.token_optimizer import TokenOptimizer
    TOKEN_OPTIMIZER_AVAILABLE = True
except ImportError:
    TOKEN_OPTIMIZER_AVAILABLE = False
    logger.debug("TokenOptimizer not available, using basic caching only")

class AIEngine:
    def __init__(self, config: Optional[Dict[str, Any]] = None, system_instruction: Optional[str] = None):
        self.config = config or {}
        self.system_instruction = system_instruction
        self.model_name = self.config.get('model_name', "gemini-2.0-flash")

        # Claude API key loading
        self.claude_api_key = None

        # [Efficiency Protocol] Setup Cache Directory
        project_root = Path("/Users/97layer/97layerOS")
        self.cache_dir = project_root / ".tmp" / "ai_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)

        # [Token Optimization] Initialize optimizer if available
        if TOKEN_OPTIMIZER_AVAILABLE:
            self.optimizer = TokenOptimizer(str(project_root))
            logger.debug("TokenOptimizer integrated successfully")
        else:
            self.optimizer = None
        
        # Search for GEMINI_API_KEY
        env_candidates = [
            project_root / ".env",
            project_root / ".env.txt"
        ]
        for env_path in env_candidates:
            if self._try_load_env(env_path):
                break

        self.api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
        self.claude_api_key = os.getenv("ANTHROPIC_API_KEY")
        if self.claude_api_key:
            logger.debug("Anthropic API key loaded for Creative Director")

        # 플레이스홀더 값 감지 및 제거
        placeholder_values = {"your_actual_api_key_here", "your_gemini_api_key_here", ""}
        if self.api_key in placeholder_values:
            self.api_key = None

        # Load Claude API key
        self.claude_api_key = os.getenv("ANTHROPIC_API_KEY") or os.getenv("CLAUDE_API_KEY")
        if self.claude_api_key in placeholder_values or not self.claude_api_key:
            self.claude_api_key = None

        self._load_master_directives()

    def _load_master_directives(self):
        """마스터 지관서(IDENTITY, SYSTEM)를 지능 엔진의 기본 컨텍스트로 로드"""
        project_root = Path(__file__).resolve().parent.parent.parent
        identity_path = project_root / "directives" / "IDENTITY.md"
        system_path = project_root / "directives" / "system" / "SYSTEM.md"

        self.identity_context = identity_path.read_text(encoding="utf-8") if identity_path.exists() else ""
        self.system_context = system_path.read_text(encoding="utf-8") if system_path.exists() else ""

    def _try_load_env(self, env_path: Path) -> bool:
        """Standard library fallback for loading .env. Returns True if valid key was found."""
        placeholder_values = {"your_actual_api_key_here", "your_gemini_api_key_here", ""}
        try:
            if env_path.exists():
                with open(env_path, "r", encoding="utf-8") as f:
                    for line in f:
                        if "=" in line and not line.startswith("#"):
                            k, v = line.strip().split("=", 1)
                            v = v.strip().strip('"').strip("'")
                            if v and v not in placeholder_values:
                                os.environ[k] = v

            key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
            if key and key not in placeholder_values:
                self.api_key = key
                return True
            return False
        except Exception as e:
            logger.debug(f"Failed to manually load .env: {e}")
            return False

    def get_cache_key(self, text: str) -> str:
        """Generate a stable MD5 hash for caching"""
        import hashlib
        return hashlib.md5(text.encode('utf-8')).hexdigest()

    def get_cached_response(self, cache_file: Path, ttl_hours: int = 24) -> Optional[str]:
        """Retrieve cached response if within TTL"""
        if cache_file.exists():
            try:
                with open(cache_file, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    if time.time() - data.get("timestamp", 0) < ttl_hours * 3600:
                        logger.info(f"✓ Cache Hit: {cache_file.name}")
                        return data.get("response")
            except Exception as e:
                logger.debug(f"Cache read error: {e}")
        return None

    def cache_response(self, cache_file: Path, response: str):
        """Save response to cache"""
        try:
            with open(cache_file, "w", encoding="utf-8") as f:
                json.dump({
                    "timestamp": time.time(),
                    "response": response
                }, f, ensure_ascii=False, indent=4)
        except Exception as e:
            logger.debug(f"Cache write error: {e}")

    def generate_response(self, prompt: str, system_instruction: Optional[str] = None) -> str:
        """에이전트 라우터 및 데몬에서 사용하는 통합 인터페이스 (urllib 전용)"""
        if system_instruction:
            self.system_instruction = system_instruction
        return self.generate_thought(prompt)

    def generate_thought(self, prompt: str) -> str:
        if not self.api_key:
            return "Error: [AIEngine] Gemini API Key Missing. Please set GEMINI_API_KEY in .env."

        # [Token Optimization] Use advanced optimizer if available
        if self.optimizer:
            full_prompt = str(self.system_instruction) + prompt
            cached = self.optimizer.get_cached_response(full_prompt, max_age_hours=24)
            if cached:
                logger.info(f"✓ Token Optimizer Cache Hit")
                return cached

        # [Efficiency Protocol] Fallback to basic cache check
        import hashlib
        prompt_hash = hashlib.md5((str(self.system_instruction) + prompt).encode()).hexdigest()
        cache_file = self.cache_dir / f"{prompt_hash}.json"

        if cache_file.exists() and not self.optimizer:
            try:
                with open(cache_file, "r", encoding="utf-8") as f:
                    cached_data = json.load(f)
                    logger.info(f"Efficiency Hit: [Cache {prompt_hash[:6]}]")
                    return cached_data["response"]
            except: pass
        
        # Google Generative AI API URL
        # Standardize on v1beta for gemini-2.0-flash stability
        url = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model_name}:generateContent?key={self.api_key}"
        
        payload = {
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {
                "temperature": self.config.get("temperature", 0.7),
                "topP": self.config.get("top_p", 0.95)
            }
        }
        if self.system_instruction:
            payload["system_instruction"] = {"parts": [{"text": self.system_instruction}]}

        headers = {"Content-Type": "application/json"}
        req = urllib.request.Request(url, data=json.dumps(payload).encode("utf-8"), headers=headers)
        
        max_retries = 4
        retry_delays = [15, 30, 60, 120]

        for attempt in range(max_retries + 1):
            try:
                with urllib.request.urlopen(req) as response:
                    res_data = json.loads(response.read().decode())
                    candidates = res_data.get("candidates", [])
                    if candidates:
                        text = candidates[0].get("content", {}).get("parts", [{}])[0].get("text", "")
                        refined_text = self._refine(text)

                        # [Token Optimization] Save with advanced optimizer
                        if self.optimizer:
                            full_prompt = str(self.system_instruction) + prompt
                            self.optimizer.cache_response(
                                full_prompt,
                                refined_text,
                                metadata={'model': self.model_name, 'context': 'ai_engine'}
                            )
                        else:
                            # [Efficiency Protocol] Fallback to basic cache
                            try:
                                with open(cache_file, "w", encoding="utf-8") as f:
                                    json.dump({
                                        "timestamp": time.time(),
                                        "prompt_hash": prompt_hash,
                                        "response": refined_text
                                    }, f, ensure_ascii=False, indent=4)
                            except: pass

                        return refined_text
                    return "Error: Empty response from Gemini API."
            except urllib.error.HTTPError as e:
                if e.code == 400 and self.system_instruction and attempt == 0:
                    logger.warning("HTTP 400 with system_instruction. Retrying without it...")
                    original_instruction = self.system_instruction
                    self.system_instruction = None
                    result = self.generate_thought(prompt)
                    self.system_instruction = original_instruction
                    return result
                
                if e.code == 429 and attempt < max_retries:
                    sleep_time = retry_delays[attempt]
                    logger.warning(f"Rate limited (429). Waiting {sleep_time}s before retry {attempt+1}/{max_retries}...")
                    time.sleep(sleep_time)
                    continue
                elif e.code in [500, 502, 503, 504] and attempt < max_retries:
                    sleep_time = 5 * (attempt + 1)
                    logger.warning(f"Server error ({e.code}). Retrying in {sleep_time}s...")
                    time.sleep(sleep_time)
                    continue
                logger.error(f"Gemini API HTTP Error: {e}")
                return f"Error: HTTP {e.code} - {e.reason}"
            except Exception as e:
                if attempt < max_retries and "reset by peer" in str(e):
                    sleep_time = 5 * (attempt + 1)
                    logger.warning(f"Connection reset. Retrying in {sleep_time}s...")
                    time.sleep(sleep_time)
                    continue
                logger.error(f"Gemini API Error: {e}")
                return f"Error: {str(e)}"

    def _refine(self, text: str) -> str:
        # 97LAYER Zero Noise 원칙: 가식적인 서론 및 금지어 제거
        # 문장 단위로 필터링하여 문맥 파손 방지
        forbidden_phrases = [
            "죄송합니다", "반갑습니다", "안녕하세요", "도움을 드리고 싶습니다", "불편을 드려",
            "Artisan에게 확인", "확인 후 보고", "복구 예정 시간", "네트워크 불안정"
        ]
        
        lines = text.split('\n')
        refined_lines = []
        for line in lines:
            # Strip lines that start with or primarily consist of forbidden fluff/hallucinations
            stripped_line = line.strip()
            if any(stripped_line.startswith(phrase) for phrase in forbidden_phrases) or \
               (any(phrase in stripped_line for phrase in forbidden_phrases) and len(stripped_line) < 100):
                continue
            
            clean_line = line.replace("😊", "").replace("🚀", "").replace("🙏", "")
            refined_lines.append(clean_line)
            
        return "\n".join(refined_lines).strip()

    def generate_claude_response(self, prompt: str, system_instruction: Optional[str] = None) -> str:
        """Generate response using Claude Opus for sovereign decisions"""
        if not self.claude_api_key:
            logger.warning("Claude API key not configured, falling back to Gemini")
            return self.generate_response(prompt, system_instruction)

        # Use cache with claude_ prefix
        cache_key = self.get_cache_key(prompt)
        cache_file = self.cache_dir / f"claude_{cache_key}.json"

        # Check cache
        cached_response = self.get_cached_response(cache_file, ttl_hours=24)
        if cached_response:
            return cached_response

        try:
            # Lazy import to avoid errors when Claude not needed
            import anthropic

            client = anthropic.Anthropic(api_key=self.claude_api_key)

            # Prepare messages
            messages = []
            if system_instruction:
                messages.append({
                    "role": "system",
                    "content": system_instruction or self.system_instruction
                })
            messages.append({"role": "user", "content": prompt})

            # Call Claude API
            response = client.messages.create(
                model="claude-3-opus-20240229",  # Verified Opus for sovereign authority
                max_tokens=2048,
                temperature=0.3,  # Lower for consistent judgment
                messages=messages
            )

            # Extract text
            result = response.content[0].text

            # Apply refinement
            refined = self._refine(result)

            # Cache the response
            self.cache_response(cache_file, refined)

            return refined

        except Exception as e:
            logger.error(f"Claude API error: {e}, falling back to Gemini")
            return self.generate_response(prompt, system_instruction)

    def generate_multimodal(self, prompt: str, image_data: bytes, system_instruction: Optional[str] = None) -> str:
        """Generate response with image analysis using Gemini Vision"""
        if not self.api_key:
            return "Gemini API key is not configured. Cannot analyze images."

        # Cache based on prompt + image hash
        import hashlib
        image_hash = hashlib.md5(image_data).hexdigest()[:8]
        cache_key = self.get_cache_key(f"{prompt}_{image_hash}")
        cache_file = self.cache_dir / f"vision_{cache_key}.json"

        # Check cache
        cached_response = self.get_cached_response(cache_file, ttl_hours=24)
        if cached_response:
            return cached_response

        import base64
        image_base64 = base64.b64encode(image_data).decode('utf-8')

        # Build request with inline_data format
        url = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model_name}:generateContent"

        data = {
            "contents": [{
                "parts": [
                    {"text": prompt},
                    {
                        "inline_data": {
                            "mime_type": "image/jpeg",
                            "data": image_base64
                        }
                    }
                ]
            }],
            "generationConfig": {
                "temperature": 0.7,
                "maxOutputTokens": 2048
            }
        }

        if system_instruction or self.system_instruction:
            data["system_instruction"] = {
                "parts": [{"text": system_instruction or self.system_instruction}]
            }

        headers = {"Content-Type": "application/json"}
        url_with_key = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model_name}:generateContent?key={self.api_key}"

        try:
            req = urllib.request.Request(
                url_with_key,
                data=json.dumps(data).encode('utf-8'),
                headers=headers
            )

            with urllib.request.urlopen(req) as response:
                result = json.loads(response.read().decode('utf-8'))

            text = result['candidates'][0]['content']['parts'][0]['text']
            refined = self._refine(text)

            # Cache the response
            self.cache_response(cache_file, refined)

            return refined

        except Exception as e:
            logger.error(f"Gemini Vision API error: {e}")
            return f"Error analyzing image: {str(e)}"

    async def generate_with_role(self, role: str, prompt: str, image_data: Optional[bytes] = None) -> str:
        """에이전트 역할(CD, CE, SA, AD, TD)에 따른 최적화된 지능 모델 및 프롬프트 라우팅"""
        
        if role == "CD":
            # CD (Creative Director) - Claude Opus + IDENTITY.md Grounding
            system_instruction = f"당신은 97layerOS의 크리에이티브 디렉터(Creative Director)입니다. 아래의 브랜드 정체성과 철학을 절대 원칙으로 삼아 의사결정하고 답변하십시오.\n\n[IDENTITY DIRECTIVE]\n{self.identity_context}"
            return self.generate_claude_response(prompt, system_instruction=system_instruction)
        
        else:
            # Worker Agents (CE, SA, AD, TD) - Gemini + SYSTEM.md Grounding
            agent_desc = {
                "CE": "편집장 (Chief Editor - 콘텐츠 서사 및 텍스트 정제)",
                "SA": "전략분석가 (Strategy Analyst - 데이터 분석 및 패턴 인식)",
                "AD": "아트 디렉터 (Art Director - 시각적 미학 및 레이아웃)",
                "TD": "기술 이사 (Technical Director - 시스템 자동화 및 기술 구현)"
            }.get(role, "전문직 에이전트")
            
            system_instruction = f"당신은 97layerOS의 {agent_desc}입니다. 아래 운영 프로토콜을 준수하며 임무를 수행하십시오.\n\n[SYSTEM PROTOCOL]\n{self.system_context}"
            
            if image_data:
                return self.generate_multimodal(prompt, image_data, system_instruction=system_instruction)
            else:
                return self.generate_response(prompt, system_instruction=system_instruction)

    async def council_debate(self, topic: str, agents: List[str] = ["CD", "SA", "TD"]) -> str:
        """선택된 에이전트들 간의 가상 토론 시연 및 결론 도출"""
        logger.info(f"Starting Council Debate on: {topic}")
        debate_history = []
        
        for agent in agents:
            context = f"현재 토론 주제: {topic}\n과거 발언들:\n" + "\n".join(debate_history)
            prompt = f"{context}\n\n당신의 페르소나를 유지하며 의견을 제시하십시오."
            opinion = await self.generate_with_role(agent, prompt)
            debate_history.append(f"[{agent}]: {opinion}")
        
        # CD가 최종 결론 요약
        summary_prompt = "위의 토론 내용을 요약하고 시스템 개선을 위한 최종 결정을 내리십시오.\n\n" + "\n".join(debate_history)
        final_decision = await self.generate_with_role("CD", summary_prompt)
        
        return "\n".join(debate_history) + f"\n\n[FINAL DECISION]\n{final_decision}"

    def generate(self, prompt: str, system_instruction: Optional[str] = None,
                model_type: str = "gemini", image_data: Optional[bytes] = None,
                role: Optional[str] = None) -> str:
        """
        Unified entry point for all AI generation.
        role이 지정된 경우 generate_with_role을 우선 사용함.
        """
        if role:
            import asyncio
            try:
                loop = asyncio.get_running_loop()
                # If already in a loop, we can't use run_until_complete
                # This part is tricky. In the autonomous_loop, we should use await instead.
                # Here we added a safeguard for synchronous contexts.
                if loop.is_running():
                    # This is problematic for sync callers in async context.
                    # We'll assume the caller should have used await generate_with_role.
                    logger.warning("Sync generate() called from within a running event loop for role-based task. Potential blockage.")
                    # Fallback to direct synchronous generation if possible or raise error
            except RuntimeError:
                loop = asyncio.new_event_loop()
                asyncio.set_event_loop(loop)
                return loop.run_until_complete(self.generate_with_role(role, prompt, image_data))

        # Legacy direct routing (Sync)
        if image_data:
            return self.generate_multimodal(prompt, image_data, system_instruction)
        if model_type == "claude":
            return self.generate_claude_response(prompt, system_instruction)
        return self.generate_response(prompt, system_instruction)
