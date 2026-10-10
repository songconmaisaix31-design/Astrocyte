# Reused extraction sources

Astrocyte Paper is an independent extension. It reuses summarize's DOM clone and Mozilla Readability extraction approach; it does not fork summarize's daemon, provider configuration, media tools or automatic tab navigation.

- Official repository: https://github.com/steipete/summarize
- Inspected upstream revision: `560197cd4b580554cccf648744c592e867b43bb5` (2026-10-10).
- Inspected source: `apps/chrome-extension/src/entrypoints/extract.content.ts`, `apps/chrome-extension/src/entrypoints/background/extractors/page-readability.ts`, `apps/chrome-extension/README.md`, `docs/chrome-extension.md` at that revision.
- Adapted snippet: clone the document, instantiate `new Readability(cloned, { keepClasses: false })`, and read `parsed.textContent`. Astrocyte requires scholarly full-body evidence and rejects size overflow instead of upstream character truncation or whole-body fallback.
- summarize license: MIT, Copyright (c) 2026 Peter Steinberger; full text in [vendor/summarize-LICENSE](vendor/summarize-LICENSE).
- Runtime server extractor remains the existing exact installed `@steipete/summarize` / `@steipete/summarize-core` **0.25.1**, pinned in the root lockfile. Inspected upstream main and installed runtime are different pins.
- Readability **0.6.0**, Apache-2.0, Arc90/Mozilla: https://github.com/mozilla/readability/tree/0.6.0. Upstream copyright/license notice in [vendor/readability-LICENSE.md](vendor/readability-LICENSE.md); full Apache license in [vendor/Apache-2.0.txt](vendor/Apache-2.0.txt). Browser build copies this existing dependency locally; no remote code or added package is needed.

Future full PDF reading must reuse the existing summarize extraction path after explicit selection; it is not implemented in this staged DOM core. Metadata and a PDF link alone are not full text. Host-family recognition indicates an adapter target, not verified site-wide coverage or permission to bypass access restrictions.
