import os
import json
import time
import requests
from bs4 import BeautifulSoup
from datetime import datetime
import hashlib
from pathlib import Path

# 설정
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
SOURCES_FILE = str(PROJECT_ROOT / "knowledge" / "brands" / "wellness_sources.json")
SIGNALS_DIR = str(PROJECT_ROOT / "knowledge" / "signals")
USER_AGENTS = [
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
]

def ensure_dir(path):
    if not os.path.exists(path):
        os.makedirs(path)

def get_headers():
    import random
    return {
        "User-Agent": random.choice(USER_AGENTS),
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
        "Referer": "https://www.google.com/",
        "Upgrade-Insecure-Requests": "1"
    }

def fetch_content(url):
    try:
        response = requests.get(url, headers=get_headers(), timeout=10)
        response.raise_for_status()
        return response.text
    except Exception as e:
        print(f"Error fetching {url}: {e}")
        return None

def parse_source(source_config):
    print(f"Scanning {source_config['name']}...")
    html = fetch_content(source_config['url'])
    if not html:
        return []

    soup = BeautifulSoup(html, 'html.parser')
    selectors = source_config['selectors']
    
    articles = []
    # 리스트 아이템 찾기 (없으면 body 전체에서 a태그 탐색)
    if 'article_list' in selectors and selectors['article_list']:
        items = soup.select(selectors['article_list'])
    else:
        items = [soup] # Fallback
    
    print(f"Found {len(items)} blocks in {source_config['name']}")
    
    count = 0
    for item in items:
        if count >= 3: break # 소스당 3개 제한

        try:
            # 제목/링크 추출 (휴리스틱)
            link_el = None
            if selectors.get('title'):
                link_el = item.select_one(selectors['title'])
                if link_el and link_el.name != 'a':
                    link_el = link_el.find('a')
            
            if not link_el:
                # h2, h3 내부 a 태그 우선 검색
                link_el = item.select_one('h2 a, h3 a, h4 a')
            
            if not link_el: continue

            title = link_el.get_text(strip=True)
            link = link_el['href']
            
            if not title or len(title) < 5: continue
            
            if not link.startswith('http'):
                # base_url과 결합 시 슬래시 처리
                if not link.startswith('/'):
                    link = '/' + link
                link = source_config['base_url'].rstrip('/') + link
            
            # 상세 페이지 크롤링
            print(f"  Fetching article: {title}")
            article_html = fetch_content(link)
            if not article_html: continue
            
            article_soup = BeautifulSoup(article_html, 'html.parser')
            
            # 본문 추출 시도
            content = ""
            if selectors.get('content_body'):
                 content_el = article_soup.select_one(selectors['content_body'])
                 if content_el:
                     content = content_el.get_text(separator='\n\n', strip=True)
            
            # Fallback: p 태그들 모음
            if not content or len(content) < 100:
                paragraphs = article_soup.find_all('p')
                content = "\n\n".join([p.get_text(strip=True) for p in paragraphs if len(p.get_text(strip=True)) > 20])

            if len(content) < 100:
                print(f"  Skipping generic content (too short): {title}")
                continue

            # 저장
            save_signal(source_config['name'], title, link, content)
            articles.append(title)
            count += 1
            
            time.sleep(2) # 부하 방지
            
        except Exception as e:
            print(f"  Error parsing item: {e}")
            continue
            
    return articles

def save_signal(source, title, url, content):
    ensure_dir(SIGNALS_DIR)

    # 중복 체크 (URL 해시)
    doc_hash = hashlib.md5(url.encode()).hexdigest()[:10]
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    signal_id = "url_content_%s_%s" % (timestamp[:8], timestamp[9:])

    # 기존 동일 URL 신호 확인
    for existing in os.listdir(SIGNALS_DIR):
        if existing.endswith(".json") and doc_hash in existing:
            print("  Skipping duplicate: %s" % title)
            return

    # 통합 스키마 JSON
    signal_data = {
        "signal_id": signal_id,
        "type": "url_content",
        "status": "captured",
        "content": content[:3000],
        "captured_at": datetime.now().isoformat(),
        "from_user": "97layer",
        "source_channel": "crawler",
        "metadata": {
            "title": title,
            "source_url": url,
            "crawler_source": source,
            "doc_hash": doc_hash,
        },
    }

    filename = "%s_%s.json" % (signal_id, doc_hash)
    filepath = os.path.join(SIGNALS_DIR, filename)
    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(signal_data, f, ensure_ascii=False, indent=2)
    print("  Saved: %s" % filename)

def main():
    if not os.path.exists(SOURCES_FILE):
        print(f"Config file not found: {SOURCES_FILE}")
        return

    with open(SOURCES_FILE, 'r') as f:
        config = json.load(f)
        
    for source in config['sources']:
        parse_source(source)

if __name__ == "__main__":
    main()
