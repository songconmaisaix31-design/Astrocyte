// 自然到期验收：签发真实 1h agent token，等待自然到期后验证 HTTP 拒绝。
// 复用 startS1Server / humanAPI；token 仅留在内存，不记录内容或片段。
// 显式运行：node tests/s1/progress-expiry.mjs
// 不在默认测试套件中执行，不需要 env 开关、skip 或 fake 到期。
import assert from 'node:assert/strict';
import { execSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { mkdir, mkdtemp, writeFile } from 'node:fs/promises';
import { join, resolve } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { startS1Server } from './server.mjs';
import { humanAPI, command } from './api.mjs';

const realRoot = resolve(process.cwd());
const now = () => new Date().toISOString();

// 通过已有人工 API 签发项目 agent token，返回 { token, expires_at, agentID }。
async function issueAgentToken(api, projectID, agentID) {
  const grantResult = await api.write(`/local-projects/${projectID}/grants`, {
    agent_id: agentID, actions: ['read_context'],
  });
  assert.ok(grantResult.grant, 'grant response must include grant payload');

  const tokenResult = await api.write(`/local-projects/${projectID}/agent-token`, {
    agent_id: agentID,
  });
  const credential = tokenResult.credential;
  assert.ok(credential && credential.token, 'token response must include credential with token');
  assert.ok(credential.expires_at, 'token must carry an explicit expires_at');
  return { token: credential.token, expires_at: credential.expires_at, agentID };
}

// 使用 Bearer token 发起 progress GET；返回 { status, body }。
async function agentProgressGet(baseURL, projectID, token) {
  const response = await fetch(`${baseURL}/api/v1/local-projects/${projectID}/progress`, {
    signal: AbortSignal.timeout(10_000),
    headers: { Authorization: `Bearer ${token}` },
  });
  const body = await response.json().catch(() => null);
  return { status: response.status, body };
}

// 确保项目根目录是 git 仓库；不是则初始化。
function ensureGitRepo(directory) {
  try {
    execSync('git rev-parse --git-dir', { cwd: directory, stdio: 'pipe' });
  } catch {
    execSync('git init', { cwd: directory, stdio: 'pipe' });
  }
}

const server = await startS1Server();
try {
  const api = await humanAPI(server.apiURL);

  // 创建空间和项目
  const space = await api.write('/project-spaces', { title: 'progress-expiry-space' }, { status: 201 });
  const project = await api.write('/local-projects', { name: 'progress-expiry', root: realRoot, space_id: space.space.id });
  const projectID = project.project.id;

  // 设置项目允许 read_context
  await api.write(`/local-projects/${projectID}/settings`, {
    expected_revision: 1,
    settings: {
      allow_directory: true,
      allowed_subdirs: ['README.md'],
      allowed_actions: ['read_context', 'observe'],
    },
  }, { method: 'PUT' });

  // 人工写入初始进度，确保 GET 有缓存可读
  await api.write(`/local-projects/${projectID}/progress`, {
    status: 'planning', summary: 'expiry verification',
  }, { method: 'PUT' });

  // 签发真实 1h agent token
  const agentID = `expiry-verify-${randomUUID().slice(0, 8)}`;
  const { token, expires_at } = await issueAgentToken(api, projectID, agentID);

  const expiresAtMs = new Date(expires_at).getTime();
  const issuedAt = now();
  const pid = process.pid;

  // 仅记录派生身份和 UTC 时间到文件；token 内容绝不写入磁盘或 stdout。
  const logLine = `PID=${pid} ISSUED=${issuedAt} EXPIRES=${expires_at} PROJECT=${projectID} AGENT=${agentID}\n`;
  const verificationDir = join(server.temporary, '.verification');
  await mkdir(verificationDir, { recursive: true });
  const logFile = join(verificationDir, 'expiry-token.log');
  await writeFile(logFile, logLine, { encoding: 'utf8', mode: 0o600 });

  console.log(`EXPIRY_VERIFY PID=${pid} ISSUED=${issuedAt} EXPIRES=${expires_at} AGENT=${agentID}`);

  // 到期前：Agent GET 必须成功
  const before = await agentProgressGet(server.apiURL, projectID, token);
  assert.equal(before.status, 200, `pre-expiry GET must succeed, got ${before.status}: ${JSON.stringify(before.body)}`);
  console.log(`PRE_EXPIRY_OK status=200 at=${now()}`);

  // 等待自然到期 + 10s 安全边距
  const waitMs = expiresAtMs - Date.now() + 10_000;
  assert.ok(waitMs > 0, `token must not already be expired at issuance; waitMs=${waitMs}`);
  console.log(`WAITING ${Math.round(waitMs / 1000)}s until natural expiry`);
  await delay(waitMs);

  // 到期后：HTTP 必须返回 403
  const afterHTTP = await agentProgressGet(server.apiURL, projectID, token);
  assert.equal(afterHTTP.status, 403, `post-expiry GET must be forbidden, got ${afterHTTP.status}: ${JSON.stringify(afterHTTP.body)}`);
  console.log(`POST_EXPIRY_HTTP_OK status=403 at=${now()}`);

  // 跨项目拒绝：用同一 token 访问另一项目
  const space2 = await api.write('/project-spaces', { title: 'progress-expiry-cross' }, { status: 201 });
  const project2 = await api.write('/local-projects', { name: 'progress-expiry-cross', root: realRoot, space_id: space2.space.id });
  const cross = await agentProgressGet(server.apiURL, project2.project.id, token);
  assert.equal(cross.status, 403, `cross-project GET must be forbidden, got ${cross.status}`);
  console.log(`CROSS_PROJECT_OK status=403`);

  // 写入结果
  const result = {
    pid, agent_id: agentID, project_id: projectID,
    issued_at: issuedAt, expires_at,
    http_rejected: true, cross_project_rejected: true,
    verified_at: now(),
  };
  const resultFile = join(verificationDir, 'expiry-result.json');
  await writeFile(resultFile, JSON.stringify(result, null, 2), { encoding: 'utf8', mode: 0o600 });
  console.log(`EXPIRY_VERIFIED at=${result.verified_at} result=${resultFile}`);

  console.log(JSON.stringify(result, null, 2));
} finally {
  await server.close();
}
