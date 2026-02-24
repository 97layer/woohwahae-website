#!/usr/bin/env python3
"""
silhouette_renderer.py — 고객 데이터 기반 헤어 실루엣 SVG 생성기 v2

입력: 고객 데이터 딕셔너리 (visits, preference_notes, hair_type)
출력: SVG 문자열 (인라인 임베드용)

알고리즘:
- 방문 이력 + 상담 텍스트에서 파라미터 누적 추출
- 무드·길이·기질 → 베지어 제어점 생성 (해부학적 헤어라인 기반)
- 방문 횟수 → 실루엣 진화 (윤곽 → 질감 → 가르마 → 정체성)
"""

import random

# ── 캔버스 상수 ─────────────────────────────────────────────────
W = 200
H = 340
CX = 100
HEAD_CY = 128
HEAD_RX = 48
HEAD_RY = 58

# ── 키워드 매핑 ─────────────────────────────────────────────────
LENGTH_MAP = {"숏": 0.12, "미디엄": 0.48, "롱": 0.88}

# bow: 옆으로 부풀어 나오는 폭 (HEAD_RX 기준 배율)
# crown_lift: 크라운이 머리 위로 얼마나 솟는가
# taper: 하단이 좁아지는 정도 (0=펼쳐짐, 1=좁아짐)
# flat_top: 크라운 직선 여부
# side_curve: 측면 호의 강도 (볼 부위 바깥 곡선)
MOOD_PARAMS = {
    "미니멀":    {"bow": 0.00, "crown_lift": 0.10, "taper": 0.70, "flat_top": True,  "side_curve": 0.10, "noise": 0.0},
    "자연스러운": {"bow": 0.12, "crown_lift": 0.18, "taper": 0.30, "flat_top": False, "side_curve": 0.28, "noise": 0.0},
    "텍스처":   {"bow": 0.18, "crown_lift": 0.20, "taper": 0.25, "flat_top": False, "side_curve": 0.32, "noise": 5.0},
    "볼드":     {"bow": 0.30, "crown_lift": 0.08, "taper": 0.10, "flat_top": True,  "side_curve": 0.15, "noise": 0.0},
    "클래식":   {"bow": 0.08, "crown_lift": 0.22, "taper": 0.45, "flat_top": False, "side_curve": 0.20, "noise": 0.0},
}

DEFAULT_PARAMS = MOOD_PARAMS["자연스러운"]


def _extract_params(client: dict) -> dict:
    """누적 상담 텍스트 + 방문 이력에서 렌더링 파라미터 추출."""
    notes = client.get("preference_notes", "")
    visits = client.get("visits", [])
    visit_count = len([v for v in visits if v.get("service")])

    length = 0.45
    for kw, val in LENGTH_MAP.items():
        if kw in notes:
            length = val

    mood_score: dict = {m: 0.0 for m in MOOD_PARAMS}
    for mood in MOOD_PARAMS:
        mood_score[mood] += notes.count(mood) * 1.0
        for v in visits:
            combined = v.get("service", "") + " " + v.get("public_note", "")
            if mood in combined:
                mood_score[mood] += 0.5

    ranked = sorted(mood_score.items(), key=lambda x: x[1], reverse=True)
    dominant = [m for m, s in ranked if s > 0][:2]
    if not dominant:
        dominant = ["자연스러운"]

    return {
        "length": length,
        "moods": dominant,
        "visit_count": visit_count,
        "seed": hash(client.get("client_id", "x")) % 9999,
    }


def _lerp(a, b, t):
    return a + (b - a) * t


def _f(v):
    return round(v, 1)


def _build_path(length: float, params: dict, seed: int) -> str:
    """
    해부학적 헤어 실루엣 베지어 패스.

    경로: 좌측 템플 → 측면 호(볼 외곽) → 좌측 하단 → 하단 호 →
          우측 하단 → 측면 호 → 우측 템플 → 크라운 호 → 좌측 템플
    """
    rng = random.Random(seed)
    bow = params["bow"]
    crown_lift = params["crown_lift"]
    taper = params["taper"]
    flat_top = params["flat_top"]
    sc = params["side_curve"]
    amp = params["noise"]

    def n(scale=1.0):
        return (rng.random() - 0.5) * 2 * amp * scale

    # ── 앵커 포인트 ──────────────────────────────────────────────

    # 템플: 헤어라인이 이마에서 시작하는 좌우 지점
    temple_y = HEAD_CY - HEAD_RY * 0.28
    temple_lx = CX - HEAD_RX * 0.88
    temple_rx = CX + HEAD_RX * 0.88

    # 볼 외곽 (측면 가장 넓은 지점)
    bow_y = HEAD_CY + HEAD_RY * 0.15
    bow_lx = temple_lx - HEAD_RX * bow
    bow_rx = temple_rx + HEAD_RX * bow

    # 하단 좌우 X: 길이에 따라 테이퍼(좁아짐) 정도가 다름
    bot_y_min = HEAD_CY + HEAD_RY + 8
    bot_y_max = HEAD_CY + HEAD_RY + 155
    bot_y = min(_lerp(bot_y_min, bot_y_max, length), H - 8)

    base_spread = HEAD_RX * _lerp(0.62, 0.82, length)
    spread_l = base_spread * _lerp(1.0, 1 - taper * 0.5, length)
    spread_r = spread_l
    bot_lx = CX - spread_l + n(0.5)
    bot_rx = CX + spread_r + n(0.5)

    # 크라운
    crown_y = HEAD_CY - HEAD_RY - 6 - crown_lift * HEAD_RY

    # ── 제어점 ────────────────────────────────────────────────────

    # 좌측 템플 → 볼 외곽
    cl1x = temple_lx - HEAD_RX * sc * 0.6 + n()
    cl1y = temple_y + (bow_y - temple_y) * 0.4 + n(0.6)
    cl2x = bow_lx - HEAD_RX * sc * 0.3 + n()
    cl2y = bow_y - (bow_y - temple_y) * 0.15 + n(0.6)

    # 볼 외곽 → 하단 좌
    cl3x = bow_lx - HEAD_RX * sc * 0.2 + n()
    cl3y = bow_y + (bot_y - bow_y) * 0.35 + n(0.6)
    cl4x = bot_lx - HEAD_RX * 0.12 + n()
    cl4y = bot_y - (bot_y - bow_y) * 0.08 + n(0.6)

    # 하단 호
    cb1x = bot_lx + spread_l * 0.35 + n(0.4)
    cb1y = bot_y + HEAD_RX * 0.12 + n(0.4)
    cb2x = bot_rx - spread_r * 0.35 + n(0.4)
    cb2y = bot_y + HEAD_RX * 0.12 + n(0.4)

    # 하단 우 → 볼 외곽
    cr1x = bot_rx + HEAD_RX * 0.12 + n()
    cr1y = bot_y - (bot_y - bow_y) * 0.08 + n(0.6)
    cr2x = bow_rx + HEAD_RX * sc * 0.2 + n()
    cr2y = bow_y + (bot_y - bow_y) * 0.35 + n(0.6)

    # 볼 외곽 → 우측 템플
    cr3x = bow_rx + HEAD_RX * sc * 0.3 + n()
    cr3y = bow_y - (bow_y - temple_y) * 0.15 + n(0.6)
    cr4x = temple_rx + HEAD_RX * sc * 0.6 + n()
    cr4y = temple_y + (bow_y - temple_y) * 0.4 + n(0.6)

    # 크라운
    if flat_top:
        cc1x = temple_rx - HEAD_RX * 0.22;  cc1y = crown_y + 8
        cc2x = CX + HEAD_RX * 0.20;          cc2y = crown_y
        cc3x = CX - HEAD_RX * 0.20;          cc3y = crown_y
        cc4x = temple_lx + HEAD_RX * 0.22;   cc4y = crown_y + 8
    else:
        cc1x = temple_rx - HEAD_RX * 0.26;   cc1y = crown_y + HEAD_RY * 0.26
        cc2x = CX + HEAD_RX * 0.50;          cc2y = crown_y - HEAD_RY * 0.06
        cc3x = CX - HEAD_RX * 0.50;          cc3y = crown_y - HEAD_RY * 0.06
        cc4x = temple_lx + HEAD_RX * 0.26;   cc4y = crown_y + HEAD_RY * 0.26

    f = _f
    return (
        # 좌측 템플 시작
        f"M {f(temple_lx)} {f(temple_y)} "
        # → 볼 외곽
        f"C {f(cl1x)} {f(cl1y)}, {f(cl2x)} {f(cl2y)}, {f(bow_lx)} {f(bow_y)} "
        # → 하단 좌
        f"C {f(cl3x)} {f(cl3y)}, {f(cl4x)} {f(cl4y)}, {f(bot_lx)} {f(bot_y)} "
        # 하단 호
        f"C {f(cb1x)} {f(cb1y)}, {f(cb2x)} {f(cb2y)}, {f(bot_rx)} {f(bot_y)} "
        # → 볼 외곽
        f"C {f(cr1x)} {f(cr1y)}, {f(cr2x)} {f(cr2y)}, {f(bow_rx)} {f(bow_y)} "
        # → 우측 템플
        f"C {f(cr3x)} {f(cr3y)}, {f(cr4x)} {f(cr4y)}, {f(temple_rx)} {f(temple_y)} "
        # 크라운: 우 → 정수리 → 좌
        f"C {f(cc1x)} {f(cc1y)}, {f(cc2x)} {f(cc2y)}, {f(CX)} {f(crown_y)} "
        f"C {f(cc3x)} {f(cc3y)}, {f(cc4x)} {f(cc4y)}, {f(temple_lx)} {f(temple_y)} Z"
    )


def _build_texture_strands(length: float, seed: int, count: int = 7) -> str:
    """텍스처 무드: 헤어 실루엣 내부 결 선 (얇은 베지어)."""
    rng = random.Random(seed + 100)
    bot_y_max = HEAD_CY + HEAD_RY + 155
    bot_y = min(HEAD_CY + HEAD_RY + 8 + (bot_y_max - HEAD_CY - HEAD_RY - 8) * length, H - 8)
    strands = []
    for _ in range(count):
        sx = CX + (rng.random() - 0.5) * HEAD_RX * 1.4
        sy = HEAD_CY - HEAD_RY * 0.1
        ex = sx + (rng.random() - 0.5) * HEAD_RX * 0.5
        ey = HEAD_CY + HEAD_RY + (bot_y - HEAD_CY - HEAD_RY) * (0.4 + rng.random() * 0.5)
        cx1 = sx + (rng.random() - 0.5) * 12
        cy1 = sy + (ey - sy) * 0.3
        cx2 = ex + (rng.random() - 0.5) * 12
        cy2 = sy + (ey - sy) * 0.7
        strands.append(
            f'<path d="M {_f(sx)} {_f(sy)} C {_f(cx1)} {_f(cy1)}, {_f(cx2)} {_f(cy2)}, {_f(ex)} {_f(ey)}" '
            f'fill="none" stroke="#8A8980" stroke-width="0.6" opacity="0.45"/>'
        )
    return "\n".join(strands)


def _build_parting(crown_y: float) -> str:
    """방문 3회 이상: 가르마 (크라운에서 내려오는 중앙 선)."""
    x = CX
    y1 = crown_y + 4
    y2 = HEAD_CY - HEAD_RY + 12
    return (
        f'<line x1="{_f(x)}" y1="{_f(y1)}" x2="{_f(x)}" y2="{_f(y2)}" '
        f'stroke="#C8C7C0" stroke-width="0.8" stroke-dasharray="2 2" opacity="0.6"/>'
    )


def generate_silhouette(client: dict) -> str:
    """
    고객 데이터 → SVG 문자열.
    Jinja2 템플릿에서 {{ silhouette|safe }} 로 사용.
    """
    params_data = _extract_params(client)
    primary_mood = params_data["moods"][0]
    mp = MOOD_PARAMS.get(primary_mood, DEFAULT_PARAMS)
    vc = params_data["visit_count"]
    length = params_data["length"]
    seed = params_data["seed"]

    path_d = _build_path(length, mp, seed)

    # 크라운 Y (가르마 위치용)
    crown_y = HEAD_CY - HEAD_RY - 6 - mp["crown_lift"] * HEAD_RY
    neck_y = HEAD_CY + HEAD_RY - 4

    # ── 진화 단계 ─────────────────────────────────────────────────
    if vc == 0:
        # 점선 윤곽 — 가능성
        hair_fill = "none"
        hair_stroke = "#B8B7B0"
        hair_sw = 0.9
        dash = 'stroke-dasharray="6 3"'
        head_fill = "none"
        head_stroke = "#C0BFBA"
        opacity = 0.60
        anim_dur = "2.2s"
    elif vc <= 2:
        # 실선 윤곽 — 시작
        hair_fill = "none"
        hair_stroke = "#7A7972"
        hair_sw = 1.1
        dash = ""
        head_fill = "#FAF9F6"
        head_stroke = "#7A7972"
        opacity = 0.85
        anim_dur = "1.8s"
    elif vc <= 5:
        # 연한 면 — 축적
        hair_fill = "#E8E7E0"
        hair_stroke = "#4A4A44"
        hair_sw = 0.7
        dash = ""
        head_fill = "#FAF9F6"
        head_stroke = "#4A4A44"
        opacity = 1.0
        anim_dur = "1.4s"
    else:
        # 진한 면 — 정체성
        hair_fill = "#1B2D4F"
        hair_stroke = "#1B2D4F"
        hair_sw = 0.5
        dash = ""
        head_fill = "#FAF9F6"
        head_stroke = "#1B2D4F"
        opacity = 1.0
        anim_dur = "1.2s"

    # ── 추가 디테일 레이어 ────────────────────────────────────────
    texture = ""
    if primary_mood == "텍스처" and vc >= 1:
        texture = _build_texture_strands(length, seed)

    parting = ""
    if vc >= 3:
        parting = _build_parting(crown_y)

    # ── SVG 조립 ─────────────────────────────────────────────────
    svg_id = f"sil_{seed}"

    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W} {H}" '
        f'id="{svg_id}" style="width:100%;max-width:180px;display:block;">'
        f'<style>'
        f'@keyframes draw_{svg_id} {{'
        f'  from {{ stroke-dashoffset: 800; opacity: 0; }}'
        f'  to   {{ stroke-dashoffset: 0;   opacity: {opacity}; }}'
        f'}}'
        f'#{svg_id} .hair-path {{'
        f'  stroke-dasharray: 800;'
        f'  stroke-dashoffset: 0;'
        f'  animation: draw_{svg_id} {anim_dur} ease-out forwards;'
        f'}}'
        f'</style>'

        # 헤어 실루엣
        f'<path class="hair-path" d="{path_d}" fill="{hair_fill}" stroke="{hair_stroke}" '
        f'stroke-width="{hair_sw}" stroke-linejoin="round" stroke-linecap="round" '
        f'{dash} opacity="{opacity}"/>'

        # 텍스처 결 선
        f'{texture}'

        # 가르마
        f'{parting}'

        # 목
        f'<rect x="88" y="{_f(neck_y)}" width="24" height="16" '
        f'fill="{head_fill}" stroke="{head_stroke}" stroke-width="0.7" opacity="{opacity}"/>'

        # 얼굴 타원
        f'<ellipse cx="{CX}" cy="{HEAD_CY}" rx="{HEAD_RX}" ry="{HEAD_RY}" '
        f'fill="{head_fill}" stroke="{head_stroke}" stroke-width="0.7" opacity="{opacity}"/>'

        f'</svg>'
    )
