import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const pageSource = await readFile(new URL('../app/admin/(console)/knowledge/page.js', import.meta.url), 'utf8');
const viewModelSource = await readFile(new URL('./runtime/view-model.js', import.meta.url), 'utf8');

test('knowledge runtime splits summary and detail reads', () => {
  assert.match(viewModelSource, /export async function getKnowledgeSummaryView\(\)/);
  assert.match(viewModelSource, /fetchLayerOs\('\/api\/layer-os\/knowledge'\)/);
  assert.match(viewModelSource, /function summarizeKnowledgeAttention\(knowledge\)/);
  assert.match(viewModelSource, /attention:\s*summarizeKnowledgeAttention\(knowledge\)/);
  assert.match(viewModelSource, /export async function getKnowledgeDetailView\(query = ''\)/);
  assert.match(viewModelSource, /fetchLayerOs\('\/api\/layer-os\/corpus\/entries'\)/);
  assert.match(viewModelSource, /query\.trim\(\) \? fetchLayerOs\(`\/api\/layer-os\/knowledge\/search\?q=\$\{encodeURIComponent\(query\.trim\(\)\)\}`\)/);
  assert.match(viewModelSource, /export async function getKnowledgeView\(query = ''\)/);
  assert.match(viewModelSource, /const \[summary, detail\] = await Promise\.all/);
});

test('knowledge admin page streams corpus detail after summary', () => {
  assert.match(pageSource, /import \{ Suspense \} from 'react'/);
  assert.match(pageSource, /getKnowledgeDetailView, getKnowledgeSummaryView/);
  assert.match(pageSource, /async function KnowledgeDetailSection\(\{ query \}\)/);
  assert.match(pageSource, /function LoadingKnowledgeDetail\(\{ query \}\)/);
  assert.match(pageSource, /지금 중요한 지식 신호를 먼저 보고, 최근 corpus와 검색 결과는 뒤에서 이어서 불러옵니다/);
  assert.match(pageSource, /summary\.attention\?\.summary/);
  assert.match(pageSource, /summary\.attention\?\.detail/);
  assert.match(pageSource, /<Suspense fallback=\{<LoadingKnowledgeDetail query=\{query\} \/>\}>/);
  assert.doesNotMatch(pageSource, /const knowledge = await getKnowledgeView\(query\)/);
  assert.match(pageSource, /const summary = await getKnowledgeSummaryView\(\)/);
});
