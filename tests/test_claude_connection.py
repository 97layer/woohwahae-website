import os
from anthropic import Anthropic
from dotenv import load_dotenv

load_dotenv()

def test_claude():
    api_key = os.getenv("ANTHROPIC_API_KEY")
    if not api_key or "your_claude" in api_key:
        print("Error: Anthropic API Key not set or invalid in .env")
        return

    client = Anthropic(api_key=api_key)
    try:
        message = client.messages.create(
            model="claude-3-5-sonnet-20241022",
            max_tokens=100,
            messages=[{"role": "user", "content": "97layerOS 시스템 정상 연동 확인을 위한 인사말을 한국어로 짧게 해주세요."}]
        )
        print(f"Claude Response: {message.content[0].text}")
    except Exception as e:
        print(f"Connection failed: {e}")

if __name__ == "__main__":
    test_claude()
