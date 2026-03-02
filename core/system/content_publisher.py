#!/usr/bin/env python3
"""
LAYER OS Content Publisher
CD 승인 후 최종 아웃풋 패키징 + Telegram push

아웃풋:
1. Instagram 패키지 (캡션 + 해시태그 + 이미지)
2. Archive 에세이 (Notion/블로그용 롱폼)
3. 이미지 소스: 순호 제공 → Gemini Imagen → Unsplash fallback

저장 위치: knowledge/assets/published/YYYY-MM-DD/
- instagram_caption.txt
- hashtags.txt
- archive_essay.txt
- image.jpg (또는 image_prompt.txt)
- meta.json

Author: LAYER OS
Created: 2026-02-17
"""

import os
import sys
import json
import logging
import requests
from pathlib import Path
from datetime import datetime
from typing import Dict, Optional, Any

from core.system.bot_templates import PUBLISH_ALERT

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

logger = logging.getLogger(__name__)

# Unsplash fallback 키워드 매핑
UNSPLASH_KEYWORD_MAP = {
    "슬로우라이프": "slow life aesthetic",
    "본질": "minimal japanese aesthetic",
    "디지털 피로": "calm nature minimal",
    "브랜드 아이덴티티": "brand identity studio",
    "헤어": "hair salon minimal",
    "default": "slow living minimal"
}


class ContentPublisher:
    """
    최종 콘텐츠 패키징 및 배포.
    순호에게 Telegram으로 결과 전달.
    """

    def __init__(self, base_path: Optional[Path] = None):
        self.base_path = base_path or Path("/app")
        self.published_root = self.base_path / "knowledge" / "assets" / "published"
        self.telegram_token = os.getenv("TELEGRAM_BOT_TOKEN", "")
        self.admin_id = int(os.getenv("ADMIN_TELEGRAM_ID", "0"))
        self.gemini_api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY", "")
        self.unsplash_key = os.getenv("UNSPLASH_ACCESS_KEY", "")

    def publish(self, payload: Dict) -> Dict:
        """
        CD 승인 완료 payload → 최종 패키지 생성 + Telegram push

        Args:
            payload: {signal_id, ce_result, cd_result, instagram_caption, hashtags, archive_essay, ad_result, ...}

        Returns:
            {status, published_path, telegram_sent}
        """
        signal_id = payload.get("signal_id", "unknown")
        today = datetime.now().strftime("%Y-%m-%d")
        timestamp = datetime.now().strftime("%H%M%S")
        publish_dir = self.published_root / today / f"{signal_id}_{timestamp}"
        publish_dir.mkdir(parents=True, exist_ok=True)

        logger.info("[Publisher] 발행 시작: %s", signal_id)

        # 콘텐츠 추출 — corpus 멀티유즈 포맷 포함
        instagram_caption = payload.get("instagram_caption", "")
        hashtags = payload.get("hashtags", "")
        archive_essay = payload.get("archive_essay", "")
        cd_result = payload.get("cd_result", {})
        sa_result = payload.get("sa_result", {})
        ad_result = payload.get("ad_result", {})

        # CE result에서도 추출 시도 (corpus 멀티유즈 포맷 포함)
        ce_result = payload.get("ce_result", {})
        if not instagram_caption:
            instagram_caption = ce_result.get("instagram_caption", "")
        if not hashtags:
            hashtags = ce_result.get("hashtags", "")
        if not archive_essay:
            archive_essay = ce_result.get("archive_essay", "")

        # corpus 전용 포맷
        pull_quote = ce_result.get("pull_quote", payload.get("pull_quote", ""))
        carousel_slides = ce_result.get("carousel_slides", payload.get("carousel_slides", []))
        # CE가 list로 반환하는 필드 → str 정규화
        _raw_tg = ce_result.get("telegram_summary", payload.get("telegram_summary", ""))
        telegram_summary = "\n".join(_raw_tg) if isinstance(_raw_tg, list) else (_raw_tg or "")
        _raw_ig = instagram_caption
        if isinstance(_raw_ig, list):
            instagram_caption = " ".join(_raw_ig)
        _raw_hash = hashtags
        if isinstance(_raw_hash, list):
            hashtags = " ".join(_raw_hash)
        essay_title = ce_result.get("essay_title", payload.get("essay_title", ""))
        is_corpus = payload.get("mode") == "corpus_essay" or bool(pull_quote)

        # 1. 텍스트 파일 저장
        (publish_dir / "instagram_caption.txt").write_text(instagram_caption, encoding="utf-8")
        (publish_dir / "hashtags.txt").write_text(hashtags, encoding="utf-8")
        (publish_dir / "archive_essay.txt").write_text(archive_essay, encoding="utf-8")

        # corpus 멀티유즈 추가 파일
        if pull_quote:
            (publish_dir / "pull_quote.txt").write_text(pull_quote, encoding="utf-8")
        if carousel_slides:
            slides_text = "\n---\n".join(carousel_slides) if isinstance(carousel_slides, list) else str(carousel_slides)
            (publish_dir / "carousel_slides.txt").write_text(slides_text, encoding="utf-8")
        if telegram_summary:
            (publish_dir / "telegram_summary.txt").write_text(telegram_summary, encoding="utf-8")

        # 2. 이미지 소스 결정
        image_path = self._get_image(payload, publish_dir, sa_result, ad_result)

        # 3. 메타데이터
        meta = {
            "signal_id": signal_id,
            "published_at": datetime.now().isoformat(),
            "is_corpus": is_corpus,
            "essay_title": essay_title,
            "instagram_caption_length": len(instagram_caption),
            "archive_essay_length": len(archive_essay),
            "has_carousel": bool(carousel_slides),
            "has_pull_quote": bool(pull_quote),
            "cd_brand_score": cd_result.get("brand_score", 0),
            "cd_decision": cd_result.get("decision", "approve"),
            "sa_strategic_score": sa_result.get("strategic_score", 0),
            "themes": sa_result.get("themes", []),
            "image_source": image_path.get("source", "none") if image_path else "none",
            "image_path": str(image_path.get("path", "")) if image_path else ""
        }
        (publish_dir / "meta.json").write_text(
            json.dumps(meta, indent=2, ensure_ascii=False), encoding="utf-8"
        )

        logger.info("[Publisher] 패키지 저장: %s (corpus=%s)", publish_dir, is_corpus)

        # 4. Telegram push
        # corpus 모드: telegram_summary 우선, 없으면 instagram_caption fallback
        tg_content = telegram_summary if (is_corpus and telegram_summary) else instagram_caption
        telegram_sent = self._push_to_telegram(
            instagram_caption=tg_content,
            hashtags=hashtags,
            archive_essay=archive_essay,
            image_path=image_path,
            meta=meta
        )

        # 5. long_term_memory에 발행 이력 추가
        self._update_memory(meta, instagram_caption, archive_essay)

        # 6. woohwahae.kr 웹사이트 자동 발행
        # corpus 모드: essay_title을 제목으로, pull_quote를 서브헤드로
        title = essay_title or (meta.get("themes", ["기록"])[0] if meta.get("themes") else "기록")
        website_published = self._publish_to_website(
            signal_id=signal_id,
            title=title,
            archive_essay=archive_essay,
            instagram_caption=instagram_caption,
            image_path=image_path,
            meta={**meta, "pull_quote": pull_quote}
        )

        result = {
            "status": "published",
            "signal_id": signal_id,
            "published_path": str(publish_dir),
            "telegram_sent": telegram_sent,
            "website_published": website_published,
            "meta": meta
        }

        logger.info("[Publisher] 완료: %s, telegram=%s, website=%s", signal_id, telegram_sent, website_published)
        return result

    def _get_image(self, payload: Dict, publish_dir: Path, sa_result: Dict, ad_result: Dict) -> Optional[Dict]:
        """
        이미지 소스 우선순위:
        1. 순호가 신호에 포함한 이미지
        2. Gemini Imagen (ad_result의 visual_concept 기반)
        3. Unsplash fallback (키워드 기반)
        """
        # 1. 순호 제공 이미지 확인
        signal_image = payload.get("image_path", "") or payload.get("image_url", "")
        if signal_image and Path(signal_image).exists():
            logger.info("[Publisher] 이미지: 순호 제공 사용 (%s)", signal_image)
            return {"source": "user_provided", "path": signal_image}

        # 2. Gemini Imagen 시도
        visual_concept = ad_result.get("visual_concept", {})
        image_prompt = visual_concept.get("image_prompt", "") or visual_concept.get("image_description", "")
        if not image_prompt:
            # AD result에서 직접 추출
            image_prompt = ad_result.get("image_prompt", "")

        if image_prompt and self.gemini_api_key:
            imagen_result = self._generate_imagen(image_prompt, publish_dir)
            if imagen_result:
                return imagen_result

        # 3. Unsplash fallback
        themes = sa_result.get("themes", [])
        keyword = self._themes_to_unsplash_keyword(themes)
        unsplash_result = self._fetch_unsplash(keyword, publish_dir)
        if unsplash_result:
            return unsplash_result

        # 4. 이미지 프롬프트만 저장 (나중에 순호가 직접 생성)
        if image_prompt:
            prompt_file = publish_dir / "image_prompt.txt"
            prompt_file.write_text(image_prompt, encoding="utf-8")
            logger.info("[Publisher] 이미지 없음, 프롬프트만 저장")
            return {"source": "prompt_only", "path": str(prompt_file)}

        return None

    def _generate_imagen(self, prompt: str, publish_dir: Path) -> Optional[Dict]:
        """Gemini Imagen으로 이미지 생성 (Gemini API)"""
        try:
            import google.genai as genai
            client = genai.Client(api_key=self.gemini_api_key)

            # Imagen 3 사용
            response = client.models.generate_images(
                model="imagen-3.0-generate-002",
                prompt=f"Minimal slow life aesthetic, WOOHWAHAE hair atelier mood. {prompt}",
                config={"number_of_images": 1, "aspect_ratio": "1:1"}
            )

            if response.generated_images:
                image_data = response.generated_images[0].image.image_bytes
                image_path = publish_dir / "image.jpg"
                image_path.write_bytes(image_data)
                logger.info("[Publisher] Imagen 생성 성공")
                return {"source": "gemini_imagen", "path": str(image_path)}

        except Exception as e:
            logger.warning("[Publisher] Imagen 실패: %s", e)

        return None

    def _fetch_unsplash(self, keyword: str, publish_dir: Path) -> Optional[Dict]:
        """Unsplash API로 이미지 다운로드"""
        if not self.unsplash_key:
            logger.info("[Publisher] Unsplash 키 없음, 스킵")
            return None

        try:
            url = "https://api.unsplash.com/photos/random"
            params = {
                "query": keyword,
                "orientation": "squarish",
                "client_id": self.unsplash_key
            }
            resp = requests.get(url, params=params, timeout=10)
            if resp.status_code == 200:
                data = resp.json()
                image_url = data.get("urls", {}).get("regular", "")
                if image_url:
                    img_resp = requests.get(image_url, timeout=15)
                    if img_resp.status_code == 200:
                        image_path = publish_dir / "image.jpg"
                        image_path.write_bytes(img_resp.content)
                        # 사진 출처 저장 (크레딧)
                        credit = f"Photo by {data.get('user', {}).get('name', 'Unknown')} on Unsplash"
                        (publish_dir / "image_credit.txt").write_text(credit, encoding="utf-8")
                        logger.info("[Publisher] Unsplash 이미지 다운로드 성공: %s", keyword)
                        return {"source": "unsplash", "path": str(image_path), "credit": credit}

        except Exception as e:
            logger.warning("[Publisher] Unsplash 실패: %s", e)

        return None

    def _themes_to_unsplash_keyword(self, themes: list) -> str:
        """SA 테마 → Unsplash 키워드"""
        for theme in themes:
            for ko_keyword, en_keyword in UNSPLASH_KEYWORD_MAP.items():
                if ko_keyword in theme:
                    return en_keyword
        return UNSPLASH_KEYWORD_MAP["default"]

    def _push_to_telegram(self, instagram_caption: str, hashtags: str,
                          archive_essay: str, image_path: Optional[Dict], meta: Dict) -> bool:
        """순호 텔레그램에 발행 결과 전송"""
        if not self.telegram_token or not self.admin_id:
            logger.warning("[Publisher] Telegram 설정 없음, 스킵")
            return False

        try:
            themes_str = ", ".join(meta.get("themes", []))
            score = meta.get("cd_brand_score", 0)
            sa_score = meta.get("sa_strategic_score", 0)

            # 메시지 구성
            essay_preview = archive_essay[:600] + ("..." if len(archive_essay) > 600 else "")
            message = PUBLISH_ALERT.format(
                themes=themes_str,
                sa_score=sa_score,
                cd_score=score,
                caption=instagram_caption,
                hashtags=hashtags,
                essay_preview=essay_preview,
                image_source=meta.get('image_source', '없음'),
            )

            api_url = f"https://api.telegram.org/bot{self.telegram_token}"

            # 이미지 있으면 이미지와 함께 전송
            if image_path and image_path.get("source") not in ("none", "prompt_only", None):
                img_file = Path(image_path.get("path", ""))
                if img_file.exists():
                    files = {"photo": open(img_file, "rb")}
                    data = {
                        "chat_id": self.admin_id,
                        "caption": message[:1024],  # Telegram caption 제한
                        "parse_mode": "Markdown"
                    }
                    resp = requests.post(f"{api_url}/sendPhoto", data=data, files=files, timeout=30)
                    if resp.status_code == 200:
                        logger.info("[Publisher] Telegram 이미지+텍스트 전송 성공")
                        # 에세이가 길면 별도 메시지
                        if len(archive_essay) > 600:
                            self._send_text(api_url, f"📝 *Archive 전문*\n\n{archive_essay}")
                        return True

            # 텍스트만 전송
            return self._send_text(api_url, message)

        except Exception as e:
            logger.error("[Publisher] Telegram 전송 실패: %s", e)
            return False

    def _send_text(self, api_url: str, text: str) -> bool:
        """텍스트 메시지 전송"""
        try:
            # 4096자 제한 분할
            chunks = [text[i:i+4000] for i in range(0, len(text), 4000)]
            for chunk in chunks:
                resp = requests.post(f"{api_url}/sendMessage", json={
                    "chat_id": self.admin_id,
                    "text": chunk,
                    "parse_mode": "Markdown"
                }, timeout=15)
                if resp.status_code != 200:
                    logger.warning("[Publisher] Telegram 메시지 전송 실패: %s", resp.text[:200])
                    return False
            logger.info("[Publisher] Telegram 텍스트 전송 성공")
            return True
        except Exception as e:
            logger.error("[Publisher] Telegram 전송 오류: %s", e)
            return False

    def _publish_to_website(self, signal_id: str, title: str, archive_essay: str,
                             instagram_caption: str, image_path: Optional[Dict],
                             meta: Dict) -> bool:
        """
        woohwahae.kr GitHub Pages에 아카이브 에세이 자동 발행.

        1. website/archive/[slug]/ 폴더에 index.html 생성
        2. website/archive/index.json 목록 업데이트
        3. git add/commit/push → GitHub Pages 자동 반영

        환경변수: GITHUB_TOKEN, GITHUB_REPO (예: 97layer/97layerOS)
        """
        github_token = os.getenv("GITHUB_TOKEN", "")
        github_repo = os.getenv("GITHUB_REPO", "")

        # 웹사이트 루트 경로
        website_root = self.base_path / "website"
        if not website_root.exists():
            logger.warning("[Publisher] website/ 폴더 없음 — 웹사이트 발행 스킵")
            return False

        try:
            # slug 생성 (날짜 + signal_id 앞 10자)
            today = datetime.now().strftime("%Y-%m-%d")
            slug = f"{today}-{signal_id[:20].replace('_', '-')}"

            # 개별 글 HTML 생성
            article_dir = website_root / "archive" / slug
            article_dir.mkdir(parents=True, exist_ok=True)

            # 에세이 줄바꿈 → HTML 단락
            essay_text = str(archive_essay)
            paragraphs = [p.strip() for p in essay_text.split('\n\n') if p.strip()]
            essay_html = "\n".join(f"<p>{p}</p>" for p in paragraphs) if paragraphs else f"<p>{essay_text}</p>"

            # 커버 이미지
            cover_html = ""
            if image_path and image_path.get("source") not in ("none", "prompt_only", None):
                img_src = image_path.get("path", "")
                if img_src and Path(img_src).exists():
                    # 이미지를 website/archive/slug/ 로 복사
                    import shutil
                    dest_img = article_dir / "cover.jpg"
                    shutil.copy2(img_src, dest_img)
                    cover_html = '<img class="article-cover" src="cover.jpg" alt="">'

            themes_str = ", ".join(meta.get("themes", []))

            article_html = f"""<!DOCTYPE html>
<html lang="ko">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="{instagram_caption[:120]}">
  <meta property="og:title" content="{title} — WOOHWAHAE">
  <meta property="og:description" content="{instagram_caption[:120]}">
  <title>{title} — WOOHWAHAE</title>
  <link rel="stylesheet" href="../../assets/css/style.css">
</head>
<body>

  <nav>
    <a href="/" class="nav-logo">WOOHWAHAE</a>
    <ul class="nav-links">
      <li><a href="../../practice/">Practice</a></li>
      <li><a href="../" class="active">아카이브</a></li>
      <li><a href="../../contact.html">연락</a></li>
    </ul>
  </nav>

  <div class="container">
    <header class="article-header">
      <p class="article-meta fade-in">{today} · {themes_str}</p>
      <h1 class="article-title fade-in">{title}</h1>
      <p class="article-lead fade-in">{instagram_caption}</p>
    </header>
  </div>

  {cover_html}

  <div class="container">
    <div class="article-body fade-in">
      {essay_html}
    </div>
  </div>

  <footer>
    <p class="footer-copy">© 2026 WOOHWAHAE</p>
    <ul class="footer-links">
      <li><a href="../">Archive</a></li>
      <li><a href="../../contact.html">Contact</a></li>
    </ul>
  </footer>

  <script src="../../assets/js/main.js"></script>
</body>
</html>"""

            (article_dir / "index.html").write_text(article_html, encoding="utf-8")

            # archive/index.json 업데이트
            index_json_path = website_root / "archive" / "index.json"
            try:
                posts = json.loads(index_json_path.read_text(encoding="utf-8"))
            except Exception:
                posts = []

            # 중복 방지
            if not any(p.get("slug") == slug for p in posts):
                posts.insert(0, {
                    "slug": slug,
                    "title": title,
                    "date": today,
                    "preview": paragraphs[0][:120] if paragraphs else instagram_caption[:120],
                    "signal_id": signal_id
                })
                index_json_path.write_text(
                    json.dumps(posts, indent=2, ensure_ascii=False), encoding="utf-8"
                )

            logger.info("[Publisher] 웹사이트 HTML 생성: %s", slug)

            # items.json 자동 갱신 (Dependency Graph 연동)
            self._update_items_json(signal_id, title, slug, today, meta)

            # Cascade Manager 실행 (의존성 전파)
            self._trigger_cascade("knowledge/service/items.json")

            # git push (GITHUB_TOKEN 있을 때만)
            if github_token and github_repo:
                return self._git_push_website(website_root, slug, today)
            else:
                logger.info("[Publisher] GITHUB_TOKEN 없음 — 로컬 저장만 완료")
                return True  # 로컬 생성은 성공

        except Exception as e:
            logger.error("[Publisher] 웹사이트 발행 실패: %s", e, exc_info=True)
            return False

    def _trigger_cascade(self, changed_file: str):
        """
        Cascade Manager 호출 → 의존성 전파
        items.json 변경 시 영향받는 파일 자동 추적
        """
        try:
            # Lazy import (순환 참조 방지)
            import sys
            sys.path.insert(0, str(self.base_path))
            from core.system.cascade_manager import CascadeManager

            manager = CascadeManager()
            report = manager.on_file_change(changed_file)

            logger.info(
                "[Publisher] Cascade triggered: %s (Tier: %s, Affected: %d nodes)",
                changed_file, report.tier, len(report.affected_nodes)
            )

        except Exception as e:
            logger.warning("[Publisher] Cascade Manager 실행 실패 (비치명적): %s", e)

    def _update_items_json(self, signal_id: str, title: str, slug: str, date: str, meta: Dict) -> bool:
        """
        items.json에 새 에세이 자동 추가
        Dependency Graph 연동: items.json 변경 → cascade_manager 실행
        """
        items_path = self.base_path / 'knowledge/service/items.json'

        try:
            # 기존 items 로드
            if items_path.exists():
                with open(items_path, 'r', encoding='utf-8') as f:
                    items = json.load(f)
            else:
                items = []

            # 중복 확인
            if any(item.get('item_id') == f"essay-{signal_id}" for item in items):
                logger.info("[Publisher] items.json에 이미 존재: essay-%s", signal_id)
                return True

            # 새 에세이 추가
            new_item = {
                "item_id": f"essay-{signal_id}",
                "name": title,
                "category": "essay",
                "price": 0,
                "active": True,
                "url": f"/archive/{slug}/",
                "created_at": date,
                "themes": meta.get("themes", []),
                "signal_id": signal_id
            }
            items.append(new_item)

            # 저장
            with open(items_path, 'w', encoding='utf-8') as f:
                json.dump(items, f, indent=2, ensure_ascii=False)

            logger.info("[Publisher] items.json 갱신: essay-%s 추가", signal_id)
            return True

        except Exception as e:
            logger.error("[Publisher] items.json 갱신 실패: %s", e)
            return False

    def _git_push_website(self, website_root: Path, slug: str, date: str) -> bool:
        """website/ 변경사항 git push → CF Pages 자동 배포"""
        import subprocess

        repo_root = website_root.parent  # LAYER OS 루트

        # Commit message with Co-Authored-By
        commit_msg = (
            "archive: %s (%s)\n\n"
            "Co-Authored-By: Claude <noreply@anthropic.com>"
        ) % (slug, date)

        github_token = os.getenv('GITHUB_TOKEN', '')
        github_repo = os.getenv('GITHUB_REPO', '97layer/97layerOS')

        try:
            stage_cmd = ["git", "-C", str(repo_root), "add", "website/", "knowledge/service/items.json"]
            commit_cmd = ["git", "-C", str(repo_root), "commit", "-m", commit_msg]

            for cmd in [stage_cmd, commit_cmd]:
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
                if result.returncode != 0:
                    if "nothing to commit" in result.stdout or "nothing to commit" in result.stderr:
                        continue
                    logger.warning("[Publisher] git 명령 실패: %s\n%s", ' '.join(cmd), result.stderr[:200])

            # push: GITHUB_TOKEN이 있으면 HTTPS 토큰 인증, 없으면 기존 방식
            if github_token:
                auth_url = "https://%s@github.com/%s.git" % (github_token, github_repo)
                push_cmd = ["git", "-C", str(repo_root), "push", auth_url, "main"]
                push_result = subprocess.run(push_cmd, capture_output=True, text=True, timeout=60)
                if push_result.returncode != 0:
                    logger.warning("[Publisher] git push 실패 (인증 오류 또는 네트워크): returncode=%d", push_result.returncode)
                    return False
            else:
                logger.warning("[Publisher] GITHUB_TOKEN 미설정 — 로컬 저장만 완료, push 생략")
                return True

            logger.info("[Publisher] git push 완료 → CF Pages 자동 배포 예정")
            return True

        except subprocess.TimeoutExpired:
            logger.error("[Publisher] git push 타임아웃")
            return False
        except OSError as e:
            logger.error("[Publisher] git push 실패: %s", e)
            return False

    def _update_memory(self, meta: Dict, instagram_caption: str, archive_essay: str):
        """long_term_memory.json에 발행 이력 추가"""
        memory_path = self.base_path / "knowledge" / "long_term_memory.json"
        try:
            if memory_path.exists():
                memory = json.loads(memory_path.read_text(encoding="utf-8"))
            else:
                memory = {}

            if "experiences" not in memory:
                memory["experiences"] = []
            if "published_content" not in memory:
                memory["published_content"] = []

            # 발행 이력 추가
            memory["published_content"].append({
                "signal_id": meta.get("signal_id"),
                "published_at": meta.get("published_at"),
                "themes": meta.get("themes", []),
                "cd_brand_score": meta.get("cd_brand_score", 0),
                "instagram_caption_preview": instagram_caption[:100],
                "archive_essay_preview": str(archive_essay)[:200]
            })

            # 최근 20개만 유지
            memory["published_content"] = memory["published_content"][-20:]
            memory["last_published_at"] = meta.get("published_at")

            memory_path.parent.mkdir(parents=True, exist_ok=True)
            memory_path.write_text(json.dumps(memory, indent=2, ensure_ascii=False), encoding="utf-8")
            logger.info("[Publisher] memory 업데이트 완료")

        except Exception as e:
            logger.warning("[Publisher] memory 업데이트 실패: %s", e)


if __name__ == "__main__":
    # 테스트 실행
    import sys
    logging.basicConfig(level=logging.INFO)

    publisher = ContentPublisher()
    test_payload = {
        "signal_id": "test_001",
        "instagram_caption": "느림 속에서 발견하는 것들. 서두르지 않아도 된다.",
        "hashtags": "#woohwahae #slowlife #아카이브",
        "archive_essay": "우리는 종종 속도를 미덕으로 착각한다. 하지만 진정한 깊이는 느림에서 온다. 헤어 아틀리에 WOOHWAHAE가 추구하는 것도 그런 느림이다.",
        "sa_result": {"strategic_score": 85, "themes": ["슬로우라이프", "본질"]},
        "cd_result": {"brand_score": 88, "decision": "approve"}
    }

    result = publisher.publish(test_payload)
    print(json.dumps(result, indent=2, ensure_ascii=False))
