import { expect, it, beforeAll } from 'vitest';
import { request } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { createApiClient } from './client';
import { createRelaySessionApi, type RelaySessionEvents } from './relaySession';

let origin: string, taskID: string;
beforeAll(async () => {
  origin = process.env.ASTROCYTE_TEST_RELAY_PUBLIC_URL || '';
  expect(origin).toMatch(/^http:\/\/127\.0\.0\.1:\d+$/);
  const file = process.env.ASTROCYTE_TEST_RELAY_L1_GETEVENT_PUBLIC_FILE;
  expect(file, '需要实际 GetEvent 公开事实').toBeTruthy();
  taskID = (JSON.parse(await readFile(file!, 'utf8')) as { task_id: string }).task_id;
});

it('uses actual session bootstrap and generated GET query against the owned L1 service', async () => {
  const context = await request.newContext({ baseURL: origin });
  const requests: URL[] = [];
  try {
    const transport: typeof fetch = async input => {
      expect(input).toBeInstanceOf(Request);
      const outgoing = input as Request;
      expect(outgoing.method).toBe('GET');
      requests.push(new URL(outgoing.url));
      const response = await context.get(outgoing.url, { headers: Object.fromEntries(outgoing.headers) });
      return new Response(await response.text(), { status: response.status(), headers: response.headers() });
    };
    const api = createRelaySessionApi(createApiClient(`${origin}/api/v1`, transport));
    const data = await api.events({ task_id: taskID, page: 1, page_size: 1 });
    expect(requests[0].pathname).toBe('/api/v1/auth/session');
    expect(Object.fromEntries(requests[1].searchParams)).toEqual({ task_id: taskID, page: '1', page_size: '1' });
    expect(data.events).toHaveLength(1);
    const direct = await context.get(`${origin}/api/v1/relay/session-events?task_id=${taskID}&page=1&page_size=1`);
    expect(direct.status()).toBe(200);
    expect(data).toEqual(await direct.json() as RelaySessionEvents);
  } finally { await context.dispose(); }
});

it('preserves actual Origin refusal as a read error without command or forbidden-read retries', async () => {
  const context = await request.newContext({ baseURL: origin });
  let deniedOrigin = false, forbiddenRequests = 0;
  try {
    const transport: typeof fetch = async input => {
      expect(input).toBeInstanceOf(Request);
      const outgoing = input as Request;
      expect(outgoing.method).toBe('GET');
      if (deniedOrigin) forbiddenRequests++;
      const response = await context.get(outgoing.url, { headers: { ...Object.fromEntries(outgoing.headers), ...(deniedOrigin ? { Origin: 'http://127.0.0.1:1' } : {}) } });
      return new Response(await response.text(), { status: response.status(), headers: response.headers() });
    };
    const api = createRelaySessionApi(createApiClient(`${origin}/api/v1`, transport));
    await api.events({ task_id: taskID });
    deniedOrigin = true;
    await expect(api.events({ task_id: taskID })).rejects.toMatchObject({ status: 403 });
    expect(forbiddenRequests).toBe(1);
  } finally { await context.dispose(); }
});
