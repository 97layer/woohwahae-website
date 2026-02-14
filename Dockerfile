# Google Cloud Run을 위한 Dockerfile
FROM python:3.9-slim

WORKDIR /app

# 시스템 의존성 설치
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Python 의존성 파일 복사 및 설치
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 프로젝트 파일 복사
COPY . .

# 환경변수 설정
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Flask 서버 실행
CMD ["python", "execution/telegram_webhook.py"]
