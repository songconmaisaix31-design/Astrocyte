import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { createWriteStream } from 'node:fs';
import { lstat, mkdir, mkdtemp, readFile } from 'node:fs/promises';
import { isAbsolute, join } from 'node:path';
import { createInterface } from 'node:readline';
import { DatabaseSync } from 'node:sqlite';
import { promisify } from 'node:util';
import test from 'node:test';
import Ajv2020 from 'ajv/dist/2020.js';
import addFormats from 'ajv-formats';
import { parse } from 'yaml';
import { waitJob } from '../s1/api.mjs';
import { root, start, stop } from '../../scripts/process.mjs';

const git = promisify(execFile);

function contractValidator(spec) {
  const ajv = new Ajv2020({ allErrors: true, strict: false });
  addFormats(ajv);
  const definitions = JSON.parse(JSON.stringify(spec.components.schemas).replaceAll('#/components/schemas/', '#/$defs/'));
  const validators = new Map();
  return (method, path, status, data) => {
    const pathname = new URL(path, 'http://localhost').pathname;
    const route = Object.keys(spec.paths).find(candidate => new RegExp(`^${candidate.replace(/\{[^}]+\}/g, '[^/]+')}$`).test(pathname));
    let response = spec.paths[route]?.[method.toLowerCase()]?.responses?.[String(status)];
    if (response?.$ref) response = response.$ref.slice(2).split('/').reduce((value, key) => value[key], spec);
    const schema = response?.content?.['application/json']?.schema;
    assert.ok(schema, `Exact source must document ${method} ${path} ${status}`);
    const key = `${method} ${route} ${status}`;
    if (!validators.has(key)) {
      const normalized = JSON.parse(JSON.stringify(schema).replaceAll('#/components/schemas/', '#/$defs/'));
      validators.set(key, ajv.compile({ ...normalized, $defs: definitions }));
    }
    const validate = validators.get(key);
    assert.ok(validate(data), `${method} ${path} ${status} violates exact OpenAPI: ${ajv.errorsText(validate.errors)}`);
  };
}

async function listenerURL(child) {
  return await new Promise((resolve, reject) => {
    const lines = createInterface({ input: child.stdout, crlfDelay: Infinity });
    const timer = setTimeout(() => finish(new Error('Actual server listener announcement timed out')), 30_000);
    const onExit = () => finish(new Error('Actual server exited before publishing its listener'));
    const finish = (error, url) => {
      clearTimeout(timer);
      child.removeListener('error', onError);
      child.removeListener('exit', onExit);
      lines.removeAllListeners();
      lines.close();
      child.stdout.resume();
      if (error) reject(error); else resolve(url);
    };
    const onError = error => finish(error);
    child.once('error', onError);
    child.once('exit', onExit);
    lines.once('line', line => {
      try {
        const record = JSON.parse(line);
        const url = new URL(record.api_url);
        assert.equal(record.type, 'http_listening');
        assert.equal(Object.keys(record).length, 2);
        assert.equal(url.protocol, 'http:');
        assert.equal(url.hostname, '127.0.0.1');
        assert.equal(record.api_url, url.origin);
        assert.ok(Number(url.port) > 0);
        finish(undefined, url.origin);
      } catch (error) { finish(error); }
    });
  });
}

async function authenticatedAPI(baseURL, validate) {
  const bootstrap = await fetch(`${baseURL}/api/v1/auth/session`, { signal: AbortSignal.timeout(5000) });
  const session = await bootstrap.json();
  assert.equal(bootstrap.status, 200);
  validate('GET', '/auth/session', bootstrap.status, session);
  assert.equal(session.actor_kind, 'human');
  const cookie = bootstrap.headers.getSetCookie().map(value => value.split(';')[0]).join('; ');
  assert.ok(cookie);
  return {
    async request(path, { method = 'GET', body, token, key = randomUUID(), status = 200 } = {}) {
      const response = await fetch(`${baseURL}/api/v1${path}`, {
        method, signal: AbortSignal.timeout(10_000),
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : { Cookie: cookie }),
          ...(method === 'GET' ? {} : { Origin: baseURL, 'Content-Type': 'application/json', 'X-CSRF-Token': session.csrf_token, 'Idempotency-Key': key }),
        },
        ...(body === undefined ? {} : { body: JSON.stringify({ schema_version: 1, request_id: key, expected_version: 1, ...body }) }),
      });
      const result = await response.json();
      validate(method, path, response.status, result);
      assert.equal(response.status, status, `${method} ${path} unexpected response status`);
      return result;
    },
    async get(path) { return await this.request(path); },
    async write(path, body, options = {}) { return await this.request(path, { method: 'POST', body, ...options }); },
  };
}

test('actual imported public source enters a fixed ContextPacket with current source permission and process restart', { timeout: 120_000 }, async () => {
  const executable = process.env.S2_CONTEXT_BINARY;
  const source = process.env.S2_CONTEXT_SOURCE;
  const outputRoot = process.env.S2_CONTEXT_OUTPUT_ROOT;
  assert.ok(executable && isAbsolute(executable), 'Explicit real production binary is required');
  assert.match(source ?? '', /^[a-f0-9]{40}$/, 'Exact production SOURCE commit is required');
  assert.ok(outputRoot && isAbsolute(outputRoot), 'Explicit owner-controlled output directory is required');
  assert.ok((await lstat(executable)).isFile());
  const spec = parse((await git('git', ['show', `${source}:contracts/openapi.yaml`], { cwd: root, maxBuffer: 4 * 1024 * 1024 })).stdout);
  const validate = contractValidator(spec);
  const head = (await git('git', ['rev-parse', 'HEAD'], { cwd: root })).stdout.trim();
  const sourceText = (await git('git', ['show', `${head}:README.md`], { cwd: root, maxBuffer: 2 * 1024 * 1024 })).stdout;
  assert.ok(sourceText.length > 0);
  await mkdir(outputRoot, { recursive: true });
  const directory = await mkdtemp(join(outputRoot, 'context-material-'));
  const dataDir = join(directory, 'data');
  const toolTemporary = join(directory, 'tool-temp');
  await mkdir(toolTemporary);
  const stderr = createWriteStream(join(directory, 'server.log'), { flags: 'a', mode: 0o600 });
  const environmentID = randomUUID();
  const env = {
    ...Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.toUpperCase().startsWith('ASTROCYTE_'))),
    ASTROCYTE_DATA_DIR: dataDir, ASTROCYTE_PORT: '0', ASTROCYTE_ENABLE_CODEX_DISTILLATION: 'false',
    ASTROCYTE_ENABLE_SUMMARIZE: 'false', ASTROCYTE_SUMMARIZE_CLI: '', ASTROCYTE_RELAY: '0',
    ASTROCYTE_PROJECT_DISCOVERY_ROOTS: JSON.stringify([root]), ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: environmentID,
    ASTROCYTE_REPOSITORY_MAX_FILES: '2048', ASTROCYTE_REPOSITORY_MAX_BYTES: '67108864',
    ASTROCYTE_REPOSITORY_MAX_RESULTS: '100', ASTROCYTE_REPOSITORY_MAX_DEPTH: '5',
    TEMP: toolTemporary, TMP: toolTemporary,
  };
  let child;
  let apiURL;
  let phase = 'startup';
  let failure;
  const launch = async () => {
    child = start(executable, [], { env, stdio: ['ignore', 'pipe', 'pipe'] });
    child.stderr.pipe(stderr, { end: false });
    apiURL = await listenerURL(child);
    env.ASTROCYTE_PORT = new URL(apiURL).port;
    const response = await fetch(`${apiURL}/api/v1/health`, { signal: AbortSignal.timeout(10_000) });
    assert.equal(response.status, 200);
  };
  const query = (sql, ...parameters) => {
    const db = new DatabaseSync(join(dataDir, 'state.sqlite'), { readOnly: true });
    try { return db.prepare(sql).all(...parameters); } finally { db.close(); }
  };
  try {
    await launch();
    let api = await authenticatedAPI(apiURL, validate);
    phase = 'public source import and explicit project authorization';
    const importReceipt = await api.write('/materials/imports', {
      kind: 'text', adapter: 'manual', source_key: '', content_digest: '',
      source_locator: `https://github.com/songconmaisaix31-design/Astrocyte/blob/${head}/README.md`,
      export_text: sourceText, title: 'Public fixed Git README for ContextPacket', source_spans: [`Git ${head}:README.md`], collection_reason: null,
    }, { status: 202 });
    const job = await waitJob(api, importReceipt.job_id);
    const material = (await api.get(`/materials/${job.material_id}`)).material;
    const reference = { material_id: material.id, revision: material.current_revision };
    const space = (await api.write('/project-spaces', { title: 'Public fixed ContextPacket source' }, { status: 201 })).space;
    let project = (await api.write('/local-projects', { name: 'Public fixed ContextPacket source', root, space_id: space.id })).project;
    project = (await api.write(`/local-projects/${project.id}/settings`, {
      expected_revision: project.settings.revision,
      settings: { ...project.settings, allow_directory: true, allowed_subdirs: ['internal/workspace/domain'], allowed_actions: ['read_context'] },
    }, { method: 'PUT' })).project;
    const selected = (await api.write(`/project-spaces/${space.id}/references`, { expected_version: space.version, ...reference })).space;
    const goal = (await api.write('/goals', { expected_version: 0, goal: { project_id: project.id, title: 'Review a fixed public source in an immutable context', status: 'active', success_criteria: ['Current source permission gates new handoff context'], priority: 1 } })).goal;
    const deadline = new Date(Date.now() + 180_000).toISOString();
    const content = {
      title: 'Review the actual selected public README', goal: { id: goal.id, revision: goal.revision }, inputs: [reference],
      scope: { project_id: project.id, settings_revision: project.settings.revision, roots: [join(project.root, 'internal', 'workspace', 'domain')], actions: ['read_context'], tools: [], external_targets: [], write_paths: [], max_concurrency: 1 },
      budget: { calls: 1, tokens: null, money_minor: null, currency: '' },
      stops: { deadline, conditions: ['Stop after context permission verification'], max_work_items: 1, max_attempts: 1, no_progress_seconds: 0, max_no_progress_rounds: 1 },
      deliverables: ['Fixed source reference'], success_criteria: ['Existing packet and object remain immutable'],
    };
    let proposal = (await api.write('/proposals', { expected_version: 0, content })).proposal;
    proposal = (await api.write(`/proposals/${proposal.id}/submit`, { expected_version: proposal.version })).proposal;
    const grant = (await api.write(`/proposals/${proposal.id}/approvals`, { expected_version: proposal.version, expected_revision: proposal.revision, expires_at: deadline })).approval;
    const build = { expected_version: 0, grant_id: grant.id, grant_epoch: grant.epoch, max_bytes: 128 * 1024, checks: ['Actual fixed public Git object'], next_steps: ['Review approved source'] };
    phase = 'actual fixed-source context creation';
    const packet = (await api.write('/context-packets', build)).context_packet;
    assert.equal(packet.snapshot.environment_id, environmentID);
    assert.equal(packet.snapshot.head, head);
    assert.equal(Object.hasOwn(packet.snapshot, 'root'), false);
    assert.deepEqual(packet.scope.roots, ['internal/workspace/domain']);
    assert.deepEqual(packet.sources.map(value => value.reference), [reference]);
    assert.match(packet.sources[0].object_ref, /^[a-f0-9]{64}$/);
    assert.ok(await readFile(join(dataDir, 'objects', packet.sources[0].object_ref), 'utf8') === sourceText, 'Stored source must equal the actual fixed public Git object');
    const packetBytes = await readFile(join(dataDir, 'objects', packet.object_ref));
    const saved = JSON.parse(query('SELECT data FROM workspace_context_packets WHERE id=?', packet.id)[0].data);
    assert.deepEqual(saved.sources, packet.sources);
    assert.ok(saved.snapshot.root === project.root, 'Private snapshot must retain the actual registered root');
    assert.equal((await api.get(`/materials/${material.id}`)).material.human_usage_count, material.human_usage_count);
    const memberID = randomUUID();
    await api.write(`/local-projects/${project.id}/grants`, { agent_id: memberID, actions: ['read_context'] });
    const memberToken = (await api.write(`/local-projects/${project.id}/agent-token`, { agent_id: memberID })).credential.token;
    assert.deepEqual((await api.request(`/context-packets/${packet.id}`, { token: memberToken })).context_packet, packet);
    phase = 'first process restart and immutable object read';
    await stop(child);
    await launch();
    api = await authenticatedAPI(apiURL, validate);
    assert.deepEqual((await api.get(`/context-packets/${packet.id}`)).context_packet, packet);
    assert.ok((await readFile(join(dataDir, 'objects', packet.object_ref))).equals(packetBytes), 'Restart must retain the exact immutable packet object');
    phase = 'current source permission removal and rejection';
    await api.write(`/project-spaces/${space.id}/references/remove`, { expected_version: selected.version, material_id: material.id });
    const denied = await api.write('/context-packets', build, { status: 403 });
    assert.equal(denied.error.code, 'scope_denied');
    assert.equal(query('SELECT count(*) AS n FROM workspace_context_packets')[0].n, 1);
    assert.equal(query('SELECT count(*) AS n FROM foundation_outbox WHERE event_type=?', 'workspace.context_packet_built')[0].n, 1);
    assert.deepEqual((await api.get(`/context-packets/${packet.id}`)).context_packet, packet);
    phase = 'second process restart and persistent source rejection';
    await stop(child);
    await launch();
    api = await authenticatedAPI(apiURL, validate);
    assert.equal((await api.write('/context-packets', build, { status: 403 })).error.code, 'scope_denied');
    assert.ok((await readFile(join(dataDir, 'objects', packet.object_ref))).equals(packetBytes), 'Source permission changes must preserve the original packet object');
    assert.equal(query('SELECT count(*) AS n FROM local_agent_sessions')[0].n, 0);
    assert.equal(query('SELECT count(*) AS n FROM swarm_missions')[0].n, 0);
    console.log(JSON.stringify({ productionSource: source, publicSourceHead: head, database: join(dataDir, 'state.sqlite'), result: 'actual_context_material_pass' }));
  } catch (error) {
    failure = new Error(`Actual Context verification failed during ${phase}; retained local evidence: ${directory}`, { cause: error });
    throw failure;
  } finally {
    const cleanup = [];
    if (child) cleanup.push(...await Promise.allSettled([stop(child)]));
    cleanup.push(...await Promise.allSettled([new Promise((resolve, reject) => { stderr.once('error', reject); stderr.end(resolve); })]));
    const failed = cleanup.filter(value => value.status === 'rejected');
    if (failed.length) throw new AggregateError([...(failure ? [failure] : []), ...failed.map(value => value.reason)], `Actual Context cleanup failed; retained ${directory}`);
  }
});
