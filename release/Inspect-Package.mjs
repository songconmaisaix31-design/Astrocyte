import { readdir, lstat } from 'node:fs/promises';
import { createReadStream } from 'node:fs';
import { resolve, join, relative } from 'node:path';
import { homedir } from 'node:os';
import { StringDecoder } from 'node:string_decoder';

const root = resolve(process.argv[2]);
const escape = value => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
const homePattern = homedir().split(/[\\/]/).map(escape).join('[\\\\/]+');
const pattern = new RegExp(`${homePattern}[\\\\/]+|astrocyte-validation[\\\\/]+v1-20261012|formal-20261013-first|v1r-w1[\\\\/]+replay-002|-----BEGIN (?:RSA |EC |OPENSSH |ENCRYPTED )?PRIVATE KEY-----\\r?\\n[A-Za-z0-9+/]{32,}={0,2}\\r?\\n`, 'i');
let files = 0;
let bytes = 0;
async function visit(dir) {
  for (const item of await readdir(dir, { withFileTypes: true })) {
    const path = join(dir, item.name);
    const name = relative(root, path).replaceAll('\\', '/');
    const info = await lstat(path);
    if (info.isSymbolicLink()) throw new Error(`Package link rejected: ${name}`);
    if (/(^|\/)(\.git|\.env[^/]*|\.npmrc|\.netrc|auth\.json|credentials[^/]*|test-results|playwright-report)(\/|$)|\.(sqlite|db|log|jsonl)(-|$|\.)/i.test(name)) throw new Error(`Private or development member: ${name}`);
    if (item.isDirectory()) { await visit(path); continue; }
    if (!info.isFile()) throw new Error(`Nonregular package member: ${name}`);
    const decoders = [new StringDecoder('utf8'), new StringDecoder('utf16le')];
    const tails = ['', ''];
    for await (const chunk of createReadStream(path, { highWaterMark: 1024 * 1024 })) {
      for (let index = 0; index < decoders.length; index++) {
        const text = tails[index] + decoders[index].write(chunk);
        if (pattern.test(text)) throw new Error(`Private production string found in ${name}`);
        tails[index] = text.slice(-4096);
      }
    }
    for (let index = 0; index < decoders.length; index++) {
      if (pattern.test(tails[index] + decoders[index].end())) throw new Error(`Private production string found in ${name}`);
    }
    files++;
    bytes += info.size;
  }
}
await visit(root);
console.log(JSON.stringify({ scope: 'all regular package bytes; streaming UTF8 and UTF16LE; current home, formal constants and potential PEM key material signatures; filenames and links', files, bytes, private_matches: 0 }));
