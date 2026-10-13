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
