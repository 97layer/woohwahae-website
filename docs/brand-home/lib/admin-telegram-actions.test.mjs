import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const runtimeSource = await readFile(new URL('./runtime/telegram.js', import.meta.url), 'utf8');
const componentSource = await readFile(new URL('../components/admin-telegram-actions.js', import.meta.url), 'utf8');

test('telegram runtime turns packet and route state into a single founder-facing attention summary', () => {
  assert.match(runtimeSource, /function buildTelegramAttention\(status, packet\)/);
  assert.match(runtimeSource, /mode:\s*'restore_delivery'/);
  assert.match(runtimeSource, /mode:\s*'review_room'/);
  assert.match(runtimeSource, /summary:\s*primaryAction/);
});

test('telegram admin surface shows next action before packet body detail', () => {
  assert.match(componentSource, /const attention = runtime\?\.attention \|\| null/);
  assert.match(componentSource, /payload\?\.telegram\?\.attention\?\.summary \|\| 'telegram packet sent'/);
  assert.match(componentSource, /<span className=\"signal-card__source\">다음 액션<\/span>/);
  assert.match(componentSource, /Telegram preview는 founder에게 지금 가장 먼저 볼 액션을 한 줄로 올려주는 운영면입니다/);
});
