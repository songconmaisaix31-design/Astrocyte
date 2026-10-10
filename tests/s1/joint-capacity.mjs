import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { copyFile, lstat, mkdir, readFile, readdir, realpath, writeFile } from 'node:fs/promises';
import { join, resolve, sep } from 'node:path';
import { backup, DatabaseSync } from 'node:sqlite';
import { command, humanAPI, waitJob } from './api.mjs';
import { startS1Server } from './server.mjs';

// One explicitly authorized NEW joint operation. This runner has no default
// test entry, media import, fixture source, automatic model retry or failure
// injection. Do not rerun it after a delivered or unknown model effect.
assert.equal(process.env.ASTROCYTE_RUN_JOINT_CAPACITY, '1', 'Requires coordinator-authorized single joint model slot');
const originalRoot = await realpath(join(process.env.LOCALAPPDATA, 'Temp', 'astrocyte-s1-4pmMWO', 'data'));
const original = new DatabaseSync(join(originalRoot, 'state.sqlite'), { readOnly: true });
const unknownID = '4GQ7GKDXHZIND7CAB7GIZU3W4V';
const sourceDefinitions = [
  { id: 'IJNSLKOAR6LH6XT2KDZXJE6CCC', revision: 1, bytes: 95314, digest: '05acdc763e444fa7b9762df6dc2c879867d6a2e38d42884dcf3b80e07cf89d64' },
  { id: '3766OCH7SX3O5AVIDRPWYPCL3G', revision: 2, bytes: 32908, digest: 'eb1106b97e3db4ddb6465ae066d1efe29affb809aa4841d52014ede3ab45a3fc' },
];
const originalJobs = original.prepare('SELECT id,data,payload FROM attention_jobs ORDER BY id').all();
const originalUnknown = originalJobs.find(job => job.id === unknownID);
assert.ok(originalUnknown && JSON.parse(originalUnknown.data).delivery_unknown);
const sources = [];
for (const expected of sourceDefinitions) {
  const row = original.prepare('SELECT data FROM attention_material_revisions WHERE material_id=? AND revision=?').get(expected.id, expected.revision);
  assert.ok(row);
  const revision = JSON.parse(row.data);
  assert.equal(revision.content_digest, expected.digest);
  assert.match(revision.object_ref, /^[a-f0-9]{64}$/);
  const path = join(originalRoot, 'objects', revision.object_ref);
  assert.ok((await lstat(path)).isFile() && !(await lstat(path)).isSymbolicLink());
  const body = await readFile(path);
  assert.equal(body.byteLength, expected.bytes);
  assert.equal(createHash('sha256').update(body).digest('hex'), expected.digest);
  sources.push({ expected, revision, text: body.toString('utf8') });
}
const projectID = '4f8ca4e3-2353-4e2a-9c48-ce70fcbf9cce';
const server = await startS1Server({ env: async ({ dataDir, temporary }) => {
  assert.ok(resolve(dataDir).startsWith(resolve(temporary) + sep));
  await mkdir(join(dataDir, 'objects'), { recursive: true });
  // SQLite online backup produces a consistent copy without modifying source.
  await backup(original, join(dataDir, 'state.sqlite'));
  for (const name of await readdir(join(originalRoot, 'objects'))) {
    assert.match(name, /^[a-f0-9]{64}$/, 'Only ordinary digest objects belong to this approved store');
    const path = join(originalRoot, 'objects', name);
    const info = await lstat(path);
    assert.ok(info.isFile() && !info.isSymbolicLink(), 'No redirected/private filesystem objects are followed');
    await copyFile(path, join(dataDir, 'objects', name));
  }
  return { ASTROCYTE_ENABLE_SUMMARIZE: 'false', ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false', ASTROCYTE_JOB_TIMEOUT_SECONDS: '1800' };
} });
let primaryError;
let receipt;
try {
  let api = await humanAPI(server.apiURL);
  for (const source of sources) {
    const content = await api.get(`/materials/${source.expected.id}/revisions/${source.expected.revision}/content`);
    assert.equal(content.text, source.text);
  }
  const projects = await api.get('/local-projects');
  const project = projects.items.find(item => item.id === projectID);
  assert.ok(project && project.settings.external_model_cli === 'codex');
  assert.equal(project.settings.allow_directory, false);
  const refs = sources.map(source => ({ material_id: source.expected.id, revision: source.expected.revision, locator: source.revision.source_locator }));
  const space = await api.get(`/project-spaces/${project.space_id}`);
  for (const ref of refs) assert.ok(space.space.material_refs.some(link => link.material_id === ref.material_id && link.revision === ref.revision), 'Fixed sources must already be human-approved in the selected project space');
  const missions = await api.get('/missions');
  // Fresh registry configuration requires an explicit no-model native probe.
  // This verifies the already approved CLI; it does not grant new permissions.
  const probe = await api.write(`/local-projects/${projectID}/sessions/probe`, { cli: 'codex' });
  assert.ok(probe.session.stop_confirmed);
  const request = command({ input_refs: refs, prior_distillation_ids: [], stage: 'topic', question: 'Use both complete selected source versions; retain uncertainty.', processing_config: 's1-full-joint-capacity-20261010', project_id: projectID, cli: 'codex' });
  console.log(JSON.stringify({ phase: 'before_single_joint_submission', directory: server.temporary, sources: sourceDefinitions, request }));
  receipt = await api.write('/distillations/jobs', request, { status: 202, key: 's1-full-joint-capacity-20261010' });
  await writeFile(join(server.temporary, 'joint-request.json'), JSON.stringify({ request, receipt }, null, 2));
  // Polling observes this original job. There is no resubmission or paid retry.
  let job;
  try {
    job = await waitJob(api, receipt.job_id, 'succeeded', 1_800_000);
  } catch (firstError) {
    const failedRow = server.query('SELECT data,payload FROM attention_jobs WHERE id=?', receipt.job_id)[0];
    const failedJob = JSON.parse(failedRow.data);
    const saved = JSON.parse(failedRow.payload);
    if (failedJob.status !== 'failed' || failedJob.delivery_unknown || !saved.Result) throw firstError;
    await writeFile(join(server.temporary, 'joint-local-publication-first.json'), JSON.stringify({ first_error: String(firstError), job: failedJob, saved_result: saved.Result }, null, 2));
    console.error('First local publication failure preserved; recovering the saved delivered result without another model call');
    await api.write(`/jobs/${receipt.job_id}/retry`, {});
    job = await waitJob(api, receipt.job_id, 'succeeded', 30_000);
  }
  const row = server.query('SELECT data,payload FROM attention_jobs WHERE id=?', receipt.job_id)[0];
  const payload = JSON.parse(row.payload);
  assert.ok(payload.Result);
  assert.ok(job.attempts === 1 || (job.attempts === 2 && payload.Result), 'Only saved-result publication recovery may add a local attempt');
  assert.notEqual(job.delivery_unknown, true);
  const distillations = await api.get('/distillations');
  const result = distillations.items.find(item => item.input_refs.length === 2 && item.question === request.question && item.stage === 'topic');
  assert.ok(result && result.provenance.processor === 'codex' && result.output_text.length > 0);
  assert.deepEqual(result.input_refs.map(ref => [ref.material_id, ref.revision]).sort(), refs.map(ref => [ref.material_id, ref.revision]).sort());
  assert.equal(await readFile(join(server.dataDir, 'objects', result.output_ref), 'utf8'), result.output_text);
  assert.deepEqual(await api.get('/missions'), missions);
  assert.deepEqual(server.query('SELECT id,data,payload FROM attention_jobs WHERE id=?', unknownID)[0], originalUnknown);
  await server.restart();
  api = await humanAPI(server.apiURL);
  job = await api.get(`/jobs/${receipt.job_id}`);
  assert.equal(job.status, 'succeeded');
  assert.deepEqual((await api.get('/distillations')).items.find(item => item.id === result.id), result);
  for (const source of sources) assert.equal((await api.get(`/materials/${source.expected.id}/revisions/${source.expected.revision}/content`)).text, source.text);
  assert.deepEqual(await api.get('/missions'), missions);
  assert.deepEqual(original.prepare('SELECT id,data,payload FROM attention_jobs ORDER BY id').all(), originalJobs);
  const report = { scope: 'task_live_full_selected_joint_input', directory: server.temporary, sources: sourceDefinitions, job, result, preserved_unknown: unknownID, native_usage_cost: null };
  await writeFile(join(server.temporary, 'joint-result.json'), JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report));
} catch (error) {
  primaryError = error;
  if (receipt) console.error(JSON.stringify({ phase: 'joint_first_failure_preserved', directory: server.temporary, receipt, job: server.query('SELECT data,payload FROM attention_jobs WHERE id=?', receipt.job_id) }));
  else console.error(`Joint preparation first failure; no model submission: ${server.temporary}`);
  throw error;
} finally {
  original.close();
  await server.close({ preserveData: true, primaryError });
}
