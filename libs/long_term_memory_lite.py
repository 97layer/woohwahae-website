#!/usr/bin/env python3
"""
장기 기억 엔진 (Lite Version - No ML Dependencies)
에이전트들의 경험과 지식을 축적하고 실시간으로 참조

Features:
- 작업 경험 자동 기록 및 카테고리화
- 키워드 기반 유사 경험 검색
- 성공/실패 패턴 분석
- 지식 진화 및 최적화
"""

import json
import os
import hashlib
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any
from collections import defaultdict
import logging

logger = logging.getLogger(__name__)


class LongTermMemoryLite:
    """장기 기억 관리 엔진 (경량 버전)"""

    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.memory_dir = self.project_root / "knowledge" / "long_term_memory"
        self.memory_dir.mkdir(parents=True, exist_ok=True)

        # 메모리 파일 경로
        self.memory_file = self.memory_dir / "long_term_memory.json"
        self.index_file = self.memory_dir / "memory_index.json"
        self.patterns_file = self.memory_dir / "learned_patterns.json"

        # 메모리 로드
        self.memories = self._load_memories()
        self.index = self._load_index()
        self.patterns = self._load_patterns()

        # 키워드 인덱스
        self.keyword_index = self._build_keyword_index()

    def _load_memories(self) -> List[Dict]:
        """장기 기억 로드"""
        if self.memory_file.exists():
            try:
                with open(self.memory_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"메모리 로드 실패: {e}")
                return []
        return []

    def _load_index(self) -> Dict:
        """인덱스 로드"""
        if self.index_file.exists():
            try:
                with open(self.index_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"인덱스 로드 실패: {e}")
                return self._create_default_index()
        return self._create_default_index()

    def _load_patterns(self) -> Dict:
        """학습된 패턴 로드"""
        if self.patterns_file.exists():
            try:
                with open(self.patterns_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"패턴 로드 실패: {e}")
                return self._create_default_patterns()
        return self._create_default_patterns()

    def _create_default_index(self) -> Dict:
        """기본 인덱스 구조"""
        return {
            "by_category": {},
            "by_agent": {},
            "by_success": {"true": [], "false": []},
            "by_date": {},
            "total_count": 0
        }

    def _create_default_patterns(self) -> Dict:
        """기본 패턴 구조"""
        return {
            "success_patterns": [],
            "failure_patterns": [],
            "optimization_rules": [],
            "learned_strategies": []
        }

    def _build_keyword_index(self) -> Dict[str, List[str]]:
        """키워드 인덱스 구축"""
        keyword_index = defaultdict(list)

        for memory in self.memories:
            memory_id = memory.get("id", "")
            text = f"{memory.get('task', '')} {memory.get('context', '')} {memory.get('result', '')}"
            keywords = self._extract_keywords(text)

            for keyword in keywords:
                keyword_index[keyword.lower()].append(memory_id)

        return dict(keyword_index)

    def store_experience(self, experience: Dict[str, Any]) -> str:
        """
        새로운 경험 저장

        Args:
            experience: 경험 데이터

        Returns:
            메모리 ID
        """
        # 메모리 ID 생성
        memory_id = self._generate_memory_id(experience)

        # 타임스탬프 추가
        experience["id"] = memory_id
        experience["timestamp"] = datetime.now().isoformat()
        experience["accessed_count"] = 0
        experience["last_accessed"] = None

        # 중복 체크
        if not self._is_duplicate(experience):
            # 메모리에 추가
            self.memories.append(experience)

            # 인덱스 업데이트
            self._update_index(experience)

            # 패턴 학습
            self._learn_from_experience(experience)

            # 키워드 인덱스 업데이트
            self._update_keyword_index(experience)

            # 저장
            self._save_memories()
            self._save_index()
            self._save_patterns()

            logger.info(f"새로운 경험 저장: {memory_id}")
        else:
            logger.info(f"중복 경험 스킵: {memory_id}")

        return memory_id

    def retrieve_similar_experiences(self, query: str, top_k: int = 5,
                                    category: Optional[str] = None,
                                    success_only: bool = False) -> List[Dict]:
        """
        유사한 경험 검색 (키워드 기반)

        Args:
            query: 검색 쿼리
            top_k: 반환할 경험 수
            category: 카테고리 필터
            success_only: 성공 사례만 검색

        Returns:
            유사한 경험 리스트
        """
        if not self.memories:
            return []

        # 쿼리에서 키워드 추출
        query_keywords = set(self._extract_keywords(query.lower()))

        # 각 메모리의 유사도 계산
        scores = {}

        for memory in self.memories:
            # 필터링
            if category and memory.get("category") != category:
                continue
            if success_only and not memory.get("success", False):
                continue

            # 메모리 텍스트에서 키워드 추출
            memory_text = f"{memory.get('task', '')} {memory.get('context', '')} {memory.get('result', '')}"
            memory_keywords = set(self._extract_keywords(memory_text.lower()))

            # Jaccard 유사도 계산
            if memory_keywords:
                intersection = len(query_keywords & memory_keywords)
                union = len(query_keywords | memory_keywords)
                similarity = intersection / union if union > 0 else 0

                if similarity > 0:
                    scores[memory["id"]] = similarity

        # 상위 K개 선택
        sorted_ids = sorted(scores.keys(), key=lambda x: scores[x], reverse=True)[:top_k]

        results = []
        for memory_id in sorted_ids:
            memory = next((m for m in self.memories if m["id"] == memory_id), None)
            if memory:
                result = memory.copy()
                result["similarity_score"] = scores[memory_id]

                # 접근 횟수 업데이트
                self._update_access_count(memory_id)

                results.append(result)

        return results

    def get_success_patterns(self, category: Optional[str] = None) -> List[Dict]:
        """성공 패턴 반환"""
        patterns = self.patterns.get("success_patterns", [])

        if category:
            patterns = [p for p in patterns if p.get("category") == category]

        return patterns

    def get_failure_patterns(self, category: Optional[str] = None) -> List[Dict]:
        """실패 패턴 반환"""
        patterns = self.patterns.get("failure_patterns", [])

        if category:
            patterns = [p for p in patterns if p.get("category") == category]

        return patterns

    def get_optimization_rules(self) -> List[Dict]:
        """최적화 규칙 반환"""
        return self.patterns.get("optimization_rules", [])

    def consolidate_memories(self, max_age_days: int = 30):
        """오래된 메모리 압축 및 정리"""
        cutoff_date = datetime.now() - timedelta(days=max_age_days)
        old_memories = []
        recent_memories = []

        for memory in self.memories:
            memory_date = datetime.fromisoformat(memory["timestamp"])
            if memory_date < cutoff_date:
                old_memories.append(memory)
            else:
                recent_memories.append(memory)

        if old_memories:
            # 오래된 메모리를 카테고리별로 요약
            summaries = self._summarize_memories(old_memories)

            # 요약을 새로운 압축 메모리로 저장
            for summary in summaries:
                summary["type"] = "consolidated"
                summary["original_count"] = len(old_memories)
                recent_memories.append(summary)

            # 메모리 업데이트
            self.memories = recent_memories
            self._save_memories()

            # 인덱스 재구축
            self._rebuild_index()
            self.keyword_index = self._build_keyword_index()

            logger.info(f"{len(old_memories)}개 메모리를 {len(summaries)}개로 압축")

    def _generate_memory_id(self, experience: Dict) -> str:
        """메모리 ID 생성"""
        content = f"{experience.get('task', '')}{experience.get('agent', '')}{datetime.now().isoformat()}"
        return hashlib.md5(content.encode()).hexdigest()[:16]

    def _is_duplicate(self, experience: Dict) -> bool:
        """중복 체크"""
        for memory in self.memories[-10:]:  # 최근 10개만 체크
            if (memory.get("task") == experience.get("task") and
                memory.get("agent") == experience.get("agent") and
                memory.get("result") == experience.get("result")):
                return True
        return False

    def _update_index(self, experience: Dict):
        """인덱스 업데이트"""
        memory_id = experience["id"]

        # 카테고리별 인덱싱
        category = experience.get("category", "unknown")
        if category not in self.index["by_category"]:
            self.index["by_category"][category] = []
        self.index["by_category"][category].append(memory_id)

        # 에이전트별 인덱싱
        agent = experience.get("agent", "unknown")
        if agent not in self.index["by_agent"]:
            self.index["by_agent"][agent] = []
        self.index["by_agent"][agent].append(memory_id)

        # 성공/실패별 인덱싱
        success = "true" if experience.get("success", False) else "false"
        self.index["by_success"][success].append(memory_id)

        # 날짜별 인덱싱
        date = experience["timestamp"][:10]  # YYYY-MM-DD
        if date not in self.index["by_date"]:
            self.index["by_date"][date] = []
        self.index["by_date"][date].append(memory_id)

        # 총 개수
        self.index["total_count"] += 1

    def _update_keyword_index(self, experience: Dict):
        """키워드 인덱스 업데이트"""
        memory_id = experience["id"]
        text = f"{experience.get('task', '')} {experience.get('context', '')} {experience.get('result', '')}"
        keywords = self._extract_keywords(text)

        for keyword in keywords:
            keyword_lower = keyword.lower()
            if keyword_lower not in self.keyword_index:
                self.keyword_index[keyword_lower] = []
            self.keyword_index[keyword_lower].append(memory_id)

    def _learn_from_experience(self, experience: Dict):
        """경험으로부터 패턴 학습"""
        if experience.get("success", False):
            # 성공 패턴 추출
            pattern = {
                "category": experience.get("category"),
                "context_keywords": self._extract_keywords(experience.get("context", "")),
                "successful_actions": experience.get("actions", []),
                "metrics": experience.get("metrics", {}),
                "confidence": 0.8  # 초기 신뢰도
            }

            # 기존 패턴과 병합 또는 추가
            self._merge_or_add_pattern(self.patterns["success_patterns"], pattern)

        else:
            # 실패 패턴 추출
            pattern = {
                "category": experience.get("category"),
                "context_keywords": self._extract_keywords(experience.get("context", "")),
                "failed_actions": experience.get("actions", []),
                "error_type": experience.get("error_type", "unknown"),
                "avoidance_strategy": experience.get("learnings", [])
            }

            self._merge_or_add_pattern(self.patterns["failure_patterns"], pattern)

        # 최적화 규칙 도출
        if experience.get("learnings"):
            for learning in experience["learnings"]:
                rule = {
                    "category": experience.get("category"),
                    "condition": experience.get("context"),
                    "optimization": learning,
                    "source_memory": experience["id"]
                }
                self.patterns["optimization_rules"].append(rule)

    def _extract_keywords(self, text: str, top_n: int = 10) -> List[str]:
        """텍스트에서 키워드 추출"""
        if not text:
            return []

        # 간단한 키워드 추출
        words = text.lower().split()
        word_freq = defaultdict(int)

        for word in words:
            # 불용어 제거
            if len(word) > 3 and word not in {'this', 'that', 'with', 'from', 'have', 'been', 'were', 'would'}:
                word_freq[word] += 1

        # 빈도순 정렬
        sorted_words = sorted(word_freq.items(), key=lambda x: x[1], reverse=True)
        return [word for word, _ in sorted_words[:top_n]]

    def _merge_or_add_pattern(self, pattern_list: List[Dict], new_pattern: Dict):
        """패턴 병합 또는 추가"""
        # 유사한 패턴 찾기
        for existing_pattern in pattern_list:
            if self._patterns_similar(existing_pattern, new_pattern):
                # 병합
                self._merge_patterns(existing_pattern, new_pattern)
                return

        # 유사한 패턴이 없으면 추가
        pattern_list.append(new_pattern)

    def _patterns_similar(self, pattern1: Dict, pattern2: Dict) -> bool:
        """패턴 유사도 체크"""
        if pattern1.get("category") != pattern2.get("category"):
            return False

        # 키워드 유사도 체크
        keywords1 = set(pattern1.get("context_keywords", []))
        keywords2 = set(pattern2.get("context_keywords", []))

        if not keywords1 or not keywords2:
            return False

        overlap = len(keywords1.intersection(keywords2))
        similarity = overlap / min(len(keywords1), len(keywords2))

        return similarity > 0.6

    def _merge_patterns(self, existing: Dict, new: Dict):
        """패턴 병합"""
        # 키워드 병합
        existing_keywords = set(existing.get("context_keywords", []))
        new_keywords = set(new.get("context_keywords", []))
        existing["context_keywords"] = list(existing_keywords.union(new_keywords))[:10]

        # 신뢰도 업데이트
        if "confidence" in existing:
            existing["confidence"] = min(1.0, existing["confidence"] + 0.1)

        # 액션 병합
        for key in ["successful_actions", "failed_actions"]:
            if key in existing and key in new:
                existing[key] = list(set(existing[key] + new[key]))

    def _rebuild_index(self):
        """인덱스 재구축"""
        self.index = self._create_default_index()

        for memory in self.memories:
            self._update_index(memory)

        self._save_index()

    def _update_access_count(self, memory_id: str):
        """메모리 접근 횟수 업데이트"""
        for memory in self.memories:
            if memory["id"] == memory_id:
                memory["accessed_count"] = memory.get("accessed_count", 0) + 1
                memory["last_accessed"] = datetime.now().isoformat()
                break

    def _summarize_memories(self, memories: List[Dict]) -> List[Dict]:
        """메모리 요약"""
        summaries = []

        # 카테고리별로 그룹화
        by_category = defaultdict(list)
        for memory in memories:
            category = memory.get("category", "unknown")
            by_category[category].append(memory)

        for category, category_memories in by_category.items():
            # 성공률 계산
            success_count = sum(1 for m in category_memories if m.get("success", False))
            success_rate = success_count / len(category_memories) if category_memories else 0

            # 공통 패턴 추출
            all_actions = []
            all_learnings = []
            for memory in category_memories:
                all_actions.extend(memory.get("actions", []))
                all_learnings.extend(memory.get("learnings", []))

            # 요약 생성
            summary = {
                "id": self._generate_memory_id({"task": f"summary_{category}"}),
                "timestamp": datetime.now().isoformat(),
                "category": category,
                "task": f"{category} 카테고리 요약",
                "context": f"{len(category_memories)}개 경험 종합",
                "success_rate": success_rate,
                "common_actions": list(set(all_actions))[:10],
                "key_learnings": list(set(all_learnings))[:10],
                "memory_count": len(category_memories)
            }

            summaries.append(summary)

        return summaries

    def _save_memories(self):
        """메모리 저장"""
        try:
            with open(self.memory_file, 'w', encoding='utf-8') as f:
                json.dump(self.memories, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"메모리 저장 실패: {e}")

    def _save_index(self):
        """인덱스 저장"""
        try:
            with open(self.index_file, 'w', encoding='utf-8') as f:
                json.dump(self.index, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"인덱스 저장 실패: {e}")

    def _save_patterns(self):
        """패턴 저장"""
        try:
            with open(self.patterns_file, 'w', encoding='utf-8') as f:
                json.dump(self.patterns, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"패턴 저장 실패: {e}")

    def get_statistics(self) -> Dict:
        """메모리 통계 반환"""
        return {
            "total_memories": len(self.memories),
            "categories": list(self.index.get("by_category", {}).keys()),
            "agents": list(self.index.get("by_agent", {}).keys()),
            "success_rate": len(self.index.get("by_success", {}).get("true", [])) / len(self.memories) if self.memories else 0,
            "patterns_learned": {
                "success": len(self.patterns.get("success_patterns", [])),
                "failure": len(self.patterns.get("failure_patterns", [])),
                "optimization_rules": len(self.patterns.get("optimization_rules", []))
            },
            "most_accessed": self._get_most_accessed_memories(5),
            "keyword_count": len(self.keyword_index)
        }

    def _get_most_accessed_memories(self, top_n: int) -> List[Dict]:
        """가장 많이 접근된 메모리"""
        sorted_memories = sorted(
            self.memories,
            key=lambda m: m.get("accessed_count", 0),
            reverse=True
        )

        return [
            {
                "id": m["id"],
                "task": m.get("task", ""),
                "accessed_count": m.get("accessed_count", 0)
            }
            for m in sorted_memories[:top_n]
        ]