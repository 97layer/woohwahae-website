# 97layerOS Cortex Engine Dockerfile
FROM python:3.9-slim

# OS 의존성 및 주기적 태스크를 위한 cron 등 설치
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 의존성 설치
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 프로젝트 전체 복사
COPY . .

# 실행 권한 부여
RUN chmod +x core/system/signal_processor.py \
    core/daemons/telegram_secretary.py \
    core/admin/app.py \
    core/daemons/dashboard_server.py

# 환경 변수 기본값
ENV PYTHONPATH=/app
ENV PORT=5001

EXPOSE 5001 8000 8080

# 컴포즈에서 덮어쓸 예정이나 기본 실행은 엔트리포인트로 설정 가능
CMD ["python3", "core/admin/app.py"]
