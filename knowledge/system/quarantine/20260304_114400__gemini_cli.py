#!/Users/97layer/97layerOS/.venv/bin/python3
import google.genai as genai
import os
import sys

def get_api_key():
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        print("\033[91mError\033[0m: GEMINI_API_KEY 환경변수가 설정되지 않았습니다.")
        print("💡 \033[93m해결 방법\033[0m: ~/.zshrc 파일에 다음 줄을 추가하세요.")
        print("export GEMINI_API_KEY='대표님의_API_키'")
        sys.exit(1)
    return api_key

def read_input():
    prompt = ""
    # 파이프(Piped)나 리다이렉션으로 입력이 들어오는 경우 (예: cat file.txt | gemini)
    if not sys.stdin.isatty():
        prompt += sys.stdin.read().strip() + "\n\n"

    # 인자(Argument)로 입력된 텍스트 처리 (예: gemini "코드 요약해줘")
    if len(sys.argv) > 1:
        prompt += " ".join(sys.argv[1:]).strip()
        
    if not prompt:
        print("\033[93m사용법\033[0m: gemini '질문 내용'")
        print("예시 1) gemini 코드 리뷰 좀 해줘")
        print("예시 2) cat file.txt | gemini '요약해'")
        sys.exit(1)
        
    return prompt.strip()

def main():
    api_key = get_api_key()
    client = genai.Client(api_key=api_key)
    prompt = read_input()
    
    try:
        print("\033[90m■ 시스템 사유 연산 중... (gemini-2.5-pro)\033[0m\n")
        response = client.models.generate_content(
            model='gemini-2.5-pro',
            contents=prompt,
        )
        
        # 출력 꾸미기
        print("\033[36m────────────────────────────────────────────────────────────────────────\033[0m")
        print(response.text)
        print("\033[36m────────────────────────────────────────────────────────────────────────\033[0m")
    except Exception as e:
         print(f"\n\033[91m오류가 발생했습니다\033[0m: {e}")

if __name__ == "__main__":
    main()
