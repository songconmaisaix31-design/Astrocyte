import { spawn } from 'node:child_process';
import { readFile, writeFile } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { createHash } from 'node:crypto';
import assert from 'node:assert/strict';

const root = resolve(import.meta.dirname, '../..');
const rpcEntry = join(root, '.verification/s5/ipc/node_modules/vscode-jsonrpc/lib/node/main.js');
const { createMessageConnection, StreamMessageReader, StreamMessageWriter } = await import(pathToFileURL(rpcEntry).href);
const sdkEntry = join(root, '.verification/s5/upstream/laya/laya-ts/dist/index.js');
const sdkText = await readFile(join(root, '.verification/s5/upstream/laya/laya-ts/README.md'), 'utf8');
const english = sdkText.slice(0, sdkText.indexOf('## Node'));
const chineseSource = await readFile(join(root, '.verification/s5/upstream/astrocyte-readme.md'), 'utf8');
const reports = [];

async function loadAssets(checkpoint) {
  const assetRoot = process.env.ASTROCYTE_S5_ASSET_ROOT;
  if (!assetRoot) throw new Error('ASTROCYTE_S5_ASSET_ROOT must explicitly name the owned local assets');
  const modelDir = join(assetRoot, checkpoint);
  const files = {};
  for (const name of ['encoder.onnx', 'encoder.onnx.data', 'head.onnx', 'head.onnx.data', 'tokenizer.json', 'rl_agent_config.json']) {
    files[name] = createHash('sha256').update(await readFile(join(modelDir, name))).digest('hex');
  }
  const assets = { repo: 'convaiinnovations/laya', model_revision: '7b928d828b7b0e022f929d9bd2e44165aa270148', source_revision: '432312b36714d8889eba405507b37eda25257d9c', device: 'cpu', checkpoint, files };
  const assetsPath = join(root, `.verification/s5/laya-${checkpoint}-assets.json`);
  await writeFile(assetsPath, JSON.stringify(assets, null, 2));
  return { assets, assetsPath, modelDir };
}

async function session(checkpoint, language, runIndex) {
  const bundle = await loadAssets(checkpoint);
  const child = spawn(process.execPath, [join(root, 'internal/adapters/judgment/laya-worker.mjs'), '--model-dir', bundle.modelDir, '--assets', bundle.assetsPath, '--sdk-entry', sdkEntry, '--rpc-entry', rpcEntry, '--language', language], { windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] });
  let stderr = '';
  child.stderr.setEncoding('utf8');
  child.stderr.on('data', text => { stderr += text; });
  const connection = createMessageConnection(new StreamMessageReader(child.stdout), new StreamMessageWriter(child.stdin));
  const exited = new Promise((resolveExit, rejectExit) => { child.on('exit', (code, signal) => resolveExit({ code, signal })); child.on('error', rejectExit); });
  connection.listen();
  const timeout = setTimeout(() => { child.kill(); }, 120000);
  const health = await connection.sendRequest('judgment/health');
  assert.equal(health.device, 'cpu');
  assert.equal(health.language, language);
  assert.equal(health.pid, child.pid);
  assert.equal(health.sdk_reported_revision, null);
  const state = language === 'en' ? english : chineseSource.slice(0, 600);
  const questions = language === 'en' ? [
    { question_id: 'runtime', question_type: 'choice', instructions: 'Which language implements the documented local inference runtime?', options: [{ id: 'typescript', description: 'TypeScript implementation' }, { id: 'python', description: 'Python implementation' }], levels: [], rule: '' },
    { question_id: 'verify', question_type: 'truth', instructions: 'Does the official export verify torch versus ONNX to within 1e-4?', options: [], levels: [], rule: '' },
    { question_id: 'detail', question_type: 'score', instructions: 'How explicit is the export numerical agreement?', options: [], levels: ['Not documented', 'Documented without tolerance', 'A numerical tolerance is documented'], rule: '' },
  ] : [
    { question_id: 'purpose', question_type: 'choice', instructions: '文档描述的平台用途是什么？', options: [{ id: 'agent_platform', description: 'Agent 和资料的工作平台' }, { id: 'music_player', description: '音乐播放软件' }], levels: [], rule: '' },
    { question_id: 'go', question_type: 'truth', instructions: '文档是否涉及 Agent 或人工批准？', options: [], levels: [], rule: '' },
    { question_id: 'detail', question_type: 'score', instructions: '文档对项目用途的描述程度是什么？', options: [], levels: ['没有说明', '只写名称', '说明了用途和工作方式'], rule: '' },
  ];
  const answers = [];
  for (const question of questions) {
    const response = await connection.sendRequest('judgment/evaluate', { question, state, language, model_revision: health.model_revision, budget: { timeout_ms: 60000, max_input_bytes: 65536, max_input_tokens: 2048, max_output_tokens: 0 } });
    assert.equal(response.model_revision, health.model_revision);
    assert.equal(response.raw.model, 'laya-rl-agent');
    assert.equal(response.raw.usage.truncated, false);
    answers.push(response);
  }
  const longResult = await connection.sendRequest('judgment/evaluate', { question: questions[0], state: language === 'en' ? sdkText : chineseSource, language, model_revision: health.model_revision, budget: { timeout_ms: 60000, max_input_bytes: 131072, max_input_tokens: 2048, max_output_tokens: 0 } });
  assert.equal(longResult.raw.usage.truncated, true);
  const sameHealth = await connection.sendRequest('judgment/health');
  assert.equal(sameHealth.pid, child.pid);
  child.stdin.end();
  const exit = await exited;
  clearTimeout(timeout);
  connection.dispose();
  assert.equal(exit.code, 0);
  const report = { checkpoint, language, runIndex, health, answers, longResult, stderr, exit };
  await writeFile(join(root, `.verification/s5/rpc-${checkpoint}-${language}-${runIndex}.json`), JSON.stringify(report, null, 2));
  reports.push({ checkpoint, language, runIndex, pid: health.pid, model_revision: health.model_revision, questions: answers.length, truncated: longResult.raw.usage.truncated, exit });
}

for (const [checkpoint, language] of [['english', 'en'], ['multilingual', 'en'], ['multilingual', 'zh']]) {
  for (const runIndex of [1, 2]) await session(checkpoint, language, runIndex);
}
await writeFile(join(root, '.verification/s5/rpc-live-safe.json'), JSON.stringify(reports, null, 2));
console.log(JSON.stringify(reports));
