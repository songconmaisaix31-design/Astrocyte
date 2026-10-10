# Astrocyte Paper

Independent summarize-inspired paper extraction extension. Build with `node extensions/paper/build.mjs`; it reuses the pinned installed summarize core to resolve the vendored Mozilla Readability **0.6.0** (Apache-2.0) and copies the browser-native extraction assets into `extensions/paper/dist`, a loadable Manifest V3 extension. No new package, lockfile, build framework, provider credential, automatic browsing or daemon is shipped.

## Install

Load `extensions/paper/dist` as an unpacked extension in Chrome (`chrome://extensions` → Developer mode → Load unpacked). No provider account, API key or sign-in is required. The extension is **not installed in the personal Chrome profile**; it has only been loaded into an isolated Chromium profile for verification.

## Permissions

- `activeTab` — reads the current tab only after a human click on the action; no other-tab, history, bookmark or background-tab access.
- `scripting` — injects the vendored Readability + extraction core into the current tab only.
- `clipboardWrite` — copies the snapshot JSON on an explicit "Copy snapshot" click.

No `storage`, `cookies`, `history`, host or `tabs` read permission is requested. The extension stores no provider credentials and never imports anything on its own.

## Behavior

- A human click on the extension action grants `activeTab` access to the **current page only**. The popup injects Readability + `extract.js`, then shows title, abstract, DOI/arXiv identity, PDF candidate links and the observed `content_state`.
- `content_state` is one of `readable_fulltext | abstract_only | paywall | restricted`. Full text is returned only when the page openly publishes a complete scholarly body; a PDF link or an abstract is never presented as full text; restrictions are never bypassed.
- The "Copy snapshot" button copies the snapshot JSON (including the extracted `text` for `readable_fulltext` pages) to the clipboard for human review.

## Import path (human-reviewed, zero-network)

The popup does not import directly. The flow is:

1. Click **读取当前页并提取** ("Read current page and extract").
2. Click **复制快照** ("Copy snapshot") to copy the snapshot JSON.
3. Paste the JSON into the Astrocyte **Attention** import form for human review, then confirm.

The server ingests this pasted JSON through the **`paper_snapshot`** adapter (`POST /materials/imports`, `adapter=paper_snapshot`), which performs **no network access and no model call**. It stores the original snapshot JSON verbatim as the `paper-snapshot.json` attachment plus the reviewed `text`/version, and re-derives the material `source_key` from the observed arXiv ID / ACL URL / DOI / canonical URL — never trusting the snapshot's self-reported `source_key`. Provenance is recorded as `Processor=paper_snapshot` with mode `browser_snapshot_fulltext | browser_snapshot_abstract_only | browser_snapshot_truncated`: a browser snapshot, **not a publisher-verified online export**.

- `readable_fulltext` → ingested as `browser_snapshot_fulltext` (or `browser_snapshot_truncated` when `truncated`).
- `abstract_only` → ingested as `browser_snapshot_abstract_only`; any captured body is deliberately dropped and never presented as full text.
- `paywall` / `restricted` → rejected (`EvidenceMissing`); no entitlement bypass.

### Optional online adapters

`paper_snapshot` is the offline, human-reviewed path and needs no network. The server additionally supports two optional online paths — `adapter=paper_url` (fetch the page body) and `adapter=paper_pdf` (fetch a public PDF and extract its body). These are distinct from the plugin snapshot, require public-source network access, and may be blocked by the current local fake-IP DNS; the `paper_snapshot` path does not depend on them.

## Snapshot shape (for W0/W3 review UI)

```
{ schema_version, source_url, host_family, source_key, title, abstract,
  authors[], doi, arxiv_id, observed_version, pdf_urls[], content_state,
  text, warning, truncated, provenance{ processor, version, mode, source } }
```

`text` is the extracted full body for `readable_fulltext` pages (empty otherwise); it is the reviewed body ingested via `paper_snapshot`, not display-only text. The plugin's self-reported `source_key` is advisory: the server re-derives identity from DOI/arXiv/URL so a hand-edited key cannot merge different texts.

## Application transport

There is no `externally_connectable` message handoff and no automatic import. The extension only exposes the snapshot to its own popup (`background.js` accepts messages solely from the extension's own popup; no external sender is trusted) and to the clipboard. The application-side import is the explicit paste-and-review step above, so a plugin token can never substitute for human approval.

## Licensing

See [THIRD_PARTY.md](THIRD_PARTY.md). summarize (MIT, Peter Steinberger) and Mozilla Readability (Apache-2.0, Arc90/Mozilla) attributions and full license texts are copied into the built extension. Personal browser installation is reserved to the coordinator; loading tests must use an isolated profile.
