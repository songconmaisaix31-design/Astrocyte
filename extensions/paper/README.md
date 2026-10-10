# Astrocyte Paper

Independent summarize-inspired paper extraction core. Build with `node extensions/paper/build.mjs` using existing pinned project dependencies; no package installation or new lockfile. The core performs only DOM inspection, no automatic browsing, import, model call, provider credential storage, cookie access or personal tab enumeration.

Current build stages local Readability plus Astrocyte scholarly metadata/provenance parsing. Search versus current-page/both and plugin action/permission scope remain **ASKED PENDING**. Final MV3 manifest and application transport will be implemented after the root relays those decisions; staging assets are not an installable extension.

Current-page design for review: a user click acquires only that page, displays title/abstract/identity and observed extraction state, and requires explicit selection before import or model processing. PDF candidates are displayed as links, not automatically fetched or described as read full text. DOM readable text requires scholarly body sections; access restrictions/unknown text stay explicit. Application writes must use a root-approved human identity path owned by W0.

See [THIRD_PARTY.md](THIRD_PARTY.md) for exact upstream pins and licensing. Personal browser installation is reserved to the coordinator. Browser loading tests must use an isolated profile with W3 slot coordination.
