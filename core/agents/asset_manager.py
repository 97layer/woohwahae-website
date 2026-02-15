#!/usr/bin/env python3
"""
97layerOS Asset Manager
자산 추적, 분류, 개선 사이클 관리

Author: 97layerOS Technical Director
"""

import json
from pathlib import Path
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional
import sys

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Knowledge paths
KNOWLEDGE_PATHS = {
    'signals': PROJECT_ROOT / 'knowledge' / 'signals',
    'insights': PROJECT_ROOT / 'knowledge' / 'insights',
    'content': PROJECT_ROOT / 'knowledge' / 'content',
    'system': PROJECT_ROOT / 'knowledge' / 'system',
    'archive': PROJECT_ROOT / 'knowledge' / 'archive',
}


class AssetManager:
    """
    자산 관리자

    Features:
    - 자산 등록, 조회, 업데이트
    - 생명주기 추적 (captured → analyzed → refined → validated → approved → published)
    - 품질 점수 관리
    - 연관 자산 링크
    - 통계 및 보고서
    """

    # 자산 타입
    ASSET_TYPES = ["insight", "content", "visual", "code", "report"]

    # 자산 상태 (생명주기)
    ASSET_STATUSES = [
        "captured",   # 신호 포착됨
        "analyzed",   # SA가 분석 완료
        "refined",    # CE가 정제 완료
        "validated",  # Ralph Loop 검증 통과
        "approved",   # CD가 승인
        "published",  # 외부 발행 완료
        "archived"    # 아카이브됨
    ]

    # 자산 소스
    ASSET_SOURCES = ["telegram", "clipboard", "file", "agent", "parallel_orchestrator"]

    def __init__(self):
        self.registry_path = KNOWLEDGE_PATHS["system"] / "asset_registry.json"
        self.registry_path.parent.mkdir(parents=True, exist_ok=True)

        # 레지스트리 로드 또는 초기화
        if not self.registry_path.exists():
            self._init_registry()

    def _init_registry(self):
        """레지스트리 초기화"""
        registry = {
            "version": "1.0",
            "created_at": datetime.now().isoformat(),
            "assets": [],
            "stats": {
                "total": 0,
                "by_type": {},
                "by_status": {},
                "by_source": {}
            }
        }

        # Ensure directory exists
        self.registry_path.parent.mkdir(parents=True, exist_ok=True)

        with open(self.registry_path, 'w', encoding='utf-8') as f:
            json.dump(registry, f, indent=2, ensure_ascii=False)

    def _load_registry(self) -> Dict:
        """레지스트리 로드"""
        with open(self.registry_path, 'r', encoding='utf-8') as f:
            registry = json.load(f)

        # Ensure all required keys exist (backward compatibility)
        if 'stats' not in registry:
            registry['stats'] = {}

        if 'by_source' not in registry['stats']:
            registry['stats']['by_source'] = {}

        if 'by_type' not in registry['stats']:
            registry['stats']['by_type'] = {}

        if 'by_status' not in registry['stats']:
            registry['stats']['by_status'] = {}

        return registry

    def _save_registry(self, registry: Dict):
        """레지스트리 저장"""
        with open(self.registry_path, 'w', encoding='utf-8') as f:
            json.dump(registry, f, indent=2, ensure_ascii=False)

    def register_asset(
        self,
        path: str,
        asset_type: str,
        source: str,
        metadata: Optional[Dict] = None,
        initial_status: str = "captured"
    ) -> str:
        """
        자산 등록

        Args:
            path: 파일 경로
            asset_type: insight, content, visual, code, report
            source: telegram, clipboard, file, agent, parallel_orchestrator
            metadata: 추가 메타데이터
            initial_status: 초기 상태 (기본: captured)

        Returns:
            asset_id (e.g., AST-2026-02-001)
        """
        # 검증
        if asset_type not in self.ASSET_TYPES:
            raise ValueError(f"Invalid asset_type: {asset_type}. Must be one of {self.ASSET_TYPES}")

        if initial_status not in self.ASSET_STATUSES:
            raise ValueError(f"Invalid initial_status: {initial_status}. Must be one of {self.ASSET_STATUSES}")

        registry = self._load_registry()

        # Generate asset ID
        asset_id = f"AST-{datetime.now().strftime('%Y-%m')}-{registry['stats']['total']+1:03d}"

        # Create asset entry
        asset = {
            "id": asset_id,
            "type": asset_type,
            "source": source,
            "created_at": datetime.now().isoformat(),
            "updated_at": datetime.now().isoformat(),
            "path": path,
            "status": initial_status,
            "quality_score": 0.0,
            "metadata": metadata or {},
            "lifecycle": [
                {
                    "stage": initial_status,
                    "at": datetime.now().isoformat(),
                    "by": source
                }
            ],
            "linked_assets": []
        }

        # Update registry
        registry['assets'].append(asset)
        registry['stats']['total'] += 1
        registry['stats']['by_type'][asset_type] = registry['stats']['by_type'].get(asset_type, 0) + 1
        registry['stats']['by_status'][initial_status] = registry['stats']['by_status'].get(initial_status, 0) + 1
        registry['stats']['by_source'][source] = registry['stats']['by_source'].get(source, 0) + 1

        self._save_registry(registry)

        print(f"📦 Asset registered: {asset_id} ({asset_type}, {initial_status})")
        return asset_id

    def get_asset(self, asset_id: str) -> Optional[Dict]:
        """자산 조회"""
        registry = self._load_registry()

        for asset in registry['assets']:
            if asset['id'] == asset_id:
                return asset

        return None

    def update_asset_status(
        self,
        asset_id: str,
        new_status: str,
        updated_by: str,
        quality_score: Optional[float] = None
    ) -> bool:
        """
        자산 상태 업데이트

        Args:
            asset_id: 자산 ID
            new_status: 새 상태
            updated_by: 업데이트 주체 (에이전트 ID)
            quality_score: 품질 점수 (0-100)

        Returns:
            성공 여부
        """
        if new_status not in self.ASSET_STATUSES:
            raise ValueError(f"Invalid new_status: {new_status}")

        registry = self._load_registry()

        for asset in registry['assets']:
            if asset['id'] == asset_id:
                old_status = asset['status']

                # 상태 업데이트
                asset['status'] = new_status
                asset['updated_at'] = datetime.now().isoformat()

                # 생명주기 기록
                asset['lifecycle'].append({
                    "stage": new_status,
                    "at": datetime.now().isoformat(),
                    "by": updated_by
                })

                # 품질 점수 업데이트
                if quality_score is not None:
                    asset['quality_score'] = quality_score

                # 통계 업데이트
                registry['stats']['by_status'][old_status] -= 1
                registry['stats']['by_status'][new_status] = registry['stats']['by_status'].get(new_status, 0) + 1

                self._save_registry(registry)

                print(f"✅ Asset {asset_id}: {old_status} → {new_status}")
                if quality_score is not None:
                    print(f"   Quality score: {quality_score}/100")

                return True

        print(f"❌ Asset not found: {asset_id}")
        return False

    def link_assets(self, asset_id: str, related_asset_ids: List[str]) -> bool:
        """
        자산 간 연관 관계 설정

        Args:
            asset_id: 기준 자산 ID
            related_asset_ids: 관련 자산 ID 리스트

        Returns:
            성공 여부
        """
        registry = self._load_registry()

        for asset in registry['assets']:
            if asset['id'] == asset_id:
                asset['linked_assets'].extend(related_asset_ids)
                asset['linked_assets'] = list(set(asset['linked_assets']))  # 중복 제거
                asset['updated_at'] = datetime.now().isoformat()

                self._save_registry(registry)

                print(f"🔗 Linked {len(related_asset_ids)} assets to {asset_id}")
                return True

        print(f"❌ Asset not found: {asset_id}")
        return False

    def get_assets_by_status(self, status: str) -> List[Dict]:
        """특정 상태의 자산 조회"""
        registry = self._load_registry()

        return [asset for asset in registry['assets'] if asset['status'] == status]

    def get_assets_by_type(self, asset_type: str) -> List[Dict]:
        """특정 타입의 자산 조회"""
        registry = self._load_registry()

        return [asset for asset in registry['assets'] if asset['type'] == asset_type]

    def get_recent_assets(self, days: int = 7, limit: int = 10) -> List[Dict]:
        """최근 자산 조회"""
        registry = self._load_registry()

        cutoff_date = datetime.now() - timedelta(days=days)

        recent = [
            asset for asset in registry['assets']
            if datetime.fromisoformat(asset['created_at']) > cutoff_date
        ]

        # 최신순 정렬
        recent.sort(key=lambda x: x['created_at'], reverse=True)

        return recent[:limit]

    def get_stats(self) -> Dict:
        """통계 조회"""
        registry = self._load_registry()
        return registry['stats']

    def generate_report(self, output_path: Optional[Path] = None) -> str:
        """
        자산 관리 보고서 생성

        Args:
            output_path: 출력 경로 (없으면 콘솔 출력)

        Returns:
            보고서 내용
        """
        registry = self._load_registry()
        stats = registry['stats']

        report = f"""# 97layerOS Asset Management Report
Generated: {datetime.now().strftime('%Y-%m-%d %H:%M')}

## 📊 Overall Statistics

- **Total Assets**: {stats['total']}
- **By Type**:
{self._format_dict(stats['by_type'])}
- **By Status**:
{self._format_dict(stats['by_status'])}
- **By Source**:
{self._format_dict(stats['by_source'])}

## 🔥 Recent Activity (Last 7 days)

{self._format_recent_assets(self.get_recent_assets())}

## ⏳ Pending Assets

{self._format_pending_assets()}

## ✅ Top Quality Assets

{self._format_top_quality_assets()}

---
Report generated by 97layerOS Asset Manager
"""

        if output_path:
            output_path.parent.mkdir(parents=True, exist_ok=True)
            output_path.write_text(report, encoding='utf-8')
            print(f"📄 Report saved to: {output_path}")

        return report

    def _format_dict(self, d: Dict) -> str:
        """딕셔너리를 Markdown 리스트로 변환"""
        if not d:
            return "  - None"
        return '\n'.join([f"  - {k}: {v}" for k, v in d.items()])

    def _format_recent_assets(self, assets: List[Dict]) -> str:
        """최근 자산을 Markdown 테이블로 변환"""
        if not assets:
            return "No recent assets."

        lines = ["| ID | Type | Status | Created |", "|:---|:---|:---|:---|"]

        for asset in assets:
            created = datetime.fromisoformat(asset['created_at']).strftime('%m-%d %H:%M')
            lines.append(f"| {asset['id']} | {asset['type']} | {asset['status']} | {created} |")

        return '\n'.join(lines)

    def _format_pending_assets(self) -> str:
        """미완료 자산 포맷"""
        pending_statuses = ["captured", "analyzed", "refined", "validated"]
        pending_assets = []

        registry = self._load_registry()
        for asset in registry['assets']:
            if asset['status'] in pending_statuses:
                pending_assets.append(asset)

        if not pending_assets:
            return "All assets completed or archived."

        lines = ["| ID | Type | Status | Age (hours) |", "|:---|:---|:---|:---|"]

        for asset in pending_assets:
            created = datetime.fromisoformat(asset['created_at'])
            age_hours = int((datetime.now() - created).total_seconds() / 3600)
            lines.append(f"| {asset['id']} | {asset['type']} | {asset['status']} | {age_hours}h |")

        return '\n'.join(lines)

    def _format_top_quality_assets(self, limit: int = 5) -> str:
        """상위 품질 자산 포맷"""
        registry = self._load_registry()

        # 품질 점수 순 정렬
        sorted_assets = sorted(
            [a for a in registry['assets'] if a['quality_score'] > 0],
            key=lambda x: x['quality_score'],
            reverse=True
        )[:limit]

        if not sorted_assets:
            return "No assets with quality scores yet."

        lines = ["| ID | Type | Score | Status |", "|:---|:---|:---|:---|"]

        for asset in sorted_assets:
            lines.append(f"| {asset['id']} | {asset['type']} | {asset['quality_score']}/100 | {asset['status']} |")

        return '\n'.join(lines)


# ─────────────────────────────────────────────────────────────────
# CLI Interface
# ─────────────────────────────────────────────────────────────────

def main():
    import argparse

    parser = argparse.ArgumentParser(description="97layerOS Asset Manager")
    parser.add_argument('--register', nargs=3, metavar=('PATH', 'TYPE', 'SOURCE'), help='자산 등록')
    parser.add_argument('--get', type=str, help='자산 조회')
    parser.add_argument('--update-status', nargs=3, metavar=('ASSET_ID', 'STATUS', 'UPDATED_BY'), help='상태 업데이트')
    parser.add_argument('--quality-score', type=float, help='품질 점수 (--update-status와 함께 사용)')
    parser.add_argument('--link', nargs='+', metavar='ASSET_ID', help='자산 연결 (첫 번째가 기준)')
    parser.add_argument('--stats', action='store_true', help='통계 조회')
    parser.add_argument('--report', type=str, help='보고서 생성 (출력 경로)')

    args = parser.parse_args()

    manager = AssetManager()

    if args.register:
        path, asset_type, source = args.register
        asset_id = manager.register_asset(path, asset_type, source)
        print(f"✅ Registered: {asset_id}")

    elif args.get:
        asset = manager.get_asset(args.get)
        if asset:
            print(json.dumps(asset, indent=2, ensure_ascii=False))
        else:
            print(f"❌ Asset not found: {args.get}")

    elif args.update_status:
        asset_id, status, updated_by = args.update_status
        quality_score = args.quality_score if args.quality_score else None
        manager.update_asset_status(asset_id, status, updated_by, quality_score)

    elif args.link:
        if len(args.link) < 2:
            print("❌ Need at least 2 asset IDs to link")
        else:
            base_id = args.link[0]
            related_ids = args.link[1:]
            manager.link_assets(base_id, related_ids)

    elif args.stats:
        stats = manager.get_stats()
        print(json.dumps(stats, indent=2, ensure_ascii=False))

    elif args.report:
        manager.generate_report(Path(args.report))

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
