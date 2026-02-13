# 97layerOS Core Configuration & Mercenary Standard Recovery
# Author: 97LAYER Mercenary Standard Applied

SYSTEM_CONFIG = {
    "PROJECT_NAME": "97layerOS",
    "VERSION": "1.1.0 (Hybrid Engine Ready)",
    "BASE_DIRECTORY": "/Users/97layer/97layerOS",
    "CLOUD_SYNC_DIRECTORY": "Google Drive/내 드라이브/97layerOS"
}

# 1. 5인 크루 직책 및 도메인 매핑 (Updated naming)
AGENT_CREW = {
    "Creative_Director": {
        "legacy_name": "Sovereign",
        "domain": "브랜드 철학 및 최종 의사결정",
        "directive_path": "directives/agents/creative_director.md"
    },
    "Strategy_Analyst": {
        "legacy_name": "Scout",
        "domain": "시장 조사 및 데이터 온톨로지화",
        "directive_path": "directives/agents/strategy_analyst.md"
    },
    "Technical_Director": {
        "legacy_name": "Architect",
        "domain": "시스템 아키텍처 및 오케스트레이션",
        "directive_path": "directives/agents/technical_director.md"
    },
    "Chief_Editor": {
        "legacy_name": "Narrator",
        "domain": "콘텐츠 서사 집필 및 텍스트 정제",
        "directive_path": "directives/agents/chief_editor.md"
    },
    "Art_Director": {
        "legacy_name": "Artisan",
        "domain": "시각적 미학 및 레이아웃 설계",
        "directive_path": "directives/agents/art_director.md"
    }
}

# 2. 텔레그램 인프라 자격 증명
TELEGRAM_CONFIG = {
    "BOT_NAME": "studio_97layer_official_bot",
    "BOT_TOKEN": "8271602365:AAE3zKgYnwg_C_lizxfZbXM5JtVymZfrNGk"
}

# 3. 97LAYER Mercenary Standard (코드 작성 원칙)
MERCENARY_STANDARD = {
    "TYPE_HINTING": True,  # 모든 함수에 타입 힌트 적용
    "LOGGING": "Lazy Formatting (%s)",  # 성능 최적화 로그 포맷
    "ENCODING": "utf-8",  # 파일 입출력 시 한글 깨짐 방지
    "ERROR_HANDLING": ["HttpError", "IOError", "RequestException"],  # 구체적 예외 처리
    "ZERO_NOISE": "이모지 사용 금지, 가식적 공감 배제, 냉철한 논리 유지"
}

# 4. 차세대 스킬 아키텍처 (New)
SKILLS_STRUCTURE = {
    "PATH": ".agent/skills/",
    "FORMAT": "YAML Front Matter + Markdown Description",
    "OBJECTIVE": "토큰 소모 최적화 및 기능 단위 호출"
}

# 5. 작업 상태 관리자 초기화 템플릿 (task_status.json)
INITIAL_TASK_STATUS = {
    "current_mission": "System Recovery & Infrastructure Re-alignment",
    "completed_tasks": ["Agent Naming Reset", "Cloud Sync Setup"],
    "pending_tasks": ["Telegram Daemon Integration", "Skills Architecture Porting"],
    "system_entropy": "Low"
}

# 6. 시냅스 (신경망) 설정 (New)
SYNAPSE_CONFIG = {
    "MAX_DEBATE_TURNS": 3,
    "DEFAULT_COUNCIL": ["Creative_Director", "Strategy_Analyst", "Technical_Director"],
    "AUTONOMY_INTERVAL": 3600,  # 1 hour
    "DEBATE_MODEL": "gemini-2.0-flash-thinking-exp-01-21"  # Thinking model for complex reasoning
}
