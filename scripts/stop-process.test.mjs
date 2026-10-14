import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import test from 'node:test';
import { stop } from './process.mjs';

async function ownedNode(t, source) {
  const child = spawn(process.execPath, ['-e', source], {
    stdio: 'ignore', windowsHide: true, detached: process.platform !== 'win32',
  });
  t.after(async () => {
    await new Promise(resolve => setImmediate(resolve));
    if (child.exitCode !== null || child.signalCode !== null) return;
    const exited = once(child, 'exit');
    child.kill('SIGKILL'); // 只终止本测试原先持有的子进程。
    await exited;
  });
  await once(child, 'spawn');
  return child;
}

test('stop confirms a queued exit without misreporting a missing Windows PID', async t => {
  const child = await ownedNode(t, 'process.exit(0)');
  // 控制本测试的事件循环暂停，使实际子进程退出事件尚未分发。
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 1000);
  assert.equal(child.exitCode, null);
  await stop(child);
  assert.equal(child.exitCode, 0);
  await stop(child);
});

test('stop confirms termination of its live owned child and is idempotent', async t => {
  const child = await ownedNode(t, 'setInterval(() => {}, 1000)');
  await stop(child);
  assert.ok(child.exitCode !== null || child.signalCode !== null);
  await stop(child);
});
