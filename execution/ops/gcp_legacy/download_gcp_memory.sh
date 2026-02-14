#!/bin/bash
# GCP chat_memory 수동 다운로드

echo "📥 GCP chat_memory 다운로드 중..."

# GCP 브라우저 SSH에서 실행할 명령어
cat << 'EOF'

===========================================
GCP 브라우저 SSH에서 다음 명령어 실행:
===========================================

cd ~/97layerOS
tar czf /tmp/knowledge_latest.tar.gz knowledge/
ls -lh /tmp/knowledge_latest.tar.gz

# 그 다음:
# 1. GCP SSH 톱니바퀴 → Download file
# 2. 경로: /tmp/knowledge_latest.tar.gz
# 3. Mac 다운로드 폴더에 저장됨

===========================================

다운로드 후 Mac에서 실행:

tar xzf ~/Downloads/knowledge_latest.tar.gz -C /tmp/
rsync -av /tmp/knowledge/ ~/97layerOS/knowledge/
echo "✅ 동기화 완료!"
tail -50 ~/97layerOS/knowledge/chat_memory/7565534667.json

===========================================

EOF
