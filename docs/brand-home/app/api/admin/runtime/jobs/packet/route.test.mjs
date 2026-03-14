import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const source = await readFile(new URL('./route.js', import.meta.url), 'utf8');

test('job packet admin route exposes founder/admin read access only', () => {
  assert.match(source, /authorizeRequest\(request,\s*\['founder', 'admin'\]\)/);
  assert.match(source, /getControlTowerPacket\(jobId\)/);
  assert.match(source, /job_id is required/);
});
