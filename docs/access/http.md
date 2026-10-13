# Works Relay HTTP / assembly owner

Client: Codex API; session identifies GPT-6 family only, exact deployed model name unavailable. Owner branch: `songconmaisaix31-design/works-relay-http`; baseline `6ebf523`, business equivalent to accepted `d46f9f4`. Specification: Works Relay v1.2, not older revisions.

## Contract candidate

All routes share the existing server, loopback, human session, Origin, CSRF and ErrorV1. `/api/v1/relay/agents` is a typed registry query; `/local-agents` remains the independently verified CLI inventory. `/works` remains the real persisted scan/catalog, existing manual and file verification paths remain authoritative. No scan or CLI launch on a relay GET.

`POST /relay/intents` stores an explicit human intent as `ready`, without launch. `POST /relay/intents/{id}/confirm` requires explicit permission confirmation, A/B/C and current project settings revision. Confirmation is checked again by the sole runtime Service, using the transport-established human Caller. Bodies cannot choose trusted actor, command, raw path, native session or grant. Idempotency-Key names a single operation; the UI must never automatically retry an ambiguous launch or send. Default A does not create a grant or change existing project settings.

`GET /relay/tasks`, task detail, event and receipt queries read the runtime Service. `POST /relay/tasks/{id}/send`, `/read`, `/stop`, `/resume` and `/transition` also call that Service. Terminal, task and worktree writes have no HTTP/store shortcut. Dispatch/send/read/stop preserve their original receipt, including first RED and delivery_unknown. Transport rejection happens before a runtime action; runtime rejection must have a receipt. Agent bearer credentials cannot change human permissions or confirm a launch.

Task state remains ready → accepted → working → done → released. A terminal exit is separate display evidence (`terminal_exited`) for unfinished tasks; silence is `unverifiable`, never online/exited/done. Terminal identity survives restart. Workspace identity remains independent of Agent. No timed, chained or autonomous launch.

Events are newest first and expose exactly `event_id`, `kind`, `topic`, `tokens`, `scope`, `ts`. Null tokens remain unknown. No native text, raw, payload, tool arguments, messages, local SQL errors or source paths belong in event DTOs. Task done is not model success; Git activity is not done.

## Event and Today seams

Baseline has `internal/adapters/events/mapping.go` native observations, `native_reader.go` selected-file reader and `attention/app/event_service.go` PendingEventService. `NewPendingEventService` records observations only and rejects mapped events requiring an undecided signal strength. `cmd/server/main.go` already assembles WorksEventSink into application events. There is no running generic Git watcher or git_record/repository_observation file in this baseline (bounded Go search of internal and cmd). R-runtime will supply a scoped fixed Git observation port; H will bridge existing event vocabulary through the entrypoint, never import another app from workspace/app or transplant validation mappers/engine code.

Relay events write only isolated application event state, never engine/cognitive state, formal source or published snapshot. The existing Workbench Reader continues `mode=ro`, `immutable=1`, `query_only(1)` against atomically published daily snapshots. Today numeric snapshot fields and source provenance are unchanged. Whether Today may append local live events, dirty/untracked and non-Git workspace treatment, and the recent liveness window remain user decisions; this candidate does not set those policies.

## Local operation and validation

All access-layer state, configuration, raw evidence and databases stay under `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access`. The formal V4 source is read only by the existing snapshot program. No edits to `scripts/workbench`, 00:30 tasks, migrations, engine, cognitive schema or weights. Prepare configuration/run instructions here; replacing the current permanent service at `http://127.0.0.1:17321` waits for root acceptance.

Early candidate is for schema generation, not assembled HTTP acceptance. Required owner checks: OpenAPI lint/examples, real HTTP session/Origin/CSRF/human permission rejection, actual runtime persistence/receipts and mutex, six-field canary, unknown/no-retry boundaries, scoped Go build/tests and import boundaries. Full root check/build/E2E, actual CLI handover/terminal-only, administrator Procmon and the two real scheduled dates retain their separate outcomes.

Current transport tests use a real loopback TCP listener and the actual session/Origin/CSRF chain with a recording port: unauthenticated/cross-origin/agent/invalid command denial; ready intent invokes no confirm; missing confirmation reaches Service for its denial receipt; provider canary redaction; descending six-field event projection and null tokens; delivery_unknown invokes send once and does not count as a confirmed relayed work. Actual isolated SQLite EventService test records a stable derived observation twice, preserves one pending receipt and zero cognitive signals. These do not establish CLI launch or runtime store acceptance.

Registration/run configuration prepared in `cmd/server/relay_config.go`: `ASTROCYTE_RELAY=1` selects isolated v0.2 app-state; trusted registration is fixed `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/agents.json`. Entries contain AgentSpec name/launch_cmd/capabilities/drive_mode; no config means no registered launch, while independent `/local-agents` inventory remains visible. `ASTROCYTE_RELAY_LIVE_SECONDS` has no implicit value; unset remains unverifiable. Its eventual selected value is a user policy, not this owner's default. B (`allow_directory`) and C (`expand_references`) are independent existing settings, both explicitly projected; neither authorizes unrestricted CLI tools.

## Runtime assembly checkpoint

Consumed exact R sources `d2cf9d3c13be32369bfe353ef9907c4aa2e4cf6c` and `bcc60c0681f92855fd334008b95a7c73a2d7c509` by ordinary no-ff dependency merges. `cmd/server/relay_assembly.go` wires the sole Service to real Works cache, existing project repository/native registry, independent relay.sqlite and pending observation sink. Unanswered workspace and unrestricted tool policies remain disabled. Main resolves the access state before migration and uses the existing protected database guard.

Coordinator explicitly approved extending this owner's write domain to `cmd/server/workbench_config.go` only to extract `guardWorkbenchExplicitDataDir`: the old v0.1 default remains unchanged, while v0.2 selection reuses every formal/snapshot/path/hardlink check without being overwritten by the old default. New regression in relay_config_test creates a fixture snapshot hardlink and rejects it before SQLite opens; bytes and both paths' WAL/SHM remain unchanged. No real formal source was written or opened by this test.

Real loopback HTTP + actual R Service + real isolated SQLite + controlled execution ports: ready intent, mandatory confirm, exact duplicate dispatch with one launch, accepted dispatch work count, exclusive-workspace refusal, explicit test-only live window/quiet unverifiable/confirmed exited, retained terminal identity after reopen, one original rejected confirmation and delivery_unknown send receipt after reopen, read/stop receipts, current grant revocation denial all PASS. Fourteen successful actual HTTP responses independently validated against OpenAPI with AJV PASS. Receipt status vocabulary is `pending | accepted | rejected | delivery_unknown`; accepted is action acceptance, never model or task success. Initial task audit edge is empty → ready; task progression then follows the specified state machine.

Native real model CLI, terminal-only real coding, Today live append and administrator formal zero-write acceptance remain NOT_RUN or BLOCKED by the respective undecided policy/owner handoff. No existing permanent service has been replaced.

## Independent HTTP engineering report

Validated checkpoint SOURCE: `c3533d5554403f2523e7c5b1d63b88e268fa9ecc`, pushed to the owner branch. It includes exact R SOURCE `a3f47921f4627a8b3e48e44c7cabe0d2b021b617` and REPORT `2e8112669deffd9b85017fc81cc665e7863bb030` through ordinary no-ff merges. Subsequent repairs and their validation are recorded below, without changing this checkpoint's results. The H report is a separate documentation commit; this is an engineering handoff, not Works Relay v1.2 acceptance.

`cmd/server/relay_observer.go` assembles the optional runtime lifecycle port outside both application contexts. A bounded two-second observer calls only existing owned-task observation: it cannot create intent, confirm, launch, send or resume. Cancellation is joined before store cleanup; Shutdown receives a bounded cleanup context. This cadence is an internal observation loop, not the undecided liveness threshold or Today policy. Observer cancellation/join and actual runtime HTTP observation tests PASS.

Task `event_publication` projects `none | pending | published | error` and pending/published counts. Here published means only an isolated application observation was recorded. A failing sink retains its original error and pending event, repeated observation neither retries publication nor renews stale evidence, and only an explicit human `/events/reconcile` with the current observe grant attempts reconciliation. The real HTTP flaky-sink case verifies published-after-reconcile without relaunch, stable origin, retained failure and no automatic replay. Pending observation storage does not imply cognitive normalization, model success or formal publication.

Validation on the exact SOURCE:

| Command/check | Result |
| --- | --- |
| `go test -mod=readonly ./...` | PASS: 29 packages, including actual TCP HTTP + runtime + isolated SQLite fixtures; controlled CLI fixture is not a real model case. |
| `go vet -mod=readonly ./...` | PASS. |
| `go build -mod=readonly -o <access-root>/verification/build/astrocyte-relay-http.exe ./cmd/server` | PASS; output outside Git. |
| `node scripts/check-architecture.mjs` | PASS: 29 packages. |
| `pnpm exec redocly lint contracts/openapi.yaml --config contracts/redocly.yaml` | PASS with one existing unused EventV1 warning. |
| `node scripts/validate-examples.mjs` | PASS: 287 examples. |
| Actual HTTP JSON independently validated with AJV against this OpenAPI | PASS: 14 response samples; evidence stays under access-root/verification. |
| `git diff --check` | PASS. |
| Missing LOCALAPPDATA simulation, with GOCACHE pinned before removing it, `go test -mod=readonly ./cmd/server ./internal/adapters/events -run Relay -count=1` | PASS. Only fixtures use a temporary absolute root when the variable is absent; production still rejects missing/relative LOCALAPPDATA. |
| Local Linux validation via WSL | NOT_RUN: installed Ubuntu has no Go executable. |
| Remote complete CI | RED; details below. |

First CI `38455064422` on `3e9dc58` exposed the Linux fixture's absent-LOCALAPPDATA failure. It remains retained; fixture repair SOURCE `8d353e3381cd620834ee64aea357611267f0d27d` and the missing-variable check are separate later results. The first attempt to simulate the missing variable failed before tests because Go's default GOCACHE also requires LOCALAPPDATA; the corrected test explicitly pins the existing cache first. Production's configuration rejection was not relaxed.

CI `38457889686` on `8d353e3` and `38458172696` on the final SOURCE both remain RED. Ubuntu reached the frontend owner's generated-client schema drift; this branch intentionally does not write `web/src/api/schema.d.ts`. Windows additionally failed R's `TestNeutralWorkspacePrecedesAgentAndReusePreservesChanges` at `workspaces_test.go:42` (`scope_denied: work must match the repository root`); the final-source run also failed `TestRelayControlledCLIFullStateAndGitObservation` at `relay_process_test.go:72` with the same error. Exact failures were handed back to the R owner for runner path/canonical-root diagnosis. Remote race/build/E2E stages were skipped after check failure. Local PASS does not overwrite these remote failures or establish full CI acceptance.

## Formal collector handoff and remaining gates

Read-only source inspection of validation ref `98a13b43e76240436fb17d3246d02507bf959417` identifies the concrete existing entrypoint: `cmd/validation/live.go/runLiveCLI` takes explicit `--live-claude-root`, `--live-codex-root` and the three `--live-*-root` Git roots. `live_capture.go/captureLive` selects authorized JSONL files, verifies native session/CWD association and timestamps, and passes real native tool blocks or Git commits through `NativeSemanticMapper.MapTool/MapGit`, `NormalizeValidationSignal` and the existing production `EventRecorder.Record`. No collector entrypoint consumes Relay's isolated pending-observation database. No private transcript, database, deployed configuration or credentials were opened during this inspection.

For a neutral Relay worktree to enter that existing chain, the real CLI must archive a verifiable native session under an already authorized collector root, its original CWD/session must associate with the worktree, and repository common-dir/remote identity must support the existing mapper's topic classification. Git collection requires the selected explicit Git root and actual commit evidence. Whether those conditions hold is NOT_RUN; current source provides no evidence of a real coding launch or formal ingestion. The collector/root owner must verify authorized source coverage and repository identity for the chosen real worktree and perform any necessary authorized-root configuration change. H does not add default signal strength, import validation mappers/engine code, change the seven weights, or turn task-done/git activity/pending observations into cognitive success.

Current independent HTTP assembly keeps committed-Git workspaces disabled (`AllowCommittedGit:false`) and unrestricted terminal tool policy unset (`LaunchPolicy:nil`) while those choices await the user. Dirty/untracked/non-Git treatment, recent liveness window and Today live append also remain undecided. The observer and cached queries are independently testable, but these safe gates block a real coding dispatch; they are not a reduced acceptance criterion. Formal event-to-cognition acceptance therefore remains BLOCKED/NOT_RUN for the collector handoff and real case.

Root/integration still owns generated-type integration, full build/race/E2E and real Works scan/Agent handover on the integrated SHA. Real native/terminal-only coding and cross-Agent takeover, administrator Procmon, the two real 00:30 scheduled dates, formal cognitive-source/snapshot zero-write evidence and final same-URL permanent deployment remain NOT_RUN or BLOCKED. Existing Workbench Reader remains `mode=ro`, `immutable=1`, `query_only(1)`; no snapshot pipeline, schedule, formal source, mapper, engine, migrations or weights were modified by H.

## Original-owner repairs after the checkpoint

Exact R Windows identity repair `af6e76d1083c73b3fb3d28d0e3c915e99ee014cb` was consumed by ordinary merge SOURCE `0c21364218eb4dc0ae379a5a66942532e986b460`. Existing directory identity now uses `os.Stat`/`os.SameFile`, accepting a real Windows 8.3 alias without accepting an unrelated directory or guessed prefix. `go test -mod=readonly ./internal/adapters/relayworktree ./internal/workspace/app ./cmd/server -count=1` PASS; the additional verbose `TestNeutralWindowsActualShortRoot` ran against an actual distinct alias and PASS, not skip. Later exact-source CI `38459252341` reached generated schema drift on both operating systems after their Go checks, without the earlier root failures; the complete CI still failed and subsequent race/build/E2E were skipped. This later result does not overwrite either earlier Windows RED.

H repair SOURCE `7e167026593cd510f2f6e62aa4b3d737531ade64` preserves the mutex hint in immediate ErrorV1 and saved receipt ErrorV1. Only `version_conflict` with exact `view_existing_task:<canonical nonnil UUID>` and an exact task ID verified through the same human caller's runtime `GetTask` public projection retains that action. Provider messages/paths/raw are discarded, and other conflicts retain the fixed refresh action. Real TCP HTTP with actual Service/SQLite verifies the first refusal, persistent receipt and receipt after reopen show the existing task, with only one launch. Unknown task, malformed UUID, private suffix, noncanonical spelling, unrelated conflict and provider canary cases PASS. `go test -mod=readonly ./internal/adapters/httpapi ./cmd/server -run Relay -count=1` and `git diff --check` PASS; 16 actual HTTP public responses independently pass AJV. OpenAPI shape is unchanged; F's original owner received the exact source and safe-action semantics for an explicit read-only lookup.

## Initial local registration candidate

No active registration was changed. A four-entry candidate is stored only at `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/registry-candidates/20261014/agents.candidate.json`; all four entries pass AJV `RelayAgentV1` validation. It contains trusted absolute direct argv only, uses existing native-registry names and `drive_mode: structured`, and launches nothing. Root/integration can inspect the candidate before selecting the active `agents.json` for the accepted service. Registration is data, not launch permission or an all-Agents-ready claim. The trusted file is loaded once at server startup: an appended entry requires a safe restart to appear in `ListAgents`; there is no hot reload. I's real JSON append/restart/query/UI extension acceptance remains NOT_RUN here, and a mock list or decoder test cannot establish it.

R's bounded public version/help investigation and H's public entry existence check support these candidates:

| Registry name | Public version/help | Trusted direct entry | Configuration / real start / model / native runtime capabilities |
| --- | --- | --- | --- |
| `codex` | 0.162.0, PASS | absolute Node executable + installed `@openai/codex/bin/codex.js` | unknown / NOT_RUN / NOT_RUN / unknown |
| `claude` | 2.1.238, PASS | installed `@anthropic-ai/claude-code/bin/claude.exe`; old cli.js absent | unknown / NOT_RUN / NOT_RUN / unknown |
| `pi` | 1.0.1, PASS | absolute Node executable + installed `@earendil-works/pi-coding-agent/dist/cli.js` | unknown / NOT_RUN / NOT_RUN / unknown |
| `opencode` | 1.18.35, PASS | installed `opencode-ai/bin/opencode.exe` | unknown / NOT_RUN / NOT_RUN / unknown |

The eight native capability observations remain unknown. The additional `workspace_tools: unsupported` describes the existing selected-context adapter's restriction, not the CLI's intrinsic capability: Codex/Claude/Pi disable tools and OpenCode applies deny rules in current source. A coding task must not treat these context-only adapters as a coding-capable Relay. Main's neutral-workspace and unrestricted-tool gates remain unchanged.

The independent inventory already distinguishes PATH installation, private configuration, native startability and native capabilities. A bounded new PATH-only lookup found entries for Codex, Claude, OpenCode, Pi, Grok, Kimi, Qwen and Cursor; Gemini and Cursor Agent were not found on PATH, which is not a whole-machine absence claim. Qwen 0.24.6 has R's public help/version evidence but no verified trusted direct entry or native adapter. Grok/Kimi/Cursor have PATH-entry evidence only here. These four are therefore queryable inventory observations, not executable registration candidates. No configuration, credentials or private native history was inspected.

Terminal-only native candidates remain NOT_RUN: public help did not establish incremental input with an open stdin, EOF submission semantics or TTY compatibility for any real model CLI. OpenCode's JSON format describes output, not an input framing protocol. The terminal pipe helper's successful fixture does not establish these native cases; no PTY was introduced.

Root subsequently approved the optional trusted `AgentSpec.input_mode` field: `line | stdin_eof`, omitted/legacy empty means line. Exact interface SOURCE `556c6f1645778286a171ea02668ed0456cde3a98` uses `omitempty`; H's configuration accepts the two modes and rejects unknown values, while the versioned OpenAPI keeps the field optional and validates both an old four-field entry and an EOF entry. The runtime EOF implementation is still the R owner's work at this contract checkpoint. Intended semantics are one complete initial input followed by closing stdin, normal read/stop and explicit unsupported later Send with no write. This host framing field grants no tool, workspace or launch permission and does not prove real Codex/Claude compatibility. Candidate data remains inactive; root's undecided permission/source/Today/liveness gates are unchanged.

Two optional EOF alternatives are stored only in the same local candidate directory as `terminal-eof.candidate.json`: Codex's trusted direct entry with `exec --json -` and Claude's trusted executable with `--print --input-format text --output-format stream-json --verbose`. Both are explicit terminal/stdin_eof candidates; all runtime capabilities and workspace tools remain unknown and unapproved. They pass AJV as registration data only. Select an alternative for a given name rather than concatenating duplicate names with the structured candidates. Neither active configuration nor a model process was changed.

## Final independent HTTP engineering handoff

Final SOURCE: `c7c3255d0530b93312bb541a477b9ec47e0afc15`, including exact R EOF/noop SOURCE `b19d83635fe10cde9d42642d7444f3771d02f513` and REPORT `4848a98eef45656f30326054f3da8298f61044b1` by ordinary merges. H contract SOURCE `f52a5b84c5e36679c60a7bdbc7605c535dc212cb` was handed to F for the existing generator. H's actual EOF HTTP test SOURCE is `9323154c7b026991fed0ac9fecf0c8167251c8d6`; the final-source merge changes only R's report from that tested source, with identical Go and contract files. Final H REPORT is a separate documentation commit. Root explicitly authorized settlement of this independent engineering phase; complete v1.2 acceptance remains outside that settlement.

Consumed R repairs preserve monotonic maximum event time across native/Git sources, do not refresh unchanged evidence and prioritize real exit. Reusing a released task's terminal ID clears the old native conversation, while original-task Resume does not replay input. `drive_mode` chooses the default path: both structured and both may reach terminal only after a proven zero-start Unsupported result, where the trusted terminal policy is still checked. A missing policy refuses launch. Operation idempotency also compares the target task. Background unchanged observation does not append receipts or update persisted state; fresh evidence/first capture error is recorded, and failed capture requires explicit reconciliation rather than automatic reprobe. Explicit human Read continues to record each receipt.

The real EOF HTTP case uses an actual child executable and stdin/stdout pipes, actual runtime Service, actual isolated SQLite, shared TCP/session/Origin/CSRF and a controlled workspace port. It verifies cached registry `input_mode`, ready intent without launch, human-confirmed complete first manual/task input followed by EOF, safe HTTP 501 Unsupported for subsequent Send, one original rejected receipt for duplicate operation, explicit accepted Read/Stop receipts, raw stdout canary exclusion and real exit distinct from done/released. It is a public helper fixture, not a model CLI or coding demonstration.

Final applicable validation:

| Command/check | Result |
| --- | --- |
| `go test -mod=readonly ./internal/workspace/app ./internal/adapters/agents ./internal/adapters/sqlite ./internal/adapters/relayworktree ./internal/adapters/httpapi ./cmd/server -count=1` after consuming b19d836 | PASS: all six affected packages. |
| `go test -mod=readonly ./cmd/server ./internal/adapters/httpapi -count=1` after the new real EOF HTTP test | PASS: entire two packages, including observer, configuration, origin/auth, mutex, receipts, unknown/sink and EOF. |
| `go test -mod=readonly ./internal/adapters/events -run Relay -count=1` | PASS: six-field pending bridge, stable observation origin and no cognitive signal inference. |
| `go vet -mod=readonly ./cmd/server ./internal/adapters/httpapi ./internal/workspace/app` | PASS. |
| `node scripts/check-architecture.mjs` | PASS: 29 packages. |
| `go build -mod=readonly -o <access-root>/verification/build/astrocyte-relay-http.exe ./cmd/server` | PASS. |
| OpenAPI lint and `node scripts/validate-examples.mjs` on f52a5b8 contract | PASS: existing unused EventV1 warning, 289 examples. |
| Independent AJV on actual HTTP and candidate DTOs | PASS: 16 runtime HTTP samples + 10 EOF HTTP samples (including the actual 501 ErrorV1); four structured candidates + two alternative EOF candidates. |
| `git diff --check` | PASS. |

The EOF HTTP test's first two failures are retained outside Git as `verification/eof-http-first-red.txt` and `eof-http-second-red.txt`. First, its fixture used an arbitrary child path instead of the required LOCALAPPDATA-derived access root and was refused before opening the Relay DB. Second, the recording-only workspace fixture marked Prepared without creating a directory, so the real process could not start and confirmation returned 409. The repaired EOF-specific fixture uses the required root and creates its directory through the Service's workspace port; each rerun uses a new isolated fixture. No production guard was weakened and no original unknown receipt was overwritten or replayed. Later PASS does not replace those failures. The one transient ordinary push rejection at 377b1f5 was followed by read-only remote verification and a successful ordinary push; no force push was used.

The last inspected newer CI `38462618912` on 9323154 was still in progress; earlier 7e16702 CI `38459655342` failed only at F-owned generated schema drift after Go checks on both systems. Complete remote CI is not claimed PASS, and its race/build/E2E remain NOT_RUN until the integrated F-generated client is tested by I/root. Full local Go results at the earlier checkpoint and R's own final full suite remain separately labeled; the table above states H's actual applicable final commands.

Final limitations: no active `agents.json` change, registration hot reload, real native/terminal-only model launch, tool authorization, dirty/non-Git source decision, default liveness window, Today live-append decision, formal collector/normalize/cognition completion, administrator Procmon or two real 00:30 dates. No replacement of the existing 17321 permanent service. I/root still owns same-URL integrated acceptance, real registration extension after safe restart, full UI/build/race/E2E and final deployment; original owners remain responsible for their subsequent fixes.

## Original H owner: Root deep-isolation EOF fixture diagnosis

Task `task_06ea6371c814`, Dispatch `ctx_e1fe34e57251`, same terminal/worktree/branch; no additional worker. Client remains Codex API. Read the local development protocol and this report, plus Root's decisions at `b0280b3`. Scope is the public EOF helper and the specified isolated Root fixture/log; no real Agent, private transcript, credentials or formal database is involved.

Root's independent check on SOURCE `3387019f7c349957ef8f4c75d98ecd7fbbe42200` first failed after 316.5 seconds: only `TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt`, confirmation HTTP 409 / expected 200 / public `delivery_unknown`, test duration 0.13 seconds. The original `root-3387019-20261014/check-first.*` evidence and failed fixture remain retained. This is a distinct later RED, not a replacement of H's earlier stage results. Root's later build PASS does not clear that check failure.

Bounded causal findings:

- Root's valid child environment points LOCALAPPDATA into its deeper isolated directory and all three native discovery variables into its public-empty-cli directory. With `GOFLAGS=-p=1`, `GOMAXPROCS=2`, `NODE_OPTIONS=--max-old-space-size=512`, H reproduced the same target failure in a fresh fixture: EXIT 1, 10.38 seconds overall. The existing compiler cache was pinned separately to avoid recompiling unrelated code; all runtime isolation and resource settings stayed unchanged.
- Both retained failing fixtures have an existing owned workspace path of length 281, without a first-input proof file. The old H fixture appended `access/verification/eof-http/<UUID>` to the caller's LOCALAPPDATA, then set that as its own LOCALAPPDATA and appended another access-root/worktrees/UUID scaffold. This unnecessary layout inflation came from the test fixture, not Root's external isolation wrapper.
- An independently compiled public Go test helper, with stdin/stdout redirected and no model entrypoint, could not start in a fresh existing directory of the same length 281: Windows process creation returned native error 267 (`ERROR_DIRECTORY`) before Start succeeded. The same helper started in the length-126 control directory, received complete input followed by EOF, created its proof and exited 0. Only static error class, lengths, status and timing are reported; private diagnostics are never emitted.
- The evidence locates the failure before EOF submission, rather than a closed-input race or a demonstrated runtime domain defect. The existing conservative `delivery_unknown` mapping and original receipt were not changed. The observed Windows long-CWD process restriction remains a limitation; this is a fixture repair, not a product fix for arbitrary long process paths.

Repair SOURCE `19c83e39b65e9a0cebc5bad20f7dba3f6c82cd2a` changes only `cmd/server/relay_eof_http_test.go`: create unique child appdata with `os.MkdirTemp` directly under the same caller-derived access root. It remains outside Git and global temporary storage. The external Root directory and all native-empty/resource settings, actual HTTP/Service/SQLite/child executable, workspace identity, input and receipt assertions, permissions, operation identity and deadlines are unchanged. No R driver, specification, OpenAPI, workflow or business code changed; no original unknown action was replayed.

| Exact same deep Root environment check | Later result |
| --- | --- |
| `go test -mod=readonly ./cmd/server -run '^TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt$' -count=1 -v` | PASS / EXIT 0 / 3.50 seconds; owned CWD length 233 and full first-input proof exists. |
| `go test -mod=readonly ./cmd/server -count=1 -v` | PASS / EXIT 0 / 19.16 seconds; 23 top-level tests, zero failures, including the original target. |
| `go vet -mod=readonly ./cmd/server` | PASS / EXIT 0. |
| `gofmt` / `git diff --check` | PASS. |

Separate evidence stays under `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/h-eof-root338-diagnosis-20261014`: `exact-root-reproduction-first.log`, `process-controls-derived.json`, `exact-root-after-fixture-fix.log`, `exact-root-entire-cmd-server.log` and `exact-root-cmd-server-vet.log`. The first Root log/result and original failed workspace/store are untouched; fresh case results do not overwrite the original rejected/unknown receipt. Source and this REPORT are separate ordinary commits and pushed, with clean/remote parity verified at handoff.

The exact repair was handed to I/root for an ordinary merge. H did not rerun unrelated full suites, modify I's CI resource configuration or weaken any check. Root still must independently pass complete check/build on the integrated source, and I owns the full dual-platform CI with the generated frontend types and approved Windows resource change. Real native/terminal-only coding, formal cognition, Today, tool/source/liveness policy decisions, administrator Procmon, two real scheduled dates and permanent deployment retain their previous NOT_RUN/undecided outcomes.
