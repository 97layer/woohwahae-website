#!/usr/bin/env python3
"""
DirectiveEditor — 97layerOS 텔레그램 양방향 파일 수정 실행기

텔레그램에서 순호가 직접 지시하면 실제 파일을 수정한다.
Gardener 주기 밖에서 즉시 실행.

권한 정책:
  FROZEN  — 순호 확인 후 실행 (confirm_token 방식)
  나머지  — 즉시 실행

Author: 97layerOS
Created: 2026-02-16
"""

import json
import logging
import re
import secrets
import os
import sys
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Optional, Tuple

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

try:
    from dotenv import load_dotenv
    load_dotenv(PROJECT_ROOT / '.env')
except ImportError:
    pass

try:
    import google.genai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)

# ── 파일 경로 매핑 ─────────────────────────────────────────────
DIRECTIVE_FILES = {
    "IDENTITY.md":  PROJECT_ROOT / "directives" / "IDENTITY.md",
    "CD_SUNHO.md":  PROJECT_ROOT / "directives" / "agents" / "CD_SUNHO.md",
    "JOON.md":      PROJECT_ROOT / "directives" / "agents" / "JOON.md",
    "MIA.md":       PROJECT_ROOT / "directives" / "agents" / "MIA.md",
    "RAY.md":       PROJECT_ROOT / "directives" / "agents" / "RAY.md",
    "SYSTEM.md":    PROJECT_ROOT / "directives" / "system" / "SYSTEM.md",
}

# 확인 필요한 파일 (FROZEN)
REQUIRES_CONFIRM = {"IDENTITY.md", "CD_SUNHO.md"}

# 확인 토큰 만료 시간 (분)
CONFIRM_TTL_MINUTES = 10


class DirectiveEditor:
    """
    텔레그램 지시 → 실제 directives/ 파일 수정

    사용법:
        editor = DirectiveEditor()

        # 즉시 실행 가능한 파일
        result = editor.execute(edit_params)
        # result: {'success': bool, 'message': str, 'diff': str}

        # FROZEN 파일 — 2단계 확인
        result = editor.request_confirm(edit_params)
        # result: {'needs_confirm': True, 'token': '...', 'preview': '...'}

        # 순호가 /confirm [token] 보내면
        result = editor.confirm_and_execute(token)
    """

    def __init__(self):
        self._pending_confirms: Dict[str, Dict] = {}  # token → {params, expires_at}

        api_key = os.getenv('GOOGLE_API_KEY') or os.getenv('GEMINI_API_KEY')
        if GEMINI_AVAILABLE and api_key:
            self.client = genai.Client(api_key=api_key)
            self._model = 'gemini-2.5-flash'
            self.use_ai = True
        else:
            self.client = None
            self.use_ai = False

    # ── 퍼블릭 API ──────────────────────────────────────────────

    def handle(self, edit_params: Dict) -> Dict:
        """
        edit_params 처리 진입점.

        Returns:
            {
                'success': bool,
                'needs_confirm': bool,  # True면 순호가 /confirm [token] 필요
                'token': str,           # needs_confirm=True일 때
                'message': str,         # 사용자에게 보낼 메시지
                'diff': str,            # 변경 내용 미리보기
            }
        """
        target = edit_params.get('target_file', '')
        if not target:
            return {'success': False, 'needs_confirm': False,
                    'message': "어떤 파일을 수정할지 파악하지 못했습니다."}

        if target not in DIRECTIVE_FILES:
            return {'success': False, 'needs_confirm': False,
                    'message': f"'{target}'은 수정 가능한 파일이 아닙니다.\n"
                               f"가능한 파일: {', '.join(DIRECTIVE_FILES.keys())}"}

        # 실제 수정 내용 계산 (Gemini 보조)
        prepared = self._prepare_edit(edit_params)
        if not prepared['feasible']:
            return {'success': False, 'needs_confirm': False,
                    'message': prepared['reason']}

        if target in REQUIRES_CONFIRM:
            # FROZEN — 확인 토큰 발급
            token = secrets.token_hex(4).upper()
            expires = datetime.now() + timedelta(minutes=CONFIRM_TTL_MINUTES)
            self._pending_confirms[token] = {
                'params': edit_params,
                'prepared': prepared,
                'expires_at': expires,
            }
            return {
                'success': False,
                'needs_confirm': True,
                'token': token,
                'message': (
                    f"⚠️ <b>{target}</b>은 핵심 파일입니다. 변경을 확인해주세요.\n\n"
                    f"<b>변경 내용:</b>\n{prepared['diff_preview']}\n\n"
                    f"/confirm {token} 으로 확인하면 즉시 반영됩니다.\n"
                    f"({CONFIRM_TTL_MINUTES}분 내 확인 필요)"
                ),
                'diff': prepared['diff_preview'],
            }
        else:
            # 일반 파일 — 즉시 실행
            return self._apply_edit(edit_params, prepared)

    def confirm_and_execute(self, token: str) -> Dict:
        """
        /confirm [token] 처리 — FROZEN 파일 변경 최종 실행
        """
        pending = self._pending_confirms.get(token)
        if not pending:
            return {'success': False, 'needs_confirm': False,
                    'message': f"확인 토큰 '{token}'을 찾을 수 없습니다. 이미 사용됐거나 만료됐을 수 있습니다."}

        if datetime.now() > pending['expires_at']:
            del self._pending_confirms[token]
            return {'success': False, 'needs_confirm': False,
                    'message': "확인 토큰이 만료되었습니다. 다시 수정 요청을 보내주세요."}

        del self._pending_confirms[token]
        return self._apply_edit(pending['params'], pending['prepared'])

    # ── 내부 로직 ──────────────────────────────────────────────

    def _prepare_edit(self, edit_params: Dict) -> Dict:
        """
        실제로 파일에 적용할 수정 내용을 계산.
        Returns: {feasible, old_text, new_text, diff_preview, reason}
        """
        target = edit_params.get('target_file', '')
        action = edit_params.get('action', 'remove')
        content = edit_params.get('content', '')
        replace_with = edit_params.get('replace_with', '')

        path = DIRECTIVE_FILES[target]
        if not path.exists():
            return {'feasible': False, 'reason': f"{target} 파일을 찾을 수 없습니다."}

        current_text = path.read_text(encoding='utf-8')

        if action == 'remove':
            result = self._find_and_remove(current_text, content, target)
        elif action == 'add':
            result = self._find_and_add(current_text, content, target)
        elif action == 'replace':
            result = self._find_and_replace(current_text, content, replace_with, target)
        else:
            return {'feasible': False, 'reason': f"알 수 없는 action: {action}"}

        return result

    def _find_and_remove(self, text: str, content: str, target: str) -> Dict:
        """
        content와 일치하는 줄/블록을 찾아 제거.
        Gemini를 사용해 퍼지 매칭.
        """
        if not content:
            return {'feasible': False, 'reason': "삭제할 내용이 지정되지 않았습니다."}

        # 직접 매칭 시도
        if content in text:
            new_text = text.replace(content, '', 1).strip()
            # 연속 빈줄 정리
            new_text = re.sub(r'\n{3,}', '\n\n', new_text)
            diff = f"삭제: \"{content[:80]}\""
            return {'feasible': True, 'old_text': text, 'new_text': new_text,
                    'diff_preview': diff, 'reason': ''}

        # Gemini 퍼지 매칭
        if self.use_ai:
            return self._ai_find_and_remove(text, content, target)

        return {'feasible': False,
                'reason': f"'{content[:50]}' 내용을 {target}에서 찾을 수 없습니다."}

    def _find_and_add(self, text: str, content: str, target: str) -> Dict:
        """내용 추가 — 파일 끝에 추가"""
        if not content:
            return {'feasible': False, 'reason': "추가할 내용이 지정되지 않았습니다."}

        new_text = text.rstrip() + '\n\n' + content + '\n'
        diff = f"추가: \"{content[:80]}\""
        return {'feasible': True, 'old_text': text, 'new_text': new_text,
                'diff_preview': diff, 'reason': ''}

    def _find_and_replace(self, text: str, content: str, replace_with: str, target: str) -> Dict:
        """content → replace_with 교체"""
        if not content:
            return {'feasible': False, 'reason': "교체할 내용이 지정되지 않았습니다."}

        if content in text:
            new_text = text.replace(content, replace_with, 1)
            diff = f"교체: \"{content[:60]}\" → \"{replace_with[:60]}\""
            return {'feasible': True, 'old_text': text, 'new_text': new_text,
                    'diff_preview': diff, 'reason': ''}

        if self.use_ai:
            return self._ai_find_and_replace(text, content, replace_with, target)

        return {'feasible': False,
                'reason': f"'{content[:50]}' 내용을 {target}에서 찾을 수 없습니다."}

    def _ai_find_and_remove(self, text: str, content: str, target: str) -> Dict:
        """Gemini로 퍼지 매칭 후 제거"""
        prompt = f"""파일 내용에서 '{content}'와 관련된 줄/블록을 찾아 제거해라.

파일 내용:
{text[:3000]}

제거 대상 (사용자 표현): "{content}"

응답 형식 (JSON):
{{
  "found": true/false,
  "matched_text": "실제로 파일에서 찾은 정확한 텍스트 (줄 단위)",
  "reason": "왜 이 텍스트를 매칭했는지"
}}

파일에서 정확히 찾을 수 없으면 found: false.
JSON만 출력."""
        try:
            resp = self.client.models.generate_content(
                model=self._model, contents=[prompt]
            )
            m = re.search(r'\{.*\}', resp.text.strip(), re.DOTALL)
            if not m:
                return {'feasible': False, 'reason': "AI 매칭 실패"}
            r = json.loads(m.group())
            if not r.get('found'):
                return {'feasible': False,
                        'reason': f"'{content[:50]}'와 일치하는 내용을 {target}에서 찾을 수 없습니다."}
            matched = r['matched_text']
            if matched not in text:
                return {'feasible': False,
                        'reason': f"AI가 찾은 내용을 파일에서 확인할 수 없습니다: {matched[:50]}"}
            new_text = text.replace(matched, '', 1)
            new_text = re.sub(r'\n{3,}', '\n\n', new_text)
            diff = f"삭제: \"{matched[:80]}\""
            return {'feasible': True, 'old_text': text, 'new_text': new_text,
                    'diff_preview': diff, 'reason': ''}
        except Exception as e:
            return {'feasible': False, 'reason': f"AI 처리 오류: {e}"}

    def _ai_find_and_replace(self, text: str, content: str, replace_with: str, target: str) -> Dict:
        """Gemini로 퍼지 매칭 후 교체"""
        prompt = f"""파일 내용에서 '{content}'와 관련된 줄/블록을 찾아라.

파일 내용:
{text[:3000]}

찾을 내용 (사용자 표현): "{content}"

응답 형식 (JSON):
{{
  "found": true/false,
  "matched_text": "실제로 파일에서 찾은 정확한 텍스트"
}}

JSON만 출력."""
        try:
            resp = self.client.models.generate_content(
                model=self._model, contents=[prompt]
            )
            m = re.search(r'\{.*\}', resp.text.strip(), re.DOTALL)
            if not m:
                return {'feasible': False, 'reason': "AI 매칭 실패"}
            r = json.loads(m.group())
            if not r.get('found'):
                return {'feasible': False,
                        'reason': f"'{content[:50]}'와 일치하는 내용을 {target}에서 찾을 수 없습니다."}
            matched = r['matched_text']
            if matched not in text:
                return {'feasible': False,
                        'reason': f"AI가 찾은 내용을 파일에서 확인할 수 없습니다: {matched[:50]}"}
            new_text = text.replace(matched, replace_with, 1)
            diff = f"교체: \"{matched[:60]}\" → \"{replace_with[:60]}\""
            return {'feasible': True, 'old_text': text, 'new_text': new_text,
                    'diff_preview': diff, 'reason': ''}
        except Exception as e:
            return {'feasible': False, 'reason': f"AI 처리 오류: {e}"}

    def _apply_edit(self, edit_params: Dict, prepared: Dict) -> Dict:
        """준비된 수정을 실제 파일에 기록"""
        target = edit_params.get('target_file', '')
        path = DIRECTIVE_FILES[target]

        try:
            # 백업
            backup_dir = PROJECT_ROOT / '.infra' / 'directive_backups'
            backup_dir.mkdir(parents=True, exist_ok=True)
            ts = datetime.now().strftime('%Y%m%d_%H%M%S')
            backup_path = backup_dir / f"{target}.{ts}.bak"
            backup_path.write_text(prepared['old_text'], encoding='utf-8')

            # 실제 파일 수정
            path.write_text(prepared['new_text'], encoding='utf-8')

            logger.info("✅ 파일 수정 완료: %s (%s)", target, prepared['diff_preview'][:60])

            return {
                'success': True,
                'needs_confirm': False,
                'message': (
                    f"✅ <b>{target}</b> 수정 완료\n\n"
                    f"{prepared['diff_preview']}\n\n"
                    f"백업: <code>{backup_path.name}</code>"
                ),
                'diff': prepared['diff_preview'],
            }
        except Exception as e:
            logger.error("파일 수정 실패: %s", e)
            return {'success': False, 'needs_confirm': False,
                    'message': f"파일 수정 중 오류가 발생했습니다: {e}"}


# ── 싱글턴 ────────────────────────────────────────────────────

_editor_instance: Optional[DirectiveEditor] = None


def get_directive_editor() -> DirectiveEditor:
    global _editor_instance
    if _editor_instance is None:
        _editor_instance = DirectiveEditor()
    return _editor_instance
