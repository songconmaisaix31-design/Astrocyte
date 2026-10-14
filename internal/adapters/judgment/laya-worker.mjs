import { readFile } from 'node:fs/promises';
import { join, isAbsolute, basename } from 'node:path';
import { pathToFileURL } from 'node:url';
import { parseArgs } from 'node:util';
import { createHash } from 'node:crypto';

const { values } = parseArgs({ options: {
  'model-dir': { type: 'string' }, assets: { type: 'string' },
  'sdk-entry': { type: 'string' }, 'rpc-entry': { type: 'string' },
  language: { type: 'string' },
} });
for (const name of ['model-dir', 'assets', 'sdk-entry', 'rpc-entry', 'language']) {
  if (!values[name]) throw new Error(`Laya worker requires explicit --${name}`);
}
for (const name of ['model-dir', 'assets', 'sdk-entry', 'rpc-entry']) {
  if (!isAbsolute(values[name])) throw new Error(`Laya worker --${name} must be an absolute configured path`);
}
const assets = JSON.parse(await readFile(values.assets, 'utf8'));
if (assets.repo !== 'convaiinnovations/laya' || assets.model_revision !== '7b928d828b7b0e022f929d9bd2e44165aa270148' || assets.source_revision !== '432312b36714d8889eba405507b37eda25257d9c') {
  throw new Error('Laya assets must match the reviewed fixed source and model revisions');
}
if (assets.device !== 'cpu' || !['english', 'multilingual'].includes(assets.checkpoint)) throw new Error('Laya assets require an explicit reviewed CPU checkpoint');
if (assets.checkpoint === 'english' && values.language !== 'en') throw new Error('English Laya checkpoint only supports explicit en');
const expectedNames = ['encoder.onnx', 'encoder.onnx.data', 'head.onnx', 'head.onnx.data', 'tokenizer.json', 'rl_agent_config.json'];
if (expectedNames.some(name => !assets.files[name])) throw new Error('Laya assets require all graphs, sidecars, tokenizer and configuration digests');
for (const [name, expected] of Object.entries(assets.files)) {
  if (basename(name) !== name) throw new Error('Laya asset path escapes the model directory');
  const actual = createHash('sha256').update(await readFile(join(values['model-dir'], name))).digest('hex');
  if (actual !== expected) throw new Error(`Laya asset digest mismatch: ${name}`);
}
const { Agent } = await import(pathToFileURL(values['sdk-entry']).href);
const { createMessageConnection, StreamMessageReader, StreamMessageWriter, ResponseError } = await import(pathToFileURL(values['rpc-entry']).href);
const cfg = JSON.parse(await readFile(join(values['model-dir'], 'rl_agent_config.json'), 'utf8'));
// 官方校准值的无效题型桶在推理前拒绝，不让 SDK 的 clamp 改写该题型输出。
const invalidBuckets = Object.entries(cfg.temperature_by_options ?? {}).filter(([, value]) => typeof value !== 'number' || !Number.isFinite(value) || value < 0.5 || value > 5).map(([key]) => key);
if (!Array.isArray(cfg.temperature) || cfg.temperature.length !== 3 || cfg.temperature.some(value => typeof value !== 'number' || !Number.isFinite(value) || value < 0.5 || value > 5)) {
  throw new Error('Laya checkpoint base temperature is invalid');
}
const agent = await Agent.load(values['model-dir'], { device: 'cpu', expectedSha256: assets.files });
const assetDigest = createHash('sha256').update(JSON.stringify(Object.fromEntries(Object.entries(assets.files).sort(([left], [right]) => left.localeCompare(right))))).digest('hex');
const modelRevision = `${assets.model_revision}/${assets.checkpoint}/${assets.source_revision}/${assetDigest}`;
const connection = createMessageConnection(new StreamMessageReader(process.stdin), new StreamMessageWriter(process.stdout));
let busy = false;
connection.onRequest('judgment/health', () => ({
  model_revision: modelRevision, provider_model: 'laya-rl-agent', device: 'cpu',
  checkpoint: assets.checkpoint, language: values.language, sdk_reported_revision: agent.revision,
  invalid_calibration_buckets: invalidBuckets, pid: process.pid,
}));
connection.onRequest('judgment/evaluate', async input => {
  if (busy) throw new ResponseError(-32001, 'Laya worker already has an active inference');
  if (input.model_revision !== modelRevision || input.language !== values.language) throw new ResponseError(-32602, 'Laya model revision or explicitly configured language differs');
  const question = input.question;
  const type = question.question_type === 'truth' ? 'noul' : question.question_type;
  const count = type === 'choice' ? question.options.length : type === 'score' ? question.levels.length : 2;
  const bucket = `${type}:${count <= 2 ? '2' : count <= 5 ? '3-5' : count <= 10 ? '6-10' : '11+'}`;
  if (invalidBuckets.includes(bucket)) throw new ResponseError(-32602, `Laya checkpoint has invalid calibration for ${bucket}`);
  if (Buffer.byteLength(JSON.stringify(input), 'utf8') > input.budget.max_input_bytes) throw new ResponseError(-32602, 'Laya complete question, criteria and state exceed explicit byte budget');
  const definition = { type, instructions: question.instructions };
  if (type === 'choice') definition.criteria = Object.fromEntries(question.options.map(option => [option.id, option.description]));
  if (type === 'score') definition.criteria = question.levels;
  busy = true;
  try {
    const raw = await agent.predict(input.state, { [question.question_id]: definition }, { lang: values.language });
    if (raw.model !== 'laya-rl-agent') throw new Error('Unexpected actual Laya SDK model');
    return { model_revision: modelRevision, raw };
  } finally {
    busy = false;
  }
});
connection.onError(error => { throw error[0]; });
connection.onClose(() => { connection.dispose(); process.exit(0); });
connection.listen();
