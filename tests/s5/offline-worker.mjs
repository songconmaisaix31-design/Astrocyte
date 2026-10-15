import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const evidence = process.env.ASTROCYTE_S5_NETWORK_EVIDENCE;
if (!evidence) throw new Error('An explicit owned network evidence path is required');
const url = 'https://huggingface.co/api/models/convaiinnovations/laya/revision/7b928d828b7b0e022f929d9bd2e44165aa270148';
let failure;
try {
  const response = await fetch(url, { signal: AbortSignal.timeout(10000) });
  await response.arrayBuffer();
  throw new Error(`The official HTTP outlet remained reachable: ${response.status}`);
} catch (error) {
  if (error.message.startsWith('The official HTTP outlet remained reachable')) throw error;
  failure = { name: error.name, message: error.message, stack: error.stack, cause: { code: error.cause?.code, address: error.cause?.address, port: error.cause?.port, message: error.cause?.message } };
}
// 只确认本进程真实 HTTP 出口拒绝；OS 隔离使用独立审查步骤。
if (process.env.ASTROCYTE_S5_NETWORK_MODE === 'http-proxy' && failure.cause.code !== 'ECONNREFUSED') {
  throw new Error(`Expected an actual unavailable proxy refusal: ${JSON.stringify(failure)}`);
}
await writeFile(evidence, JSON.stringify({ pid: process.pid, url, mode: process.env.ASTROCYTE_S5_NETWORK_MODE, failure, observed_at: new Date().toISOString() }, null, 2));
await import(pathToFileURL(resolve(import.meta.dirname, '../../internal/adapters/judgment/laya-worker.mjs')).href);
