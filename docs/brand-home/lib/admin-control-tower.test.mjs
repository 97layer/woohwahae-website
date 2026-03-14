import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const jobsPageSource = await readFile(new URL('../app/admin/(console)/jobs/page.js', import.meta.url), 'utf8');
const shellSource = await readFile(new URL('../components/admin-shell.js', import.meta.url), 'utf8');
const controlTowerSource = await readFile(new URL('../components/admin-control-tower.js', import.meta.url), 'utf8');

test('admin jobs page mounts the control tower inside the same /admin app', () => {
  assert.match(jobsPageSource, /import AdminControlTower from '\.\.\/\.\.\/\.\.\/\.\.\/components\/admin-control-tower'/);
  assert.match(jobsPageSource, /import \{ getAdminOverviewView \} from '\.\.\/\.\.\/\.\.\/\.\.\/lib\/runtime\/view-model'/);
  assert.match(jobsPageSource, /8081\/admin 작업 운영판/);
  assert.match(jobsPageSource, /첫 화면은 바로 열리고, 실시간 작업 상태는 아래 작업판이 즉시 이어서 불러옵니다/);
  assert.match(jobsPageSource, /const founderAttention = overview\.founderAttention \|\| null/);
  assert.match(jobsPageSource, /founderAttention\?\.summary \|\| overview\.jobs\.attentionHint\?\.summary/);
  assert.match(jobsPageSource, /<AdminControlTower canWrite=\{session\.canWrite\} initialData=\{null\} initialJobId=\{jobId\} \/>/);
});

test('admin shell exposes the control tower as a first-class admin lane', () => {
  assert.match(shellSource, /href: '\/admin\/jobs', label: '컨트롤타워'/);
  assert.match(shellSource, /`\/admin` 아래에서 Layer OS 백엔드 하나만 기준으로 운영합니다/);
});

test('control tower client polls the admin jobs route and exposes live mutations', () => {
  assert.match(controlTowerSource, /fetch\('\/api\/admin\/runtime\/jobs/);
  assert.match(controlTowerSource, /fetch\(`\/api\/admin\/runtime\/jobs\/packet\?\$\{params\.toString\(\)\}`/);
  assert.match(controlTowerSource, /fetch\('\/api\/admin\/runtime\/token/);
  assert.match(controlTowerSource, /setInterval\(\(\) => \{/);
  assert.match(controlTowerSource, /if \(!refreshLockRef\.current\)/);
  assert.match(controlTowerSource, /runAction\('dispatch'/);
  assert.match(controlTowerSource, /runAction\('cancel'/);
  assert.match(controlTowerSource, /runAction\('promote'/);
  assert.match(controlTowerSource, /async function loadSelectedPacket\(jobId = selectedJobId\)/);
  assert.doesNotMatch(controlTowerSource, /useEffect\(\(\) => \{\s*if \(!selectedJobId \|\| data\?\.selectedJob\?\.packet \|\| refreshLockRef\.current\)/);
  assert.match(controlTowerSource, /작업 규칙 불러오기/);
  assert.match(controlTowerSource, /필요할 때만 packet을 읽습니다/);
  assert.match(controlTowerSource, /다음 액션/);
  assert.match(controlTowerSource, /보고 후속/);
  assert.match(controlTowerSource, /해당 작업 보기/);
  assert.match(controlTowerSource, /report가 남긴 다음 액션을 먼저 보고, 더 넓은 맥락이 필요할 때만 packet을 읽습니다/);
  assert.match(controlTowerSource, /작업 배정/);
  assert.match(controlTowerSource, /다음 안건 실행/);
  assert.match(controlTowerSource, /일시 정지/);
  assert.match(controlTowerSource, /syncJobIdToUrl/);
  assert.match(controlTowerSource, /실시간 작업 상태를 불러오는 중입니다/);
  assert.match(controlTowerSource, /조작 권한 연결/);
  assert.match(controlTowerSource, /토큰 연결/);
  assert.match(controlTowerSource, /선택한 작업/);
});
