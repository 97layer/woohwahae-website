#!/usr/bin/env python3
"""
97layerOS Brand Scout Agent
Magazine B 방향 전환 — 슬로우라이프/미니멀 브랜드 자동 발굴 시스템

Role:
1. Discovery: 슬로우라이프 브랜드 자동 발굴 (SNS/웹 크롤링)
2. Screening: 브랜드 철학 적합성 평가 (WOOHWAHAE 5 Pillars 기준)
3. Data Gathering: 승인된 브랜드 컨텐츠 수집
4. Dossier 생성: knowledge/brands/[brand_name]/ 구조화

LLM: Gemini 2.5 Flash (분석), Gemini Pro (심층 평가)
Output: 브랜드 도시에 → Magazine CE Agent 소재

Author: 97layerOS
Created: 2026-02-17
"""

import os
import sys
import json
import time
import re
import requests
from pathlib import Path
from typing import Dict, Any, Optional, List
from datetime import datetime, timedelta

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

try:
    import google.genai as genai
except ImportError:
    print("⚠️  google-genai not installed. Run: pip install google-genai")
    sys.exit(1)

# 크롤링 도구 (선택: BeautifulSoup, Playwright, Apify)
try:
    from bs4 import BeautifulSoup
except ImportError:
    BeautifulSoup = None
    print("⚠️  beautifulsoup4 not installed. Web scraping limited.")


class BrandScout:
    """
    슬로우라이프 브랜드 발굴 및 분석 에이전트.

    주요 기능:
    1. 브랜드 후보 자동 발견 (Instagram, 웹사이트, 큐레이션 사이트)
    2. WOOHWAHAE 철학 적합성 평가 (5 Pillars 기준)
    3. 승인된 브랜드 심층 데이터 수집
    4. Magazine Issue 소재 도시에 생성
    """

    # WOOHWAHAE 5 Pillars (IDENTITY.md 기준)
    BRAND_PILLARS = {
        "authenticity": "진정성 — 가면 없는 대화, 취약함의 수용",
        "practicality": "실용성 — 지속 가능한 관리, 과시 없는 실질",
        "elegance": "우아함 — 절제된 표현, 여백의 미학",
        "precision": "정밀함 — 구조적 사고, 기반의 견고함",
        "innovation": "혁신 — 관습의 파괴, Anti-Uniformity"
    }

    # 발굴 소스별 키워드
    DISCOVERY_SOURCES = {
        "instagram_hashtags": [
            "slowliving", "slowlife", "minimalism", "minimallifestyle",
            "슬로우라이프", "미니멀라이프", "독립브랜드", "로컬브랜드",
            "제주살이", "시골살이", "귀촌", "자급자족"
        ],
        "curation_sites": [
            "kinfolk.com",
            "monocle.com",
            "cereal-mag.com",
            "thekinfolktable.com"
        ],
        "reddit_communities": [
            "r/simpleliving",
            "r/minimalism",
            "r/slowliving"
        ]
    }

    def __init__(self, api_key: Optional[str] = None):
        """
        Brand Scout 초기화

        Args:
            api_key: Google API key (또는 env GOOGLE_API_KEY)
        """
        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found in environment")

        self.client = genai.Client(api_key=api_key)
        self.model_flash = 'gemini-2.5-flash'  # 빠른 스크리닝
        self.model_pro = 'gemini-2.0-flash-001'      # 심층 분석

        # 브랜드 저장 경로
        self.brands_root = PROJECT_ROOT / "knowledge" / "brands"
        self.brands_root.mkdir(parents=True, exist_ok=True)

        # 후보 큐
        self.candidates_file = self.brands_root / "candidates.json"

        print(f"Brand Scout: 준비됨. 저장소: {self.brands_root}")

    # ===== 1. DISCOVERY: 브랜드 후보 발굴 =====

    def discover_brands(self, source: str = "instagram", limit: int = 10) -> List[Dict]:
        """
        브랜드 후보 자동 발굴

        Args:
            source: "instagram", "web", "reddit"
            limit: 최대 후보 수

        Returns:
            [{"name", "url", "source", "discovered_at"}]
        """
        print(f"[Scout] {source}에서 브랜드 발굴 시작 (limit={limit})")

        if source == "instagram":
            return self._discover_instagram(limit)
        elif source == "web":
            return self._discover_web(limit)
        elif source == "reddit":
            return self._discover_reddit(limit)
        else:
            print(f"[Scout] 지원하지 않는 소스: {source}")
            return []

    def _discover_instagram(self, limit: int) -> List[Dict]:
        """
        Instagram 해시태그 기반 브랜드 발굴

        방법:
        1. Apify Instagram Scraper (유료) — 현재 미구현
        2. 수동 입력 모드: 순호가 텔레그램으로 후보 제안

        TODO: Apify/Playwright 통합
        """
        print("[Scout] Instagram 크롤링 — 미구현 (수동 입력 모드)")
        return []

    def _discover_web(self, limit: int) -> List[Dict]:
        """
        큐레이션 사이트에서 브랜드 발굴

        방법:
        - Kinfolk, Monocle 등 사이트 크롤링
        - Beautiful Soup으로 브랜드명/링크 추출

        TODO: 사이트별 파서 구현
        """
        print("[Scout] 웹 크롤링 — 미구현")
        return []

    def _discover_reddit(self, limit: int) -> List[Dict]:
        """
        Reddit 커뮤니티에서 브랜드 추천 수집

        방법:
        - Reddit API로 r/simpleliving 등 서브레딧 검색
        - "brand", "store", "shop" 키워드 포스트 필터링

        TODO: Reddit API 통합
        """
        print("[Scout] Reddit 크롤링 — 미구현")
        return []

    # ===== 2. SCREENING: 브랜드 적합성 평가 =====

    def screen_candidate(self, candidate: Dict) -> Dict[str, Any]:
        """
        브랜드 후보 스크리닝 (WOOHWAHAE 철학 적합성)

        Args:
            candidate: {"name", "url", "description"(optional)}

        Returns:
            {
                "name": str,
                "url": str,
                "screening_score": int (0-100),
                "pillars_match": {pillar: score},
                "decision": "approve" | "reject" | "review",
                "reasoning": str
            }
        """
        name = candidate.get("name", "Unknown")
        url = candidate.get("url", "")
        description = candidate.get("description", "")

        print(f"[Scout] 스크리닝 시작: {name}")

        # 웹사이트 크롤링 (About/Philosophy 섹션)
        raw_content = self._fetch_brand_content(url)
        if not raw_content and not description:
            return {
                "name": name,
                "url": url,
                "screening_score": 0,
                "decision": "reject",
                "reasoning": "브랜드 정보 수집 실패 (웹사이트 접근 불가)"
            }

        # Gemini Flash로 빠른 평가
        prompt = self._build_screening_prompt(name, url, raw_content or description)

        try:
            response = self.client.models.generate_content(
                model=self.model_flash,
                contents=prompt
            )
            analysis_text = response.text.strip()

            # JSON 파싱 시도
            analysis = self._parse_screening_result(analysis_text)
            analysis["name"] = name
            analysis["url"] = url

            print(f"[Scout] {name} 스크리닝 완료: {analysis['decision']} (점수: {analysis['screening_score']})")
            return analysis

        except Exception as e:
            print(f"[Scout] 스크리닝 실패: {e}")
            return {
                "name": name,
                "url": url,
                "screening_score": 0,
                "decision": "error",
                "reasoning": str(e)
            }

    def _fetch_brand_content(self, url: str) -> Optional[str]:
        """
        브랜드 웹사이트에서 About/Philosophy 섹션 크롤링

        Args:
            url: 브랜드 웹사이트 URL

        Returns:
            추출된 텍스트 (또는 None)
        """
        if not url or not BeautifulSoup:
            return None

        try:
            headers = {
                'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
            }
            resp = requests.get(url, headers=headers, timeout=15)
            if resp.status_code != 200:
                return None

            soup = BeautifulSoup(resp.text, 'html.parser')

            # About/Philosophy 섹션 우선 탐색
            about_sections = soup.find_all(['div', 'section'], class_=re.compile(r'about|philosophy|story', re.I))
            if about_sections:
                text = " ".join([sec.get_text(strip=True) for sec in about_sections[:3]])
                return text[:5000]  # 최대 5000자

            # 전체 텍스트 fallback
            body = soup.find('body')
            if body:
                return body.get_text(strip=True)[:5000]

            return None

        except Exception as e:
            print(f"[Scout] 크롤링 실패 ({url}): {e}")
            return None

    def _build_screening_prompt(self, name: str, url: str, content: str) -> str:
        """스크리닝 프롬프트 생성"""
        pillars_desc = "\n".join([f"- {k}: {v}" for k, v in self.BRAND_PILLARS.items()])

        return f"""당신은 슬로우라이프/미니멀 라이프 브랜드 큐레이터입니다.
WOOHWAHAE 매거진에 소개할 브랜드를 평가합니다.

**WOOHWAHAE 5 Pillars:**
{pillars_desc}

**평가 대상 브랜드:**
- 이름: {name}
- URL: {url}

**브랜드 소개 (웹사이트 내용):**
{content[:3000]}

**평가 기준:**
1. 각 Pillar별 적합도 (0-20점, 총 100점)
2. 전체 점수 (screening_score)
3. 결정: approve (80점 이상) | review (50-79점) | reject (50점 미만)
4. 이유 (reasoning): 2-3문장

**출력 형식 (JSON):**
{{
  "screening_score": 85,
  "pillars_match": {{
    "authenticity": 18,
    "practicality": 16,
    "elegance": 20,
    "precision": 15,
    "innovation": 16
  }},
  "decision": "approve",
  "reasoning": "절제된 시각 언어와 지속 가능한 제품 철학이 WOOHWAHAE와 공명. 여백을 활용한 웹사이트 디자인이 우아함 기준 충족."
}}

JSON만 출력하세요."""

    def _parse_screening_result(self, response_text: str) -> Dict:
        """Gemini 응답에서 JSON 파싱"""
        # JSON 블록 추출 시도
        json_match = re.search(r'\{[\s\S]*\}', response_text)
        if json_match:
            try:
                return json.loads(json_match.group(0))
            except json.JSONDecodeError:
                pass

        # 파싱 실패 시 기본값
        return {
            "screening_score": 0,
            "pillars_match": {},
            "decision": "error",
            "reasoning": "응답 파싱 실패"
        }

    # ===== 3. DATA GATHERING: 브랜드 심층 데이터 수집 =====

    def gather_brand_data(self, brand_name: str, brand_url: str) -> Dict[str, Any]:
        """
        승인된 브랜드의 심층 데이터 수집

        수집 항목:
        1. 웹사이트 전체 텍스트 (About, Products, Story)
        2. Instagram 최근 포스트 (6개월)
        3. 외부 언론 리뷰/인터뷰

        Args:
            brand_name: 브랜드명
            brand_url: 공식 웹사이트

        Returns:
            {"raw_content": str, "social_posts": list, "press": list}
        """
        print(f"[Scout] {brand_name} 데이터 수집 시작")

        # 1. 웹사이트 크롤링
        raw_content = self._fetch_brand_content(brand_url) or ""

        # 2. Instagram (미구현 — Apify 통합 필요)
        social_posts = []

        # 3. 언론 리뷰 (Google Search API or Serper.dev)
        press = self._search_press(brand_name)

        return {
            "raw_content": raw_content,
            "social_posts": social_posts,
            "press": press,
            "collected_at": datetime.now().isoformat()
        }

    def _search_press(self, brand_name: str) -> List[Dict]:
        """Google Search로 브랜드 언론 보도 검색 (미구현)"""
        # TODO: Serper.dev API 통합
        print(f"[Scout] {brand_name} 언론 검색 — 미구현")
        return []

    # ===== 4. DOSSIER 생성 =====

    def create_dossier(self, screening: Dict, data: Dict) -> Path:
        """
        브랜드 도시에 생성 → knowledge/brands/[slug]/

        구조:
        - profile.json: 메타데이터 + 스크리닝 결과
        - raw_content.md: 수집한 원본 컨텐츠
        - sa_analysis.json: SA Agent 전략 분석 (나중에)

        Args:
            screening: screen_candidate() 결과
            data: gather_brand_data() 결과

        Returns:
            도시에 디렉토리 경로
        """
        brand_name = screening["name"]
        slug = self._slugify(brand_name)

        dossier_dir = self.brands_root / slug
        dossier_dir.mkdir(parents=True, exist_ok=True)

        # profile.json
        profile = {
            "name": brand_name,
            "url": screening["url"],
            "screening_score": screening["screening_score"],
            "pillars_match": screening.get("pillars_match", {}),
            "decision": screening["decision"],
            "reasoning": screening.get("reasoning", ""),
            "collected_at": data.get("collected_at", datetime.now().isoformat()),
            "status": "pending_sa_analysis"  # SA Agent 분석 대기
        }
        (dossier_dir / "profile.json").write_text(
            json.dumps(profile, indent=2, ensure_ascii=False), encoding="utf-8"
        )

        # raw_content.md
        raw_md = f"""# {brand_name} — Raw Content

**URL**: {screening['url']}
**수집 일시**: {data.get('collected_at', 'N/A')}

---

## Website Content

{data.get('raw_content', '(수집 실패)')}

---

## Social Media Posts

{len(data.get('social_posts', []))}개 포스트 수집 (미구현)

---

## Press & Reviews

{len(data.get('press', []))}개 리뷰 수집 (미구현)
"""
        (dossier_dir / "raw_content.md").write_text(raw_md, encoding="utf-8")

        print(f"[Scout] 도시에 생성 완료: {dossier_dir}")
        return dossier_dir

    def _slugify(self, text: str) -> str:
        """브랜드명 → 파일시스템 안전 slug"""
        slug = re.sub(r'[^\w\s-]', '', text.lower())
        slug = re.sub(r'[\s_]+', '-', slug)
        return slug.strip('-')

    # ===== 5. 후보 큐 관리 =====

    def add_candidate(self, name: str, url: str, description: str = "", source: str = "manual"):
        """후보 브랜드 수동 추가 (순호 입력)"""
        candidates = self._load_candidates()

        # 중복 체크
        if any(c["name"] == name for c in candidates):
            print(f"[Scout] 중복 후보: {name}")
            return

        candidate = {
            "name": name,
            "url": url,
            "description": description,
            "source": source,
            "added_at": datetime.now().isoformat(),
            "status": "pending"
        }
        candidates.append(candidate)
        self._save_candidates(candidates)
        print(f"[Scout] 후보 추가: {name}")

    def process_candidates(self, auto_approve: bool = False) -> List[Path]:
        """
        큐에 있는 모든 후보 브랜드 처리

        Args:
            auto_approve: True이면 80점 이상 자동 승인 + 도시에 생성

        Returns:
            생성된 도시에 경로 리스트
        """
        candidates = self._load_candidates()
        pending = [c for c in candidates if c["status"] == "pending"]

        print(f"[Scout] 처리 대기 중인 후보: {len(pending)}개")

        dossiers = []
        for candidate in pending:
            # 스크리닝
            screening = self.screen_candidate(candidate)
            candidate["screening_result"] = screening
            candidate["status"] = "screened"

            # auto_approve 모드: 80점 이상 자동 도시에 생성
            if auto_approve and screening["decision"] == "approve":
                data = self.gather_brand_data(screening["name"], screening["url"])
                dossier_path = self.create_dossier(screening, data)
                dossiers.append(dossier_path)
                candidate["status"] = "approved"

            self._save_candidates(candidates)
            time.sleep(2)  # API rate limit

        print(f"[Scout] 처리 완료. 도시에 생성: {len(dossiers)}개")
        return dossiers

    def _load_candidates(self) -> List[Dict]:
        """candidates.json 로드"""
        if not self.candidates_file.exists():
            return []
        try:
            return json.loads(self.candidates_file.read_text(encoding="utf-8"))
        except Exception:
            return []

    def _save_candidates(self, candidates: List[Dict]):
        """candidates.json 저장"""
        self.candidates_file.write_text(
            json.dumps(candidates, indent=2, ensure_ascii=False), encoding="utf-8"
        )


# ===== CLI =====

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Brand Scout — 슬로우라이프 브랜드 발굴")
    parser.add_argument("--add", nargs=3, metavar=("NAME", "URL", "DESC"), help="후보 수동 추가")
    parser.add_argument("--process", action="store_true", help="큐 처리 (스크리닝)")
    parser.add_argument("--auto-approve", action="store_true", help="80점 이상 자동 승인")
    parser.add_argument("--discover", choices=["instagram", "web", "reddit"], help="브랜드 자동 발굴")

    args = parser.parse_args()

    scout = BrandScout()

    if args.add:
        name, url, desc = args.add
        scout.add_candidate(name, url, desc)

    elif args.process:
        scout.process_candidates(auto_approve=args.auto_approve)

    elif args.discover:
        candidates = scout.discover_brands(source=args.discover, limit=10)
        print(f"발견한 후보: {len(candidates)}개")
        for c in candidates:
            scout.add_candidate(c["name"], c["url"], source=args.discover)

    else:
        parser.print_help()
