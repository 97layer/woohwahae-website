#!/usr/bin/env python3
"""
Dependency Analyzer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename: execution/system/dependency_analyzer.py
Purpose: Analyze code dependencies to minimize file reads
         during refactoring and development tasks
Author: 97LAYER System
Date: 2026-02-15
"""

import ast
import json
import logging
from pathlib import Path
from typing import Dict, List, Set, Optional
from collections import defaultdict

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class DependencyAnalyzer:
    """
    코드베이스의 의존성을 분석하여 필요한 파일만 읽을 수 있도록 지원
    """

    def __init__(self, project_root: str = None):
        self.root = Path(project_root) if project_root else Path(__file__).parent.parent.parent
        self.cache_file = self.root / ".tmp" / "dependency_cache.json"
        self.cache_file.parent.mkdir(parents=True, exist_ok=True)

    def analyze_file(self, file_path: str) -> Dict[str, any]:
        """
        Python 파일의 의존성 분석

        Returns:
            {
                'imports': [...],
                'classes': [...],
                'functions': [...],
                'dependencies': [...]
            }
        """
        path = Path(file_path)
        if not path.exists():
            logger.error(f"File not found: {file_path}")
            return {}

        try:
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read()

            tree = ast.parse(content)

            analysis = {
                'file': str(path),
                'imports': self._extract_imports(tree),
                'classes': self._extract_classes(tree),
                'functions': self._extract_functions(tree),
                'global_vars': self._extract_globals(tree)
            }

            return analysis

        except SyntaxError as e:
            logger.error(f"Syntax error in {file_path}: {e}")
            return {'file': str(path), 'error': str(e)}
        except Exception as e:
            logger.error(f"Error analyzing {file_path}: {e}")
            return {'file': str(path), 'error': str(e)}

    def _extract_imports(self, tree: ast.AST) -> List[Dict]:
        """Import 문 추출"""
        imports = []

        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                for alias in node.names:
                    imports.append({
                        'type': 'import',
                        'module': alias.name,
                        'alias': alias.asname,
                        'line': node.lineno
                    })
            elif isinstance(node, ast.ImportFrom):
                module = node.module or ''
                for alias in node.names:
                    imports.append({
                        'type': 'from_import',
                        'module': module,
                        'name': alias.name,
                        'alias': alias.asname,
                        'line': node.lineno
                    })

        return imports

    def _extract_classes(self, tree: ast.AST) -> List[Dict]:
        """클래스 정의 추출"""
        classes = []

        for node in ast.walk(tree):
            if isinstance(node, ast.ClassDef):
                bases = [self._get_name(base) for base in node.bases]
                methods = [
                    {
                        'name': n.name,
                        'line': n.lineno,
                        'args': [arg.arg for arg in n.args.args]
                    }
                    for n in node.body if isinstance(n, ast.FunctionDef)
                ]

                classes.append({
                    'name': node.name,
                    'line': node.lineno,
                    'bases': bases,
                    'methods': methods,
                    'decorators': [self._get_name(d) for d in node.decorator_list]
                })

        return classes

    def _extract_functions(self, tree: ast.AST) -> List[Dict]:
        """함수 정의 추출 (클래스 메소드 제외)"""
        functions = []

        for node in tree.body:  # 최상위 레벨만
            if isinstance(node, ast.FunctionDef):
                functions.append({
                    'name': node.name,
                    'line': node.lineno,
                    'args': [arg.arg for arg in node.args.args],
                    'decorators': [self._get_name(d) for d in node.decorator_list],
                    'returns': self._get_name(node.returns) if node.returns else None
                })

        return functions

    def _extract_globals(self, tree: ast.AST) -> List[Dict]:
        """전역 변수 추출"""
        globals_vars = []

        for node in tree.body:
            if isinstance(node, ast.Assign):
                for target in node.targets:
                    if isinstance(target, ast.Name):
                        globals_vars.append({
                            'name': target.id,
                            'line': node.lineno,
                            'type': 'assign'
                        })
            elif isinstance(node, ast.AnnAssign):
                if isinstance(node.target, ast.Name):
                    globals_vars.append({
                        'name': node.target.id,
                        'line': node.lineno,
                        'type': 'annotated',
                        'annotation': self._get_name(node.annotation)
                    })

        return globals_vars

    def _get_name(self, node) -> str:
        """AST 노드에서 이름 추출"""
        if node is None:
            return None
        elif isinstance(node, ast.Name):
            return node.id
        elif isinstance(node, ast.Attribute):
            return f"{self._get_name(node.value)}.{node.attr}"
        elif isinstance(node, ast.Subscript):
            return f"{self._get_name(node.value)}[...]"
        else:
            return ast.unparse(node) if hasattr(ast, 'unparse') else str(node)

    def build_dependency_graph(self, file_patterns: List[str] = None) -> Dict:
        """
        프로젝트 전체의 의존성 그래프 생성

        Args:
            file_patterns: Glob 패턴 리스트 (기본: ["**/*.py"])

        Returns:
            의존성 그래프
        """
        if file_patterns is None:
            file_patterns = ["**/*.py"]

        graph = {
            'files': {},
            'module_map': defaultdict(list),  # module name -> files
            'reverse_deps': defaultdict(set)  # file -> files that import it
        }

        # 모든 Python 파일 수집
        all_files = []
        for pattern in file_patterns:
            all_files.extend(self.root.glob(pattern))

        logger.info(f"Analyzing {len(all_files)} files...")

        # 각 파일 분석
        for file_path in all_files:
            if '.venv' in str(file_path) or '__pycache__' in str(file_path):
                continue

            rel_path = file_path.relative_to(self.root)
            analysis = self.analyze_file(file_path)

            if 'error' in analysis:
                continue

            graph['files'][str(rel_path)] = analysis

            # 모듈 매핑 구축
            module_name = str(rel_path).replace('/', '.').replace('.py', '')
            graph['module_map'][module_name].append(str(rel_path))

        # 역의존성 구축
        for file_path, analysis in graph['files'].items():
            for imp in analysis.get('imports', []):
                module = imp.get('module', '')
                if module in graph['module_map']:
                    for dep_file in graph['module_map'][module]:
                        graph['reverse_deps'][dep_file].add(file_path)

        # Set을 list로 변환 (JSON 직렬화 위해)
        graph['reverse_deps'] = {
            k: list(v) for k, v in graph['reverse_deps'].items()
        }

        logger.info(f"✓ Dependency graph built: {len(graph['files'])} files")

        return graph

    def find_affected_files(self, target_file: str, graph: Dict = None) -> Set[str]:
        """
        특정 파일을 변경할 때 영향받는 파일들 찾기

        Args:
            target_file: 변경 대상 파일
            graph: 의존성 그래프 (없으면 새로 생성)

        Returns:
            영향받는 파일 경로 집합
        """
        if graph is None:
            graph = self.build_dependency_graph()

        affected = set()
        to_check = [target_file]
        checked = set()

        while to_check:
            current = to_check.pop()
            if current in checked:
                continue

            checked.add(current)
            affected.add(current)

            # 이 파일에 의존하는 파일들 추가
            dependents = graph['reverse_deps'].get(current, [])
            to_check.extend(d for d in dependents if d not in checked)

        logger.info(f"✓ {len(affected)} files affected by changes to {target_file}")
        return affected

    def get_file_summary(self, file_path: str) -> str:
        """
        파일의 요약 정보 (시그니처만) 생성
        전체 파일을 읽지 않고도 구조 파악 가능
        """
        analysis = self.analyze_file(file_path)

        if 'error' in analysis:
            return f"# Error analyzing {file_path}: {analysis['error']}"

        summary_lines = [
            f"# File: {analysis['file']}",
            "",
            "## Imports",
        ]

        for imp in analysis.get('imports', [])[:5]:  # 처음 5개만
            if imp['type'] == 'import':
                summary_lines.append(f"  - import {imp['module']}")
            else:
                summary_lines.append(f"  - from {imp['module']} import {imp['name']}")

        if len(analysis.get('imports', [])) > 5:
            summary_lines.append(f"  ... and {len(analysis['imports']) - 5} more")

        summary_lines.extend(["", "## Classes"])
        for cls in analysis.get('classes', []):
            bases_str = f"({', '.join(cls['bases'])})" if cls['bases'] else ""
            summary_lines.append(f"  - class {cls['name']}{bases_str} [line {cls['line']}]")
            for method in cls['methods'][:3]:
                args_str = ', '.join(method['args'])
                summary_lines.append(f"      - {method['name']}({args_str})")

        summary_lines.extend(["", "## Functions"])
        for func in analysis.get('functions', []):
            args_str = ', '.join(func['args'])
            summary_lines.append(f"  - def {func['name']}({args_str}) [line {func['line']}]")

        return '\n'.join(summary_lines)

    def cache_graph(self, graph: Dict):
        """의존성 그래프 캐싱"""
        try:
            with open(self.cache_file, 'w', encoding='utf-8') as f:
                json.dump(graph, f, ensure_ascii=False, indent=2)
            logger.info(f"✓ Dependency graph cached to {self.cache_file}")
        except Exception as e:
            logger.error(f"Cache write error: {e}")

    def load_cached_graph(self) -> Optional[Dict]:
        """캐시된 의존성 그래프 로드"""
        if not self.cache_file.exists():
            return None

        try:
            with open(self.cache_file, 'r', encoding='utf-8') as f:
                graph = json.load(f)
            logger.info(f"✓ Loaded cached dependency graph")
            return graph
        except Exception as e:
            logger.error(f"Cache read error: {e}")
            return None


def main():
    """CLI 인터페이스"""
    import sys

    analyzer = DependencyAnalyzer()

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python dependency_analyzer.py analyze <file>     # Analyze single file")
        print("  python dependency_analyzer.py summary <file>     # Get file summary")
        print("  python dependency_analyzer.py graph [pattern]    # Build dependency graph")
        print("  python dependency_analyzer.py affected <file>    # Find affected files")
        return

    command = sys.argv[1]

    if command == "analyze":
        file_path = sys.argv[2]
        result = analyzer.analyze_file(file_path)
        print(json.dumps(result, indent=2, ensure_ascii=False))

    elif command == "summary":
        file_path = sys.argv[2]
        summary = analyzer.get_file_summary(file_path)
        print(summary)

    elif command == "graph":
        patterns = sys.argv[2:] if len(sys.argv) > 2 else ["**/*.py"]
        graph = analyzer.build_dependency_graph(patterns)
        analyzer.cache_graph(graph)
        print(f"\n✓ Dependency graph built and cached")
        print(f"  Files analyzed: {len(graph['files'])}")
        print(f"  Modules mapped: {len(graph['module_map'])}")

    elif command == "affected":
        file_path = sys.argv[2]
        graph = analyzer.load_cached_graph()
        if not graph:
            print("Building dependency graph first...")
            graph = analyzer.build_dependency_graph()
            analyzer.cache_graph(graph)

        affected = analyzer.find_affected_files(file_path, graph)
        print(f"\n Files affected by changes to {file_path}:")
        for f in sorted(affected):
            print(f"  - {f}")


if __name__ == "__main__":
    main()
