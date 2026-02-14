#!/usr/bin/env python3
import json
import logging
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Any, Optional

class MemoryEngine:
    """97LAYER OS - 장기 기억 및 지식 인덱싱 엔진"""
    
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.memory_file = self.project_root / "knowledge" / "long_term_memory.json"
        self.assets_dir = self.project_root / "knowledge" / "assets" / "ready_to_publish"
        self.logger = logging.getLogger("MemoryEngine")
        
        if not self.memory_file.exists():
            self._init_memory()

    def _init_memory(self):
        """기억 저장소 초기화"""
        self.memory_file.parent.mkdir(parents=True, exist_ok=True)
        initial_data = {
            "metadata": {
                "created_at": datetime.now().isoformat(),
                "last_indexed": datetime.now().isoformat(),
                "total_entries": 0
            },
            "experiences": [], # 성공/실패 사례
            "concepts": {},    # 브랜드 철학적 키워드 인덱스
            "error_patterns": [] # 자가 수정을 위한 오류 로그
        }
        self.memory_file.write_text(json.dumps(initial_data, indent=4, ensure_ascii=False))

    def archive_experience(self, signal_id: str, content: str, judgment: Dict, status: str = "success"):
        """새로운 작업 경험을 기억에 추가 (자가 순환의 기초)"""
        try:
            memory = json.loads(self.memory_file.read_text())
            entry = {
                "signal_id": signal_id,
                "timestamp": datetime.now().isoformat(),
                "content_preview": content[:200],
                "judgment": judgment,
                "status": status
            }
            memory["experiences"].append(entry)
            memory["metadata"]["total_entries"] += 1
            memory["metadata"]["last_indexed"] = datetime.now().isoformat()
            
            # 용량 관리: 최신 500개 유지
            if len(memory["experiences"]) > 500:
                memory["experiences"] = memory["experiences"][-500:]
                
            self.memory_file.write_text(json.dumps(memory, indent=4, ensure_ascii=False))
        except Exception as e:
            self.logger.error(f"Failed to archive experience: {e}")

    def retrieve_relevant(self, query: str, limit: int = 3) -> List[Dict]:
        """현재 요청과 유사한 과거 기억 검색 (에이전트 참조용)"""
        try:
            memory = json.loads(self.memory_file.read_text())
            # 단순 키워드 매칭 (추후 임베딩 기반 검색으로 확장 가능)
            relevant = []
            query_words = set(query.lower().split())
            
            for exp in reversed(memory["experiences"]):
                content = exp.get("content_preview", "").lower()
                if any(word in content for word in query_words):
                    relevant.append(exp)
                if len(relevant) >= limit:
                    break
            return relevant
        except:
            return []

    def log_error(self, error_msg: str, context: Dict):
        """시스템 오류를 장기 기억에 기록 (Self-Healing 기초 데이터)"""
        try:
            memory = json.loads(self.memory_file.read_text())
            error_entry = {
                "timestamp": datetime.now().isoformat(),
                "error": error_msg,
                "context": context,
                "resolved": False
            }
            memory["error_patterns"].append(error_entry)
            self.memory_file.write_text(json.dumps(memory, indent=4, ensure_ascii=False))
        except:
            pass
