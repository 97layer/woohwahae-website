import os
import json
import time
import urllib.request
import urllib.error
import logging
from pathlib import Path
from typing import Optional, Dict, Any

logger = logging.getLogger(__name__)

class AIEngine:
    def __init__(self, config: Optional[Dict[str, Any]] = None, system_instruction: Optional[str] = None):
        self.config = config or {}
        self.system_instruction = system_instruction
        self.model_name = self.config.get('model_name', "gemini-1.5-pro-latest")
        
        # [Efficiency Protocol] Setup Cache Directory
        project_root = Path(__file__).resolve().parent.parent
        self.cache_dir = project_root / ".tmp" / "ai_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        
        # Search for GEMINI_API_KEY or GOOGLE_API_KEY
        # .env 파일을 환경변수보다 우선 로드 (쉘 환경에 플레이스홀더가 설정된 경우 대비)
        project_root = Path(__file__).resolve().parent.parent
        env_candidates = [
            project_root / ".env",
            project_root / ".env.txt",
            project_root / "env.txt"
        ]
        for env_path in env_candidates:
            if self._try_load_env(env_path):
                break

        self.api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")

        # 플레이스홀더 값 감지 및 제거
        placeholder_values = {"your_actual_api_key_here", "your_gemini_api_key_here", ""}
        if self.api_key in placeholder_values:
            self.api_key = None

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

    def generate_response(self, prompt: str, system_instruction: Optional[str] = None) -> str:
        """에이전트 라우터 및 데몬에서 사용하는 통합 인터페이스 (urllib 전용)"""
        if system_instruction:
            self.system_instruction = system_instruction
        return self.generate_thought(prompt)

    def generate_thought(self, prompt: str) -> str:
        if not self.api_key:
            return "Error: [AIEngine] Gemini API Key Missing. Please set GEMINI_API_KEY in .env."

        # [Efficiency Protocol] Cache Check
        import hashlib
        prompt_hash = hashlib.md5((str(self.system_instruction) + prompt).encode()).hexdigest()
        cache_file = self.cache_dir / f"{prompt_hash}.json"
        
        if cache_file.exists():
            try:
                with open(cache_file, "r", encoding="utf-8") as f:
                    cached_data = json.load(f)
                    logger.info(f"Efficiency Hit: [Cache {prompt_hash[:6]}]")
                    return cached_data["response"]
            except: pass
        
        # Google Generative AI API URL
        # Ensure model_id starts with 'models/' if not already present
        model_id = self.model_name
        if not model_id.startswith("models/"):
            model_id = f"models/{model_id}"
            
        url = f"https://generativelanguage.googleapis.com/v1beta/{model_id}:generateContent?key={self.api_key}"
        
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
                        
                        # [Efficiency Protocol] Save to Cache
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
