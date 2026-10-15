import assert from 'node:assert/strict';
import { readFile, realpath } from 'node:fs/promises';
import { createRequire, findPackageJSON } from 'node:module';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { runInNewContext } from 'node:vm';
import { root } from '../../scripts/process.mjs';
import { paperDOM } from '../../internal/adapters/importers/paper-dom.mjs';

const cli = await realpath(join(root, 'node_modules/@steipete/summarize/dist/cli.js'));
const require = createRequire(findPackageJSON('@steipete/summarize-core', pathToFileURL(cli)));
const { parseHTML } = require('linkedom');
const { Readability } = require('@mozilla/readability');
const original = await readFile(join(root, '.verification/acl-public.html'), 'utf8');
const extension = await readFile(join(root, 'extensions/paper/extract.js'), 'utf8');
const locator = 'https://aclanthology.org/2024.acl-long.1/';

// 坏 JSON/URL 是明确校验参数，不声明公开原件包含这些错误。
for (const input of [
  { name: 'malformed_json_ld', error: 'SyntaxError', change(document) {
    const script = document.createElement('script');
    script.type = 'application/ld+json';
    script.textContent = '{';
    document.head.append(script);
  } },
  { name: 'malformed_pdf_url', error: 'TypeError', change(document) {
    const meta = document.querySelector('meta[name="citation_pdf_url"]');
    assert.ok(meta, 'Actual ACL PDF metadata is required');
    meta.setAttribute('content', 'https://[invalid');
  } },
]) {
  for (const environment of ['node', 'extension']) {
    const document = parseHTML(original).document;
    input.change(document);
    if (environment === 'node') assert.throws(() => paperDOM(document, locator, Readability), { name: input.error });
    else {
      const context = { document, Readability, URL };
      runInNewContext(extension, context);
      assert.throws(() => context.astrocytePaperExtract(locator), { name: input.error });
    }
    console.log(JSON.stringify({ input: input.name, environment, rejected_with: input.error }));
  }
}
