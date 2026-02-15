import json
import os
import sys
from datetime import datetime
from pathlib import Path

# Path Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

from libs.ai_engine import AIEngine
from libs.core_config import AI_MODEL_CONFIG, AGENT_CREW

def generate_autonomous_tasks():
    status_file = PROJECT_ROOT / "task_status.json"
    vision_file = PROJECT_ROOT / "VISION.md"
    
    if not status_file.exists() or not vision_file.exists():
        print("Missing required files for autonomous development.")
        return

    status = json.loads(status_file.read_text())
    vision = vision_file.read_text()
    
    completed = status.get("completed_tasks", [])
    pending = status.get("pending_tasks", [])
    
    ai = AIEngine(AI_MODEL_CONFIG)
    
    prompt = f"""
    당신은 97LAYER의 수석 오케스트레이터입니다. 시스템의 자율적 확장을 위해 다음 데이터를 분석하고 신규 태스크를 제안하십시오.

    [PROJECT VISION]
    {vision}

    [COMPLETED TASKS]
    {completed[-10:]}

    [PENDING TASKS]
    {pending}

    [지침]
    1. 현재 진행 상황을 바탕으로, 비전에 다가가기 위해 '지금 꼭 필요한' 또는 '다음에 해야 할' 고도의 태스크를 1~2개 도출하십시오.
    2. 모든 태스크의 지시문(instruction)은 반드시 **한국어**로 작성하십시오. 영어 사용은 엄격히 금지됩니다.
    3. 태스크 형식은 JSON 리스트여야 하며, 각 항목은 다음 필드를 포함해야 합니다:
       - id: "auto_dev_[timestamp]"
       - type: "R&D" 또는 "DEVELOPMENT"
       - agent: "{list(AGENT_CREW.keys())}" 중 적절한 에이전트
       - instruction: 구체적이고 실행 가능한 한국어 지시문
       - council: 복잡한 논의가 필요하면 true, 아니면 false
    3. 중복되거나 사소한 작업은 배제하고, 시스템의 지능적 확장에 집중하십시오.
    4. 오직 JSON 코드 블록만 출력하십시오.
    """
    
    response = ai.generate_response(prompt)
    
    # Extract JSON from response
    try:
        json_str = response.strip()
        if "```json" in json_str:
            json_str = json_str.split("```json")[1].split("```")[0].strip()
        elif "```" in json_str:
            json_str = json_str.split("```")[1].split("```")[0].strip()
            
        new_tasks = json.loads(json_str)
        
        # Add timestamp to IDs
        timestamp = datetime.now().strftime("%H%M%S")
        for i, task in enumerate(new_tasks):
            task["id"] = f"auto_dev_{timestamp}_{i}"
            status.setdefault("pending_tasks", []).append(task)
            print(f"Added autonomous task: {task['instruction']}")
            
        # Save status
        status_file.write_text(json.dumps(status, indent=4, ensure_ascii=False))
        
    except Exception as e:
        print(f"Failed to parse or add autonomous tasks: {e}\nResponse: {response}")

if __name__ == "__main__":
    generate_autonomous_tasks()
