import test from 'node:test';
import assert from 'node:assert/strict';
import { findPackageJSON, createRequire } from 'node:module';
import { realpathSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
import { paperDOM, normalizeDOI, paperHostFamily } from './paper-dom.mjs';
const cli = realpathSync(new URL('../../../node_modules/@steipete/summarize/dist/cli.js', import.meta.url));
const require = createRequire(findPackageJSON('@steipete/summarize-core', pathToFileURL(cli)));
const { parseHTML } = require('linkedom');
const { Readability } = require('@mozilla/readability');
const extract = (html, url = 'https://journals.plos.org/plosone/article?id=10.1371/journal.pone.0000001') => paperDOM(parseHTML(html).document, url, Readability);

test('scholarly title, abstract and DOI are metadata; PDF is not full text', () => {
  const result = extract('<html><head><meta name="citation_title" content="Paper"><meta name="citation_doi" content="10.1371/Journal.Pone.0000001"><meta name="citation_author" content="Author"><meta name="citation_abstract" content="Abstract only"><meta name="citation_pdf_url" content="/paper.pdf"></head><body><article><h1>Paper</h1><p>Abstract only</p></article></body></html>');
  assert.equal(result.source_key, 'doi:10.1371/journal.pone.0000001');
  assert.equal(result.text, '');
  assert.notEqual(result.content_state, 'readable_fulltext');
  assert.equal(result.pdf_urls[0], 'https://journals.plos.org/paper.pdf');
  assert.deepEqual(result.authors, ['Author']);
});

test('actual parser distinguishes full body and preserves source document', () => {
  const body = 'A measured experimental observation with explicit evidence and limits. '.repeat(40);
  const html = `<html><head><meta name="citation_title" content="Original"></head><body><nav>Unrelated site navigation</nav><article><h1>Original</h1><h2>1 Introduction</h2><p>${body}</p><h2>2 Results</h2><p>${body}</p></article></body></html>`;
  const doc = parseHTML(html).document;
  const result = paperDOM(doc, 'https://arxiv.org/html/2501.12948v2', Readability);
  assert.equal(result.content_state, 'readable_fulltext');
  assert.equal(result.source_key, 'arxiv:2501.12948');
  assert.equal(result.observed_version, 'v2');
  assert.equal(result.truncated, false);
  assert.ok(result.text.includes(body));
  assert.ok(doc.querySelector('nav'));
  assert.ok(!result.text.includes('Unrelated site navigation'));
});

test('JSON-LD and DOI aliases dedupe; invalid URLs/identities remain absent', () => {
  const result = extract('<html><head><script type="application/ld+json">{"@type":"ScholarlyArticle","headline":"LD title","abstract":"metadata","identifier":"https://doi.org/10.1000/ABC"}</script></head><body></body></html>', 'https://dl.acm.org/doi/10.1000/ABC');
  assert.equal(result.source_key, 'doi:10.1000/abc');
  assert.equal(normalizeDOI('https://doi.org/10.1000/ABC'), '10.1000/abc');
  assert.equal(normalizeDOI('10.1000/abc?token=x'), '');
  assert.throws(() => extract('', 'https://user:pass@nature.com/paper'));
  assert.equal(paperHostFamily('evilnature.com'), 'generic');
  assert.equal(paperHostFamily('www.nature.com'), 'nature');
});

test('access restriction never produces full paper evidence', () => {
  const result = extract('<article><h2>Introduction</h2><p>Access through your institution. ' + 'Restricted. '.repeat(180) + '</p><h2>Results</h2><p>Restricted.</p></article>');
  assert.equal(result.content_state, 'restricted');
  assert.equal(result.text, '');
});
