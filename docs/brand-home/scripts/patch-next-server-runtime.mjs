import { readFile, writeFile } from 'node:fs/promises';

/**
 * Purpose: Next.js 15 emits a server runtime that attempts to load chunk files from
 * "./${u(chunkId)}" next to `webpack-runtime.js`, but app chunks live under
 * `./chunks/` while vendor chunks still live under `./vendor-chunks/`. This patch
 * rewrites the loader to route chunk paths to the correct directory so dev/build
 * starts don't fail with `MODULE_NOT_FOUND` when resolving dynamic chunks.
 *
 * Behavior: If the runtime is already patched, the script exits quietly. If a new
 * format appears that doesn't match either pattern, we throw to flag the change
 * instead of silently producing a broken runtime.
 */

const runtimePath = new URL('../.next/server/webpack-runtime.js', import.meta.url);
const source = await readFile(runtimePath, 'utf8');

const chunkRequirePattern = /require\(\s*["']\.\/(?:chunks\/)?["']\s*\+\s*([A-Za-z_$][\w$]*)\.u\(([^)]*)\)\s*\)/;
const alreadyPatched =
  /require\(\s*\(\s*\(\s*[A-Za-z_$][\w$]*\.u\([^)]+\)\s*\)\.startsWith\(["']vendor-chunks\/["']\)\s*\?\s*["']\.\/["']\s*:\s*["']\.\/chunks\/["']\s*\)\s*\+\s*[A-Za-z_$][\w$]*\.u\([^)]+\)\s*\)/;

if (alreadyPatched.test(source)) {
  process.exit(0);
}

const match = chunkRequirePattern.exec(source);
if (match) {
  const [, ident, args] = match;
  const replacement = `require(((${ident}.u(${args})).startsWith("vendor-chunks/") ? "./" : "./chunks/") + ${ident}.u(${args}))`;
  const next = source.replace(chunkRequirePattern, replacement);
  await writeFile(runtimePath, next, 'utf8');
  process.exit(0);
}

throw new Error('Unexpected Next server runtime format; chunk loader patch was not applied.');
