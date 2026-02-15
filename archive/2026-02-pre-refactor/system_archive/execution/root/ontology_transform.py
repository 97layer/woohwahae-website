import sys
import json
import os
import re
from datetime import datetime
from pathlib import Path

# Fix path to allow imports from libs
BASE_DIR = Path(__file__).resolve().parent.parent
if str(BASE_DIR) not in sys.path:
    sys.path.insert(0, str(BASE_DIR))

from libs.ai_engine import AIEngine
from libs.core_config import AI_MODEL_CONFIG

class OntologyEngine:
    def __init__(self):
        self.ai = AIEngine(AI_MODEL_CONFIG)
        self.raw_signals_dir = BASE_DIR / "knowledge" / "raw_signals"
        self.raw_signals_dir.mkdir(parents=True, exist_ok=True)

    def _generate_id(self) -> str:
        files = os.listdir(self.raw_signals_dir)
        rs_files = [f for f in files if f.startswith("rs-") and f.endswith(".md")]
        
        numbers = []
        for f in rs_files:
            try:
                # Extract number from rs-XXX format
                match = re.search(r"rs-(\d+)", f)
                if match:
                    numbers.append(int(match.group(1)))
            except:
                continue
        
        next_num = max(numbers) + 1 if numbers else 1
        return f"rs-{next_num:03d}"

    def process_content(self, content: str, source: str = "unknown") -> str:
        """
        Analyzes content via LLM and generates a structured Markdown file.
        """
        rs_id = self._generate_id()
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        prompt = f"""
        당신은 97LAYER의 지식 분류 엔진(Ontology Engine)입니다.
        아래 입력된 Raw Data를 분석하여 시스템이 이해할 수 있는 구조화된 지식(Raw Signal)으로 변환하십시오.

        [Input Data]
        Source: {source}
        Content (Truncated if too long):
        {content[:4000]}

        [Transformation Rules]
        1. **Keywords**: 본문에서 가장 중요한 핵심 키워드 3~5개를 추출하십시오 (해시태그 형식).
        2. **Summary**: 내용을 3줄 요약하십시오.
        3. **Strategic Alignment**: 이 내용이 97LAYER의 철학(Essentialism, Minimalism, high-agency)과 어떻게 연결되는지 분석하십시오.
        4. **Actionable Insight**: 즉시 실행 가능한 아이디어나 적용점이 있다면 명시하십시오. 없으면 'N/A'.

        [Output Format]
        Start directly with the Frontmatter.
        ---
        id: {rs_id}
        source: {source}
        captured_at: {now}
        tags: [tag1, tag2, tag3]
        ---

        # Raw Signal: {rs_id}

        ## 1. Executive Summary
        (Your summary here)

        ## 2. Key Insights
        - (Point 1)
        - (Point 2)

        ## 3. Strategic Alignment
        (Analysis)

        ## 4. Actionable Items
        (Items)
        """

        try:
            # Generate structured content
            structured_md = self.ai.generate_response(prompt)
            
            # Clean up potential markdown code block artifacts
            structured_md = structured_md.replace("```markdown", "").replace("```", "").strip()

            # Save to file
            filename = f"{rs_id}_{self._sanitize_filename(source)}.md"
            filepath = self.raw_signals_dir / filename
            
            with open(filepath, "w", encoding="utf-8") as f:
                f.write(structured_md)
                
            return str(filepath)

        except Exception as e:
            print(f"[OntologyEngine Error] {e}")
            return None

    def _sanitize_filename(self, name: str) -> str:
        # Keep only alphanumeric and standard separators
        clean = re.sub(r'[^a-zA-Z0-9_\-\.]', '_', name)
        # Remove extension for the ID part if present, but we append .md later
        return Path(clean).stem

# Backward compatibility for direct script usage (e.g. piping JSON)
def main():
    if sys.stdin.isatty():
        print("OntologyEngine: Use as module or pipe JSON input.")
        return

    try:
        input_data = json.load(sys.stdin)
        engine = OntologyEngine()
        
        # Handle YouTube parser output format
        content = input_data.get("transcript", "")
        source = f"youtube_{input_data.get('metadata', {}).get('id', 'unknown')}"
        
        if not content:
             content = str(input_data)

        result_path = engine.process_content(content, source)
        if result_path:
            # Extract id from result_path filename
            rs_id = Path(result_path).name.split("_")[0]
            print(json.dumps({"status": "success", "file": result_path, "id": rs_id}))
        else:
            print(json.dumps({"status": "error"}))
            
    except Exception as e:
        print(json.dumps({"status": "error", "message": str(e)}))

if __name__ == "__main__":
    main()
