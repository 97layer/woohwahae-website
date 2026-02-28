#!/bin/bash
# Quick deploy to Cloudflare Pages

# 빌드 (컴포넌트 + 캐시 버스팅)
python3 core/scripts/build.py

# Git 커밋 & 푸시
git add website/
git commit -m "quick: $(date +%H:%M) 웹 업데이트"
git push origin main

echo "✅ 배포 완료! 30초 후 확인:"
echo "   https://woohwahae.kr"
echo ""
echo "캐시 문제시: Cmd+Shift+R (강제 새로고침)"