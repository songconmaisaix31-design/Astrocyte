# Works Relay v1.2 frontend

F-ui owns the existing React/Vite album, relay pages, generated client schema and new relay tests. Business source is `C:/Users/DW/orca/Astrocyte/tasks/Works-Relay-v1.2.md`; base is `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`, with the same business tree as accepted `d46f9f4`. Actual client: Codex in Orca. Runtime identifies GPT-6 family; the specific model panel is inaccessible in this API session, so no concrete model variant is claimed.

## D1 assembly

`/workspace/works` reuses the current Nav, design tokens, ActiveTab and WorksAlbumPanel. The existing `/workspace` album remains available. Five metrics consume saved works observations; unobserved counts and unverified recent activity remain unknown. Platform names come from returned scanner data. Nonempty handbooks are labeled “有手册”; they never imply successful relay dispatch. True relayed works await the service task projection. Filters, grouping, timeline, hidden works, metadata editing and scan/import remain available.

Cards expose file verification and handbooks directly. Verification submits only an exactly matched approved registered root; missing approval refuses before scanning. It uses the existing generated client and session/CSRF chain. No command retries or automatic dispatch are introduced.

## Pending seams

H-http must publish an exact immutable OpenAPI SHA before schema generation and deep API integration. Today live append versus timely snapshot publication, dirty/untracked/non-Git/single-file workspace handling, visual reference and configurable liveness window remain undecided until the coordinator relays the user's answers. UI does not select their semantics.

## D2 contract integration checkpoint

Consumed HTTP sources by ordinary `--no-ff` merges: `f9ab2a5390b09bf4fe23df1c87cf76afb9cb4fb8`, `3e9dc58d4dd89bb352eab9d662c025bd3d9e53d7`, and `5d9c643da7ed420932253ea66bebf615ee11e38f`. Each schema update uses `pnpm generate`. Latest runtime receipts are `pending / accepted / rejected / delivery_unknown`; receipt acceptance acknowledges an action and does not finish its task. Earlier candidate `succeeded / failed` mock results are historical evidence for that candidate only.

The user explicitly creates an intent; the server returns `ready`; a separate mandatory page reads current project scope and records confirmation before dispatch. A/B/C are existing material scopes, with independent B directory and C reference switches. The confirmation shows the service's current permission summary read only: fresh projects default to A, existing B/C grants remain visible and are not silently reduced. Project settings revisions changing invalidate confirmation. No CLI workspace/tool authorization is inferred from material permission; that separate human decision is pending.

Task cards and detail poll GET projections only. They display IDs, actual Agent, work name/ID, branch, state, elapsed time, terminal evidence, transitions, six-field newest-first events and receipts. Three liveness values remain distinct, and terminal exit does not imply task completion or redispatch. Unknown dispatch stays blocked after receipt reads and reopening; unknown send retains input and blocks that send while an independently authorized stop remains possible. Failed and unknown receipts remain listed.

Coordinator requested reachable manual `working → done` and confirmed-exit `done → released` controls; their component is prepared but not wired at this checkpoint while the active browser regression finishes. HTTP source `25cef120229bd8abbba66eda457da308985b293b` adds event-publication and human event reconciliation; it is received and awaits consumption at that same checkpoint. Publication to isolated application event storage is not a cognitive snapshot update or Today acceptance.

## Verification ledger

- `pnpm install --frozen-lockfile`: PASS, dependency locks unchanged.
- First `pnpm --dir web typecheck`: RED, new PageFrame omitted required `onRefresh` and `rail`; preserved in local evidence. Follow-up typecheck: PASS after correction; dedicated album now uses existing ActiveTab and layout classes directly.
- `pnpm --dir web lint`: PASS for initial assembly.
- `pnpm --dir web test relayAlbumMetrics.test.ts`: PASS, 3 tests preserving unknowns and separating handbook/dispatch evidence.
- `pnpm --dir web test:e2e relay-album.spec.ts works-album.spec.ts --workers=1`: PASS, 18 tests across both existing Chromium viewport projects; new test also checks width 390. Evidence is contract-local mock/fixture display and refusal, not a real dispatch.
- Real scanning demo, native CLI relay/handoff, terminal-only observation, durable runtime state migration, receipts and Today closed loop: NOT_RUN pending runtime/HTTP assembly. Mock presentation checks are not live acceptance.

D2 checks: generated drift, current typecheck/lint and all 103 Vitest tests PASS. Earlier full 104 includes a candidate-only permission-options helper test removed when the actual service summary contract superseded the selector. Initial runtime mock suite: first RED from card overflow obscuring the relay button, then a second RED from imprecise accessible select names; both interrupted runs and traces remain local. After real layout/a11y repair, both viewport projects passed all 10 candidate runtime tests without timeout or assertion relaxation. Current 5d9-aligned run covers 150 tests (new permission-summary and revision checks plus existing album/navigation/cognition regressions and public real scanner); it is still running. Public scanner at 1920: PASS, real marker-derived platform, five counts, two file facts, filtering, handbook save and reload; 1280 remains pending.

For public scan, coordinator `msg_a9c0712069c2` authorized one self-owned Git fixture registered in Orca. It contains only README and the existing scanner marker, with no Agent execution. H verified that nonempty `CODEX_HOME`, `CLAUDE_CONFIG_DIR` and `PI_CODING_AGENT_DIR` pointing to empty test directories suppress native private header reads. The isolated server and current browser run use those child-only configurations and select only the exact fixture root for `works/refresh`. Earlier mock/default-server runs inherited baseline metadata discovery; they are not a private-source zero-read audit. No raw session content was inspected or published by F. CLI handover, terminal-only completion and actual Today closure remain NOT_RUN.

All browser screenshots/traces/logs and real configuration belong only under `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`. Reports contain derived summaries. Coordinator message `msg_46775738d8a5` extended F's exclusive scope for `web/e2e/works-live.spec.ts` only to change the initial four unknown metrics to five and add platform-label verification. All remaining assertions, timeout, skip and execution paths are retained. Original live opt-in: NOT_RUN.

## Three-minute demonstration script

At one origin, open `/workspace/works` and check five scanned metrics/cards within 5 seconds; filter a scanned platform; read one handbook; choose relay, enter the task and preferred data-registered Agent; inspect the mandatory permission confirmation page (A by default, B/C only within current project grant), root and allowed actions; confirm once; observe the service task card, terminal liveness, receipts and newest derived events; navigate to `/cognition/today` and verify the same work's event. This is an acceptance script, not a completed demonstration.
