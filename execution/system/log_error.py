# Filename: execution/system/log_error.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import os
import json
import logging
import traceback
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any

logger = logging.getLogger(__name__)

class ErrorLogger:
    """Structured error logging and pattern analysis."""
    
    def __init__(self, log_dir: Optional[str] = None):
        if log_dir is None:
            # Reusing /tmp strategy due to previous permission success
            log_dir = Path('/tmp') / '97layer_error_logs'
        
        self.log_dir = Path(log_dir)
        self.log_dir.mkdir(parents=True, exist_ok=True)

    def log_error(self, script: str, error: Exception, context: Optional[Dict] = None):
        """Logs an error as a structured JSON file."""
        timestamp = datetime.now().isoformat()
        log_file = self.log_dir / f"error_{datetime.now().strftime('%Y%m%d_%H%M%S_%f')}.json"
        
        error_data = {
            "timestamp": timestamp,
            "script": script,
            "error_type": type(error).__name__,
            "message": str(error),
            "traceback": traceback.format_exc(),
            "context": context or {}
        }
        
        try:
            with open(log_file, 'w', encoding='utf-8') as f:
                json.dump(error_data, f, indent=4)
        except Exception as e:
            logger.error(f"Failed to log error: {e}")

    def get_error_history(self, days: int = 7) -> List[Dict]:
        """Retrieves error history for the last N days."""
        errors = []
        limit = datetime.now() - timedelta(days=days)
        
        for log_file in self.log_dir.glob("error_*.json"):
            try:
                with open(log_file, 'r') as f:
                    data = json.load(f)
                    if datetime.fromisoformat(data['timestamp']) > limit:
                        errors.append(data)
            except: continue
        return sorted(errors, key=lambda x: x['timestamp'], reverse=True)

    def analyze_patterns(self, days: int = 7) -> Dict[str, Any]:
        """Analyzes repeated failure patterns."""
        errors = self.get_error_history(days)
        patterns = {}
        for err in errors:
            key = f"{err['script']}:{err['error_type']}"
            patterns[key] = patterns.get(key, 0) + 1
            
        return {k: v for k, v in patterns.items() if v >= 2} # Repeated at least twice
