import assert from 'node:assert/strict';
import { readFile, realpath } from 'node:fs/promises';
import { createRequire, findPackageJSON } from 'node:module';
import { isAbsolute, join } from 'node:path';
import { pathToFileURL } from 'node:url';
import { paperDOM } from '../../internal/adapters/importers/paper-dom.mjs';
import { root } from '../../scripts/process.mjs';

// 调用者明确选择已经完整取得的公开原件；解析与来源错误直接失败。
export async function loadPublicPaperInputs(directory) {
  assert.ok(isAbsolute(directory), 'Select the absolute directory containing the actual public HTML originals');
  const cli = await realpath(join(root, 'node_modules/@steipete/summarize/dist/cli.js'));
  const require = createRequire(findPackageJSON('@steipete/summarize-core', pathToFileURL(cli)));
  const { parseHTML } = require('linkedom');
  const { Readability } = require('@mozilla/readability');
  const inputs = [
    { name: 'original', file: 'arxiv-v1-public.html', url: 'https://arxiv.org/html/2501.12948v1', key: 'arxiv:2501.12948', version: 'v1' },
    { name: 'revised', file: 'arxiv-v2-public.html', url: 'https://arxiv.org/html/2501.12948v2', key: 'arxiv:2501.12948', version: 'v2' },
    { name: 'secondary', file: 'plos-public.html', url: 'https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514', key: 'doi:10.1371/journal.pdig.0000514' },
  ];
  const result = {};
  for (const input of inputs) {
    const html = await readFile(join(directory, input.file), 'utf8');
    const parsed = paperDOM(parseHTML(html).document, input.url, Readability);
    assert.equal(parsed.source_url, input.url);
    assert.equal(parsed.source_key, input.key);
    assert.equal(parsed.content_state, 'readable_fulltext');
    assert.equal(parsed.truncated, false);
    assert.ok(parsed.text.length > 1000, 'Actual complete paper text is required');
    if (input.version) assert.equal(parsed.observed_version, input.version);
    result[input.name] = {
      source: parsed,
      text: parsed.text,
      import: {
        kind: 'paper', adapter: 'manual', source_locator: parsed.source_url,
        source_key: '', content_digest: '', export_text: parsed.text,
        title: parsed.title, source_spans: [`whole HTML document: ${parsed.source_url}`],
        collection_reason: null,
      },
    };
  }
  assert.notEqual(result.original.text, result.revised.text, 'Actual source versions must retain their distinct text');
  assert.notEqual(result.original.source.source_key, result.secondary.source.source_key);
  return result;
}
