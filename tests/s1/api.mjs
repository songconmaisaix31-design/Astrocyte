import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { setTimeout as delay } from 'node:timers/promises';
import { validateResponse } from './contracts.mjs';

export function command(body = {}) {
  return { schema_version: 1, request_id: randomUUID(), expected_version: 1, ...body };
}

/** Real loopback requests. Writes are sent once; job polling never resubmits them. */
export async function humanAPI(baseURL) {
  const bootstrap = await fetch(`${baseURL}/api/v1/auth/session`, { signal: AbortSignal.timeout(5000) });
  assert.equal(bootstrap.status, 200, await bootstrap.clone().text());
  const session = await bootstrap.json();
  validateResponse('GET', '/auth/session', bootstrap.status, session);
  assert.equal(session.actor_kind, 'human');
  const cookie = bootstrap.headers.getSetCookie().map(value => value.split(';')[0]).join('; ');
  assert.ok(cookie, 'Local human session must set a cookie');
  return {
    session, cookie,
    async request(path, { method = 'GET', body, key = randomUUID(), headers = {} } = {}) {
      const response = await fetch(`${baseURL}/api/v1${path}`, {
        method, signal: AbortSignal.timeout(10_000),
        headers: {
          Cookie: cookie, ...(method === 'GET' ? {} : { Origin: baseURL, 'Content-Type': 'application/json', 'X-CSRF-Token': session.csrf_token, 'Idempotency-Key': key }),
          ...headers,
        },
        ...(body === undefined ? {} : { body: JSON.stringify(body) }),
      });
      const text = await response.text();
      let data;
      try { data = JSON.parse(text); } catch { throw new Error(`${method} ${path} returned non-JSON ${response.status}: ${text}`); }
      validateResponse(method, path, response.status, data);
      return { status: response.status, data };
    },
    async get(path) {
      const result = await this.request(path);
      assert.equal(result.status, 200, JSON.stringify(result.data));
      return result.data;
    },
    async write(path, body, { status = 200, key = randomUUID(), method = 'POST' } = {}) {
      const result = await this.request(path, { method, body: command(body), key });
      assert.equal(result.status, status, JSON.stringify(result.data));
      return result.data;
    },
  };
}

export async function waitJob(api, id, expected = 'succeeded', timeoutMs = 10_000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const job = await api.get(`/jobs/${id}`);
    if (['succeeded', 'failed', 'cancelled'].includes(job.status)) {
      assert.equal(job.status, expected, JSON.stringify(job));
      return job;
    }
    await delay(timeoutMs > 10_000 ? 500 : 50);
  }
  throw new Error(`Job ${id} did not reach ${expected}`);
}

export async function importMaterial(api, body, options) {
  const receipt = await api.write('/materials/imports', body, { status: 202, ...options });
  const job = await waitJob(api, receipt.job_id, 'succeeded', options?.timeoutMs);
  assert.ok(job.material_id);
  const detail = await api.get(`/materials/${job.material_id}`);
  return { receipt, job, detail };
}
