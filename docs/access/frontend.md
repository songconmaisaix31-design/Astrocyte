# Works Relay v1.2 frontend engineering delivery

## Scope and lineage

F-ui owns the existing React/Vite album, relay components, generated schema and new relay tests. Business source: `C:/Users/DW/orca/Astrocyte/tasks/Works-Relay-v1.2.md`. Baseline: `6ebf523ad4abf4f80d78ee4207fa30168f299ba6`, business-equivalent to accepted `d46f9f4`. Actual client: Codex in Orca, GPT-6 family; the exact model variant is not exposed by this API session.

Branch: `songconmaisaix31-design/works-relay-frontend`. Initial engineering phase UI SOURCE: `2cdcf9dc49722c63aad69ee59f54da36bc17c4f2`, REPORT `ab3d8e27d182d2c28e368454fa9e66af6e8272d5`. Earlier SOURCE stages: `f6433c6`, `9465f3b`, `6cc4a80`, `d44a7e8`; checkpoint REPORT stages: `9d370e0`, `dc58340`. Initial HTTP sources were consumed through ordinary exact-SHA `--no-ff` merges, ending with `c3533d5554403f2523e7c5b1d63b88e268fa9ecc`; its schema equals `25cef120229bd8abbba66eda457da308985b293b`, consumed through `8d353e3381cd620834ee64aea357611267f0d27d`. All client declarations use the existing `pnpm generate`, with no parallel handwritten DTOs. Backend changes enter only through owner merges.

Coordinator `msg_3a71167fe888` authorized settling this independent engineering phase while unanswered human policies remain NOT_RUN. This report is not full product acceptance.

## Delivered behavior

`/workspace/works` reuses current Nav, tokens, ActiveTab and WorksAlbumPanel. Five metrics are platform, work, folder, genuinely dispatched work and recent change; unknown observations stay unknown. Platforms come from scanner data. Handbooks are labeled “有手册” and do not count as dispatch. Filters, grouping, timeline, hidden works and metadata editing remain available. Cards expose exact-root file verification, handbook and relay actions.

An explicit human intent is followed by a mandatory permission confirmation page. It shows the current root, allowed directories/actions, revision, A/B/C material scope and independent B/C switches read only. Fresh projects use A; existing grants are not silently reduced. Exact project/work roots must match. Revision changes and pending/unknown dispatch receipts block confirmation. Data-registered Agent names and inventory availability are shown separately; installation never implies native readiness. All requests use the existing session/origin/CSRF client chain.

Task cards and details read service projections: task/Agent/work/branch/state/elapsed time, distinct live/unverifiable/exited evidence, persistent transitions, receipts and six-field newest-first events. No local task-state promotion occurs. Unknown send retains input and blocks repetition; an independently authorized stop remains available. Receipt statuses are the real pending/accepted/rejected/delivery_unknown enum; accepted acknowledges the action, not task completion. Terminal exit never triggers redispatch.

Human working-to-done and confirmed-exit done-to-released actions are wired, with cancellation submitting nothing. Done still permits authorized read/stop; send is disabled. Release requires service exit evidence. Event publication and human reconciliation are visible; published means isolated application observations, not a cognition snapshot. Today remains a link to the existing view, with no unapproved overlay or event-value substitution.

## Verification

- `pnpm install --frozen-lockfile`: PASS; lock files unchanged.
- `pnpm generate`: PASS, deterministic against the final consumed OpenAPI; generated declarations committed.
- `pnpm --dir web typecheck`: PASS, repeated after final owner merge.
- `pnpm --dir web lint`: PASS.
- `pnpm --dir web test`: PASS, 103 tests in 17 files, repeated on initial engineering SOURCE. Earlier 104 included a superseded permission-selector helper test, removed when current read-only service scope became the contract. Immediate-repair scoped checks are recorded below; this whole suite was not rerun for the repair.
- `pnpm --dir web build`: PASS; existing Vite warning for a chunk above 500 kB remains.
- `pnpm --dir web test:e2e relay-runtime.spec.ts relay-album.spec.ts relay-scan-live.spec.ts works-album.spec.ts --workers=1 --output <local-evidence>`: PASS, 42 across existing viewport projects (runtime 22, relay album 2, existing works 16, real public scan 2). This run used the final frontend and 8d353e contract before the final c3533d5 runtime-only owner merge; schema/frontend are unchanged by that merge. No broad runtime re-acceptance is claimed.
- Earlier 5d9-aligned combined navigation/cognition/album/runtime/public-scan regression: PASS, 150; valid for that source, not upgraded into later full acceptance.
- `git diff --check`: PASS. Coordinator `msg_46775738d8a5` authorized only the two-line old works-live initial four-to-five unknown metrics/label update. Existing timeout, skip and remaining assertions were unchanged. Original opt-in works-live test: NOT_RUN.

First evidence remains preserved: required PageFrame props typecheck RED; card overflow intercepting the relay button in the first runtime E2E RED; imprecise accessible select names in the second runtime E2E RED; generated-contract type mismatches corrected before PASS. The interrupted browser runs and traces are retained, and later targeted PASS does not erase them. Candidate mock results for old succeeded/failed receipts are historical only.

## Public scanning and visual evidence

Coordinator `msg_a9c0712069c2` authorized one self-owned public Git sample and its Orca registration, with no Agent execution. Only its exact root was submitted to works/refresh. Real scanning returned one marker-derived codex platform, one work/folder, two file facts, verified recent change and persisted handbook after reload; true dispatch remains unknown. Both viewport scan cases passed. Child-only native configuration roots suppress private CLI header discovery; earlier default-server mock runs inherited baseline metadata discovery and are not a private-source zero-read audit. No raw private session content was inspected or published by F.

All artifacts stay under `C:/Users/DW/AppData/Local/astrocyte-workbench/v0.2-access/frontend`:

- `public-album-desktop.png` and `public-album-narrow.png`: actual owned sample, width 1920/390, no horizontal overflow. The isolated scanner server's relay Agent/task endpoints show read-failure degradation; these screenshots do not depict real permission/liveness or Agent dispatch.
- `final-e2e-first.log`, `final-results-first`, `final-html-first`: latest 42-case evidence.
- `aligned-e2e-first.log`, `aligned-results-first`: earlier 150-case evidence.
- `first-typecheck-red.txt`, `runtime-first-e2e.log`, `runtime-second-e2e.log`, corresponding result directories: retained first failures.
- `final-build-first.log`, `final-lint-first.log`: build/lint evidence.

The public sample remains locally registered (repository ID `8e938f27-488c-4c20-9a15-242364790b60`), has no remote, and was not deleted. No raw logs, screenshots, configuration, DB or trace entered Git.

## Remaining gates and three-minute script

Unanswered decisions stay open: Today local append versus timely snapshot publication; dirty/untracked/non-Git/single-file workspace handling; reference visual/liveness-window policy; explicit CLI tool/workspace execution permission distinct from material A/B/C. Native Codex/Claude readiness, real CLI handover in the same worktree, durable runtime migration, terminal-only coding and the three-minute same-origin Today loop remain NOT_RUN by F. Mock refusal, receipt, trivalence, transition, registry, duplicate and safe-degradation tests support engineering behavior only. Final backend native acceptance/CI remains with its owner and coordinator; H reported a separate Windows exact-root first RED. Its owner fix enters through the consumed HTTP ancestry; F does not independently upgrade that backend validation to PASS.

Pending acceptance script: open `/workspace/works`, check real scanned metrics/cards within five seconds, filter platform, read handbook, select work/registered Agent, save explicit intent, inspect and confirm current root/material permissions/actions, observe real service task/terminal/receipt/derived events, then verify the same work's event in `/cognition/today` at the same origin. This track has not deployed. The user has authorized persistent deployment at the same URL after root acceptance, to be performed by root/integration; the real acceptance gates are not yet satisfied. No external message or private-source discovery was performed by this track.

## Immediate owner repair: known existing task

Same owner/client/branch, dispatch `ctx_52ef7f772849`; no new model was selected. Mutex SOURCE `c95d0299c1c663892985e7e3a08f7ab30a7d2dc1` consumes exact H SOURCE `7e167026593cd510f2f6e62aa4b3d737531ade64` by ordinary merge. OpenAPI is unchanged for this repair, so schema was not edited or regenerated for the mutex hint.

HTTP now preserves `view_existing_task:<UUID>` only for VersionConflict with a canonical lowercase, nonnil UUID verified through the same human's public GetTask projection. Generic/provider/unknown-ID/unrelated conflicts remain sanitized. F parses only that structured action, never message text or paths. The command rejection and persisted rejected receipt show the existing ID and a manual GET-only task-summary entry. Opening it does not change the original intent, submit an action, grant permission or redispatch. Unknown-command locks remain; arbitrary transport exception messages are replaced with a fixed unknown-result notice, while sanitized ErrorV1 retains its public code/action/request evidence.

Affected verification: `pnpm --dir web test relayPresentation.test.ts relayCommand.test.ts` PASS (5); typecheck, lint and build PASS (existing >500 kB bundle warning). `pnpm --dir web test:e2e relay-runtime.spec.ts --grep 'duplicate dispatch|sanitized generic|actual HTTP|unknown dispatch|unknown send' --workers=1 --output <local-evidence>` second run PASS (12 across the two existing viewport projects). Four cases replay H's actual TCP/HTTP public receipt DTOs (initial and store-reopened rejection); eight use contract-local mocks, including immediate 409, generic/unknown-ID refusal, extra canary fields and unknown dispatch/send. Manual lookup produces GET only, retains the rejected receipt and ready intent, and does not create a second writer.

The opt-in actual-response replay reads only H-owned `LOCALAPPDATA/astrocyte-workbench/v0.2-access/verification/http-runtime-contract-samples.json`. The file contains 200 responses; it does not contain the immediate 409, whose actual HTTP assertion belongs to H's `TestRelayPersistentRealHTTPReceiptsMutexAndRestart`. Browser replay is not live browser-to-runtime or native CLI acceptance, and fixture worktree paths are not published in this report.

Repair first E2E remains RED (6 failed / 6 passed): new locators conflated identical existing-task entries in the permission receipt section and general task receipt section. Only the new locators were scoped to the named permission receipt region; no timeout or old assertion was relaxed. Evidence remains `frontend/repair-mutex-e2e-first.log`, `repair-mutex-results-first`, `repair-mutex-html-first`; the distinct second PASS remains `repair-mutex-e2e-second.log`, `repair-mutex-results-second`, `repair-mutex-html-second`, all under the same local access evidence root. Earlier first REDs and 42/150 source-specific results remain intact. Deployment and all human/live gates above remain unchanged.

## Same-dispatch optional EOF contract compatibility

Coordinator `msg_b40df0499c6b` added a bounded compatibility requirement to this repair. Exact H contract SOURCE `f52a5b84c5e36679c60a7bdbc7605c535dc212cb` was consumed by ordinary merge, then the existing generator produced optional `RelayAgentV1.input_mode: line | stdin_eof`; no handwritten DTO or new business policy. EOF UI SOURCE `b48a6e5008b4da087a5d19c0a091b487197e7494` follows mutex SOURCE `c95d0299c1c663892985e7e3a08f7ab30a7d2dc1` on the same branch.

Task detail reads the existing data registry. Only an actual service-projected terminal drive and an explicitly registered stdin_eof mode disable subsequent send and explain the registered single-input behavior. Structured native drive is not disabled merely because its terminal fallback registration uses EOF. Legacy missing/line modes retain existing behavior; service-side permissions and unsupported-capability rejection remain authoritative. Read/stop controls, liveness and unknown locks are unchanged. This describes registration, not observed EOF execution or terminal exit evidence.

`pnpm generate` deterministic, `pnpm --dir web test relayPresentation.test.ts relayCommand.test.ts` (6 tests), typecheck, lint and build: PASS. `pnpm --dir web test:e2e relay-runtime.spec.ts --grep 'duplicate dispatch|sanitized generic|actual HTTP|unknown dispatch|unknown send|EOF registration' --workers=1 --output <local-evidence>`: PASS, 16 across the two existing viewport projects. This includes the prior 12 affected cases plus four terminal/structured EOF mock presentation cases. The opt-in HTTP receipt replay was enabled; no private-source discovery or real CLI launch was added. Evidence: `frontend/repair-eof-e2e-first.log`, `repair-eof-results-first`, `repair-eof-html-first`, under the same local access root. No unrelated regression suite, backend runtime acceptance, real CLI, three-minute loop, Procmon or 00:30 gate was rerun or promoted.
