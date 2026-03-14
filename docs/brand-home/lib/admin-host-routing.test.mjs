import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const middlewareSource = await readFile(new URL('../middleware.js', import.meta.url), 'utf8');
const tlsEdgeSource = await readFile(
  new URL('../../../scripts/nginx/woohwahae-public.tls.conf.example', import.meta.url),
  'utf8',
);
const dnsScriptSource = await readFile(
  new URL('../../../scripts/switch_woohwahae_dns.sh', import.meta.url),
  'utf8',
);

test('middleware keeps the admin surface on admin.woohwahae.kr', () => {
  assert.match(middlewareSource, /canonicalAdminHost = .*admin\.woohwahae\.kr/);
  assert.match(middlewareSource, /canonicalPublicHosts = new Set/);
  assert.match(middlewareSource, /function isAdminSurfacePath\(pathname\)/);
  assert.match(middlewareSource, /function resolveRequestHost\(request\)/);
  assert.match(middlewareSource, /if \(canonicalPublicHosts\.has\(hostname\) && isAdminSurfacePath\(pathname\)\)/);
  assert.match(middlewareSource, /url\.hostname = canonicalAdminHost/);
  assert.match(middlewareSource, /if \(hostname === canonicalAdminHost && pathname === '\/'\)/);
  assert.match(middlewareSource, /matcher: \['\/', '\/api\/admin\/:path\*', '\/admin\/:path\*'\]/);
});

test('TLS edge config splits the public apex and admin subdomain', () => {
  assert.match(tlsEdgeSource, /server_name woohwahae\.kr www\.woohwahae\.kr/);
  assert.match(tlsEdgeSource, /return 308 https:\/\/admin\.woohwahae\.kr\$request_uri/);
  assert.match(tlsEdgeSource, /server_name admin\.woohwahae\.kr/);
  assert.match(tlsEdgeSource, /return 302 https:\/\/admin\.woohwahae\.kr\/admin\/login/);
});

test('DNS switch script carries the dedicated admin record', () => {
  assert.match(dnsScriptSource, /ADMIN_HOST="\$\{ADMIN_HOST:-admin\.woohwahae\.kr\}"/);
  assert.match(dnsScriptSource, /keeping www\.woohwahae\.kr proxied to the apex and admin\.woohwahae\.kr pointed at/);
  assert.match(dnsScriptSource, /for name in woohwahae\.kr www\.woohwahae\.kr "\$\{ADMIN_HOST\}" edgecheck\.woohwahae\.kr/);
  assert.match(dnsScriptSource, /admin_json="\$\(find_record "\$\{ADMIN_HOST\}"\)"/);
});
