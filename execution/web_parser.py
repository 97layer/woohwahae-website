import sys
import json
import re
import urllib.request
import urllib.error
from html import unescape

def get_metadata(url):
    """
    Extracts title, description, and OG metadata from a URL using standard libraries.
    """
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
    }
    
    try:
        req = urllib.request.Request(url, headers=headers)
        with urllib.request.urlopen(req, timeout=30) as response:
            html_content = response.read().decode('utf-8', errors='ignore')
            
        # 1. Title
        title = "No Title"
        title_match = re.search(r'<title>(.*?)</title>', html_content, re.IGNORECASE | re.DOTALL)
        if title_match:
            title = unescape(title_match.group(1).strip())
            
        og_title_match = re.search(r'<meta\s+property=["\']og:title["\']\s+content=["\'](.*?)["\']', html_content, re.IGNORECASE)
        if og_title_match:
            title = unescape(og_title_match.group(1).strip())

        # 2. Description
        description = "No description available."
        og_desc_match = re.search(r'<meta\s+property=["\']og:description["\']\s+content=["\'](.*?)["\']', html_content, re.IGNORECASE)
        desc_match = re.search(r'<meta\s+name=["\']description["\']\s+content=["\'](.*?)["\']', html_content, re.IGNORECASE)
        
        if og_desc_match:
            description = unescape(og_desc_match.group(1).strip())
        elif desc_match:
            description = unescape(desc_match.group(1).strip())
            
        # 3. Image
        image = ""
        og_image_match = re.search(r'<meta\s+property=["\']og:image["\']\s+content=["\'](.*?)["\']', html_content, re.IGNORECASE)
        if og_image_match:
            image = unescape(og_image_match.group(1).strip())

        return {
            "title": title,
            "description": description[:300] + "..." if len(description) > 300 else description,
            "image": image,
            "url": url,
            "status": "success"
        }

    except urllib.error.HTTPError as e:
        # Instagram often blocks standard user-agents, but we try to return what we have
        return {
            "title": "Social Content (IG/External)",
            "description": f"URL detected but access restricted. Content may require manual fuel. (Error: {e.code})",
            "url": url,
            "status": "partial"
        }
    except Exception as e:
        return {
            "error": str(e), 
            "url": url,
            "status": "failed"
        }

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No URL provided", "status": "failed"}))
        return

    url = sys.argv[1]
    metadata = get_metadata(url)
    
    # 97layerOS Standard Output Format
    result = {
        "metadata": metadata,
        "transcript": metadata.get("description", "No content extracted.")
    }
    
    print(json.dumps(result, ensure_ascii=False))

if __name__ == "__main__":
    main()
