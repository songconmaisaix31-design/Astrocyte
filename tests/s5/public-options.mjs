import { readFile } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { pathToFileURL } from 'node:url';

const root = resolve(import.meta.dirname, '../..');
const { marked } = await import(pathToFileURL(join(root, '.verification/s5/runtime/node_modules/marked/lib/marked.esm.js')).href);
const readme = await readFile(join(root, '.verification/s5/upstream/laya/laya-ts/README.md'), 'utf8');
const card = await readFile(join(root, '.verification/s5/upstream/hf-model-card.md'), 'utf8');
const options = marked.lexer(`${readme}\n${card}`)
  .filter(token => token.type === 'heading' || token.type === 'paragraph')
  .slice(0, 100)
  .map((token, index) => ({ id: `Official Laya documentation section: ${token.text} (passage ${index})`, description: token.text }));
if (options.length < 2) throw new Error('Official documentation did not provide enough actual candidates');
process.stdout.write(JSON.stringify(options));
