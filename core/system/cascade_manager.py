"""
Cascade Manager — 의존성 기반 자동 전파 엔진

파일 변경 감지 → 의존성 그래프 BFS → 영향권 계산 → Tier별 처리

Author: LAYER OS
Created: 2026-02-26
"""

import json
import os
from pathlib import Path
from typing import Dict, List, Set
from dataclasses import dataclass


@dataclass
class ImpactReport:
    """영향권 분석 결과"""
    source: str
    tier: str
    affected_nodes: Set[str]
    cascade_actions: List[str]


class CascadeManager:
    """파일 변경 시 연쇄 영향 추적 및 자동 전파"""

    def __init__(self, graph_path: str = None):
        self.project_root = Path(os.getenv('PROJECT_ROOT', os.getcwd()))
        self.graph_path = graph_path or self.project_root / 'knowledge/system/dependency_graph.json'
        self.graph = self._load_graph()

    def _load_graph(self) -> Dict:
        """의존성 그래프 로드"""
        if not self.graph_path.exists():
            raise FileNotFoundError(f"Dependency graph not found: {self.graph_path}")

        with open(self.graph_path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def on_file_change(self, filepath: str) -> ImpactReport:
        """
        파일 변경 감지 → 영향권 계산 → Tier별 처리

        Args:
            filepath: 변경된 파일 경로 (상대 또는 절대)

        Returns:
            ImpactReport: 영향권 분석 결과
        """
        # 경로 정규화 (상대/절대 둘 다 지원)
        path = Path(filepath)
        if path.is_absolute():
            # 절대경로 → 상대경로 변환
            try:
                filepath = str(path.relative_to(self.project_root))
            except ValueError:
                # 프로젝트 외부 파일
                print(f"⚠️  {filepath} is outside project root. Skipping.")
                return ImpactReport(
                    source=str(path),
                    tier="UNKNOWN",
                    affected_nodes=set(),
                    cascade_actions=[]
                )
        else:
            # 이미 상대경로
            filepath = str(path)

        # 그래프에 없으면 스킵
        if filepath not in self.graph['nodes']:
            print(f"⚠️  {filepath} not in dependency graph. Skipping cascade.")
            return ImpactReport(
                source=filepath,
                tier="UNKNOWN",
                affected_nodes=set(),
                cascade_actions=[]
            )

        # 영향권 계산
        impact = self.calculate_impact(filepath)

        # Tier별 처리
        tier = self.graph['nodes'][filepath]['tier']
        if tier == "FROZEN":
            self._handle_frozen(impact)
        elif tier == "PROPOSE":
            self._handle_propose(impact)
        else:  # AUTO
            self._handle_auto(impact)

        return impact

    def calculate_impact(self, filepath: str) -> ImpactReport:
        """
        BFS로 영향 범위 계산

        Args:
            filepath: 변경된 파일 경로

        Returns:
            ImpactReport: 영향을 받는 노드 집합
        """
        node = self.graph['nodes'][filepath]
        tier = node['tier']
        cascade_rules = node.get('cascade_rules', {})

        # BFS 탐색
        visited = set()
        queue = [filepath]

        while queue:
            current = queue.pop(0)
            if current in visited:
                continue
            visited.add(current)

            # 현재 노드의 dependents 추가
            if current in self.graph['nodes']:
                for dependent in self.graph['nodes'][current].get('dependents', []):
                    # 와일드카드 확장 (예: website/**/*.html)
                    if '*' in dependent:
                        # 실제 구현 시 glob 패턴 매칭 필요
                        pass
                    else:
                        queue.append(dependent)

        # 액션 추출
        actions = []
        if 'on_change' in cascade_rules:
            actions.append(cascade_rules['on_change'])
        if 'propagate' in cascade_rules:
            actions.append(cascade_rules['propagate'])

        return ImpactReport(
            source=filepath,
            tier=tier,
            affected_nodes=visited,
            cascade_actions=actions
        )

    def _handle_frozen(self, impact: ImpactReport):
        """FROZEN Tier: CD 알림 + 승인 대기"""
        print(f"🔴 FROZEN 파일 변경 감지: {impact.source}")
        print(f"   영향 범위: {len(impact.affected_nodes)}개 노드")
        print(f"   → CD 승인 필요. 자동 전파 중단.")
        # TODO: CD 알림 전송 (Telegram/Email)

    def _handle_propose(self, impact: ImpactReport):
        """PROPOSE Tier: 에이전트 알림 + 검토 큐"""
        print(f"🟡 PROPOSE 파일 변경 감지: {impact.source}")
        print(f"   영향 범위: {len(impact.affected_nodes)}개 노드")
        print(f"   → 에이전트 재프롬프트 큐에 추가")
        # TODO: 에이전트 알림 + 검토 큐 추가

    def _handle_auto(self, impact: ImpactReport):
        """AUTO Tier: 자동 전파"""
        print(f"🟢 AUTO 파일 변경 감지: {impact.source}")
        print(f"   영향 범위: {len(impact.affected_nodes)}개 노드")
        print(f"   → 자동 전파 시작")

        for action in impact.cascade_actions:
            if action == "invalidate_cache":
                self._invalidate_cache(impact.affected_nodes)
            elif action == "regenerate_html":
                self._regenerate_html(impact.affected_nodes)
            elif action == "cf_pages_deploy":
                self._trigger_deploy()

    def _invalidate_cache(self, nodes: Set[str]):
        """캐시 무효화"""
        print(f"   └─ 캐시 무효화: {len(nodes)}개 노드")
        # TODO: filesystem_cache.json 갱신

    def _regenerate_html(self, nodes: Set[str]):
        """HTML 재생성"""
        print(f"   └─ HTML 재생성: {len(nodes)}개 파일")
        # TODO: content_publisher 호출

    def _trigger_deploy(self):
        """배포 트리거"""
        print(f"   └─ CF Pages 배포 예약")
        # TODO: git commit + push


# CLI 인터페이스
if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("Usage: python cascade_manager.py <filepath>")
        sys.exit(1)

    manager = CascadeManager()
    report = manager.on_file_change(sys.argv[1])

    print(f"\n📊 영향권 분석 결과:")
    print(f"   소스: {report.source}")
    print(f"   Tier: {report.tier}")
    print(f"   영향 노드: {len(report.affected_nodes)}개")
    print(f"   액션: {', '.join(report.cascade_actions)}")
