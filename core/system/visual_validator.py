"""
Visual Validator — 디자인 토큰 품질 검증

디스코드 인사이트 기반:
- WCAG AA 대비 4.5:1 이상
- spacing 8px 배수
- 토큰 외 직접 색상/간격 사용 감지

Author: LAYER OS
Created: 2026-02-26
"""

import re
import os
from pathlib import Path
from typing import List, Dict, Tuple
from dataclasses import dataclass


@dataclass
class ValidationIssue:
    """검증 이슈"""
    file: str
    line: int
    rule: str
    message: str
    severity: str  # error, warning


class VisualValidator:
    """디자인 시스템 품질 검증"""

    def __init__(self, project_root: str = None):
        self.project_root = Path(project_root or os.getcwd())
        self.issues: List[ValidationIssue] = []

    def validate_all(self) -> List[ValidationIssue]:
        """전체 검증 실행"""
        self.issues = []

        # 1. CSS 토큰 검증
        self._validate_css_tokens()

        # 2. HTML 인라인 스타일 검증
        self._validate_html_inline_styles()

        # 3. 간격 8px 배수 검증
        self._validate_spacing_scale()

        # 4. WCAG 대비 검증
        self._validate_contrast()

        return self.issues

    def _validate_css_tokens(self):
        """CSS :root 토큰 완결성 검증"""
        css_path = self.project_root / 'website/assets/css/style.css'

        if not css_path.exists():
            self.issues.append(ValidationIssue(
                file=str(css_path),
                line=0,
                rule="css_missing",
                message="style.css not found",
                severity="error"
            ))
            return

        with open(css_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 필수 토큰 체크
        required_tokens = [
            '--bg', '--text', '--text-sub',
            '--space-xs', '--space-sm', '--space-md',
            '--font-body', '--font-mono',
            '--shadow-sm', '--radius-md'
        ]

        for token in required_tokens:
            if f'{token}:' not in content:
                self.issues.append(ValidationIssue(
                    file=str(css_path),
                    line=0,
                    rule="missing_token",
                    message=f"Required token {token} not found",
                    severity="error"
                ))

    def _validate_html_inline_styles(self):
        """HTML 인라인 스타일 검증 (토큰 외 직접 값 감지)"""
        html_files = list(self.project_root.glob('website/**/*.html'))

        for html_file in html_files:
            # lab/ 디렉토리는 제외 (프로토타입)
            if '/lab/' in str(html_file):
                continue

            with open(html_file, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            for i, line in enumerate(lines, 1):
                # style= 속성 찾기
                if 'style=' not in line:
                    continue

                # 하드코딩 값 패턴
                hardcoded_patterns = [
                    (r'margin-top:\s*(\d+)rem', 'Use var(--space-*) instead of hardcoded margin'),
                    (r'padding:\s*(\d+)rem', 'Use var(--space-*) instead of hardcoded padding'),
                    (r'background:\s*#[0-9A-Fa-f]{6}', 'Use var(--bg) or color tokens instead of hex'),
                    (r'color:\s*#[0-9A-Fa-f]{6}', 'Use var(--text*) tokens instead of hex'),
                    (r'min-height:\s*(\d+)vh', 'Consider using spacing tokens or CSS class'),
                ]

                for pattern, message in hardcoded_patterns:
                    if re.search(pattern, line) and 'var(--' not in line:
                        self.issues.append(ValidationIssue(
                            file=str(html_file.relative_to(self.project_root)),
                            line=i,
                            rule="hardcoded_style",
                            message=message,
                            severity="warning"
                        ))
                        break  # 한 줄에 하나만 보고

    def _validate_spacing_scale(self):
        """spacing 값 8px 배수 검증"""
        css_path = self.project_root / 'website/assets/css/style.css'

        if not css_path.exists():
            return

        with open(css_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        for i, line in enumerate(lines, 1):
            # --space-* 토큰 확인
            match = re.search(r'--space-\w+:\s*([\d.]+)rem', line)
            if match:
                rem_value = float(match.group(1))
                px_value = rem_value * 16  # 1rem = 16px 가정

                if px_value % 8 != 0:
                    self.issues.append(ValidationIssue(
                        file=str(css_path.relative_to(self.project_root)),
                        line=i,
                        rule="spacing_scale",
                        message=f"{px_value}px is not a multiple of 8px",
                        severity="error"
                    ))

    def _validate_contrast(self):
        """WCAG AA 대비 4.5:1 검증"""
        css_path = self.project_root / 'website/assets/css/style.css'

        if not css_path.exists():
            return

        with open(css_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 색상 토큰 추출
        color_tokens = {}
        for match in re.finditer(r'--(\w+):\s*(#[0-9A-Fa-f]{6})', content):
            token_name = match.group(1)
            hex_color = match.group(2)
            color_tokens[token_name] = hex_color

        # 주요 조합 검증
        critical_pairs = [
            ('text', 'bg', 'Main text on background'),
            ('text-sub', 'bg', 'Sub text on background'),
        ]

        for fg_token, bg_token, description in critical_pairs:
            if fg_token not in color_tokens or bg_token not in color_tokens:
                continue

            fg = color_tokens[fg_token]
            bg = color_tokens[bg_token]
            ratio = self._calculate_contrast_ratio(fg, bg)

            if ratio < 4.5:
                self.issues.append(ValidationIssue(
                    file=str(css_path.relative_to(self.project_root)),
                    line=0,
                    rule="wcag_contrast",
                    message=f"{description}: {ratio:.2f}:1 (minimum 4.5:1)",
                    severity="error"
                ))

    @staticmethod
    def _calculate_contrast_ratio(hex1: str, hex2: str) -> float:
        """
        WCAG 대비 계산
        https://www.w3.org/TR/WCAG20-TECHS/G17.html
        """
        def hex_to_rgb(hex_color):
            hex_color = hex_color.lstrip('#')
            return tuple(int(hex_color[i:i+2], 16) for i in (0, 2, 4))

        def relative_luminance(rgb):
            r, g, b = [x / 255.0 for x in rgb]
            r = r / 12.92 if r <= 0.03928 else ((r + 0.055) / 1.055) ** 2.4
            g = g / 12.92 if g <= 0.03928 else ((g + 0.055) / 1.055) ** 2.4
            b = b / 12.92 if b <= 0.03928 else ((b + 0.055) / 1.055) ** 2.4
            return 0.2126 * r + 0.7152 * g + 0.0722 * b

        rgb1 = hex_to_rgb(hex1)
        rgb2 = hex_to_rgb(hex2)

        lum1 = relative_luminance(rgb1)
        lum2 = relative_luminance(rgb2)

        lighter = max(lum1, lum2)
        darker = min(lum1, lum2)

        return (lighter + 0.05) / (darker + 0.05)

    def print_report(self):
        """검증 결과 출력"""
        if not self.issues:
            print("✅ No validation issues found.")
            return

        # Severity별 그룹화
        errors = [i for i in self.issues if i.severity == "error"]
        warnings = [i for i in self.issues if i.severity == "warning"]

        if errors:
            print(f"\n🔴 {len(errors)} Errors:")
            for issue in errors[:10]:  # 최대 10개만
                print(f"   {issue.file}:{issue.line} — {issue.message}")

        if warnings:
            print(f"\n🟡 {len(warnings)} Warnings:")
            for issue in warnings[:10]:
                print(f"   {issue.file}:{issue.line} — {issue.message}")

        if len(self.issues) > 20:
            print(f"\n... and {len(self.issues) - 20} more issues")


# CLI 인터페이스
if __name__ == "__main__":
    validator = VisualValidator()
    validator.validate_all()
    validator.print_report()

    # 종료 코드 (CI/CD 통합용)
    errors = [i for i in validator.issues if i.severity == "error"]
    exit(1 if errors else 0)
