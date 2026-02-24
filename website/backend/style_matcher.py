"""
WOOHWAHAE Style Matcher
무드 키워드 → 큐레이션 레퍼런스 이미지 매칭 알고리즘

MVP: 에디터 작업 이미지 라이브러리 기반 태그 매칭
V2: THE CYCLE 에세이 콘텐츠 연동
V3: Replicate API AI 이미지 생성
"""

from typing import List, Dict

# ─── 스타일 라이브러리 ─────────────────────────────────────────
# 각 에디터 작업에 무드 태그 + 길이 태그 + 설명 부여
# 경로: /assets/img/editor/{folder}/

STYLE_LIBRARY: List[Dict] = [
    {
        "id": "wh_01",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/IMG_6589.JPG",
        "moods": ["미니멀", "자연스러운"],
        "length": ["미디엄", "롱"],
        "description": "흐르는 결, 자연스러운 무게감",
    },
    {
        "id": "wh_02",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/20251114_014807.jpg",
        "moods": ["미니멀", "클래식"],
        "length": ["숏", "미디엄"],
        "description": "정제된 실루엣, 선명한 라인",
    },
    {
        "id": "wh_03",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/20251114_014624.jpg",
        "moods": ["텍스처", "자연스러운"],
        "length": ["미디엄", "롱"],
        "description": "공기감 있는 텍스처, 움직이는 머릿결",
    },
    {
        "id": "wh_04",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/20251114_013946.jpg",
        "moods": ["볼드", "텍스처"],
        "length": ["숏"],
        "description": "강한 존재감, 뚜렷한 형태",
    },
    {
        "id": "wh_05",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/20251114_012857(1).jpg",
        "moods": ["클래식", "미니멀"],
        "length": ["미디엄"],
        "description": "시간이 지나도 유효한 균형",
    },
    {
        "id": "ref_01",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/0a641763-9207-4140-bfde-5e240354b091_rw_1200.jpeg",
        "moods": ["자연스러운", "텍스처"],
        "length": ["롱"],
        "description": "자연스럽게 쌓이는 레이어",
    },
    {
        "id": "ref_02",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/63702ded-7f95-4b24-bdb0-1e70bdc27790_rw_1200.jpeg",
        "moods": ["미니멀", "볼드"],
        "length": ["숏", "미디엄"],
        "description": "여백이 만드는 집중",
    },
    {
        "id": "ref_03",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/81f3b466-f084-4e46-8778-e75b353c230c_rw_1200.jpeg",
        "moods": ["클래식", "자연스러운"],
        "length": ["미디엄", "롱"],
        "description": "오랫동안 좋은 형태",
    },
    {
        "id": "ref_04",
        "path": "/assets/img/editor/WOOHWAHAE/Mood/c5d43b9d-7351-4155-8ad4-1059970352fa_rw_1200.jpeg",
        "moods": ["볼드", "클래식"],
        "length": ["숏"],
        "description": "결단력 있는 커트",
    },
]


def match_style(mood_keywords: List[str], length_pref: str = "") -> List[Dict]:
    """
    무드 키워드 + 선호 길이 → 최대 3개 레퍼런스 반환.

    Args:
        mood_keywords: ["미니멀", "자연스러운"] 등 복수 선택 가능
        length_pref: "숏" / "미디엄" / "롱" (선택 없으면 전체)

    Returns:
        점수 순으로 정렬된 스타일 딕셔너리 최대 3개
    """
    scored: List[tuple] = []

    for style in STYLE_LIBRARY:
        score = 0

        # 무드 키워드 매칭 (+2점/키워드)
        for mood in mood_keywords:
            if mood in style["moods"]:
                score += 2

        # 길이 선호 매칭 (+1점)
        if length_pref and length_pref in style["length"]:
            score += 1

        if score > 0:
            scored.append((score, style))

    # 점수 내림차순, 동점이면 라이브러리 순서 유지
    scored.sort(key=lambda x: x[0], reverse=True)

    # 점수 없으면 기본 3개 반환
    if not scored:
        return STYLE_LIBRARY[:3]

    return [s for _, s in scored[:3]]
