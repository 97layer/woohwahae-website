import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const reviewRoomPageSource = await readFile(new URL('../app/admin/(console)/review-room/page.js', import.meta.url), 'utf8');

test('admin review room separates fresh blockers from older unresolved items', () => {
  assert.match(reviewRoomPageSource, /import \{ Suspense \} from 'react'/);
  assert.match(reviewRoomPageSource, /getReviewRoomDetailView,\s*getReviewRoomSummaryView/);
  assert.match(reviewRoomPageSource, /triageReviewRoom/);
  assert.match(reviewRoomPageSource, /SummaryReviewItems/);
  assert.match(reviewRoomPageSource, /LoadingReviewRoomDetail/);
  assert.match(reviewRoomPageSource, /runtime 응답이 느려 review-room 세부 카드를 뒤에서 이어서 불러옵니다/);
  assert.match(reviewRoomPageSource, /summaryView\.attention\?\.summary/);
  assert.match(reviewRoomPageSource, /summaryView\.attention\?\.detail/);
  assert.match(reviewRoomPageSource, /What is actually blocking now/);
  assert.match(reviewRoomPageSource, /Runtime residue/);
  assert.match(reviewRoomPageSource, /Strategic backlog/);
  assert.match(reviewRoomPageSource, /Other unresolved/);
  assert.match(reviewRoomPageSource, /runtime residue items/);
  assert.match(reviewRoomPageSource, /active blockers are clear/);
  assert.match(reviewRoomPageSource, /local fallback/);
  assert.match(reviewRoomPageSource, /absorbed snapshot/);
  assert.match(reviewRoomPageSource, /VM continuity host remains the authority/);
  assert.match(reviewRoomPageSource, /<Suspense fallback=\{<LoadingReviewRoomDetail \/>\}>/);
});
