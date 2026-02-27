"""텔레그램 메시지 템플릿 단일 소스 (SSOT)

메시지 톤/형식 수정 = 이 파일 1개만 편집.
parse_mode 주의: PUBLISH_ALERT = Markdown, 나머지 = HTML.
"""

# ── 발행 알림 (content_publisher.py) ──────────────
# parse_mode: Markdown
PUBLISH_ALERT = """\
📦 *오늘의 콘텐츠 패키지*

🏷 테마: {themes}
📊 SA 전략점수: {sa_score} | CD 브랜드점수: {cd_score}

━━━━━━━━━━━━━━━
📸 *Instagram*

{caption}

{hashtags}

━━━━━━━━━━━━━━━
📝 *Archive Essay*

{essay_preview}

━━━━━━━━━━━━━━━
🗂 이미지: {image_source}"""

# ── 일일 브리핑 (telegram_secretary.py) ──────────────
# parse_mode: HTML
DAILY_BRIEFING = """\
☀️ <b>일일 브리핑 — {today}</b>

어젯밤 수집: {today_sigs}개 신호
Corpus 군집: {clusters_total}개 (발행가능 {ripe}개)
누적 발행: {published}개

{ripe_notice}"""

DAILY_BRIEFING_RIPE = "💡 <b>{ripe}개 군집이 발행 준비 완료</b>\n/publish 로 발행하세요."
DAILY_BRIEFING_IDLE = "Gardener가 03:00에 군집을 점검합니다."

# ── 발행 완료 알림 (telegram_secretary.py) ──────────────
# parse_mode: HTML
PUBLISH_COMPLETE = """\
✅ <b>발행 완료</b>

테마: {theme}{link_text}
website/archive/ 에 파일 저장됨
(도메인 연결 후 웹에서 확인 가능)"""

# ── 재방문 알림 (gardener.py) ──────────────
# parse_mode: HTML
REVISIT_ALERT_HEADER = "⏰ <b>재방문 예정 고객 {count}명</b>"
REVISIT_ALERT_ROW = "• {name} ({rhythm} 리듬)"

# ── 주간 리포트 (gardener.py) ──────────────
# parse_mode: HTML
WEEKLY_REPORT = """\
🌱 <b>Gardener 주간 리포트</b>

<b>지난 {period_days}일 현황</b>
신호 수집: {signal_count}개
SA 분석: {sa_analyzed}개
평균 전략점수: {avg_score}

<b>부상 테마</b>
{themes}

<b>핵심 개념</b>
{concepts}"""

WEEKLY_REPORT_PROPOSALS_HEADER = "<b>시스템 개선 제안 {count}건</b>"
WEEKLY_REPORT_PROPOSAL_ROW = "• {target_file}: {reason}"
WEEKLY_REPORT_PROPOSALS_FOOTER = "승인하려면 /approve, 거절하려면 /reject"
