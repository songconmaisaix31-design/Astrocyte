import test from 'node:test';
import assert from 'node:assert/strict';
import { access, realpath } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { root } from '../../scripts/process.mjs';
import { startS1Server } from './server.mjs';
import { humanAPI } from './api.mjs';

for (const relay of [false, true]) test(`production server keeps owned SQLite after close and reopens it without private discovery (relay layout=${relay})`, { timeout: 120_000 }, async () => {
  let server = await startS1Server({ relay });
  const original = server.temporary;
  try {
    assert.ok(relative(join(root, '.verification'), await realpath(original)).startsWith('astrocyte-s1-'));
    if (relay) {
      assert.equal(server.localAppData, join(original, 'local-appdata'));
      assert.equal(server.relayRoot, join(server.localAppData, 'astrocyte-workbench', 'v0.2-access'));
      assert.equal(server.dataDir, join(server.relayRoot, 'app-state'));
    }
    let api = await humanAPI(server.apiURL);
    assert.deepEqual((await api.get('/materials')).items, []);
    assert.deepEqual((await api.get('/local-projects')).items, []);
    assert.equal(server.query('PRAGMA integrity_check')[0].integrity_check, 'ok');
    const { snapshot: discovery } = await api.write('/local-projects/registered/refresh', {});
    assert.equal(discovery.projects.length, 1);
    assert.equal(discovery.projects[0].source, 'explicit_directory');
    assert.equal(await realpath(discovery.projects[0].root), await realpath(root));
    await assert.rejects(server.startAPI(), /still running/);
    await server.stopAPI();
    await assert.rejects(fetch(`${server.apiURL}/api/v1/health`, { signal: AbortSignal.timeout(1000) }));
    await server.startAPI();
    const oldURL = server.apiURL;
    await server.close();
    await access(join(server.dataDir, 'state.sqlite'));
    await assert.rejects(fetch(`${oldURL}/api/v1/health`, { signal: AbortSignal.timeout(1000) }));
    server = await startS1Server({ relay, reuseOwnedTemporary: { path: original, ownedRoot: original } });
    assert.equal(server.temporary, original);
    api = await humanAPI(server.apiURL);
    assert.deepEqual((await api.get('/materials')).items, []);
    assert.equal(server.query('PRAGMA integrity_check')[0].integrity_check, 'ok');
    await server.restart();
    api = await humanAPI(server.apiURL);
    assert.deepEqual((await api.get('/materials')).items, []);
  } finally { await server.close(); }
});
