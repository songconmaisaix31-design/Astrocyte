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

Real loopback HTTP + actual R Service + real isolated SQLite + controlled execution ports: ready intent, mandatory confirm, exact duplicate dispatch with one launch, accepted dispatch work count, exclusive-workspace refusal, explicit test-only live window/quiet unverifiable/confirmed exited, retained terminal identity after reopen, one original rejected confirmation and delivery_unknown send receipt after reopen, read/stop receipts, current grant revocation denial all PASS. Twelve successful actual HTTP responses independently validated against OpenAPI with AJV PASS. Receipt status vocabulary is `pending | accepted | rejected | delivery_unknown`; accepted is action acceptance, never model or task success. Initial task audit edge is empty → ready; task progression then follows the specified state machine.

Native real CLI, terminal-only real coding, live observer, Today live append and administrator formal zero-write acceptance remain NOT_RUN or BLOCKED by the respective undecided policy/owner handoff. No existing permanent service has been replaced.
