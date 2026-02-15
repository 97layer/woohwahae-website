#!/usr/bin/env python3
"""
🎯 97layerOS 파일 통합 스크립트
281개 MD 파일 → 50개 이하로 통합
"""

import os
import glob
from pathlib import Path
from datetime import datetime
import re

class FileConsolidator:
    def __init__(self):
        self.base_path = Path("/Users/97layer/97layerOS")
        self.signals_path = self.base_path / "knowledge/signals"
        self.content_path = self.base_path / "knowledge/content"
        self.consolidated_count = 0
        self.deleted_count = 0

    def consolidate_telegram_by_date(self):
        """Telegram 대화를 날짜별로 통합"""
        print("\n📱 Telegram 대화 통합 시작...")

        # 날짜별 그룹화
        date_groups = {
            "20260213": [],
            "20260214": [],
            "20260215": []
        }

        telegram_files = glob.glob(str(self.signals_path / "*telegram*.md"))

        for file in telegram_files:
            content = Path(file).read_text(encoding='utf-8')
            filename = os.path.basename(file)

            # 날짜 추출
            for date in date_groups.keys():
                if date in filename:
                    date_groups[date].append({
                        'file': filename,
                        'content': content,
                        'timestamp': self.extract_timestamp(filename)
                    })
                    break

        # 날짜별 통합 파일 생성
        for date, files in date_groups.items():
            if not files:
                continue

            # 시간순 정렬
            files.sort(key=lambda x: x['timestamp'])

            # 통합 파일 생성
            consolidated_file = self.signals_path / f"telegram_conversations_{date}.md"

            with open(consolidated_file, 'w', encoding='utf-8') as f:
                f.write(f"# Telegram Conversations - {date}\n\n")
                f.write(f"**통합일**: {datetime.now().strftime('%Y-%m-%d %H:%M')}\n")
                f.write(f"**원본 파일 수**: {len(files)}개\n\n")
                f.write("---\n\n")

                for idx, item in enumerate(files, 1):
                    f.write(f"## {idx}. {item['timestamp']} (원본: {item['file']})\n\n")
                    f.write(item['content'])
                    f.write("\n\n---\n\n")

            # 원본 파일 삭제
            for item in files:
                os.remove(self.signals_path / item['file'])
                self.deleted_count += 1

            print(f"✅ {date}: {len(files)}개 파일 → 1개로 통합")
            self.consolidated_count += 1

    def consolidate_council_logs(self):
        """Council 로그를 주제별로 통합"""
        print("\n🏛️ Council 로그 통합 시작...")

        # council_log 폴더 찾기
        council_files = []
        if (self.content_path / "council_log").exists():
            council_files = glob.glob(str(self.content_path / "council_log" / "*.md"))

        if council_files:
            # logs 폴더 생성
            logs_path = self.content_path / "logs"
            logs_path.mkdir(exist_ok=True)

            # 날짜별 그룹화
            date_groups = {}
            for file in council_files:
                content = Path(file).read_text(encoding='utf-8')
                filename = os.path.basename(file)

                # 날짜 추출
                date_match = re.search(r'(\d{8})', filename)
                if date_match:
                    date = date_match.group(1)
                    if date not in date_groups:
                        date_groups[date] = []
                    date_groups[date].append({
                        'file': filename,
                        'content': content,
                        'path': file
                    })

            # 날짜별 통합
            for date, files in date_groups.items():
                consolidated_file = logs_path / f"council_{date}_consolidated.md"

                with open(consolidated_file, 'w', encoding='utf-8') as f:
                    f.write(f"# Council Logs - {date}\n\n")
                    f.write(f"**통합일**: {datetime.now().strftime('%Y-%m-%d %H:%M')}\n")
                    f.write(f"**원본 파일 수**: {len(files)}개\n\n")
                    f.write("---\n\n")

                    for idx, item in enumerate(files, 1):
                        f.write(f"## Session {idx}: {item['file']}\n\n")
                        f.write(item['content'])
                        f.write("\n\n---\n\n")

                # 원본 파일 삭제
                for item in files:
                    os.remove(item['path'])
                    self.deleted_count += 1

                print(f"✅ {date}: {len(files)}개 council 로그 → 1개로 통합")
                self.consolidated_count += 1

    def consolidate_blueprints(self):
        """Minimal Life 블루프린트 통합"""
        print("\n📘 Minimal Life 블루프린트 통합 시작...")

        blueprint_files = glob.glob(str(self.content_path / "minimal_life*.md"))

        if blueprint_files:
            # 통합 파일 생성
            consolidated_file = self.content_path / "minimal_life_complete_guide.md"

            with open(consolidated_file, 'w', encoding='utf-8') as f:
                f.write("# Minimal Life - Complete Guide v3.0\n\n")
                f.write(f"**통합일**: {datetime.now().strftime('%Y-%m-%d')}\n")
                f.write(f"**원본 파일 수**: {len(blueprint_files)}개\n\n")
                f.write("---\n\n")

                for file in blueprint_files:
                    filename = os.path.basename(file)
                    content = Path(file).read_text(encoding='utf-8')

                    # 섹션 제목 결정
                    if "strategy" in filename.lower():
                        section = "## Part 1: Strategy & Insight"
                    elif "visual" in filename.lower():
                        section = "## Part 2: Visual Guide"
                    elif "narrative" in filename.lower():
                        section = "## Part 3: Narrative"
                    elif "tech" in filename.lower():
                        section = "## Part 4: Technical Blueprint"
                    else:
                        section = f"## {filename}"

                    f.write(f"{section}\n\n")
                    f.write(f"*원본: {filename}*\n\n")
                    f.write(content)
                    f.write("\n\n---\n\n")

            # 원본 파일 삭제
            for file in blueprint_files:
                os.remove(file)
                self.deleted_count += 1

            print(f"✅ Minimal Life: {len(blueprint_files)}개 → 1개로 통합")
            self.consolidated_count += 1

    def extract_timestamp(self, filename):
        """파일명에서 타임스탬프 추출"""
        match = re.search(r'(\d{8}_\d{6})', filename)
        if match:
            return match.group(1)
        return filename

    def final_report(self):
        """최종 보고서"""
        print("\n" + "="*50)
        print("📊 통합 완료 보고서")
        print("="*50)

        # 현재 MD 파일 수 계산
        all_md_files = glob.glob(str(self.base_path / "**/*.md"), recursive=True)

        # .git 폴더 제외
        all_md_files = [f for f in all_md_files if ".git" not in f and "node_modules" not in f]

        print(f"✅ 통합된 파일 그룹: {self.consolidated_count}개")
        print(f"🗑️ 삭제된 파일: {self.deleted_count}개")
        print(f"📁 전체 MD 파일: {len(all_md_files)}개")

        if len(all_md_files) <= 100:
            print("\n🎉 목표 달성! 100개 이하로 축소 성공")
        else:
            print(f"\n⚠️ 추가 정리 필요: {len(all_md_files) - 100}개 더 줄여야 함")

    def run(self):
        """전체 통합 프로세스 실행"""
        print("🚀 97layerOS 파일 통합 시작...\n")

        # 1. Telegram 대화 통합
        self.consolidate_telegram_by_date()

        # 2. Council 로그 통합
        self.consolidate_council_logs()

        # 3. Blueprint 통합
        self.consolidate_blueprints()

        # 4. 최종 보고
        self.final_report()


if __name__ == "__main__":
    consolidator = FileConsolidator()
    consolidator.run()