import { spawn } from 'node:child_process';
import { readFile, writeFile } from 'node:fs/promises';
import { createServer } from 'node:net';
import { join, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import assert from 'node:assert/strict';

const root = resolve(import.meta.dirname, '../..');
const mode = process.env.ASTROCYTE_S5_NETWORK_MODE;
assert.ok(['http-proxy', 'os-program-block'].includes(mode), 'Network scope must be explicit');
const node = process.env.ASTROCYTE_S5_NODE ?? process.execPath;
const assetRoot = process.env.ASTROCYTE_S5_ASSET_ROOT;
assert.ok(assetRoot, 'The reviewed local asset root is required');
const rpcEntry = join(root, '.verification/s5/ipc/node_modules/vscode-jsonrpc/lib/node/main.js');
const sdkEntry = join(root, '.verification/s5/upstream/laya/laya-ts/dist/index.js');
const { createMessageConnection, StreamMessageReader, StreamMessageWriter } = await import(pathToFileURL(rpcEntry).href);
const source = await readFile(join(root, '.verification/s5/upstream/laya/laya-ts/README.md'), 'utf8');
const state = source.slice(0, source.indexOf('## Node'));
let proxy;
if (mode === 'http-proxy') {
  const server = createServer();
  await new Promise((resolveListen, reject) => { server.once('error', reject); server.listen(0, '127.0.0.1', resolveListen); });
  proxy = `http://127.0.0.1:${server.address().port}`;
  await new Promise((resolveClose, reject) => server.close(error => error ? reject(error) : resolveClose()));
}
const reports = [];
for (const round of [1, 2]) {
  const networkEvidence = join(root, `.verification/s5/offline-${mode}-network-${round}.json`);
  const env = { ...process.env, ASTROCYTE_S5_NETWORK_EVIDENCE: networkEvidence };
  if (proxy) Object.assign(env, { HTTP_PROXY: proxy, HTTPS_PROXY: proxy, ALL_PROXY: proxy, NO_PROXY: '', http_proxy: proxy, https_proxy: proxy, no_proxy: '' });
  const child = spawn(node, [...(proxy ? ['--use-env-proxy'] : []), join(root, 'tests/s5/offline-worker.mjs'), '--model-dir', join(assetRoot, 'english'), '--assets', join(root, '.verification/s5/laya-english-assets.json'), '--sdk-entry', sdkEntry, '--rpc-entry', rpcEntry, '--language', 'en'], { env, windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] });
  let stderr = '';
  child.stderr.setEncoding('utf8');
  child.stderr.on('data', text => { stderr += text; });
  const exited = new Promise((resolveExit, reject) => { child.once('exit', (code, signal) => resolveExit({ code, signal })); child.once('error', reject); });
  const connection = createMessageConnection(new StreamMessageReader(child.stdout), new StreamMessageWriter(child.stdin));
  connection.listen();
  const timeout = setTimeout(() => child.kill(), 120000);
  try {
    const health = await connection.sendRequest('judgment/health');
    const network = JSON.parse(await readFile(networkEvidence, 'utf8'));
    assert.equal(network.pid, child.pid);
    assert.equal(health.pid, child.pid);
    assert.equal(health.device, 'cpu');
    assert.equal(health.provider_model, 'laya-rl-agent');
    if (round === 2) assert.notEqual(child.pid, reports[0].health.pid);
    const questions = [
      { question_id: 'runtime', question_type: 'choice', instructions: 'Which runtime is described?', options: [{ id: 'typescript', description: 'TypeScript Agent inference' }, { id: 'music', description: 'Music playback' }] },
      { question_id: 'export', question_type: 'truth', instructions: 'Does the documented export verify torch versus ONNX?' },
      { question_id: 'detail', question_type: 'score', instructions: 'How explicit is the export verification?', levels: ['No verification', 'Verification without tolerance', 'Verification with numerical tolerance'] },
    ];
    const answers = [];
    for (const question of questions) {
      const answer = await connection.sendRequest('judgment/evaluate', { question, state, language: 'en', model_revision: health.model_revision, budget: { timeout_ms: 60000, max_input_bytes: 131072, max_input_tokens: 2048, max_output_tokens: 0 } });
      assert.equal(answer.model_revision, health.model_revision);
      assert.equal(answer.raw.model, 'laya-rl-agent');
      assert.equal(answer.raw.usage.truncated, false);
      assert.ok(answer.raw.usage.input_tokens > 0);
      answers.push(answer);
    }
    assert.equal((await connection.sendRequest('judgment/health')).pid, child.pid);
    child.stdin.end();
    const exit = await exited;
    assert.equal(exit.code, 0);
    const report = { mode, round, network, health, answers, stderr, exit };
    await writeFile(join(root, `.verification/s5/offline-${mode}-${round}.json`), JSON.stringify(report, null, 2));
    reports.push({ mode, round, health, networkFailure: network.failure.cause.code ?? network.failure.name, exit });
  } finally {
    clearTimeout(timeout);
    connection.dispose();
    if (child.exitCode === null) child.kill();
    await writeFile(join(root, `.verification/s5/offline-${mode}-stderr-${round}.log`), stderr);
  }
}
await writeFile(join(root, `.verification/s5/offline-${mode}-safe.json`), JSON.stringify({ scope: mode === 'http-proxy' ? 'These two Node processes HTTP outlet only' : 'These two dedicated Node executables OS outbound rule only', os_full_network: 'NOT_RUN', reports }, null, 2));
console.log(JSON.stringify(reports));
