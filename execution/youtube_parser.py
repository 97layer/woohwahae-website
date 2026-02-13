import sys
import json
import re
import urllib.request
import urllib.parse

def get_video_id(url):
    pattern = r'(?:v=|\/|embed\/|youtu.be\/)([0-9A-Za-z_-]{11})'
    match = re.search(pattern, url)
    return match.group(1) if match else None

def get_metadata_portable(url):
    """oEmbed API를 사용한 라이브러리 프리 메타데이터 추출"""
    try:
        oembed_url = f"https://www.youtube.com/oembed?url={urllib.parse.quote(url)}&format=json"
        with urllib.request.urlopen(oembed_url) as response:
            data = json.loads(response.read().decode())
            return {
                "title": data.get("title", "Unknown Title"),
                "author": data.get("author_name", "Unknown Author"),
                "thumbnail": data.get("thumbnail_url", ""),
                "url": url
            }
    except Exception as e:
        return {"error": str(e), "url": url}

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No URL provided"}))
        return

    url = sys.argv[1]
    video_id = get_video_id(url)
    
    if not video_id:
        print(json.dumps({"error": "Invalid YouTube URL"}))
        return

    metadata = get_metadata_portable(url)
    metadata["id"] = video_id
    
    # 트랜스크립트는 라이브러리 부재로 인해 Gemini Web 협업 권장
    result = {
        "metadata": metadata,
        "transcript": "[Pending] 트랜스크립트는 Gemini Web에서 추출하여 시스템에 'Fuel' 해주시면 자산화가 완성됩니다."
    }
    
    print(json.dumps(result, ensure_ascii=False))

if __name__ == "__main__":
    main()
