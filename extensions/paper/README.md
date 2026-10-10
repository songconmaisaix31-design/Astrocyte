# Astrocyte Paper

Independent summarize-inspired paper extraction extension. Build with `node extensions/paper/build.mjs`; it reuses the pinned installed summarize core to resolve the vendored Mozilla Readability **0.6.0** (Apache-2.0) and copies the browser-native extraction assets into `extensions/paper/dist`, a loadable Manifest V3 extension. No new package, lockfile, build framework, provider credential, automatic browsing or daemon is shipped.

## Behavior

- A human click on the extension action grants `activeTab` access to the **current page only**. The popup injects Readability + `extract.js`, then shows title, abstract, DOI/arXiv identity, PDF candidate links and the observed `content_state`.
- `content_state` is one of `readable_fulltext | abstract_only | paywall | restricted`. Full text is returned only when the page openly publishes a complete scholarly body; a PDF link or an abstract is never presented as full text; restrictions are never bypassed.
- The popup copies a JSON snapshot to the clipboard for human review. The extension stores no provider credentials and never imports on its own.

## Snapshot shape (for W0/W3 review UI)

```
{ schema_version, source_url, host_family, source_key, title, abstract,
  authors[], doi, arxiv_id, observed_version, pdf_urls[], content_state,
  text, warning, truncated, provenance{ processor, version, mode, source } }
```

`text` is display-only (unbounded page body). The authoritative import re-fetches server-side via the existing `POST /materials/imports` with `adapter=paper_url` + `source_locator=source_url`; it dedupes on `source_key` and re-applies a 16 MiB HTML cap, so a plugin token or pasted `text` is never trusted as the library body.

## Application transport

The same-origin human review/import handoff (`externally_connectable` + human session/CSRF) is owned by W0 and is intentionally not wired here. `background.js` only accepts messages from the extension's own popup; no external sender is trusted.

## Licensing

See [THIRD_PARTY.md](THIRD_PARTY.md). summarize (MIT, Peter Steinberger) and Mozilla Readability (Apache-2.0, Arc90/Mozilla) attributions and full license texts are copied into the built extension. Personal browser installation is reserved to the coordinator; loading tests must use an isolated profile.
