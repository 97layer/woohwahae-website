import hashlib
import os
import json
from datetime import datetime
from pathlib import Path

class MemoryManager:
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        # Unified path: knowledge/chat_memory
        self.memory_dir = self.project_root / "knowledge" / "chat_memory"
        self.memory_dir.mkdir(parents=True, exist_ok=True)
        self.knowledge_dir = self.project_root / "knowledge"
        self.cache_dir = self.project_root / ".tmp" / "ai_cache"
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        self.cache = {}

    def save_chat(self, chat_id: str, text: str, role: str = "user"):
        """Persists chat history to the local knowledge base."""
        chat_file = self.memory_dir / f"{chat_id}.json"
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "role": role,
            "content": text
        }
        
        # Load all history for saving, but limit for generation
        history = self._load_all_chat(chat_id)
        history.append(log_entry)
        with open(chat_file, "w", encoding="utf-8") as f:
            json.dump(history, f, indent=4, ensure_ascii=False)

    def _load_all_chat(self, chat_id: str) -> list:
        """Helper to load all chat history from file."""
        chat_file = self.memory_dir / f"{chat_id}.json"
        if not chat_file.exists():
            return []
        try:
            with open(chat_file, "r", encoding="utf-8") as f:
                return json.load(f)
        except:
            return []

    def load_chat(self, chat_id: str, limit: int = 20) -> list:
        """Loads the last N messages of chat history. Default increased for better context."""
        history = self._load_all_chat(chat_id)
        return history[-limit:]

    def get_relevant_knowledge(self, query: str, top_n: int = 3) -> list:
        """
        [Efficiency Protocol] 도서관 전체를 읽지 않고 키워드 기반으로 조각(Snippet)만 추출.
        """
        keywords = [k for k in query.split() if len(k) > 1]
        results = []
        
        # knowledge 폴더 내 md, txt, json 파일 스캔
        for root, _, files in os.walk(self.knowledge_dir):
            if "chat_memory" in root or "archive" in root: continue
            
            for file in files:
                if file.endswith(('.md', '.txt')):
                    file_path = Path(root) / file
                    try:
                        with open(file_path, "r", encoding="utf-8") as f:
                            # 전체를 읽지 않고 줄 단위로 키워드 매칭
                            matching_lines = []
                            for i, line in enumerate(f, 1):
                                if any(k.lower() in line.lower() for k in keywords):
                                    matching_lines.append(f"L{i}: {line.strip()}")
                                if len(matching_lines) >= 5: # 파일당 최대 5개 조각
                                    break
                            
                            if matching_lines:
                                results.append({
                                    "file": file,
                                    "snippets": matching_lines
                                })
                    except: continue
                if len(results) >= top_n: break
        return results

    def compress(self, content: str) -> str:
        """중복된 대형 텍스트를 캐싱 처리하여 프로토콜 준수."""
        h = hashlib.md5(content.encode()).hexdigest()
        if h in self.cache:
            return f"[Efficiency: Cached Context Ref {h[:6]}]"
        
        # 2000자 이상인 경우 요약 후보로 등록하고 절단
        if len(content) > 2000:
            self.cache[h] = content
            return content[:1500] + f"\n... [Truncated for Efficiency, Hash: {h[:6]}]"
        return content

    def chronicle_chat(self, chat_id: str):
        """오래된 대화 내역을 하나의 요약(Chronicle)으로 압축."""
        history = self.load_chat(chat_id, limit=50)
        if len(history) < 30: return
        
        # 상위 20개 메시지를 요약 대상으로 선정 (추후 AI 엔진 연동)
        chronicle_path = self.memory_dir / f"{chat_id}_chronicle.json"
        # ... 요약 로직 구현 예정
