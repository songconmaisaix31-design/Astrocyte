import assert from 'node:assert/strict';
import { readFile, stat, readdir } from 'node:fs/promises';
import { createReadStream } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve, join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { createRequire } from 'node:module';
import { DatabaseSync } from 'node:sqlite';

const root = resolve(process.argv[2]);
process.env.CODEGRAPH_TELEMETRY = '0';
process.env.CODEGRAPH_NO_DOWNLOAD = '1';
assert.equal(process.version, 'v24.16.0');
assert.equal(process.platform, 'win32');
assert.equal(process.arch, 'x64');
const require = createRequire(join(root, 'sdk', 'package.json'));
for (const [name, version] of Object.entries({
  'laya-ts': '0.1.0', 'onnxruntime-node': '1.30.0', 'onnxruntime-common': '1.30.0',
  'vscode-jsonrpc': '8.2.1', '@colbymchenry/codegraph': '1.0.1',
})) {
  const pkg = JSON.parse(await readFile(join(root, 'sdk/node_modules', name, 'package.json'), 'utf8'));
  assert.equal(pkg.version, version, `${name} actual version`);
  if (name === '@colbymchenry/codegraph') assert.equal(pkg.main, 'dist/index.js', 'actual S4 source-built SDK entry');
}
const { Agent } = await import(pathToFileURL(join(root, 'sdk/node_modules/laya-ts/dist/index.js')).href);
assert.equal(typeof Agent.load, 'function');
const ort = require('onnxruntime-node');
assert.equal(typeof ort.InferenceSession.create, 'function');
const { CodeGraph, QueryBuilder } = require('@colbymchenry/codegraph');
assert.equal(typeof CodeGraph, 'function');
assert.equal(typeof QueryBuilder, 'function');
const rpc = require('vscode-jsonrpc/node');
assert.equal(typeof rpc.StreamMessageReader, 'function');
const db = new DatabaseSync(':memory:');
try { assert.equal(typeof db.prepare('SELECT sqlite_version() AS version').get().version, 'string'); }
finally { db.close(); }
for (const checkpoint of ['english', 'multilingual']) {
  const dir = join(root, 'models', checkpoint);
  const assets = JSON.parse(await readFile(join(dir, 'assets.json'), 'utf8'));
  assert.equal(assets.repo, 'convaiinnovations/laya');
  assert.equal(assets.model_revision, '7b928d828b7b0e022f929d9bd2e44165aa270148');
  assert.equal(assets.source_revision, '432312b36714d8889eba405507b37eda25257d9c');
  assert.equal(assets.checkpoint, checkpoint);
  assert.equal(assets.device, 'cpu');
  assert.equal(assets.artifact_kind, 'astrocyte_privacy_sanitized_derivative');
  assert.equal(assets.artifact_revision, 'metadata-clean-20261015-ctx_0a7797464f02');
  assert.equal(assets.derived_from.model_revision, assets.model_revision);
  assert.equal(assets.derived_from.source_revision, assets.source_revision);
  const names = ['encoder.onnx', 'encoder.onnx.data', 'head.onnx', 'head.onnx.data', 'tokenizer.json', 'rl_agent_config.json'];
  assert.deepEqual(Object.keys(assets.files).sort(), [...names].sort());
  assert.deepEqual(Object.keys(assets.derived_from.files).sort(), [...names].sort());
  for (const name of names) {
    if (name.endsWith('.onnx')) assert.notEqual(assets.files[name], assets.derived_from.files[name]);
    else assert.equal(assets.files[name], assets.derived_from.files[name], `${checkpoint}/${name} unchanged derivative input`);
  }
  assert.deepEqual((await readdir(dir)).sort(), [...names, 'assets.json'].sort());
  for (const name of names) {
    assert.ok((await stat(join(dir, name))).size > 0);
    const hash = createHash('sha256');
    for await (const chunk of createReadStream(join(dir, name))) hash.update(chunk);
    assert.equal(hash.digest('hex'), assets.files[name], `${checkpoint}/${name} S5 asset identity`);
  }
  JSON.parse(await readFile(join(dir, 'tokenizer.json'), 'utf8'));
  JSON.parse(await readFile(join(dir, 'rl_agent_config.json'), 'utf8'));
}
console.log('PASS: Node, native ONNX binding, official built Laya, CodeGraph, RPC, SQLite, twelve fixed S5 derivative assets and two provenance descriptors; no inference executed.');
