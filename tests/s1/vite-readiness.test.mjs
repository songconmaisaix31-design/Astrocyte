import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { startS1Server } from './server.mjs';

async function actualWebSocket(url) {
  const socket = new WebSocket(url.replace('http:', 'ws:'), 'vite-ping');
  await new Promise((resolve, reject) => {
    const timer = setTimeout(() => { socket.close(); reject(new Error('Actual Vite WebSocket did not use its owned HTTP listener')); }, 1000);
    socket.addEventListener('open', () => { clearTimeout(timer); resolve(); }, { once: true });
    socket.addEventListener('error', event => { clearTimeout(timer); reject(event.error ?? new Error('Actual Vite WebSocket failed')); }, { once: true });
  });
  socket.close();
}

test('two cold actual Vite servers serve HTML and WebSocket through their own announced listeners', { timeout: 120_000 }, async () => {
  const servers = [];
  try {
    const attempts = await Promise.allSettled([0, 1].map(async () => {
      const server = await startS1Server({ browser: true });
      servers.push(server);
      const response = await fetch(server.webURL, { signal: AbortSignal.timeout(1000) });
      assert.equal(response.status, 200);
      assert.match(await response.text(), /<html/);
      await actualWebSocket(server.webURL);
    }));
    const failures = attempts.filter(result => result.status === 'rejected');
    if (failures.length) throw new AggregateError(failures.map(result => result.reason), 'Actual concurrent Vite readiness failed');
    assert.equal(servers.length, 2);
    assert.notEqual(servers[0].webURL, servers[1].webURL);
  } finally {
    const results = await Promise.allSettled(servers.map(server => server.close()));
    const failures = results.filter(result => result.status === 'rejected');
    if (failures.length) throw new AggregateError(failures.map(result => result.reason), 'Actual concurrent Vite cleanup failed');
  }
  for (const server of servers) {
    const log = server.processLogs.find(item => item.name === 'vite');
    assert.doesNotMatch(await readFile(log.stderr, 'utf8'), /WebSocket server error|24678/);
  }
});
