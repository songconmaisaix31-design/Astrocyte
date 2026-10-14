import assert from 'node:assert/strict';
import { readFile, realpath } from 'node:fs/promises';
import { createRequire, findPackageJSON } from 'node:module';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { runInNewContext } from 'node:vm';
import { root } from '../../scripts/process.mjs';
import { paperDOM } from '../../internal/adapters/importers/paper-dom.mjs';

// 使用本次从公开出版页面完整下载的 HTML；仅输出安全聚合。
const cli = await realpath(join(root, 'node_modules/@steipete/summarize/dist/cli.js'));
const require = createRequire(findPackageJSON('@steipete/summarize-core', pathToFileURL(cli)));
const { parseHTML } = require('linkedom');
const { Readability } = require('@mozilla/readability');
const browserSource = await readFile(join(root, 'extensions/paper/extract.js'), 'utf8');
for (const input of [
  { file: 'plos-public.html', url: 'https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514', state: 'readable_fulltext', key: 'doi:10.1371/journal.pdig.0000514' },
  { file: 'acl-public.html', url: 'https://aclanthology.org/2024.acl-long.1/', state: 'abstract_only', key: 'doi:10.18653/v1/2024.acl-long.1' },
]) {
  const html = await readFile(join(root, '.verification', input.file), 'utf8');
  const document = parseHTML(html).document;
  const parsed = paperDOM(document, input.url, Readability);
  const context = { document: parseHTML(html).document, Readability, URL };
  runInNewContext(browserSource, context);
  const extension = context.astrocytePaperExtract(input.url);
  for (const result of [parsed, extension]) {
    assert.equal(result.source_key, input.key);
    assert.equal(result.content_state, input.state);
    assert.ok(result.title.length > 20);
    assert.ok(result.authors.length > 0);
    assert.ok(result.pdf_urls.length > 0);
    assert.equal(result.source_url, input.url);
    assert.equal(result.truncated, false);
    if (input.state === 'readable_fulltext') assert.ok(result.text.length > 1000);
    else assert.equal(result.text, '');
  }
  assert.equal(extension.text, parsed.text);
  assert.equal(extension.title, parsed.title);
  assert.ok(document.querySelector('script'), '调用者的 DOM 必须保留');
  console.log(JSON.stringify({ source_url: input.url, source_key: parsed.source_key, content_state: parsed.content_state, html_bytes: Buffer.byteLength(html), text_chars: parsed.text.length, authors: parsed.authors.length, pdfs: parsed.pdf_urls.length, node_extension_equal: true }));
}
