import { readFile, writeFile } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { createHash } from 'node:crypto';
import assert from 'node:assert/strict';

const root = resolve(import.meta.dirname, '../..');
const source = join(root, '.verification/s5/upstream/laya');
const modelDir = join(root, '.verification/s5/assets/onnx');
const { Agent } = await import(pathToFileURL(join(source, 'laya-ts/dist/index.js')).href);
const packageInfo = JSON.parse(await readFile(join(source, 'laya-ts/package.json'), 'utf8'));
const digests = {};
for (const name of ['encoder.onnx', 'encoder.onnx.data', 'head.onnx', 'head.onnx.data', 'tokenizer.json', 'rl_agent_config.json']) {
  digests[name] = createHash('sha256').update(await readFile(join(modelDir, name))).digest('hex');
}
const assetInfo = {
  repo: 'convaiinnovations/laya',
  model_revision: '7b928d828b7b0e022f929d9bd2e44165aa270148',
  source_revision: '432312b36714d8889eba405507b37eda25257d9c',
  package_version: packageInfo.version,
  device: 'cpu', checkpoint: 'english',
  files: digests,
};
await writeFile(join(root, '.verification/s5/laya-assets.json'), JSON.stringify(assetInfo, null, 2));

const agent = await Agent.load(modelDir, { device: 'cpu', expectedSha256: digests });
const state = await readFile(join(source, 'laya-ts/README.md'), 'utf8');
// 真实官方文档的短片段和全文分别覆盖正常输入与自然长输入。
const short = state.slice(0, state.indexOf('## Node'));
const questions = {
  runtime: { type: 'choice', instructions: 'Which runtime is used in the code example?', criteria: { node: 'Node local inference', browser: 'Browser inference', python: 'Python inference' } },
  verify: { type: 'noul', instructions: 'Does the export script verify torch versus ONNX numerical agreement within 1e-4?' },
  detail: { type: 'score', instructions: 'How explicit is the export verification described?', criteria: ['Not mentioned', 'Mentioned without tolerance', 'An exact numerical tolerance is stated'] },
};
const started = performance.now();
const shortResult = await agent.predict(short, questions, { lang: 'en' });
assert.equal(shortResult.model, 'laya-rl-agent');
assert.deepEqual(Object.keys(shortResult.answers).sort(), Object.keys(questions).sort());
assert.ok(Number.isInteger(shortResult.usage.input_tokens));
assert.equal(shortResult.usage.truncated, false);
for (const [id, answer] of Object.entries(shortResult.answers)) {
  assert.equal(answer.type, questions[id].type);
  if (answer.type === 'noul') assert.ok(answer.noul >= 0 && answer.noul <= 1);
  else assert.ok(answer.confidence >= 0 && answer.confidence <= 1);
}
const longResult = await agent.predict(state, { runtime: questions.runtime }, { lang: 'en' });
assert.equal(longResult.usage.truncated, true);
assert.ok(longResult.usage.state_tokens_dropped > 0);
assert.ok(longResult.usage.truncated_questions.includes('runtime'));

const { marked } = await import(pathToFileURL(join(root, '.verification/s5/runtime/node_modules/marked/lib/marked.esm.js')).href);
const card = await readFile(join(root, '.verification/s5/upstream/hf-model-card.md'), 'utf8');
const headingOptions = Object.fromEntries(marked.lexer(state + '\n' + card).filter(token => token.type === 'heading' || token.type === 'paragraph').slice(0, 100).map((token, index) => [`Official Laya documentation section: ${token.text} (passage ${index})`, token.text]));
// 每个候选引用真实原文段落；使用真实文档候选验证预算导致的覆盖损失。
const optionsResult = await agent.predict(short, { headings: { type: 'choice', instructions: 'Which section documents CPU Node inference?', criteria: headingOptions } }, { lang: 'en' });
assert.ok(optionsResult.usage.options?.headings);
assert.equal(optionsResult.usage.options.headings.total, Object.keys(headingOptions).length);
assert.ok(optionsResult.usage.options.headings.distinct < optionsResult.usage.options.headings.total);
const report = { scope: 'actual_official_laya_sdk_cpu', assetInfo, sdk_reported_revision: agent.revision, shortResult, longResult, optionsResult, elapsed_ms: performance.now() - started };
await writeFile(join(root, '.verification/s5/laya-live.json'), JSON.stringify(report, null, 2));
console.log(JSON.stringify({ scope: report.scope, actual_model: shortResult.model, sdk_reported_revision: agent.revision, usage: shortResult.usage, truncated: longResult.usage, options: optionsResult.usage.options, elapsed_ms: report.elapsed_ms }));
