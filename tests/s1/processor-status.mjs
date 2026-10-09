// Opt-in native version/status and default-deny checks only. Never submits authorized model work.
import assert from 'node:assert/strict';
import { startS1Server } from './server.mjs';
import { command, humanAPI, importMaterial } from './api.mjs';
import { paperImport, sourceRef } from './fixtures.mjs';

function option(name) {
  const index = process.argv.indexOf(name);
  if (index < 0 || !process.argv[index + 1]) throw new Error(`Pass ${name} explicitly; no native executable/model is guessed`);
  return process.argv[index + 1];
}
const configuredModel = option('--model');
const executable = option('--native-executable');
const server = await startS1Server({ env: {
  ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'true', ASTROCYTE_CODEX_EXECUTABLE: executable, ASTROCYTE_CODEX_MODEL: configuredModel,
  ASTROCYTE_PROCESSING_SOURCE_KEYS: JSON.stringify(['arxiv:2504.16054']),
} });
try {
  const api = await humanAPI(server.apiURL);
  const status = await api.get('/distillations/processor');
  assert.equal(status.available, true);
  assert.equal(status.processor, 'codex-cli');
  assert.equal(status.model, configuredModel, 'Status is the actual configured native model, not a completed inference receipt');
  assert.match(status.configuration_id, /codex\/0\.162\.0/);
  assert.match(status.configuration_id, /output_schema_sha256=[a-f0-9]{64}/);
  assert.deepEqual(status.allowed_source_keys, ['arxiv:2504.16054']);
  const imported = await importMaterial(api, paperImport('https://example.invalid/contract-local/native-ungranted'));
  const jobsBefore = (await api.get('/jobs')).items;
  const recordsBefore = (await api.get('/distillations')).items;
  const result = await api.request('/distillations/jobs', { method: 'POST', body: command({
    input_refs: [sourceRef(imported.detail.material)], stage: 'content', processing_config: status.configuration_id,
    question: 'contract_local ungranted source must be denied before model invocation',
  }) });
  assert.equal(result.status, 403, JSON.stringify(result.data));
  assert.equal(result.data.error.code, 'scope_denied');
  assert.deepEqual((await api.get('/jobs')).items, jobsBefore);
  assert.deepEqual((await api.get('/distillations')).items, recordsBefore);
  assert.equal(server.query('SELECT COUNT(*) AS n FROM attention_jobs WHERE json_extract(data,\'$.kind\')=?', 'distillation')[0].n, 0);
  assert.equal((await api.get('/missions')).items.length, 0);
  console.log(JSON.stringify({ native_configuration_status: 'PASS', ungranted_input_denial: 'PASS', configured_model: status.model, actual_inference_model: null, model_calls: 0, paid_automatic_acceptance: 'NOT_RUN' }, null, 2));
} finally { await server.close(); }
