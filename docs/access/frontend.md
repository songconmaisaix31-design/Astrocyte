# Works Relay v1.2 frontend

F-ui owns the existing React/Vite album, relay pages, generated client schema and new relay tests. Business source is `C:/Users/DW/orca/Astrocyte/tasks/Works-Relay-v1.2.md`; base is `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`, with the same business tree as accepted `d46f9f4`. Actual client: Codex in Orca. Runtime identifies GPT-6 family; the specific model panel is inaccessible in this API session, so no concrete model variant is claimed.

## D1 assembly

`/workspace/works` reuses the current Nav, design tokens, ActiveTab and WorksAlbumPanel. The existing `/workspace` album remains available. Five metrics consume saved works observations; unobserved counts and unverified recent activity remain unknown. Platform names come from returned scanner data. Nonempty handbooks are labeled “有手册”; they never imply successful relay dispatch. True relayed works await the service task projection. Filters, grouping, timeline, hidden works, metadata editing and scan/import remain available.

Cards expose file verification and handbooks directly. Verification submits only an exactly matched approved registered root; missing approval refuses before scanning. It uses the existing generated client and session/CSRF chain. No command retries or automatic dispatch are introduced.

## Pending seams

H-http must publish an exact immutable OpenAPI SHA before schema generation and deep API integration. Today live append versus timely snapshot publication, dirty/untracked/non-Git/single-file workspace handling, visual reference and configurable liveness window remain undecided until the coordinator relays the user's answers. UI does not select their semantics.

## Verification ledger

- `pnpm install --frozen-lockfile`: PASS, dependency locks unchanged.
- First `pnpm --dir web typecheck`: RED, new PageFrame omitted required `onRefresh` and `rail`; preserved in local evidence. Follow-up typecheck: PASS after correction; dedicated album now uses existing ActiveTab and layout classes directly.
- `pnpm --dir web lint`: PASS for initial assembly.
- `pnpm --dir web test relayAlbumMetrics.test.ts`: PASS, 3 tests preserving unknowns and separating handbook/dispatch evidence.
- `pnpm --dir web test:e2e relay-album.spec.ts works-album.spec.ts --workers=1`: PASS, 18 tests across both existing Chromium viewport projects; new test also checks width 390. Evidence is contract-local mock/fixture display and refusal, not a real dispatch.
- Real scanning demo, native CLI relay/handoff, terminal-only observation, durable runtime state migration, receipts and Today closed loop: NOT_RUN pending runtime/HTTP assembly. Mock presentation checks are not live acceptance.

All browser screenshots/traces/logs and real configuration belong only under `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`. Reports contain derived summaries. Coordinator message `msg_46775738d8a5` extended F's exclusive scope for `web/e2e/works-live.spec.ts` only to change the initial four unknown metrics to five and add platform-label verification. All remaining assertions, timeout, skip and execution paths are retained. Original live opt-in: NOT_RUN.

## Three-minute demonstration script

At one origin, open `/workspace/works` and check five scanned metrics/cards within 5 seconds; filter a scanned platform; read one handbook; choose relay, enter the task and preferred data-registered Agent; inspect the mandatory permission confirmation page (A by default, B/C only within current project grant), root and allowed actions; confirm once; observe the service task card, terminal liveness, receipts and newest derived events; navigate to `/cognition/today` and verify the same work's event. This is an acceptance script, not a completed demonstration.
