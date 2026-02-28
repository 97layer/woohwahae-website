#!/bin/bash
# 로컬 프리뷰 서버 (디자인 확인용)

echo "🌐 로컬 서버 시작..."
echo "   http://localhost:8000"
echo ""
echo "종료: Ctrl+C"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

cd website && python3 -m http.server 8000