#!/usr/bin/env python3
import os
from pathlib import Path
from typing import Optional
import logging

try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent

class KnowledgeRAG:
    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')
        self.knowledge_dir = PROJECT_ROOT / 'knowledge' / 'docs'
        
        if GEMINI_AVAILABLE and self.api_key:
            genai.configure(api_key=self.api_key)
            self.model = genai.GenerativeModel('gemini-2.0-flash-exp')
            logger.info("✅ Knowledge RAG initialized")
        else:
            self.model = None
            logger.warning("⚠️  Gemini not available")
        
        self.knowledge_base = self._load_knowledge_base()
    
    def _load_knowledge_base(self) -> str:
        documents = []
        if self.knowledge_dir.exists():
            for doc_file in sorted(self.knowledge_dir.glob('*.md')):
                try:
                    content = doc_file.read_text(encoding='utf-8')
                    documents.append(f"## {doc_file.name}\n\n{content}\n\n")
                    logger.info(f"Loaded: {doc_file.name}")
                except Exception as e:
                    logger.warning(f"Failed: {doc_file.name}: {e}")
        
        combined = "\n---\n\n".join(documents)
        logger.info(f"Knowledge base: {len(documents)} docs, {len(combined)} chars")
        return combined
    
    def query(self, question: str, max_tokens: int = 2048) -> str:
        if not self.model:
            return "❌ Gemini API 설정 필요"
        
        if not self.knowledge_base:
            return "❌ Knowledge base 비어있음"
        
        logger.info(f"RAG query: {question[:50]}...")
        
        prompt = f"""당신은 97layerOS 지식 기반 AI입니다.

Knowledge Base:
{self.knowledge_base[:50000]}

질문: {question}

답변 (간결하게 2-3문단):
"""
        
        try:
            response = self.model.generate_content(prompt)
            return response.text
        except Exception as e:
            logger.error(f"RAG failed: {e}")
            return f"❌ RAG 실패: {str(e)}"

def get_knowledge_rag():
    if not hasattr(get_knowledge_rag, 'instance'):
        get_knowledge_rag.instance = KnowledgeRAG()
    return get_knowledge_rag.instance
