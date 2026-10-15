import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { statSync } from 'node:fs';
import { mkdir, readdir, stat, writeFile } from 'node:fs/promises';
import { isAbsolute, join, relative } from 'node:path';
import test from 'node:test';
import { setTimeout as delay } from 'node:timers/promises';
import { pathToFileURL } from 'node:url';

function required(name) {
  const value = process.env[name];
  assert.ok(value, `${name} is required`);
  return value;
}

const sourceRoot = required('ASTROCYTE_S4_SOURCE_ROOT');
const outputRoot = required('ASTROCYTE_TEST_OUTPUT_ROOT');
const roots = [required('ASTROCYTE_S4_ROOT_A'), required('ASTROCYTE_S4_ROOT_B')];
for (const value of [sourceRoot, outputRoot, ...roots]) assert.ok(isAbsolute(value), 'Explicit absolute owned paths are required');
const expectedHead = '4317d610cbe80aa6c4e22ea4da5e8dde013de4ed';
const origin = 'https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015.git';
const gitEnvironment = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.toUpperCase().startsWith('GIT_')));
for (const root of roots) {
  const git = (...args) => execFileSync('git', ['--no-optional-locks', '-c', 'core.fsmonitor=false', '-C', root, ...args], { env: gitEnvironment, encoding: 'utf8' }).trim();
  assert.equal(git('remote', 'get-url', 'origin'), origin);
  assert.equal(git('rev-parse', 'HEAD'), expectedHead);
}
const { startS1Server } = await import(pathToFileURL(join(sourceRoot, 'tests/s1/server.mjs')).href);
const budget = { max_files: 100, max_bytes: 8388608, max_results: 100, max_depth: 3 };
const query = { schema_version: 1, purpose: 'verify_production_repository_context', roots: ['.'], budget };

async function humanAPI(base) {
  const bootstrap = await fetch(`${base}/api/v1/auth/session`, { signal: AbortSignal.timeout(5000) });
  assert.equal(bootstrap.status, 200);
  const session = await bootstrap.json();
  assert.equal(session.actor_kind, 'human');
  const cookie = bootstrap.headers.getSetCookie().map(value => value.split(';')[0]).join('; ');
  assert.ok(cookie);
  const request = async (path, { method = 'GET', body } = {}) => {
    const response = await fetch(`${base}/api/v1${path}`, {
      method, signal: AbortSignal.timeout(10000),
      headers: { Cookie: cookie, ...(method === 'GET' ? {} : { Origin: base, 'Content-Type': 'application/json', 'X-CSRF-Token': session.csrf_token, 'Idempotency-Key': randomUUID() }) },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
    return { status: response.status, data: await response.json() };
  };
  const success = async (path, options, status = 200) => {
    const response = await request(path, options);
    assert.equal(response.status, status, `Actual HTTP ${path}: ${JSON.stringify(response.data)}`);
    return response.data;
  };
  return {
    request,
    get: path => success(path),
    post: (path, body) => success(path, { method: 'POST', body }),
    command: (path, body, method = 'POST', status = 200) => success(path, { method, body: { schema_version: 1, request_id: randomUUID(), expected_version: 1, ...body } }, status),
  };
}

function safeResult(result) {
  assert.equal(Object.hasOwn(result.index.snapshot, 'root'), false, 'Private root entered the repository projection');
  for (const item of [...result.symbols.map(symbol => symbol.source_span), ...result.relations.map(relation => relation.source_span)]) {
    assert.deepEqual(Object.keys(item).sort(), ['end_line', 'path', 'start_line']);
    assert.ok(item.path && !isAbsolute(item.path) && !item.path.includes(':') && !item.path.startsWith('../'));
    assert.ok(item.start_line > 0 && item.end_line >= item.start_line);
  }
}

async function filesState(root) {
  const entries = [];
  async function visit(directory) {
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      assert.equal(entry.isSymbolicLink(), false, 'Owned projection contains a link');
      const path = join(directory, entry.name);
      if (entry.isDirectory()) await visit(path);
      else {
        const info = await stat(path);
        entries.push([relative(root, path), info.size, info.mtimeMs]);
      }
    }
  }
  await visit(root);
  return entries.sort((a, b) => a[0].localeCompare(b[0]));
}

test('actual production main, new SQLite, repository tools, restart and revocation', { timeout: 180000 }, async () => {
  const application = await startS1Server({
    outputRoot,
    env: async ({ temporary }) => {
      const profile = join(temporary, 'profile');
      const appdata = join(temporary, 'appdata');
      const localappdata = join(temporary, 'localappdata');
      await Promise.all([profile, appdata, localappdata].map(directory => mkdir(directory, { recursive: true })));
      return {
        HOME: profile, USERPROFILE: profile, APPDATA: appdata, LOCALAPPDATA: localappdata,
        ASTROCYTE_PROJECT_DISCOVERY_ROOTS: JSON.stringify(roots),
        ASTROCYTE_REPOSITORY_ENVIRONMENT_ID: 's4-production-public-windows',
        ASTROCYTE_CODEGRAPH_NODE: required('ASTROCYTE_CODEGRAPH_NODE'),
        ASTROCYTE_CODEGRAPH_SDK: required('ASTROCYTE_CODEGRAPH_SDK'),
        ASTROCYTE_CODEGRAPH_SCRIPT: join(sourceRoot, 'internal/adapters/repositorycontext/codegraph.mjs'),
        ASTROCYTE_AOCI_EXECUTABLE: required('ASTROCYTE_AOCI_EXECUTABLE'),
        ASTROCYTE_REPOSITORY_MAX_FILES: String(budget.max_files),
        ASTROCYTE_REPOSITORY_MAX_BYTES: String(budget.max_bytes),
        ASTROCYTE_REPOSITORY_MAX_RESULTS: String(budget.max_results),
        ASTROCYTE_REPOSITORY_MAX_DEPTH: String(budget.max_depth),
        ASTROCYTE_REPOSITORY_CLEANUP_MAX_GENERATIONS: '1',
      };
    },
  });
  let failure;
  try {
    assert.equal(application.query('SELECT count(*) AS n FROM repository_context_indexes')[0].n, 0);
    let api = await humanAPI(application.apiURL);
    const space = await api.command('/project-spaces', { title: 'Actual public S4 production verification' }, 'POST', 201);
    const projects = [];
    for (const [index, root] of roots.entries()) {
      const registered = await api.command('/local-projects', { name: `Public Unicode worktree ${index + 1}`, root, space_id: space.space.id });
      const project = registered.project;
      projects.push((await api.command(`/local-projects/${project.id}/settings`, {
        expected_revision: project.settings.revision,
        settings: { ...project.settings, allow_directory: true, allowed_subdirs: ['.'], allowed_actions: ['read_context'] },
      }, 'PUT')).project);
    }
    const prefix = project => `/local-projects/${project.id}/repository-context`;
    const graphs = [];
    for (const project of projects) {
      const indexed = (await api.post(`${prefix(project)}/index`, { ...query, action: 'index' })).repository_context;
      assert.equal(indexed.backend, 'codegraph');
      assert.equal(indexed.version, '1.0.1');
      assert.equal(indexed.snapshot.head, expectedHead);
      const found = (await api.post(`${prefix(project)}/search`, { ...query, query: 'Count' })).repository_context;
      await writeFile(join(outputRoot, `actual-search-${project.id}.json`), JSON.stringify(found), { flag: 'wx', mode: 0o600 });
      safeResult(found);
      assert.ok(found.gaps.includes('semantic_embedding_not_available'));
      const count = found.symbols.find(symbol => symbol.name === 'Count' && symbol.source_span.path === 'textstats/stats.go');
      assert.ok(count, 'Actual Count implementation was absent');
      const impact = (await api.post(`${prefix(project)}/impact`, { ...query, symbol_id: count.id })).repository_context;
      safeResult(impact);
      assert.ok(impact.relations.length > 0);
      assert.ok(impact.relations.every(relation => relation.evidence_kind === 'codegraph_static_heuristic'));
      graphs.push(indexed);
    }
    assert.equal(graphs[0].snapshot.repo_id, graphs[1].snapshot.repo_id);
    assert.notEqual(graphs[0].snapshot.worktree_id, graphs[1].snapshot.worktree_id);
    assert.notEqual(graphs[0].generation, graphs[1].generation);
    const changed = (await api.post(`${prefix(projects[0])}/search`, { ...query, query: 'countParagraphs' })).repository_context;
    assert.ok(changed.symbols.some(symbol => symbol.name === 'countParagraphs' && symbol.source_span.path === 'textstats/stats.go'));
    const unchanged = (await api.post(`${prefix(projects[1])}/search`, { ...query, query: 'countParagraphs' })).repository_context;
    assert.equal(unchanged.symbols.some(symbol => symbol.name === 'countParagraphs'), false);
    console.log('PASS production_new_database_official_indexes_and_worktree_relations');

    await api.post(`${prefix(projects[0])}/map/refresh`, query);
    const mapQuery = new URLSearchParams({ purpose: query.purpose, root: '.', ...Object.fromEntries(Object.entries(budget).map(([key, value]) => [key, String(value)])) });
    const mapPath = `${prefix(projects[0])}/map?${mapQuery}`;
    const map = (await api.get(mapPath)).repository_context;
    assert.equal(map.map_entries.length, 6);
    safeResult(map);
    const missing = await api.request(`${prefix(projects[1])}/map/refresh`, { method: 'POST', body: query });
    assert.equal(missing.status, 400);
    assert.equal(missing.data.error.code, 'evidence_missing');
    console.log('PASS production_actual_authored_map_and_missing_worktree');

    const cacheRoot = join(application.dataDir, 'repository-context');
    const before = await filesState(cacheRoot);
    const governanceBefore = await filesState(join(roots[0], '.aoci'));
    await api.post(`${prefix(projects[0])}/search`, { ...query, query: 'Count' });
    await api.get(mapPath);
    assert.deepEqual(await filesState(cacheRoot), before);
    assert.deepEqual(await filesState(join(roots[0], '.aoci')), governanceBefore);
    await application.restart();
    api = await humanAPI(application.apiURL);
    const reopened = (await api.post(`${prefix(projects[0])}/search`, { ...query, query: 'Count' })).repository_context;
    assert.equal(reopened.index.generation, graphs[0].generation);
    assert.equal((await api.get(mapPath)).repository_context.map_entries.length, 6);
    console.log('PASS production_readonly_queries_and_actual_process_restart');

    const exhausted = await api.request(`${prefix(projects[0])}/search`, { method: 'POST', body: { ...query, query: 'Count', budget: { ...budget, max_files: 1 } } });
    assert.equal(exhausted.status, 409);
    assert.equal(exhausted.data.error.code, 'budget_exhausted');
    const agent = randomUUID();
    await api.command(`/local-projects/${projects[0].id}/grants`, { agent_id: agent, actions: ['read_context'] });
    const issued = await api.command(`/local-projects/${projects[0].id}/agent-token`, { agent_id: agent });
    const agentSearch = async () => await fetch(`${application.apiURL}/api/v1${prefix(projects[0])}/search`, {
      method: 'POST', headers: { Authorization: `Bearer ${issued.credential.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...query, query: 'Count' }), signal: AbortSignal.timeout(10000),
    });
    let agentResponse = await agentSearch();
    assert.equal(agentResponse.status, 200);
    await agentResponse.arrayBuffer();
    await api.command(`/local-projects/${projects[0].id}/grants/revoke`, { agent_id: agent });
    agentResponse = await agentSearch();
    assert.equal(agentResponse.status, 403);
    await agentResponse.arrayBuffer();
    await api.post(`${prefix(projects[0])}/search`, { ...query, query: 'Count' });
    const project = projects[0];
    await api.command(`/local-projects/${project.id}/settings`, {
      expected_revision: project.settings.revision,
      settings: { ...project.settings, allow_directory: false, allowed_subdirs: [] },
    }, 'PUT');
    const forbidden = await api.request(`${prefix(project)}/search`, { method: 'POST', body: { ...query, query: 'Count' } });
    assert.equal(forbidden.status, 403);
    console.log('PASS production_budget_current_grant_revocation_and_directory_read_denial');
    const retiredDirectory = join(cacheRoot, graphs[0].snapshot.worktree_id, graphs[0].generation);
    const deadline = Date.now() + 10000;
    while (application.query('SELECT count(*) AS n FROM repository_context_indexes WHERE project_id=?', project.id)[0].n !== 0 || statSync(retiredDirectory, { throwIfNoEntry: false }) !== undefined) {
      assert.ok(Date.now() < deadline, 'Actual production cleanup did not retire the revoked projection');
      await delay(50);
    }
    await api.post(`${prefix(projects[1])}/search`, { ...query, query: 'Count' });
    await application.restart();
    api = await humanAPI(application.apiURL);
    assert.equal(application.query('SELECT count(*) AS n FROM repository_context_indexes WHERE project_id=?', project.id)[0].n, 0);
    assert.equal(statSync(retiredDirectory, { throwIfNoEntry: false }), undefined);
    const stillForbidden = await api.request(`${prefix(project)}/search`, { method: 'POST', body: { ...query, query: 'Count' } });
    assert.equal(stillForbidden.status, 403);
    const retained = (await api.post(`${prefix(projects[1])}/search`, { ...query, query: 'Count' })).repository_context;
    safeResult(retained);
    assert.equal(retained.index.generation, graphs[1].generation);
    assert.equal(application.query('SELECT count(*) AS n FROM local_agent_sessions')[0].n, 0);
    assert.equal(application.query('SELECT count(*) AS n FROM swarm_missions')[0].n, 0);
    console.log('PASS production_budget_grant_revocation_directory_cleanup_and_other_worktree');
  } catch (error) {
    failure = error;
    throw error;
  } finally {
    await application.close({ primaryError: failure });
  }
});
