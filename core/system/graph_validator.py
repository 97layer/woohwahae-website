"""
Graph Validator — Dependency Graph 안전성 검증

1. DAG 검증 (Directed Acyclic Graph) — 순환 참조 방지
2. 고아 노드 감지
3. 미등록 의존성 경고

Author: LAYER OS
Created: 2026-02-26
"""

import json
from pathlib import Path
from typing import Dict, Set, List


class GraphValidator:
    """의존성 그래프 안전성 검증"""

    def __init__(self, graph_path: Path):
        self.graph_path = graph_path
        with open(graph_path, 'r', encoding='utf-8') as f:
            self.graph = json.load(f)
        self.nodes = self.graph['nodes']

    def validate_all(self) -> bool:
        """전체 검증 실행"""
        issues = []

        # 1. DAG 검증 (순환 참조 감지)
        cycles = self.detect_cycles()
        if cycles:
            issues.append(f"🔴 CRITICAL: Circular dependencies detected!")
            for cycle in cycles:
                issues.append(f"   Cycle: {' → '.join(cycle)}")

        # 2. 고아 노드 감지
        orphans = self.detect_orphans()
        if orphans:
            issues.append(f"🟡 Warning: {len(orphans)} orphaned nodes (no dependents)")
            for orphan in orphans[:5]:
                issues.append(f"   - {orphan}")

        # 3. 미등록 의존성
        missing = self.detect_missing_dependencies()
        if missing:
            issues.append(f"🟡 Warning: {len(missing)} missing dependencies")
            for node, deps in list(missing.items())[:5]:
                issues.append(f"   {node} → {deps}")

        if issues:
            print("\n".join(issues))
            return len(cycles) == 0  # 순환만 critical
        else:
            print("✅ Dependency graph is safe (DAG validated)")
            return True

    def detect_cycles(self) -> List[List[str]]:
        """순환 참조 감지 (DFS)"""
        cycles = []
        visited = set()
        rec_stack = set()

        def dfs(node: str, path: List[str]):
            if node in rec_stack:
                # 순환 발견
                cycle_start = path.index(node)
                cycles.append(path[cycle_start:] + [node])
                return

            if node in visited:
                return

            visited.add(node)
            rec_stack.add(node)
            path.append(node)

            # dependents 순회
            for dependent in self.nodes.get(node, {}).get('dependents', []):
                # 와일드카드 제외
                if '*' not in dependent:
                    dfs(dependent, path[:])

            rec_stack.remove(node)

        for node in self.nodes:
            if node not in visited:
                dfs(node, [])

        return cycles

    def detect_orphans(self) -> List[str]:
        """고아 노드 감지 (dependents가 없는 노드)"""
        orphans = []
        for node, data in self.nodes.items():
            if not data.get('dependents'):
                orphans.append(node)
        return orphans

    def detect_missing_dependencies(self) -> Dict[str, List[str]]:
        """미등록 의존성 감지"""
        missing = {}
        for node, data in self.nodes.items():
            for dep in data.get('dependencies', []):
                if dep not in self.nodes and '*' not in dep:
                    if node not in missing:
                        missing[node] = []
                    missing[node].append(dep)

            for dep in data.get('dependents', []):
                if dep not in self.nodes and '*' not in dep:
                    if node not in missing:
                        missing[node] = []
                    missing[node].append(dep)

        return missing


if __name__ == "__main__":
    import sys
    from pathlib import Path

    project_root = Path(__file__).parent.parent.parent
    graph_path = project_root / 'knowledge/system/dependency_graph.json'

    if not graph_path.exists():
        print(f"❌ Graph not found: {graph_path}")
        sys.exit(1)

    validator = GraphValidator(graph_path)
    is_safe = validator.validate_all()

    sys.exit(0 if is_safe else 1)
