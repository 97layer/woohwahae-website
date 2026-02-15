# 97layerOS Core Configuration & Mercenary Standard Recovery
# Author: 97LAYER Mercenary Standard Applied

import os
from pathlib import Path
from dotenv import load_dotenv

# Load .env file
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
try:
    load_dotenv(PROJECT_ROOT / ".env")
except Exception:
    pass # Bypassing permission errors for environment loading

SYSTEM_CONFIG = {
    "PROJECT_NAME": "97layerOS",
    "VERSION": "1.2.0 (Autonomous Guardian Ready)",
    "BASE_DIRECTORY": "/Users/97layer/97layerOS",
    "CLOUD_SYNC_DIRECTORY": "Google Drive/내 드라이브/97layerOS",
    "DAEMONS": ["technical_daemon.py", "telegram_daemon.py", "snapshot_daemon.py"]
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

# 2. 텔레그램 인프라 자격 증명 (Hardcoded for stability due to environment restrictions)
TELEGRAM_CONFIG = {
    "BOT_NAME": "97LayerOSwoohwahae",
    "BOT_TOKEN": os.getenv("TELEGRAM_BOT_TOKEN")  # Fallback 제거 (보안)
}

# 3. 97LAYER Mercenary Standard (코드 작성 원칙)
MERCENARY_STANDARD = {
    "TYPE_HINTING": True,  # 모든 함수에 타입 힌트 적용
    "LOGGING": "Lazy Formatting (%s)",  # 성능 최적화 로그 포맷
    "ENCODING": "utf-8",  # 파일 입출력 시 한글 깨짐 방지
    "ERROR_HANDLING": ["HttpError", "IOError", "RequestException"],  # 구체적 예외 처리
    "ZERO_NOISE": "이모지 및 볼드(**) 사용 금지, 가식적 공감 배제, 냉철한 논리 유지",
    "SILENT_MODE": True  # 기본적으로 콘솔 출력 최소화 (Zero Notification)
}

# Global Log Level
LOG_LEVEL = "WARNING" if MERCENARY_STANDARD["SILENT_MODE"] else "INFO"

# 4. 차세대 스킬 아키텍처 (New)
SKILLS_STRUCTURE = {
    "PATH": "system/libs/skills/",
    "FORMAT": "YAML Front Matter + Markdown Description",
    "OBJECTIVE": "토큰 소모 최적화 및 기능 단위 호출",
    "REGISTRY": {
        "skill-001": {
            "name": "Unified Input Protocol (UIP)",
            "executor": ["execution/system/youtube_parser.py", "execution/system/ontology_transform.py"],
            "trigger_patterns": [r'youtube\.com', r'youtu\.be'],
            "output_path": "knowledge/raw_signals/"
        },
        "signal_capture": {
            "name": "Signal Capture",
            "executor": ["execution/system/web_parser.py"],
            "trigger_patterns": [r'https?://'],
            "output_path": "knowledge/inbox/"
        },
        "data_curation": {
            "name": "Data Curation",
            "executor": ["execution/system/ontology_transform.py"],
            "trigger_patterns": [],
            "output_path": "knowledge/"
        },
        "instagram_content_curator": {
            "name": "Instagram Content Curator",
            "executor": ["execution/system/instagram_generator.py"],
            "trigger_patterns": [r'instagram\.com'],
            "output_path": "assets/instagram/"
        }
    }
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

# 7. AI 엔진 모델 설정
AI_MODEL_CONFIG = {
    "model_name": "gemini-2.0-flash",
    "temperature": 0.7,
    "top_p": 0.95
}

# 8. 동기화 설정 (Sync Config)
SYNC_CONFIG = {
    "SOURCE_PATH": "/Users/97layer/97layerOS",
    "DESTINATION_PATH": "/Users/97layer/내 드라이브/97layerOS",
    "EXCLUDE_PATTERNS": [
        "*.venv*", "venv*", ".local_node", "__pycache__", 
        ".git", ".DS_Store", "*.tmp", ".mcp-source"
    ],
    "SYNC_DIRECTORIES": ["directives", "execution", "knowledge", "system"]
}

# 9. Instagram API 설정 (New)
INSTAGRAM_CONFIG = {
    "ACCESS_TOKEN": os.getenv("INSTAGRAM_ACCESS_TOKEN", ""),
    "BUSINESS_ACCOUNT_ID": os.getenv("INSTAGRAM_BUSINESS_ACCOUNT_ID", ""),
    "API_VERSION": "v18.0",
    "DEFAULT_PUBLISH_TIME": "next_monday_10am",
    "MAX_CAPTION_LENGTH": 2200,
    "PUBLISH_QUEUE_PATH": "knowledge/assets/ready_to_publish/publish_queue.json"
}

# 10. 자율 의식 (Rituals) 설정 (New)
RITUALS_CONFIG = {
    "DAILY_BRIEFING": {
        "trigger_hour": 9,  # 09:00 AM
        "agent": "Strategy_Analyst",
        "task_type": "BRIEFING",
        "instruction": "밤사이 글로벌 및 로컬 미용/패션/테크 트렌드를 분석하고, 우리 브랜드에 적용 가능한 인사이트를 'Daily Briefing' 형식으로 보고하십시오.",
        "council": False
    },
    "WEEKLY_COUNCIL": {
        "trigger_weekday": 0,  # Monday
        "trigger_hour": 10,    # 10:00 AM
        "agent": "Creative_Director",
        "task_type": "COUNCIL",
        "instruction": "주간 전략 회의(Weekly Council)를 소집합니다. [Cycle Protocol 기반] 안건: (1) 지난주 발행 회고, (2) 이번주 콘텐츠 후보 검토 (SA 제안), (3) 브랜드 일관성 점검 (MBQ 기준), (4) 사이클 병목 체크 (TD 리포트). WOOHWAHAE 브랜드 철학 5가지와 Aesop 벤치마크 준수 여부를 중점 검토하십시오.",
        "council": True
    },
    "NIGHTLY_CONSOLIDATION": {
        "trigger_hour": 23,  # 11:00 PM
        "agent": "Strategy_Analyst",
        "task_type": "CONSOLIDATION",
        "instruction": "[Junction Protocol - Connect 단계] 오늘 수집된 모든 Raw Signals(knowledge/raw_signals/)를 분석하여 연결 그래프를 생성하십시오. (1) 과거 기록과 유사성 탐색, (2) 반복 테마 식별, (3) 5가지 철학(Slow, 실용적 미학, 무언의 교감, 자기 긍정, 아카이브)과 연결, (4) 콘텐츠 후보 우선순위 제안. 97layer_identity.md와 연결 강도를 명시하십시오.",
        "council": False
    },
    "SYSTEM_HEALTH_CHECK": {
        "trigger_hour": None, # Non-scheduled, runs every loop in technical_daemon
        "agent": "Technical_Director",
        "task_type": "DIAGNOSTIC",
        "instruction": "시스템 상태를 점검하고 자가 치유 및 보고를 수행하십시오.",
        "council": False
    },
    "MIDDAY_INSIGHT": {
        "trigger_hour": 14,  # 02:00 PM
        "agent": "Strategy_Analyst",
        "task_type": "INSIGHT",
        "instruction": "오전 동안 수집된 데이터와 진행 상황을 요약하고, 브랜드 방향성에 대한 짧은 인사이트를 브로드캐스트하십시오.",
        "council": False
    },
    "DAILY_AUTONOMOUS_DEVELOPMENT": {
        "trigger_hour": 3,  # 03:00 AM
        "agent": "Technical_Director",
        "task_type": "AUTONOMOUS_DEV",
        "instruction": "현재 프로젝트 상황과 비전을 분석하여 시스템 고도화를 위한 차세대 태스크를 스스로 생성하십시오.",
        "council": False
    },
    "DRAFT_72H_CHECK": {
        "trigger_hour": None,  # Runs every loop
        "agent": "Technical_Director",
        "task_type": "PUBLISH_CHECK",
        "instruction": "[Imperfect Publish Protocol] Draft 폴더에서 72시간 경과 파일을 체크하고, 76시간 초과 시 자동 폐기하십시오.",
        "council": False
    },
    "INSTAGRAM_PUBLISH_CHECK": {
        "trigger_hour": 10,  # 10:00 AM (Monday default publish time)
        "agent": "Technical_Director",
        "task_type": "INSTAGRAM_PUBLISH",
        "instruction": "[Instagram 발행] 발행 큐에서 예약 시간이 도래한 콘텐츠를 Instagram에 자동 발행하십시오.",
        "council": False
    },
    "WEEKLY_SECURITY_UPDATE": {
        "trigger_weekday": 6,  # Sunday
        "trigger_hour": 2,  # 02:00 AM
        "agent": "Technical_Director",
        "task_type": "SECURITY_UPDATE",
        "instruction": "[Self-Maintenance] 주간 보안 업데이트를 실행하십시오. (1) pip 패키지 취약점 검사 (pip-audit), (2) 보안 업데이트 적용, (3) 시스템 무결성 테스트 (import/environment/critical files), (4) requirements.txt 갱신. 실패 시 Telegram 알림.",
        "council": False
    },
    "DAILY_AUTO_UPDATE": {
        "trigger_hour": 3,  # 03:00 AM (with AUTONOMOUS_DEV)
        "agent": "Technical_Director",
        "task_type": "AUTO_UPDATE",
        "instruction": "[Self-Maintenance] Google Drive Core 폴더에서 신규/수정 스크립트를 감지하여 시스템에 자동 반영하십시오. (1) Drive 스캔 (SHA256 해시 비교), (2) 기존 파일 Archive 백업, (3) 즉시 적용, (4) update_log.json 기록.",
        "council": False
    },
    "ENVIRONMENT_CHECK": {
        "trigger_hour": None,  # Runs before every major operation
        "agent": "Technical_Director",
        "task_type": "ENV_CHECK",
        "instruction": "[Self-Maintenance] 환경 정합성 검사를 실행하십시오. (1) Python 버전, (2) 필수 패키지, (3) 필수 디렉토리, (4) 환경변수, (5) 시스템 리소스. 누락 시 Silent Install 수행.",
        "council": False
    }
}

# ===========================================
# 하이브리드 지능망: 환경 감지 및 Low-Power Mode
# ===========================================

def detect_environment() -> str:
    """
    현재 실행 환경 감지

    Returns:
        "GCP_VM": Google Cloud VM
        "CLOUD_CONTAINER": Cloud Run Container
        "MACBOOK": 맥북 로컬
    """
    import subprocess

    # GCP VM 감지 (방법 1: 파일 존재)
    if Path("/etc/google_compute_engine").exists():
        return "GCP_VM"

    # GCP VM 감지 (방법 2: 메타데이터 서버)
    try:
        result = subprocess.run(
            ["curl", "-s", "-H", "Metadata-Flavor: Google",
             "http://metadata.google.internal/computeMetadata/v1/instance/id"],
            capture_output=True,
            timeout=2
        )
        if result.returncode == 0 and result.stdout:
            return "GCP_VM"
    except:
        pass

    # Docker/Podman Container 감지
    if Path("/.dockerenv").exists() or Path("/run/.containerenv").exists():
        return "CLOUD_CONTAINER"

    # 맥북 로컬 (기본값)
    return "MACBOOK"

# 현재 환경
ENVIRONMENT = os.getenv("ENVIRONMENT", detect_environment())

# ===========================================
# Knowledge Root: Container-First Architecture
# ===========================================
def get_knowledge_root() -> Path:
    """
    환경별 지식 저장소 경로 반환

    Container: /app/.container_data/knowledge
    Local: PROJECT_ROOT/.container_data/knowledge

    Returns:
        Path: 지식 저장소 루트 경로
    """
    if ENVIRONMENT in ["GCP_VM", "CLOUD_CONTAINER"]:
        # Container 내부: /app/.container_data/knowledge
        return Path("/app/.container_data/knowledge")
    else:
        # Macbook 로컬: PROJECT_ROOT/system/infra/container_data/knowledge
        return PROJECT_ROOT / "system" / "infra" / "container_data" / "knowledge"

KNOWLEDGE_ROOT = get_knowledge_root()

# Knowledge 하위 경로들
KNOWLEDGE_PATHS = {
    "signals": KNOWLEDGE_ROOT / "signals",
    "content": KNOWLEDGE_ROOT / "content",
    "archive": KNOWLEDGE_ROOT / "archive",
    "system": KNOWLEDGE_ROOT / "system",
    "blueprints": PROJECT_ROOT / "knowledge" / "blueprints",
    "manuals": PROJECT_ROOT / "knowledge" / "manuals",
    "docs": PROJECT_ROOT / "knowledge" / "docs",  # New Sanctuary path
}

# Core System 경로들
CORE_PATHS = {
    "directives": PROJECT_ROOT / "directives",
    "execution": PROJECT_ROOT / "execution",
    "knowledge": PROJECT_ROOT / "knowledge",
    "system": PROJECT_ROOT / "system",
    "logs": PROJECT_ROOT / "knowledge" / "logs"
}

# Low-Power Mode 설정 (GCP_VM 전용)
if ENVIRONMENT == "GCP_VM":
    # Night Guard 모드: RAM 1GB 최적화
    PROCESSING_MODE = "sequential"          # 순차 처리 (병렬 비활성화)
    MAX_CONCURRENT_AGENTS = 1               # 동시 실행 Agent 1개만
    ENABLE_MULTIMODAL = False               # 이미지 분석 비활성화
    MEMORY_LIMIT_MB = 700                   # 메모리 제한 (1GB 중 300MB는 OS용)
    AI_MODEL_PREFERENCE = "gemini-1.5-flash"  # 경량 모델
    print(f"[Low-Power Mode] GCP_VM 환경 감지 → RAM 1GB 최적화 활성화")
else:
    # Full Power 모드: 맥북 또는 Cloud Run
    PROCESSING_MODE = "parallel"            # 병렬 처리
    MAX_CONCURRENT_AGENTS = 5               # 5-agent 동시 실행
    ENABLE_MULTIMODAL = True                # 이미지 분석 활성화
    MEMORY_LIMIT_MB = None                  # 메모리 제한 없음
    AI_MODEL_PREFERENCE = "gemini-1.5-pro"  # 고성능 모델
    if ENVIRONMENT == "MACBOOK":
        print(f"[Full Power Mode] 맥북 환경 감지 → 5-Agent 병렬 처리 활성화")

