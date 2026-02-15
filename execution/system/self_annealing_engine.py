#!/usr/bin/env python3
"""
자가 치유 엔진 (Self-Annealing Engine)
연산 에러 3회 반복 시 자동 코드 수정 및 재실행

Features:
- 에러 패턴 추적 (3회 임계값)
- 자동 코드 분석 및 수정 제안
- 수정 내용 directives/self_annealing_log.md에 기록
- 재실행 (Retry) 메커니즘
"""

import os
import sys
import json
import traceback
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Callable, Any
from collections import defaultdict
import logging

# Project Root
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [SelfAnnealing] - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 로그 파일 경로
ANNEALING_LOG = PROJECT_ROOT / "directives" / "self_annealing_log.md"
ERROR_TRACKING = PROJECT_ROOT / "knowledge" / "system" / "error_tracking.json"


class ErrorTracker:
    """에러 추적 및 패턴 분석"""

    def __init__(self, threshold: int = 3):
        self.threshold = threshold
        self.error_counts: Dict[str, int] = defaultdict(int)
        self.error_history: List[Dict] = []
        self.tracking_file = ERROR_TRACKING

        # 기존 이력 로드
        self._load_tracking_data()

    def _load_tracking_data(self):
        """추적 데이터 로드"""
        if self.tracking_file.exists():
            try:
                with open(self.tracking_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    self.error_counts = defaultdict(int, data.get("error_counts", {}))
                    self.error_history = data.get("error_history", [])
            except Exception as e:
                logger.error(f"Failed to load error tracking: {e}")

    def _save_tracking_data(self):
        """추적 데이터 저장"""
        try:
            self.tracking_file.parent.mkdir(parents=True, exist_ok=True)
            with open(self.tracking_file, 'w', encoding='utf-8') as f:
                json.dump({
                    "error_counts": dict(self.error_counts),
                    "error_history": self.error_history[-100:]  # 최근 100개만
                }, f, indent=2, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to save error tracking: {e}")

    def record_error(
        self,
        error_type: str,
        error_message: str,
        context: Dict
    ) -> bool:
        """
        에러 기록 및 임계값 체크

        Returns:
            True if threshold reached (자가 치유 필요)
        """
        error_key = f"{error_type}:{context.get('function', 'unknown')}"

        # 카운트 증가
        self.error_counts[error_key] += 1
        count = self.error_counts[error_key]

        # 이력 기록
        self.error_history.append({
            "timestamp": datetime.now().isoformat(),
            "error_key": error_key,
            "error_type": error_type,
            "error_message": error_message,
            "count": count,
            "context": context
        })

        self._save_tracking_data()

        logger.warning(f"Error recorded: {error_key} (count: {count})")

        # 임계값 도달
        if count >= self.threshold:
            logger.error(f"Threshold reached for {error_key} ({count} times)")
            return True

        return False

    def reset_error(self, error_key: str):
        """에러 카운트 리셋 (수정 후)"""
        if error_key in self.error_counts:
            del self.error_counts[error_key]
            self._save_tracking_data()
            logger.info(f"Reset error count: {error_key}")

    def get_frequent_errors(self, limit: int = 10) -> List[Dict]:
        """빈번한 에러 목록 조회"""
        sorted_errors = sorted(
            self.error_counts.items(),
            key=lambda x: x[1],
            reverse=True
        )
        return [
            {"error_key": key, "count": count}
            for key, count in sorted_errors[:limit]
        ]


class SelfAnnealingEngine:
    """자가 치유 엔진"""

    def __init__(self, threshold: int = 3):
        self.error_tracker = ErrorTracker(threshold=threshold)
        self.annealing_log = ANNEALING_LOG
        self.annealing_log.parent.mkdir(parents=True, exist_ok=True)

    def _analyze_error(self, error_info: Dict) -> Dict:
        """에러 분석 및 수정 제안 생성"""
        error_type = error_info.get("error_type", "Unknown")
        error_message = error_info.get("error_message", "")
        context = error_info.get("context", {})

        # 공통 에러 패턴 매칭
        suggestions = []

        if "FileNotFoundError" in error_type:
            suggestions.append({
                "type": "FILE_CHECK",
                "description": "파일 존재 확인 로직 추가",
                "code_example": "if not Path(file_path).exists():\n    logger.error(f'File not found: {file_path}')\n    return None"
            })

        elif "KeyError" in error_type:
            suggestions.append({
                "type": "DICT_SAFETY",
                "description": "딕셔너리 안전 접근 (.get() 사용)",
                "code_example": "value = data.get('key', default_value)"
            })

        elif "IndexError" in error_type:
            suggestions.append({
                "type": "LIST_BOUNDS",
                "description": "리스트 범위 체크",
                "code_example": "if index < len(items):\n    item = items[index]"
            })

        elif "AttributeError" in error_type:
            suggestions.append({
                "type": "ATTR_CHECK",
                "description": "속성 존재 확인",
                "code_example": "if hasattr(obj, 'attr'):\n    value = obj.attr"
            })

        elif "TypeError" in error_type:
            suggestions.append({
                "type": "TYPE_VALIDATION",
                "description": "타입 체크 추가",
                "code_example": "if isinstance(value, expected_type):\n    process(value)"
            })

        elif "ConnectionError" in error_type or "TimeoutError" in error_type:
            suggestions.append({
                "type": "RETRY_LOGIC",
                "description": "재시도 로직 추가",
                "code_example": "for attempt in range(3):\n    try:\n        result = api_call()\n        break\n    except Exception as e:\n        if attempt == 2:\n            raise\n        time.sleep(2 ** attempt)"
            })

        else:
            suggestions.append({
                "type": "GENERAL_ERROR_HANDLING",
                "description": "try-except 블록 강화",
                "code_example": "try:\n    operation()\nexcept SpecificError as e:\n    logger.error(f'Error: {e}')\n    return fallback_value"
            })

        return {
            "error_info": error_info,
            "suggestions": suggestions,
            "timestamp": datetime.now().isoformat()
        }

    def _log_annealing(self, analysis: Dict):
        """자가 치유 로그 기록"""
        try:
            # 기존 로그 읽기
            if self.annealing_log.exists():
                with open(self.annealing_log, 'r', encoding='utf-8') as f:
                    content = f.read()
            else:
                content = "# Self-Annealing Log\n\n자가 치유 활동 기록\n\n---\n\n"

            # 새 항목 추가
            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            error_info = analysis["error_info"]
            suggestions = analysis["suggestions"]

            log_entry = f"""
## [{timestamp}] {error_info.get('error_type', 'Unknown Error')}

**에러 메시지**: {error_info.get('error_message', 'N/A')}

**컨텍스트**:
- Function: {error_info.get('context', {}).get('function', 'unknown')}
- File: {error_info.get('context', {}).get('file', 'unknown')}
- Line: {error_info.get('context', {}).get('line', 'unknown')}

**발생 횟수**: {error_info.get('count', 'N/A')}

**자동 수정 제안**:
"""
            for i, suggestion in enumerate(suggestions, 1):
                log_entry += f"""
{i}. **{suggestion['type']}**: {suggestion['description']}

```python
{suggestion['code_example']}
```
"""

            log_entry += "\n**상태**: 분석 완료 (수동 검토 필요)\n\n---\n\n"

            # 파일 저장
            with open(self.annealing_log, 'w', encoding='utf-8') as f:
                f.write(content + log_entry)

            logger.info(f"✓ Annealing log updated: {self.annealing_log}")

        except Exception as e:
            logger.error(f"Failed to log annealing: {e}")

    def handle_error(
        self,
        error: Exception,
        context: Dict,
        auto_fix: bool = False
    ) -> Dict:
        """
        에러 처리 및 자가 치유 판단

        Args:
            error: 발생한 예외
            context: 에러 컨텍스트 (function, file, line 등)
            auto_fix: 자동 수정 시도 여부 (현재는 로그만)

        Returns:
            처리 결과
        """
        error_type = type(error).__name__
        error_message = str(error)
        tb = traceback.extract_tb(error.__traceback__)

        # 컨텍스트 보강
        if tb:
            last_frame = tb[-1]
            context.setdefault("file", last_frame.filename)
            context.setdefault("line", last_frame.lineno)
            context.setdefault("function", last_frame.name)

        # 에러 기록
        needs_annealing = self.error_tracker.record_error(
            error_type=error_type,
            error_message=error_message,
            context=context
        )

        result = {
            "error_type": error_type,
            "error_message": error_message,
            "needs_annealing": needs_annealing,
            "context": context
        }

        # 자가 치유 필요
        if needs_annealing:
            logger.warning("Triggering self-annealing process...")

            # 에러 분석
            analysis = self._analyze_error({
                "error_type": error_type,
                "error_message": error_message,
                "context": context,
                "count": self.error_tracker.error_counts.get(
                    f"{error_type}:{context.get('function', 'unknown')}",
                    0
                )
            })

            # 로그 기록
            self._log_annealing(analysis)

            result["analysis"] = analysis
            result["status"] = "ANNEALING_LOGGED"

            # TODO: 자동 수정 (AI 기반 코드 패치)
            # if auto_fix:
            #     self._apply_auto_fix(analysis)

        return result

    def retry_with_annealing(
        self,
        func: Callable,
        max_retries: int = 3,
        context: Optional[Dict] = None,
        *args,
        **kwargs
    ) -> Any:
        """
        재시도 메커니즘 (자가 치유 포함)

        Args:
            func: 실행할 함수
            max_retries: 최대 재시도 횟수
            context: 에러 컨텍스트
            *args, **kwargs: 함수 인자

        Returns:
            함수 실행 결과
        """
        context = context or {}
        context.setdefault("function", func.__name__)

        last_error = None

        for attempt in range(max_retries):
            try:
                result = func(*args, **kwargs)

                # 성공 시 에러 카운트 리셋
                error_key = f"*:{context.get('function', 'unknown')}"
                self.error_tracker.reset_error(error_key)

                return result

            except Exception as e:
                last_error = e
                logger.warning(f"Attempt {attempt + 1}/{max_retries} failed: {e}")

                # 자가 치유 처리
                annealing_result = self.handle_error(e, context.copy())

                if attempt < max_retries - 1:
                    # 재시도 대기
                    import time
                    wait_time = 2 ** attempt
                    logger.info(f"Retrying in {wait_time}s...")
                    time.sleep(wait_time)
                else:
                    # 최종 실패
                    logger.error(f"All retries exhausted for {func.__name__}")
                    if annealing_result.get("needs_annealing"):
                        logger.error("Self-annealing log created - manual review required")

        # 최종 실패 시 예외 발생
        raise last_error


def main():
    """테스트"""
    engine = SelfAnnealingEngine(threshold=3)

    # 테스트 함수
    def failing_function():
        raise FileNotFoundError("/path/to/missing/file.txt")

    # 재시도 테스트
    try:
        result = engine.retry_with_annealing(
            failing_function,
            max_retries=3,
            context={"function": "failing_function", "file": "test.py"}
        )
    except Exception as e:
        logger.error(f"Final error: {e}")

    # 빈번한 에러 조회
    frequent = engine.error_tracker.get_frequent_errors()
    print("\nFrequent Errors:")
    print(json.dumps(frequent, indent=2))


if __name__ == "__main__":
    main()
