import { createServer as createHTTPServer } from 'node:http';
import { createServer } from '../web/node_modules/vite/dist/node/index.js';

let vite;
const listener = createHTTPServer();
async function close() {
  const results = await Promise.allSettled([
    vite?.close(),
    listener.listening ? new Promise((resolve, reject) => listener.close(error => error ? reject(error) : resolve())) : undefined,
  ]);
  if (process.connected) process.disconnect();
  const failures = results.filter(result => result.status === 'rejected');
  if (failures.length) throw new AggregateError(failures.map(result => result.reason), 'Actual Vite cleanup failed');
}
for (const signal of ['SIGINT', 'SIGTERM']) process.once(signal, () => {
  close().catch(error => { console.error(error); process.exitCode = 1; });
});

async function announce(type, webURL) {
  const record = { type, web_url: webURL };
  console.log(JSON.stringify(record));
  await new Promise((resolve, reject) => process.send(record, error => error ? reject(error) : resolve()));
}

try {
  if (!process.send) throw new Error('Actual Vite requires its owning IPC process');
  if (!process.env.ASTROCYTE_TEST_VITE_CACHE) throw new Error('Actual Vite requires an explicit cache directory');
  const requestedPort = Number(process.env.ASTROCYTE_WEB_PORT);
  if (!Number.isInteger(requestedPort) || requestedPort < 0 || requestedPort > 65535) throw new Error('Actual Vite requires an explicit valid TCP port');
  await new Promise((resolve, reject) => {
    listener.once('error', reject);
    listener.listen(requestedPort, '127.0.0.1', resolve);
  });
  const address = listener.address();
  const webURL = `http://127.0.0.1:${address.port}`;
  const configured = new Promise((resolve, reject) => {
    process.once('disconnect', () => reject(new Error('Actual Vite owner disconnected before API configuration')));
    process.once('message', record => {
      try {
        const url = new URL(record.api_url);
        if (Object.keys(record).length !== 1 || record.api_url !== url.origin || url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || !url.port || url.username || url.password) throw new Error('Actual Vite requires its observed loopback API origin');
        resolve(url);
      } catch (error) { reject(error); }
    });
  });
  await announce('vite_listening', webURL);
  const apiURL = await configured;
  process.env.ASTROCYTE_PORT = apiURL.port;
  vite = await createServer({ configLoader: 'runner', cacheDir: process.env.ASTROCYTE_TEST_VITE_CACHE, clearScreen: false, server: { middlewareMode: { server: listener } } });
  listener.on('request', vite.middlewares);
  await announce('vite_ready', webURL);
} catch (error) {
  console.error(error);
  try { await close(); }
  catch (cleanupError) { console.error(cleanupError); }
  process.exitCode = 1;
}
