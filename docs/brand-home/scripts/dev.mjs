import { spawn } from 'node:child_process';
import { stat } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const patchScriptPath = fileURLToPath(new URL('./patch-next-server-runtime.mjs', import.meta.url));
const runtimePath = fileURLToPath(new URL('../.next/server/webpack-runtime.js', import.meta.url));
const nextArgs = ['next', 'dev', '-p', '8081', ...process.argv.slice(2)];

let runtimeMtimeMs = 0;
let patchRunning = false;

function runPatch() {
  return new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [patchScriptPath], {
      env: process.env,
      stdio: 'inherit',
    });

    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }

      reject(new Error(`patch-next-server-runtime exited with code ${code ?? 'null'} signal ${signal ?? 'null'}`));
    });

    child.on('error', reject);
  });
}

async function maybePatchRuntime() {
  if (patchRunning) return;

  try {
    const info = await stat(runtimePath);
    if (info.mtimeMs <= runtimeMtimeMs) return;

    patchRunning = true;
    await runPatch();
    runtimeMtimeMs = info.mtimeMs;
  } catch (error) {
    if (error && error.code === 'ENOENT') return;
    console.error('[dev-runtime-patch]', error);
  } finally {
    patchRunning = false;
  }
}

const nextDev = spawn('npx', nextArgs, {
  env: process.env,
  stdio: 'inherit',
});

const patchTimer = setInterval(() => {
  void maybePatchRuntime();
}, 250);

void maybePatchRuntime();

function shutdown(signal) {
  clearInterval(patchTimer);
  if (!nextDev.killed) nextDev.kill(signal);
}

process.on('SIGINT', () => shutdown('SIGINT'));
process.on('SIGTERM', () => shutdown('SIGTERM'));

nextDev.on('exit', (code, signal) => {
  clearInterval(patchTimer);
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 0);
});

nextDev.on('error', (error) => {
  clearInterval(patchTimer);
  console.error('[dev-runtime-patch]', error);
  process.exit(1);
});
