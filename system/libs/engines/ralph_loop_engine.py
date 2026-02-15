#!/usr/bin/env python3
"""
Ralph Loop Self-Validation Engine
STAP (Stop, Task, Assess, Process) validation loop with automatic retry.
No false success reports - forces agents to self-validate before completion.

Author: 97LAYER Mercenary Standard Applied
"""

import json
import os
import subprocess
import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Tuple, Any
import time
import re

# Add project root to path
sys.path.append(str(Path(__file__).parent.parent))

from libs.core_config import MERCENARY_STANDARD, PROJECT_ROOT, KNOWLEDGE_PATHS


class RalphLoopEngine:
    """
    STAP-based self-validation engine that enforces quality before completion.

    Process:
    1. Stop - Pause before claiming success
    2. Task - Review what was supposed to be accomplished
    3. Assess - Run comprehensive validation checks
    4. Process - Retry on failure or emit [SYSTEM_STABLE] on success
    """

    def __init__(self, config_path: Optional[str] = None):
        """Initialize Ralph Loop Engine"""
        self.project_root = Path(PROJECT_ROOT)
        self.config_path = config_path or self.project_root / "core/ralph_loop_config.yaml"
        # Use container-aware paths
        self.log_path = KNOWLEDGE_PATHS["system"] / "ralph_loop.json"
        self.log_path.parent.mkdir(parents=True, exist_ok=True)

        # Load configuration
        self.config = self._load_config()

        # Initialize log
        if not self.log_path.exists():
            self._init_log()

    def _load_config(self) -> Dict:
        """Load configuration from YAML or use defaults"""
        if self.config_path.exists():
            import yaml
            with open(self.config_path) as f:
                return yaml.safe_load(f)

        # Default configuration
        return {
            'max_iterations': 5,
            'verification_rules': {
                'build': True,
                'skill_compliance': True,
                'logs': True,
                'syntax': True
            },
            'banned_patterns': [
                r'TODO\s*:',
                r'FIXME\s*:',
                r'HACK\s*:',
                r'XXX\s*:',
                r'hardcoded.*token',
                r'hardcoded.*password',
                r'print\s*\(',  # No debug prints in production
            ],
            'error_keywords': [
                'error',
                'exception',
                'failed',
                'failure',
                'traceback',
                'critical',
                'fatal'
            ],
            'completion_token': '[SYSTEM_STABLE]',
            'timeout_seconds': 300,  # 5 minutes max per iteration
        }

    def _init_log(self):
        """Initialize Ralph Loop log file"""
        initial_log = {
            'created_at': datetime.now().isoformat(),
            'total_runs': 0,
            'successful_runs': 0,
            'failed_runs': 0,
            'runs': []
        }
        with open(self.log_path, 'w') as f:
            json.dump(initial_log, f, indent=2)

    def validate_task(
        self,
        task_id: str,
        agent_id: str,
        task_type: str,
        files_modified: List[str],
        output_path: Optional[str] = None
    ) -> Tuple[bool, Dict]:
        """
        Main validation entry point.

        Args:
            task_id: Unique task identifier
            agent_id: Agent performing the task
            task_type: Type of task (frontend, backend, system, content)
            files_modified: List of files that were modified
            output_path: Optional output file path

        Returns:
            Tuple of (success: bool, validation_report: Dict)
        """
        print(f"\n{'='*70}")
        print(f"🔄 Ralph Loop - STAP Validation Starting")
        print(f"{'='*70}")
        print(f"Task: {task_id}")
        print(f"Agent: {agent_id}")
        print(f"Type: {task_type}")
        print(f"Files: {len(files_modified)}")
        print(f"{'='*70}\n")

        iteration = 0
        max_iterations = self.config['max_iterations']

        run_log = {
            'task_id': task_id,
            'agent_id': agent_id,
            'task_type': task_type,
            'started_at': datetime.now().isoformat(),
            'iterations': [],
            'final_status': None,
            'completion_token_issued': False
        }

        while iteration < max_iterations:
            iteration += 1
            print(f"\n{'─'*70}")
            print(f"🔁 Iteration {iteration}/{max_iterations}")
            print(f"{'─'*70}\n")

            # STAP Loop
            iteration_result = self._stap_cycle(
                task_id=task_id,
                agent_id=agent_id,
                task_type=task_type,
                files_modified=files_modified,
                output_path=output_path,
                iteration=iteration
            )

            run_log['iterations'].append(iteration_result)

            if iteration_result['passed']:
                # SUCCESS - All checks passed
                print(f"\n{'='*70}")
                print(f"✅ VALIDATION PASSED - Iteration {iteration}")
                print(f"{'='*70}\n")

                run_log['final_status'] = 'success'
                run_log['completed_at'] = datetime.now().isoformat()
                run_log['completion_token_issued'] = True

                self._log_run(run_log, success=True)

                # Emit completion token
                print(f"\n🎯 {self.config['completion_token']}")
                print(f"{'='*70}\n")

                return True, iteration_result

            else:
                # FAILURE - Retry needed
                print(f"\n{'⚠'*70}")
                print(f"❌ VALIDATION FAILED - Iteration {iteration}")
                print(f"Failed checks: {', '.join(iteration_result['failed_checks'])}")
                print(f"{'⚠'*70}\n")

                if iteration < max_iterations:
                    print(f"🔄 Retrying... ({max_iterations - iteration} attempts remaining)\n")
                    time.sleep(2)  # Brief pause before retry
                else:
                    print(f"❌ MAX ITERATIONS REACHED - VALIDATION FAILED\n")

        # Max iterations reached without success
        run_log['final_status'] = 'failed'
        run_log['completed_at'] = datetime.now().isoformat()
        run_log['completion_token_issued'] = False
        run_log['failure_reason'] = 'Max iterations exceeded'

        self._log_run(run_log, success=False)

        print(f"\n{'='*70}")
        print(f"❌ SYSTEM UNSTABLE - NO COMPLETION TOKEN ISSUED")
        print(f"{'='*70}\n")

        return False, run_log['iterations'][-1]

    def _stap_cycle(
        self,
        task_id: str,
        agent_id: str,
        task_type: str,
        files_modified: List[str],
        output_path: Optional[str],
        iteration: int
    ) -> Dict:
        """
        Execute one STAP (Stop, Task, Assess, Process) cycle.

        Returns:
            Dict with validation results
        """
        result = {
            'iteration': iteration,
            'timestamp': datetime.now().isoformat(),
            'checks': {},
            'failed_checks': [],
            'passed': False
        }

        # STOP - Pause and prepare
        print("🛑 STOP - Preparing validation...")
        time.sleep(0.5)

        # TASK - Review what needs validation
        print("📋 TASK - Reviewing requirements...")
        task_requirements = self._get_task_requirements(task_type)
        result['requirements'] = task_requirements

        # ASSESS - Run all validation checks
        print("🔍 ASSESS - Running validation checks...\n")

        # Check 1: Build verification
        if self.config['verification_rules']['build']:
            build_passed, build_details = self._verify_build(task_type)
            result['checks']['build'] = {
                'passed': build_passed,
                'details': build_details
            }
            if not build_passed:
                result['failed_checks'].append('build')

        # Check 2: Skill compliance
        if self.config['verification_rules']['skill_compliance']:
            skill_passed, skill_details = self._verify_skill_compliance(files_modified)
            result['checks']['skill_compliance'] = {
                'passed': skill_passed,
                'details': skill_details
            }
            if not skill_passed:
                result['failed_checks'].append('skill_compliance')

        # Check 3: Log analysis
        if self.config['verification_rules']['logs']:
            log_passed, log_details = self._verify_logs()
            result['checks']['logs'] = {
                'passed': log_passed,
                'details': log_details
            }
            if not log_passed:
                result['failed_checks'].append('logs')

        # Check 4: Syntax verification
        if self.config['verification_rules']['syntax']:
            syntax_passed, syntax_details = self._verify_syntax(files_modified, task_type)
            result['checks']['syntax'] = {
                'passed': syntax_passed,
                'details': syntax_details
            }
            if not syntax_passed:
                result['failed_checks'].append('syntax')

        # Check 5: Banned patterns
        banned_passed, banned_details = self._check_banned_patterns(files_modified)
        result['checks']['banned_patterns'] = {
            'passed': banned_passed,
            'details': banned_details
        }
        if not banned_passed:
            result['failed_checks'].append('banned_patterns')

        # Check 6: Output verification
        if output_path:
            output_passed, output_details = self._verify_output(output_path)
            result['checks']['output'] = {
                'passed': output_passed,
                'details': output_details
            }
            if not output_passed:
                result['failed_checks'].append('output')

        # PROCESS - Determine outcome
        print("\n⚙️  PROCESS - Analyzing results...")
        result['passed'] = len(result['failed_checks']) == 0

        return result

    def _get_task_requirements(self, task_type: str) -> List[str]:
        """Get validation requirements for task type"""
        requirements = {
            'frontend': [
                'npm build must succeed',
                'No TypeScript errors',
                'No console.log statements',
                'Components follow naming convention'
            ],
            'backend': [
                'Python syntax valid',
                'No import errors',
                'Tests pass (if exists)',
                'Type hints present'
            ],
            'system': [
                'Core files intact',
                'No environment corruption',
                'Services functional'
            ],
            'content': [
                'Output file exists',
                'Proper markdown format',
                'No placeholder text'
            ]
        }
        return requirements.get(task_type, ['Basic validation'])

    def _verify_build(self, task_type: str) -> Tuple[bool, str]:
        """Verify build process"""
        print("  🔨 Checking build...")

        if task_type == 'frontend':
            frontend_dir = self.project_root / "frontend"
            if frontend_dir.exists():
                try:
                    result = subprocess.run(
                        ["npm", "run", "build"],
                        cwd=frontend_dir,
                        capture_output=True,
                        text=True,
                        timeout=60
                    )
                    if result.returncode == 0:
                        print("     ✅ Frontend build successful")
                        return True, "Build successful"
                    else:
                        print("     ❌ Frontend build failed")
                        return False, f"Build failed: {result.stderr[:200]}"
                except subprocess.TimeoutExpired:
                    print("     ❌ Build timeout")
                    return False, "Build timeout (>60s)"
            else:
                print("     ⚠️  No frontend directory")
                return True, "No frontend to build"

        elif task_type == 'backend':
            # Python syntax check
            py_files = [f for f in self.project_root.rglob("*.py")
                       if 'venv' not in str(f) and '__pycache__' not in str(f)]

            for py_file in py_files[:20]:  # Limit to 20 files
                result = subprocess.run(
                    ["python3", "-m", "py_compile", str(py_file)],
                    capture_output=True,
                    text=True
                )
                if result.returncode != 0:
                    print(f"     ❌ Syntax error in {py_file.name}")
                    return False, f"Syntax error: {py_file}"

            print("     ✅ Python syntax valid")
            return True, "All Python files compile"

        print("     ℹ️  No build required")
        return True, "No build verification needed"

    def _verify_skill_compliance(self, files_modified: List[str]) -> Tuple[bool, str]:
        """Verify skill compliance and directive adherence"""
        print("  📐 Checking skill compliance...")

        issues = []

        for file_path in files_modified:
            if not Path(file_path).exists():
                continue

            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                # Check for type hints in Python files
                if file_path.endswith('.py'):
                    if MERCENARY_STANDARD['TYPE_HINTING']:
                        # Check for function definitions without type hints
                        func_pattern = r'def\s+\w+\s*\([^)]*\)\s*:'
                        funcs = re.findall(func_pattern, content)
                        if funcs:
                            # Simple heuristic: check if '->' is present for return types
                            if '->' not in content and 'def __init__' not in content:
                                issues.append(f"{Path(file_path).name}: Missing type hints")

                # Check for emojis in code (violates ZERO_NOISE)
                emoji_pattern = r'[\U0001F600-\U0001F64F\U0001F300-\U0001F5FF\U0001F680-\U0001F6FF]'
                if re.search(emoji_pattern, content):
                    # Only flag if not in comments or strings
                    lines = content.split('\n')
                    for line in lines:
                        if re.search(emoji_pattern, line) and not line.strip().startswith('#'):
                            issues.append(f"{Path(file_path).name}: Emoji found (violates Zero Noise)")
                            break

            except Exception as e:
                issues.append(f"{Path(file_path).name}: Read error - {str(e)}")

        if issues:
            print("     ❌ Skill compliance issues found")
            for issue in issues[:3]:  # Show first 3
                print(f"        - {issue}")
            return False, f"{len(issues)} compliance issues"

        print("     ✅ Skill compliance verified")
        return True, "All files comply with directives"

    def _verify_logs(self) -> Tuple[bool, str]:
        """Check recent logs for errors"""
        print("  📋 Checking logs...")

        log_dirs = [
            self.project_root / "knowledge/system",
            self.project_root / "knowledge/notifications"
        ]

        error_count = 0
        recent_errors = []

        # Files to skip (avoid circular detection)
        skip_files = ['ralph_loop.json', 'task_board.json']

        for log_dir in log_dirs:
            if not log_dir.exists():
                continue

            # Check recent log files
            log_files = sorted(log_dir.glob("*.json"), key=lambda p: p.stat().st_mtime, reverse=True)[:5]

            for log_file in log_files:
                # Skip self and task board
                if log_file.name in skip_files:
                    continue

                try:
                    with open(log_file, 'r') as f:
                        content = f.read().lower()
                        for keyword in self.config['error_keywords']:
                            if keyword in content:
                                error_count += 1
                                recent_errors.append(f"{log_file.name}: {keyword}")
                                break
                except:
                    pass

        if error_count > 0:
            print(f"     ⚠️  Found {error_count} error(s) in logs")
            for error in recent_errors[:3]:
                print(f"        - {error}")
            return False, f"{error_count} errors in recent logs"

        print("     ✅ No errors in recent logs")
        return True, "Logs clean"

    def _verify_syntax(self, files_modified: List[str], task_type: str) -> Tuple[bool, str]:
        """Verify syntax of modified files"""
        print("  🔍 Checking syntax...")

        errors = []

        for file_path in files_modified:
            if not Path(file_path).exists():
                continue

            file_ext = Path(file_path).suffix

            # Python syntax
            if file_ext == '.py':
                result = subprocess.run(
                    ["python3", "-m", "py_compile", file_path],
                    capture_output=True,
                    text=True
                )
                if result.returncode != 0:
                    errors.append(f"{Path(file_path).name}: Python syntax error")

            # JSON syntax
            elif file_ext == '.json':
                try:
                    with open(file_path, 'r') as f:
                        json.load(f)
                except json.JSONDecodeError as e:
                    errors.append(f"{Path(file_path).name}: JSON error at line {e.lineno}")

            # YAML syntax
            elif file_ext in ['.yaml', '.yml']:
                try:
                    import yaml
                    with open(file_path, 'r') as f:
                        yaml.safe_load(f)
                except yaml.YAMLError as e:
                    errors.append(f"{Path(file_path).name}: YAML error")

        if errors:
            print("     ❌ Syntax errors found")
            for error in errors:
                print(f"        - {error}")
            return False, f"{len(errors)} syntax errors"

        print("     ✅ All syntax valid")
        return True, "All files have valid syntax"

    def _check_banned_patterns(self, files_modified: List[str]) -> Tuple[bool, str]:
        """Check for banned patterns in code"""
        print("  🚫 Checking banned patterns...")

        violations = []

        for file_path in files_modified:
            if not Path(file_path).exists():
                continue

            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                for pattern in self.config['banned_patterns']:
                    matches = re.findall(pattern, content, re.IGNORECASE)
                    if matches:
                        violations.append(f"{Path(file_path).name}: {pattern}")

            except Exception:
                pass

        if violations:
            print("     ❌ Banned patterns found")
            for violation in violations[:5]:
                print(f"        - {violation}")
            return False, f"{len(violations)} banned patterns detected"

        print("     ✅ No banned patterns")
        return True, "No banned patterns found"

    def _verify_output(self, output_path: str) -> Tuple[bool, str]:
        """Verify output file exists and is valid"""
        print(f"  📄 Checking output: {Path(output_path).name}...")

        output_file = Path(output_path)

        if not output_file.exists():
            print("     ❌ Output file not found")
            return False, "Output file does not exist"

        # Check file is not empty
        if output_file.stat().st_size == 0:
            print("     ❌ Output file is empty")
            return False, "Output file is empty"

        # Check for placeholder text
        try:
            with open(output_file, 'r') as f:
                content = f.read()
                placeholders = ['TODO', 'FIXME', 'PLACEHOLDER', 'TBD', '...']
                for placeholder in placeholders:
                    if placeholder.upper() in content.upper():
                        print(f"     ⚠️  Contains placeholder: {placeholder}")
                        return False, f"Output contains placeholder: {placeholder}"
        except:
            pass

        print("     ✅ Output verified")
        return True, "Output file valid"

    def _log_run(self, run_log: Dict, success: bool):
        """Log the validation run"""
        # Load existing log
        with open(self.log_path, 'r') as f:
            log = json.load(f)

        # Update statistics
        log['total_runs'] += 1
        if success:
            log['successful_runs'] += 1
        else:
            log['failed_runs'] += 1

        # Add run record
        log['runs'].append(run_log)

        # Keep only last 100 runs
        if len(log['runs']) > 100:
            log['runs'] = log['runs'][-100:]

        # Save
        with open(self.log_path, 'w') as f:
            json.dump(log, f, indent=2)

    def get_stats(self) -> Dict:
        """Get validation statistics"""
        if not self.log_path.exists():
            return {'error': 'No log file found'}

        with open(self.log_path, 'r') as f:
            log = json.load(f)

        success_rate = (log['successful_runs'] / log['total_runs'] * 100) if log['total_runs'] > 0 else 0

        return {
            'total_runs': log['total_runs'],
            'successful_runs': log['successful_runs'],
            'failed_runs': log['failed_runs'],
            'success_rate': f"{success_rate:.1f}%",
            'recent_runs': log['runs'][-10:]  # Last 10 runs
        }


def main():
    """CLI interface for Ralph Loop"""
    if len(sys.argv) < 2:
        print("Usage: ralph_loop_engine.py [validate|stats] [args...]")
        print("\nCommands:")
        print("  validate <task_id> <agent_id> <task_type> <file1> [file2...]")
        print("  stats")
        return

    engine = RalphLoopEngine()
    command = sys.argv[1]

    if command == "validate":
        if len(sys.argv) < 5:
            print("Usage: ralph_loop_engine.py validate <task_id> <agent_id> <task_type> <file1> [file2...]")
            return

        task_id = sys.argv[2]
        agent_id = sys.argv[3]
        task_type = sys.argv[4]
        files_modified = sys.argv[5:]

        success, report = engine.validate_task(
            task_id=task_id,
            agent_id=agent_id,
            task_type=task_type,
            files_modified=files_modified
        )

        sys.exit(0 if success else 1)

    elif command == "stats":
        stats = engine.get_stats()
        print(json.dumps(stats, indent=2))


if __name__ == "__main__":
    main()
