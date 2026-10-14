// Generated from contracts/openapi.yaml. Run pnpm generate; do not edit.
export interface paths {
    "/relay/session-events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read local owned session admissions
         * @description Existing human session and Origin guard. Rechecks project grant, revision and owned binding on every GET, returns cached/persisted six-field derived events by source timestamp descending. GET never captures, launches or resumes. Snapshot extent is separate from complete native session coverage. No raw/native IDs/paths/commands or cognitive values.
         */
        get: operations["listRelaySessionEvents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/events/reconcile": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Explicit human reconciliation of the same derived event origins
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset. Does not re-execute Agent or clear original receipts. Published means admitted to isolated app observations, not cognitive normalization or snapshot publication.
         */
        post: operations["reconcileRelayTaskEvents"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/projects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read current human project permission scope
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["listRelayProjects"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/agents": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listRelayAgents
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["listRelayAgents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listRelayTasks
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["listRelayTasks"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["getRelayTask"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listRelayTaskEvents
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["listRelayTaskEvents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/receipts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listRelayTaskReceipts
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        get: operations["listRelayTaskReceipts"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/intents": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * createRelayIntent
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["createRelayIntent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/intents/{id}/confirm": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * confirmRelayIntent
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["confirmRelayIntent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/send": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * sendRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["sendRelayTask"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/read": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * readRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["readRelayTask"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/stop": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * stopRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["stopRelayTask"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * resumeRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["resumeRelayTask"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/relay/tasks/{id}/transition": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * transitionRelayTask
         * @description Candidate Works Relay v1.2 contract. Existing human session and Origin guard; POST also requires CSRF and Idempotency-Key. Sole runtime Service owns state and immutable action receipts. No automatic retry or dispatch. Unconfirmed user seams remain unset.
         */
        post: operations["transitionRelayTask"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/health": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench health
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchHealth"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/today": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench today
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchToday"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/focus": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench focus
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchFocus"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/topics/evidence": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench topics/evidence
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchEvidence"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/agents": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench agents
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchAgents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench events
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchEvents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/cognition/filters": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cognitive workbench filters
         * @description Read-only published snapshot. Human session, loopback and origin guard apply; no source DB fallback. Events use exactly six safe fields. Shanghai dates are inclusive; missing state is preserved.
         */
        get: operations["getWorkbenchFilters"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/source-collections/discover": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Discover explicitly selected source collections
         * @description Human CSRF command invokes a bounded public provider or an explicitly configured selected browser folder bridge and caches metadata. Failure preserves prior catalog; GET only reads cache. No binding, full body extraction, model call, credentials or automatic selection.
         */
        post: operations["discoverSourceCollections"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/github-repositories": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read public repository cache
         * @description Human cached GET only; no network, code checkout, credential access, model processing or Agent launch.
         */
        get: operations["listGitHubRepositories"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/github-repositories/sync": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Sync explicitly named public GitHub metadata
         * @description Anonymous fixed-program metadata retrieval capped at 100 repositories. No clone or README/body extraction; only explicit human input and CSRF. Account/private scope remains pending.
         */
        post: operations["syncGitHubRepositories"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/github-repositories/{id}/placement": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Place repository in top-level development space
         * @description Human selects an existing ProjectSpace; fixed program clones into its managed root and records actual HEAD. Repeated placement reuses the clone. Default A only; B/C, control, model and each action remain independently human approved. No Mission or Agent autostart.
         */
        post: operations["placeGitHubRepository"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/probe": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Verify chosen native CLI connection
         * @description Human explicitly verifies one registered native adapter in the approved project: initialize a bounded empty native session, confirm owned stop, and cache actual observations. No model turn or automatic dispatch; unknown stays unknown until observed.
         */
        post: operations["probeNativeCLI"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/source-collections": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Source Collections
         * @description Cached favorite folder metadata for one explicit owner; GET never invokes a browser, CLI or provider. Explicit human POST discovery refreshes this cache.
         */
        get: operations["listSourceCollections"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/agent-token": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * issue Project Agent Token
         * @description Human-only issuance for an already approved project Agent grant. Returned once, never stored in browser fixtures or logs; tokens recheck current grant on every operation.
         */
        post: operations["issueProjectAgentToken"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/discover": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * discover Native Sessions
         * @description Reads only explicitly registered CLI history roots and filters exact project identity. External observations do not confer process control.
         */
        post: operations["discoverNativeSessions"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/{session_id}/context": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * read Native Context
         * @description Project-scoped native context; registered history paths remain internal and do not authorize arbitrary transcript reads.
         */
        get: operations["readNativeContext"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/tracking-sources": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Tracking Sources
         * @description Public metadata only; no cookies, login, credentials or body extraction before human selection. Binding never starts execution; unavailable services return 501.
         */
        get: operations["listTrackingSources"];
        put?: never;
        /**
         * bind Tracking Source
         * @description Public metadata only; no cookies, login, credentials or body extraction before human selection. Binding never starts execution; unavailable services return 501.
         */
        post: operations["bindTrackingSource"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/tracking-sources/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * get Tracking Source
         * @description Public metadata only; no cookies, login, credentials or body extraction before human selection. Binding never starts execution; unavailable services return 501.
         */
        get: operations["getTrackingSource"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/tracking-sources/{id}/sync": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * sync Tracking Source
         * @description Sync up to 100 public metadata records; startup runs once and has no timer. Partial pages retain cached records and warnings.
         */
        post: operations["syncTrackingSource"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/tracking-sources/{id}/recommend": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * recommend Source Items
         * @description Selected metadata revisions sent only to the human-approved project CLI. Recommendations do not import content or approve themselves. Unknown model outcomes are not replayed.
         */
        post: operations["recommendSourceItems"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/tracking-sources/{id}/select": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * select Source Items
         * @description Human selection of exact metadata revisions queues normal reusable summarize imports; no full-list auto import.
         */
        post: operations["selectSourceItems"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/registered": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cached registered project observations
         * @description Human-only cached SQLite observations. GET never scans roots, invokes Git or starts a CLI. Candidates must be manually bound to an Attention space; discovery grants no context, model or control permission.
         */
        get: operations["listRegisteredProjects"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/registered/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Refresh observations of Orca registered project roots
         * @description Explicit human action runs bounded fixed discovery over actual registered roots and associated worktrees. It runs no project scripts, installs no dependencies, and never launches an Agent. Source failures retain old observations with stale status.
         */
        post: operations["refreshRegisteredProjects"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/registered/{project_id}/metadata": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * Save human notes, review, grouping and archive state
         * @description Human session and CSRF only. The revision is the expected manual metadata revision (zero before the first save). Observations and refresh never replace these fields. This command grants no project access, model consent or Agent control and accepts no inferred progress.
         */
        put: operations["setRegisteredProjectMetadata"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Local Projects
         * @description Human registered roots only; no whole-drive discovery.
         */
        get: operations["listLocalProjects"];
        put?: never;
        /**
         * register Local Project
         * @description Human registers one explicit root and Attention project space; default A selected fixed refs, B/C and external model permission disabled.
         */
        post: operations["registerLocalProject"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/discover": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * discover Local Projects
         * @description Human chooses a bounded parent root for project discovery; no arbitrary disk scanning.
         */
        post: operations["discoverLocalProjects"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/settings": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * set Local Project Settings
         * @description Human-only CAS settings. B allowlisted directories, C reference expansion and model CLI are explicit project permissions; revoke checked per operation.
         */
        put: operations["setLocalProjectSettings"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/grants": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * grant Local Project Agent
         * @description Human-only operation-scoped Agent grant; token cannot confer permissions on itself.
         */
        post: operations["grantLocalProjectAgent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/grants/revoke": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * revoke Local Project Agent
         * @description Human-only revoke; subsequent reads and native operations reload current authority.
         */
        post: operations["revokeLocalProjectAgent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/context": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * read Local Project Context
         * @description Current project scoped A fixed references plus explicitly enabled B/C only; read does not increase human attention.
         */
        post: operations["readLocalProjectContext"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Native Sessions
         * @description Native session observations for this project only; no arbitrary native histories.
         */
        get: operations["listNativeSessions"];
        put?: never;
        /**
         * start Native Session
         * @description Explicitly approved project CLI and actions only; automatic Agent dispatch remains disabled pending user clarification. Cooperative mode does not promise OS isolation.
         */
        post: operations["startNativeSession"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/{session_id}/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * resume Native Session
         * @description Native session identity and ownership must be verified; unsupported native resume never becomes a new session silently.
         */
        post: operations["resumeNativeSession"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/{session_id}/send": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * send Native Message
         * @description Operation-scoped input to this project session. Unknown effects are never resent automatically.
         */
        post: operations["sendNativeMessage"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/{session_id}/stop": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * stop Native Session
         * @description Stop only the owned native session and preserve actual stop confirmation; failed stop remains blocked.
         */
        post: operations["stopNativeSession"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/sessions/{session_id}/observe": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * observe Native Session
         * @description Read factual native output and current state; no guessed completion, no synthetic history.
         */
        get: operations["observeNativeSession"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/papers/search": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * search public academic papers
         * @description Public metadata search only. It performs no import, no full-text extraction and no model call. A human selects a result and imports it through the existing ImportMaterial path (kind=paper, adapter per site, source_key dedup, new revision). Paywalled or login-restricted content is reported, never bypassed. A server without the search service returns 501.
         */
        get: operations["searchPapers"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/progress": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * read cached project progress
         * @description Cached human-authored or model-inferred project stage read from approved fixed TASK/STATUS files only. It is never authoritative project state, a grant or a human approval; evidence carries source, version and freshness. A server without the progress service returns 501, not an empty success.
         */
        get: operations["getProjectProgress"];
        /**
         * set human-authored project progress
         * @description Human-authored project stage, summary and optional 0..100 percent. It is an advisory observation and never grants or expands project permission. A server without the progress service returns 501.
         */
        put: operations["setProjectProgress"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-projects/{id}/progress/infer": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * infer project progress from approved TASK/STATUS files
         * @description Triggers a model inference over approved fixed TASK/STATUS files only. The caller-supplied Idempotency-Key header is the verified operation identity, persisted before the model turn so a retry cannot silently re-charge; unknown delivery is never replayed. Evidence carries source, version and freshness and never reads credentials, unapproved files or grants project control. A server without the progress service returns 501.
         */
        post: operations["inferProjectProgress"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/local-agents": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read cached local CLI inventory
         * @description Human-session read of bounded PATH-based CLI observations. This GET never executes CLI probes, reads credentials or scans private sessions/projects. Installation, configuration and startability are independent observations; native capabilities require actual native runtime checks. Unknown is not support. No project access or native control is granted by this inventory. A server without the inventory service returns 501, not an empty success.
         */
        get: operations["listLocalAgents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/health": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * get Health
         * @description S0 implements this read-only operation.
         */
        get: operations["getHealth"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/foundation": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * get Foundation
         * @description S0 implements this read-only operation.
         */
        get: operations["getFoundation"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/materials": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Materials
         * @description S0 implements this read-only operation.
         */
        get: operations["listMaterials"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/opportunities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Opportunitys
         * @description S0 implements this read-only operation.
         */
        get: operations["listOpportunitys"];
        put?: never;
        /**
         * createOpportunity
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["createOpportunity"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/projects": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Projects
         * @description S0 implements this read-only operation.
         */
        get: operations["listProjects"];
        put?: never;
        /**
         * register Project
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["registerProject"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/proposals": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Proposals
         * @description S0 implements this read-only operation.
         */
        get: operations["listProposals"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/sessions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Sessions
         * @description S0 implements this read-only operation.
         */
        get: operations["listSessions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/missions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Missions
         * @description S0 implements this read-only operation.
         */
        get: operations["listMissions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/missions/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * get Mission
         * @description S0 implements this read-only operation.
         */
        get: operations["getMission"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/materials/imports": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * import Material
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["importMaterial"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/opportunities/{id}/reviews": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * review Opportunity
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["reviewOpportunity"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/opportunities/{id}/admissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * admit Opportunity
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["admitOpportunity"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/sessions/{id}/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * resume Session
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["resumeSession"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/sessions/{id}/handoff": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * handoff Session
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["handoffSession"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/proposals/{id}/submit": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * submit Proposal
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["submitProposal"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/proposals/{id}/approvals": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * approve Proposal
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["approveProposal"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/approvals/{id}/revoke": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * revoke Approval
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["revokeApproval"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/missions/{id}/pause": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * pause Mission
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["pauseMission"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/missions/{id}/cancel": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * cancel Mission
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["cancelMission"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/work-items/{id}/claims": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * claim Work Item
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["claimWorkItem"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/work-items/{id}/artifacts": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * submit Artifact
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["submitArtifact"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/artifacts/{id}/acceptance": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * accept Artifact
         * @description S0 MUST return HTTP 501 unsupported_capability without side effects. Examples describe future DTOs only.
         */
        post: operations["acceptArtifact"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Stream scoped events
         * @description Future cursor-based SSE projection of EventV1. S0 returns 501; the browser must never infer mission state from accumulated events.
         */
        get: operations["streamEvents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/session": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getLocalSession
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
         */
        get: operations["getLocalSession"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/materials/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getMaterial
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["getMaterial"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /**
         * updateMaterial
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        patch: operations["updateMaterial"];
        trace?: never;
    };
    "/materials/{id}/revisions/{revision}/content": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getMaterialContent
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["getMaterialContent"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/materials/{id}/uses": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * recordMaterialUse
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["recordMaterialUse"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/distillations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listDistillations
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["listDistillations"];
        put?: never;
        /**
         * recordDistillation
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["recordDistillation"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/opportunities/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getOpportunity
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["getOpportunity"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/opportunities/{id}/revisions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * reviseOpportunity
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["reviseOpportunity"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/jobs": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listJobs
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["listJobs"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/jobs/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getJob
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["getJob"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/jobs/{id}/retry": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * retryJob
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["retryJob"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/jobs/{id}/cancel": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * cancelJob
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        post: operations["cancelJob"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/materials/{id}/revisions/{revision}/attachments/{name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read original immutable source attachment
         * @description S1 Attention service. Human writes require local session cookie and CSRF token. Agent credentials identify the caller but deny all protected data access until explicit user material scopes are configured. Reads never increase human attention.
         */
        get: operations["getMaterialAttachment"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/material-domains": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listMaterialDomains
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        get: operations["listMaterialDomains"];
        put?: never;
        /**
         * createMaterialDomain
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        post: operations["createMaterialDomain"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/material-domains/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /**
         * reviseMaterialDomain
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        patch: operations["reviseMaterialDomain"];
        trace?: never;
    };
    "/materials/{id}/domains": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * setMaterialDomains
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        put: operations["setMaterialDomains"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/project-spaces": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listProjectSpaces
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        get: operations["listProjectSpaces"];
        put?: never;
        /**
         * createProjectSpace
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        post: operations["createProjectSpace"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/project-spaces/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * getProjectSpace
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        get: operations["getProjectSpace"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/project-spaces/{id}/references": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * referenceMaterial
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        post: operations["referenceMaterial"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/project-spaces/{id}/references/remove": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * removeMaterialReference
         * @description Human-owned classification and project-space reference. Original material remains in its classification. References pin an existing material revision; they grant no Agent access or execution authority. Agent read scopes are pending and denied by default.
         */
        post: operations["removeMaterialReference"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/distillations/jobs": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Queue automatic distillation from explicitly selected fixed source snapshots
         * @description Human-triggered durable processing. Only configured authorized source keys and real adapter provenance may be sent to the local Codex processor. Native read isolation is mandatory; unavailable processing remains explicit unsupported/failed. No source database or object paths are passed to the processor.
         */
        post: operations["requestDistillation"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/attention-ranking-profile": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read explicit versioned candidate ranking profile
         * @description No configured profile means manual order; no default composite weights. Settings never grant data access or execution permission.
         */
        get: operations["getRankingProfile"];
        /**
         * Update human-owned candidate ranking weights
         * @description Human CAS+idempotent update. All weights positive; goal_progress is primary. Unknown candidate dimensions retain unknown; no delegation/approval rights are changed.
         */
        put: operations["updateRankingProfile"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/distillations/processor": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read actual processor availability and frozen configuration
         * @description Runtime facts only. Unconfigured or failed native isolation remains unavailable with required_action; no synthetic automatic capability.
         */
        get: operations["getDistillerStatus"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/knowledge/{id}/confirm": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Confirm a KnowledgeEntry for workspace reference
         * @description Human gate that moves a KnowledgeEntry from inbox to confirmed so a Project Workspace may reference it. M0 returns 501; the assembled gate (W4 store) performs exactly one status transition and no model call.
         */
        post: operations["confirmKnowledge"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/experiences/{id}/approve": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Approve a quarantined Experience
         * @description Human gate that moves an Experience from quarantine to approved so it becomes retrievable. M0 returns 501; the assembled gate (W5 store) performs exactly one status transition and no model call.
         */
        post: operations["approveExperience"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/conflicts/{id}/adjudications": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Record the human decision on a pending Conflict
         * @description Human gate that records a decision on a pending Conflict. M0 returns 501; the assembled gate (W2 store) applies the decision value and never merges evidence automatically.
         */
        post: operations["adjudicateConflict"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/knowledge/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read project-scoped knowledge
         * @description Human-session cached read with persisted project visibility checks; never starts discovery.
         */
        get: operations["readEngineknowledge"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/knowledge/{id}/shares": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Human shares knowledge
         * @description Human-session command with CSRF, caller-scoped idempotency and version checks. Service implementation owns persisted mutation and audit.
         */
        post: operations["sharesEngineknowledge"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/knowledge/{id}/demote": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Human demote knowledge
         * @description Human-session command with CSRF, caller-scoped idempotency and version checks. Service implementation owns persisted mutation and audit.
         */
        post: operations["demoteEngineknowledge"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/experiences/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Read project-scoped experiences
         * @description Human-session cached read with persisted project visibility checks; never starts discovery.
         */
        get: operations["readEngineexperiences"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/experiences/{id}/shares": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Human shares experiences
         * @description Human-session command with CSRF, caller-scoped idempotency and version checks. Service implementation owns persisted mutation and audit.
         */
        post: operations["sharesEngineexperiences"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/experiences/{id}/demote": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Human demote experiences
         * @description Human-session command with CSRF, caller-scoped idempotency and version checks. Service implementation owns persisted mutation and audit.
         */
        post: operations["demoteEngineexperiences"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/knowledge/{id}/references": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Reference confirmed knowledge
         * @description Human-session command with CSRF, caller-scoped idempotency and version checks. Service implementation owns persisted mutation and audit.
         */
        post: operations["referenceEngineKnowledge"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * listWorks
         * @description Read the persisted Works album without filesystem access.
         */
        get: operations["listWorks"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * refreshWorks
         * @description Scan only registered approved metadata roots; omitted roots use the cached registered roots. Supplied roots cannot extend that boundary.
         */
        post: operations["refreshWorks"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/delivery": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * importWorksDelivery
         * @description Human-session Works command with CSRF, idempotency and independent metadata revision.
         */
        post: operations["importWorksDelivery"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/{work_id}/attitude": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * setWorkAttitude
         * @description Human-session Works command with CSRF, idempotency and independent metadata revision.
         */
        put: operations["setWorkAttitude"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/{work_id}/group": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * setWorkGroup
         * @description Human-session Works command with CSRF, idempotency and independent metadata revision.
         */
        put: operations["setWorkGroup"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/{work_id}/handoff": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * setWorkHandoff
         * @description Human-session Works command with CSRF, idempotency and independent metadata revision.
         */
        put: operations["setWorkHandoff"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/works/{work_id}/hidden": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /**
         * setWorkHidden
         * @description Human-session Works command with CSRF, idempotency and independent metadata revision.
         */
        put: operations["setWorkHidden"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        RelaySessionCoverageV1: {
            /**
             * @description archive counts complete JSONL lines and byte extents; native_notifications counts public app-server notifications, with zero byte counters. Session completeness is bounded to the selected source protocol; internal events not exposed by it are outside that claim.
             * @enum {string}
             */
            source?: "unknown" | "archive" | "native_notifications";
            /** Format: uuid */
            task_id: string;
            /** Format: uuid */
            terminal_id: string;
            /** Format: uuid */
            worktree_id: string;
            project_id: string;
            /** @enum {string} */
            status: "unknown" | "partial" | "complete";
            gap: string;
            lines: number;
            projected: number;
            duplicates: number;
            unknown: number;
            errors: number;
            unknown_tokens: number;
            remaining_bytes: number;
            tail_bytes: number;
            snapshot_complete: boolean;
            session_complete: boolean;
        };
        /**
         * @example {
         *       "events": [
         *         {
         *           "event_id": "ae6c15f2-394b-402c-9f53-b75fc2cd0783",
         *           "kind": "edit_file",
         *           "topic": "relay:cf1497b0-e092-43c0-80a9-33449bb57a40",
         *           "tokens": null,
         *           "scope": "public-project",
         *           "ts": "2026-10-14T01:00:01Z"
         *         }
         *       ],
         *       "coverage": [
         *         {
         *           "task_id": "cf1497b0-e092-43c0-80a9-33449bb57a40",
         *           "terminal_id": "c48d6c05-3b58-4bdc-8a7b-46a471612a0b",
         *           "worktree_id": "0623c675-ddb7-4d62-bd02-831f6cb6d51d",
         *           "project_id": "public-project",
         *           "status": "partial",
         *           "gap": "prefix_unverified",
         *           "lines": 2,
         *           "projected": 2,
         *           "duplicates": 0,
         *           "unknown": 0,
         *           "errors": 0,
         *           "unknown_tokens": 2,
         *           "remaining_bytes": 0,
         *           "tail_bytes": 0,
         *           "snapshot_complete": true,
         *           "session_complete": false
         *         }
         *       ],
         *       "total": 1,
         *       "page": 1,
         *       "page_size": 50
         *     }
         */
        RelaySessionEventsV1: {
            events: components["schemas"]["WorkbenchEventV1"][];
            coverage: components["schemas"]["RelaySessionCoverageV1"][];
            total: number;
            page: number;
            page_size: number;
        };
        RelayProjectV1: {
            project_id: string;
            name: string;
            root: string;
            /** @enum {string} */
            permission_level: "A" | "B" | "C";
            project_settings_revision: number;
            allowed_subdirs: string[];
            allowed_actions: string[];
            allow_agent_control: boolean;
            allow_directory: boolean;
            expand_references: boolean;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "projects": []
         *     }
         */
        RelayProjectsV1: {
            /** @constant */
            schema_version: 1;
            projects: components["schemas"]["RelayProjectV1"][];
        };
        /**
         * @example {
         *       "name": "fixture-cli",
         *       "launch_cmd": [
         *         "fixture-cli"
         *       ],
         *       "capabilities": {},
         *       "drive_mode": "terminal"
         *     }
         * @example {
         *       "name": "fixture-cli",
         *       "launch_cmd": [
         *         "fixture-cli"
         *       ],
         *       "capabilities": {},
         *       "drive_mode": "terminal",
         *       "input_mode": "stdin_eof"
         *     }
         */
        RelayAgentV1: {
            name: string;
            launch_cmd: string[];
            capabilities: {
                [key: string]: components["schemas"]["NativeCapabilityObservationV1"];
            };
            /** @enum {string} */
            drive_mode: "structured" | "terminal" | "both";
            /**
             * @description Optional trusted registration input framing. Omitted means existing line mode. stdin_eof closes stdin after the full first input and does not support later Send. This field does not authorize launch or prove native CLI compatibility.
             * @enum {string}
             */
            input_mode?: "line" | "stdin_eof";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "agents": [],
         *       "inventory": []
         *     }
         */
        RelayAgentsV1: {
            /** @constant */
            schema_version: 1;
            agents: components["schemas"]["RelayAgentV1"][];
            inventory: components["schemas"]["LocalAgentV1"][];
        };
        RelayTaskV1: {
            task_id: string;
            work_id: string;
            project_id: string;
            worktree_id: string;
            terminal_id: string;
            agent: string;
            description: string;
            /** @enum {string} */
            state: "ready" | "accepted" | "working" | "done" | "released";
            condition: string;
            /** @enum {string} */
            permission_level: "A" | "B" | "C";
            project_settings_revision: number;
            degraded_launch: boolean;
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            updated_at: string;
        };
        RelayTerminalV1: {
            terminal_id: string;
            worktree_id: string;
            agent: string;
            /** @enum {string} */
            drive_mode: "structured" | "terminal" | "both";
            /** @enum {string} */
            liveness: "live" | "unverifiable" | "exited";
            /** Format: date-time */
            last_event_at: string | null;
            /** Format: date-time */
            exited_at: string | null;
            last_sequence: number;
            /** Format: date-time */
            created_at: string;
        };
        RelayWorkspaceV1: {
            worktree_id: string;
            project_id: string;
            root: string;
            branch: string;
            prepared: boolean;
            active_task_id: string;
            /** Format: date-time */
            created_at: string;
        };
        RelayTransitionV1: {
            task_id: string;
            /** @enum {string} */
            from: "" | "ready" | "accepted" | "working" | "done";
            /** @enum {string} */
            to: "ready" | "accepted" | "working" | "done" | "released";
            /** Format: date-time */
            ts: string;
            operator: string;
        };
        RelayTaskViewV1: {
            task: components["schemas"]["RelayTaskV1"];
            terminal: components["schemas"]["RelayTerminalV1"] | null;
            worktree: components["schemas"]["RelayWorkspaceV1"] | null;
            transitions: components["schemas"]["RelayTransitionV1"][];
            event_publication: {
                /** @enum {string} */
                status: "none" | "pending" | "published" | "error";
                pending_events: number;
                published_events: number;
            };
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "tasks": [],
         *       "dispatched_work_ids": []
         *     }
         */
        RelayTasksV1: {
            /** @constant */
            schema_version: 1;
            tasks: components["schemas"]["RelayTaskViewV1"][];
            dispatched_work_ids: string[];
        };
        RelayTaskResultV1: {
            /** @constant */
            schema_version: 1;
            view: components["schemas"]["RelayTaskViewV1"];
        };
        RelayReceiptErrorV1: {
            /** @enum {string} */
            code: "validation_failed" | "unsupported_capability" | "version_conflict" | "context_stale" | "index_stale" | "scope_denied" | "approval_required" | "approval_revoked" | "lease_lost" | "budget_exhausted" | "provider_unavailable" | "evidence_missing" | "delivery_unknown" | "not_found" | "internal_error";
            message: string;
            request_id: string;
            retryable: boolean;
            required_action: string;
        };
        RelayReceiptV1: {
            receipt_id: string;
            task_id: string;
            /** Format: uuid */
            operation_id: string;
            /** @enum {string} */
            action: "intent" | "dispatch" | "send" | "read" | "stop" | "resume" | "transition" | "observe" | "reconcile_events";
            operator: string;
            /** @enum {string} */
            status: "pending" | "accepted" | "rejected" | "delivery_unknown";
            error: components["schemas"]["RelayReceiptErrorV1"] | null;
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            finished_at: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "receipts": []
         *     }
         */
        RelayReceiptsV1: {
            /** @constant */
            schema_version: 1;
            receipts: components["schemas"]["RelayReceiptV1"][];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "events": [
         *         {
         *           "event_id": "fixture-relay-event",
         *           "kind": "edit_file",
         *           "topic": "fixture-work",
         *           "tokens": null,
         *           "scope": "fixture-project",
         *           "ts": "2026-10-14T00:00:00Z"
         *         }
         *       ]
         *     }
         */
        RelayEventsV1: {
            /** @constant */
            schema_version: 1;
            events: components["schemas"]["WorkbenchEventV1"][];
        };
        /**
         * @example {
         *       "work_id": "fixture-work",
         *       "project_id": "fixture-project",
         *       "agent": "fixture-agent",
         *       "task": "Fixture human task"
         *     }
         */
        RelayIntentRequestV1: {
            work_id: string;
            project_id: string;
            agent: string;
            task: string;
        };
        /**
         * @example {
         *       "permission_confirmed": true,
         *       "permission_level": "A",
         *       "project_settings_revision": 0
         *     }
         */
        RelayConfirmRequestV1: {
            /** @constant */
            permission_confirmed: true;
            /** @enum {string} */
            permission_level: "A" | "B" | "C";
            project_settings_revision: number;
        };
        /**
         * @example {
         *       "input": "Fixture human input"
         *     }
         */
        RelaySendRequestV1: {
            input: string;
        };
        /** @example {} */
        RelayEmptyRequestV1: Record<string, never>;
        /**
         * @example {
         *       "from": "working",
         *       "to": "done",
         *       "completion_confirmed": true
         *     }
         */
        RelayTransitionRequestV1: {
            /** @enum {string} */
            from: "ready" | "accepted" | "working" | "done";
            /** @enum {string} */
            to: "accepted" | "working" | "done" | "released";
            completion_confirmed: boolean;
        };
        RelayActionResultV1: {
            /** @constant */
            schema_version: 1;
            view: components["schemas"]["RelayTaskViewV1"];
            receipt: components["schemas"]["RelayReceiptV1"];
        };
        /**
         * @example {
         *       "event_id": "fixture-event",
         *       "kind": "agent_task",
         *       "topic": "fixture-topic",
         *       "tokens": null,
         *       "scope": "fixture-scope",
         *       "ts": "2026-10-12T16:00:00Z"
         *     }
         */
        WorkbenchEventV1: {
            event_id: string;
            kind: string;
            topic: string;
            tokens: number | null;
            scope: string;
            /** Format: date-time */
            ts: string;
        };
        WorkbenchTopicV1: {
            topic: string;
            attention: number;
            momentum: number;
            confidence: number;
        };
        WorkbenchCardV1: {
            topic: string;
            attention: number;
            momentum: number;
            confidence: number;
            attention_delta: number | null;
        };
        /**
         * @example {
         *       "status": "ok",
         *       "snapshot_at": "2026-10-13T08:00:00Z",
         *       "timezone": "Asia/Shanghai"
         *     }
         */
        WorkbenchHealthV1: {
            /** @constant */
            status: "ok";
            /** Format: date-time */
            snapshot_at: string;
            /** @constant */
            timezone: "Asia/Shanghai";
        };
        WorkbenchTodayV1: {
            /** Format: date */
            day: string;
            focus: components["schemas"]["WorkbenchCardV1"][];
            activity: components["schemas"]["WorkbenchEventV1"][];
            sleep: {
                /** Format: date */
                day: string;
                /** @enum {string} */
                status: "not_produced" | "running" | "completed";
                /** Format: date-time */
                completed_at: string | null;
            };
            daily_counts: {
                /** Format: date */
                day: string;
                count: number;
            }[];
            sources: {
                source: string;
                count: number;
            }[];
        };
        WorkbenchFocusV1: {
            topics: components["schemas"]["WorkbenchTopicV1"][];
            series: {
                topic: string;
                points: {
                    /** Format: date-time */
                    ts: string;
                    attention: number;
                }[];
            }[];
        };
        WorkbenchAgentsV1: {
            groups: {
                source: string;
                today: number;
                days7: number;
                days30: number;
                recent: components["schemas"]["WorkbenchEventV1"][];
            }[];
        };
        WorkbenchEventsV1: {
            events: components["schemas"]["WorkbenchEventV1"][];
            total: number;
            page: number;
            page_size: number;
        };
        WorkbenchEvidenceV1: {
            topic: string;
            aggregates: {
                /** Format: date */
                day: string;
                kind: string;
                count: number;
            }[];
            events: components["schemas"]["WorkbenchEventV1"][];
            total: number;
            page: number;
            page_size: number;
        };
        WorkbenchFiltersV1: {
            topics: string[];
            kinds: string[];
            sources: string[];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "error": {
         *         "code": "unsupported_capability",
         *         "message": "Not implemented in S0",
         *         "request_id": "example-request",
         *         "retryable": false,
         *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
         *       }
         *     }
         */
        ErrorV1: {
            /** @constant */
            schema_version: 1;
            error: {
                /** @enum {string} */
                code: "validation_failed" | "unsupported_capability" | "version_conflict" | "context_stale" | "index_stale" | "scope_denied" | "approval_required" | "approval_revoked" | "lease_lost" | "budget_exhausted" | "provider_unavailable" | "evidence_missing" | "delivery_unknown" | "not_found" | "internal_error";
                message: string;
                request_id: string;
                retryable: boolean;
                required_action: string;
            };
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "status": "ok",
         *       "service": "astrocyte"
         *     }
         */
        HealthV1: {
            /** @constant */
            schema_version: 1;
            /** @enum {string} */
            status: "ok";
            /** @enum {string} */
            service: "astrocyte";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "stage": "S0",
         *       "capabilities": {
         *         "imports": false,
         *         "approvals": false,
         *         "execution": false,
         *         "native_resume": false,
         *         "handoff": false
         *       },
         *       "storage": {
         *         "engine": "sqlite",
         *         "schema_version": 1
         *       },
         *       "fixture": false
         *     }
         * @example {
         *       "schema_version": 1,
         *       "stage": "S1",
         *       "capabilities": {
         *         "imports": true,
         *         "approvals": false,
         *         "execution": false,
         *         "native_resume": false,
         *         "handoff": false
         *       },
         *       "storage": {
         *         "engine": "sqlite",
         *         "schema_version": 5
         *       },
         *       "fixture": false
         *     }
         */
        FoundationV1: {
            /** @constant */
            schema_version: 1;
            /** @enum {string} */
            stage: "S0" | "S1";
            /** @description Availability of assembled HTTP operations, not evidence about installed local agent native abilities or completion of S1 acceptance. */
            capabilities: {
                /** @description True when the S1 material import service is assembled. */
                imports: boolean;
                /** @constant */
                approvals: false;
                /** @constant */
                execution: false;
                /** @constant */
                native_resume: false;
                /** @constant */
                handoff: false;
            };
            storage: {
                /** @enum {string} */
                engine: "sqlite";
                schema_version: number;
            };
            /** @constant */
            fixture: false;
        };
        /**
         * @description Independent readiness observation; no credential inspection is implied.
         * @example {
         *       "status": "unknown",
         *       "reason": "configuration_not_inspected",
         *       "checked_at": null
         *     }
         */
        AgentObservationV1: {
            /** @enum {string} */
            status: "available" | "unavailable" | "unknown" | "unsupported";
            reason: string;
            /** Format: date-time */
            checked_at: string | null;
        };
        /**
         * @description SPEC 8.1 native capability; CLI help or installation is insufficient evidence of support.
         * @example {
         *       "status": "unknown",
         *       "reason": "native_runtime_not_tested",
         *       "checked_at": null
         *     }
         */
        NativeCapabilityObservationV1: {
            /** @enum {string} */
            status: "supported" | "unsupported" | "unknown";
            reason: string;
            /** Format: date-time */
            checked_at: string | null;
        };
        /**
         * @description CLI installation identity, independent of any private native session, project or Mission.
         * @example {
         *       "id": "example-cli",
         *       "display_name": "Example CLI (schema example only)",
         *       "version": null,
         *       "installed": {
         *         "status": "unknown",
         *         "reason": "cli_not_probed",
         *         "checked_at": null
         *       },
         *       "configured": {
         *         "status": "unknown",
         *         "reason": "configuration_not_inspected",
         *         "checked_at": null
         *       },
         *       "startable": {
         *         "status": "unknown",
         *         "reason": "native_runtime_not_tested",
         *         "checked_at": null
         *       },
         *       "capabilities": {
         *         "discover": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "read_context": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "start": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "resume": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "send": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "stop": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "observe": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         },
         *         "reconcile": {
         *           "status": "unknown",
         *           "reason": "native_runtime_not_tested",
         *           "checked_at": null
         *         }
         *       }
         *     }
         */
        LocalAgentV1: {
            /** @description Stable CLI adapter identity, not executable path or session ID. */
            id: string;
            display_name: string;
            /** @description Observed version only; null if not observed. */
            version: string | null;
            installed: components["schemas"]["AgentObservationV1"];
            configured: components["schemas"]["AgentObservationV1"];
            startable: components["schemas"]["AgentObservationV1"];
            capabilities: {
                discover: components["schemas"]["NativeCapabilityObservationV1"];
                read_context: components["schemas"]["NativeCapabilityObservationV1"];
                start: components["schemas"]["NativeCapabilityObservationV1"];
                resume: components["schemas"]["NativeCapabilityObservationV1"];
                send: components["schemas"]["NativeCapabilityObservationV1"];
                stop: components["schemas"]["NativeCapabilityObservationV1"];
                observe: components["schemas"]["NativeCapabilityObservationV1"];
                reconcile: components["schemas"]["NativeCapabilityObservationV1"];
            };
            /** @description Backend registry implements this CLI protocol. This is independent of installed/configured/startable and every observed native capability; a human may request explicit verification. */
            native_adapter_registered?: boolean;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        LocalAgentListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["LocalAgentV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "material_id": "fixture-material",
         *       "revision": 1,
         *       "locator": "arxiv:fixture",
         *       "span": null
         *     }
         */
        SourceRefV1: {
            material_id: string;
            revision: number;
            locator: string;
            span?: string | null;
        };
        /**
         * @example {
         *       "value": null,
         *       "reason": "unknown"
         *     }
         */
        DimensionScoreV1: {
            value: number | null;
            reason: string;
        };
        /**
         * @example {
         *       "goal_progress": {
         *         "value": null,
         *         "reason": "unknown"
         *       },
         *       "current_interest": {
         *         "value": null,
         *         "reason": "unknown"
         *       },
         *       "project_improvement": {
         *         "value": null,
         *         "reason": "unknown"
         *       },
         *       "originality": {
         *         "value": null,
         *         "reason": "unknown"
         *       }
         *     }
         */
        DimensionsV1: {
            goal_progress: components["schemas"]["DimensionScoreV1"];
            current_interest: components["schemas"]["DimensionScoreV1"];
            project_improvement: components["schemas"]["DimensionScoreV1"];
            originality: components["schemas"]["DimensionScoreV1"];
        };
        /**
         * @example {
         *       "id": "fixture-material",
         *       "source_locator": "arxiv:fixture",
         *       "kind": "paper",
         *       "current_revision": 1,
         *       "lifecycle": "active",
         *       "title": "Fixture only",
         *       "collection_reason": null
         *     }
         */
        MaterialV1: {
            id: string;
            source_locator: string;
            /** @enum {string} */
            kind: "paper" | "video" | "text" | "file";
            current_revision: number;
            /** @enum {string} */
            lifecycle: "active" | "archived" | "withdrawn";
            title?: string;
            collection_reason?: string | null;
            source_spans?: string[];
            /** @enum {string} */
            import_status?: "queued" | "running" | "succeeded" | "failed" | "cancelled";
            human_usage_count?: number;
            agent_usage_count?: number;
            attention_score?: number;
            long_term_value?: number;
            /** Format: date-time */
            created_at?: string;
            version?: number;
            pinned?: boolean;
            domain_ids?: string[];
            ranking_strategy?: string;
            ranking_reason?: string;
            attention_half_life_seconds?: number;
            attention_weights?: {
                [key: string]: number;
            };
        };
        /**
         * @example {
         *       "id": "fixture-opportunity",
         *       "revision": 1,
         *       "state": "incubating",
         *       "evidence_refs": [],
         *       "dimensions": {
         *         "goal_progress": {
         *           "value": null,
         *           "reason": "unknown"
         *         },
         *         "current_interest": {
         *           "value": null,
         *           "reason": "unknown"
         *         },
         *         "project_improvement": {
         *           "value": null,
         *           "reason": "unknown"
         *         },
         *         "originality": {
         *           "value": null,
         *           "reason": "unknown"
         *         }
         *       },
         *       "next_step": "Select a real case"
         *     }
         */
        OpportunityV1: {
            id: string;
            revision: number;
            /** @enum {string} */
            state: "incubating" | "ready_for_review" | "admitted" | "deferred" | "rejected" | "withdrawn";
            title?: string;
            evidence_refs: components["schemas"]["SourceRefV1"][];
            goal_refs?: string[];
            dimensions: components["schemas"]["DimensionsV1"];
            next_step: string;
            missing_evidence?: string[];
            purpose?: string;
            distillation_ids?: string[];
            /** Format: date-time */
            created_at?: string;
            version?: number;
            composite_score?: number | null;
            ranking_strategy?: string;
            ranking_reason?: string;
            ranking_profile_version?: number | null;
        };
        /**
         * @example {
         *       "id": "fixture-project",
         *       "name": "Fixture only",
         *       "environment_id": "fixture",
         *       "root_path": "/fixture"
         *     }
         */
        ProjectV1: {
            id: string;
            name: string;
            environment_id: string;
            root_path: string;
            repo_id?: string | null;
            worktree_id?: string | null;
        };
        /**
         * @example {
         *       "calls": null,
         *       "tokens": null,
         *       "money": null,
         *       "currency": null
         *     }
         */
        BudgetV1: {
            calls: number | null;
            tokens: number | null;
            money: number | null;
            currency: string | null;
        };
        /**
         * @example {
         *       "project_ids": [],
         *       "allowed_roots": [],
         *       "allowed_actions": [],
         *       "data_egress_rules": []
         *     }
         */
        ScopeV1: {
            project_ids: string[];
            allowed_roots: string[];
            allowed_actions: string[];
            data_egress_rules: string[];
        };
        /**
         * @example {
         *       "id": "fixture-proposal",
         *       "revision": 1,
         *       "status": "draft",
         *       "title": "Fixture only",
         *       "goal": "Select real case",
         *       "scope": {
         *         "project_ids": [],
         *         "allowed_roots": [],
         *         "allowed_actions": [],
         *         "data_egress_rules": []
         *       },
         *       "deliverables": [],
         *       "budget": {
         *         "calls": null,
         *         "tokens": null,
         *         "money": null,
         *         "currency": null
         *       },
         *       "stop_conditions": []
         *     }
         */
        ProposalV1: {
            id: string;
            revision: number;
            /** @enum {string} */
            status: "draft" | "in_review" | "approved" | "declined" | "superseded";
            title: string;
            goal: string;
            scope: components["schemas"]["ScopeV1"];
            deliverables: string[];
            budget: components["schemas"]["BudgetV1"];
            stop_conditions: string[];
            candidate_refs?: components["schemas"]["SourceRefV1"][];
        };
        /**
         * @example {
         *       "id": "fixture-session",
         *       "adapter": "fixture",
         *       "native_session_id": null,
         *       "project_id": "fixture-project",
         *       "binding_status": "unavailable",
         *       "capabilities": {
         *         "native_resume": false,
         *         "handoff": false
         *       },
         *       "context_state": "unknown"
         *     }
         */
        SessionV1: {
            id: string;
            adapter: string;
            native_session_id: string | null;
            project_id: string;
            worktree_id?: string | null;
            /** @enum {string} */
            binding_status: "observed" | "bound" | "unavailable";
            capabilities: {
                /** @description Null means unprobed or unknown; false means verified unsupported. */
                native_resume: boolean | null;
                /** @description Null means unprobed or unknown; false means verified unsupported. */
                handoff: boolean | null;
            };
            context_packet_id?: string | null;
            /** @enum {string} */
            context_state: "unknown" | "current" | "stale";
        };
        /**
         * @example {
         *       "id": "fixture-mission",
         *       "version": 1,
         *       "grant_id": "fixture-grant",
         *       "goal": "Fixture only",
         *       "status": "pending"
         *     }
         */
        MissionV1: {
            id: string;
            version: number;
            grant_id: string;
            goal: string;
            /** @enum {string} */
            status: "pending" | "running" | "paused" | "blocked" | "completed" | "cancelled" | "failed";
            work_items?: components["schemas"]["WorkItemV1"][];
            artifact_ids?: string[];
            blockers?: string[];
        };
        /**
         * @example {
         *       "id": "fixture-work",
         *       "version": 1,
         *       "status": "open",
         *       "holder_id": null,
         *       "context_packet_id": null,
         *       "artifact_ids": []
         *     }
         */
        WorkItemV1: {
            id: string;
            version: number;
            /** @enum {string} */
            status: "open" | "claimed" | "in_progress" | "submitted" | "revision_needed" | "blocked" | "accepted" | "cancelled";
            holder_id: string | null;
            context_packet_id: string | null;
            artifact_ids: string[];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        MaterialListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["MaterialV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        OpportunityListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["OpportunityV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        ProjectListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["ProjectV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        ProposalListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["ProposalV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        SessionListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["SessionV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [],
         *       "next_cursor": null
         *     }
         */
        MissionListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["MissionV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "mission": {
         *         "id": "fixture-mission",
         *         "version": 1,
         *         "grant_id": "fixture-grant",
         *         "goal": "Fixture only",
         *         "status": "pending"
         *       }
         *     }
         */
        MissionResponseV1: {
            /** @constant */
            schema_version: 1;
            mission: components["schemas"]["MissionV1"];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "operation_id": "fixture-operation",
         *       "status": "queued"
         *     }
         */
        OperationV1: {
            /** @constant */
            schema_version: 1;
            operation_id: string;
            /** @enum {string} */
            status: "queued" | "running" | "succeeded" | "failed" | "cancelled" | "delivery_unknown";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "job_id": "fixture-job",
         *       "status": "queued"
         *     }
         */
        ImportJobV1: {
            /** @constant */
            schema_version: 1;
            job_id: string;
            /** @enum {string} */
            status: "queued" | "running" | "succeeded" | "failed" | "cancelled";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "id": "fixture-resource",
         *       "version": 1
         *     }
         */
        VersionResultV1: {
            /** @constant */
            schema_version: 1;
            id: string;
            version: number;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "grant_id": "fixture-grant",
         *       "proposal_id": "fixture-proposal",
         *       "proposal_revision": 1,
         *       "epoch": 1,
         *       "status": "active"
         *     }
         */
        ApprovalV1: {
            /** @constant */
            schema_version: 1;
            grant_id: string;
            proposal_id: string;
            proposal_revision: number;
            epoch: number;
            /** @enum {string} */
            status: "active" | "revoked" | "expired" | "completed";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "work_item_id": "fixture-work",
         *       "attempt_id": "fixture-attempt",
         *       "holder_id": "fixture-holder",
         *       "fence": 1,
         *       "lease_token": "fixture-not-valid",
         *       "expires_at": "2026-10-09T00:00:00Z"
         *     }
         */
        LeaseV1: {
            /** @constant */
            schema_version: 1;
            work_item_id: string;
            attempt_id: string;
            holder_id: string;
            fence: number;
            lease_token: string;
            /** Format: date-time */
            expires_at: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "artifact_id": "fixture-artifact",
         *       "status": "submitted"
         *     }
         */
        ArtifactResultV1: {
            /** @constant */
            schema_version: 1;
            artifact_id: string;
            /** @enum {string} */
            status: "submitted" | "revision_needed" | "blocked";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "acceptance_id": "fixture-acceptance",
         *       "artifact_id": "fixture-artifact",
         *       "experience_id": "fixture-experience"
         *     }
         */
        AcceptanceV1: {
            /** @constant */
            schema_version: 1;
            acceptance_id: string;
            artifact_id: string;
            experience_id: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "source_locator": "arxiv:fixture",
         *       "source_key": "fixture-source",
         *       "kind": "paper",
         *       "content_digest": "fixture-digest",
         *       "export_text": "Fixture only"
         *     }
         */
        ImportMaterialRequestV1: {
            /** @description Ordinary import reuses the existing extraction; only explicit human refresh obtains new content. */
            refresh?: boolean;
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            source_locator: string;
            /** @description Empty string requests adapter canonical source key; nonempty keys are verified by the adapter. */
            source_key: string;
            /** @enum {string} */
            kind: "paper" | "video" | "text" | "file";
            /** @description Empty string requests server digest computation; a nonempty digest is verified against actual imported bytes. */
            content_digest: string;
            export_text?: string;
            local_file_ref?: string;
            collection_reason?: string | null;
            source_spans?: string[];
            title?: string;
            /**
             * @description summarize_url extracts a selected public video URL without export_text or local_file_ref. summarize, summarize_json and summarize_markdown import existing exports. paper_url extracts a selected public paper HTML body; paper_pdf extracts a selected public paper PDF; paper_snapshot ingests a human-reviewed browser-plugin snapshot JSON supplied as export_text (no network or model call). Extraction does not imply model distillation or Agent authorization.
             * @enum {string}
             */
            adapter?: "arxiv" | "paper_url" | "paper_pdf" | "paper_snapshot" | "summarize_url" | "summarize" | "summarize_json" | "summarize_markdown" | "manual";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "feedback": "later",
         *       "reason": "Fixture only"
         *     }
         */
        ReviewOpportunityRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            feedback: "adopt" | "later" | "reject" | "revise" | "already_solved";
            reason: string;
            dimensions?: components["schemas"]["DimensionsV1"];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "reason": "Fixture only",
         *       "policy_version": null
         *     }
         */
        AdmitOpportunityRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            reason: string;
            policy_version: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "name": "Fixture only",
         *       "environment_id": "fixture",
         *       "root_path": "/fixture"
         *     }
         */
        RegisterProjectRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            name: string;
            environment_id: string;
            root_path: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "context_packet_id": null
         *     }
         */
        ResumeSessionRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            context_packet_id: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "context_packet_id": "fixture-packet",
         *       "target_adapter": "fixture",
         *       "target_project_id": "fixture-project"
         *     }
         */
        HandoffSessionRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            context_packet_id: string;
            target_adapter: string;
            target_project_id: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1
         *     }
         */
        SubmitProposalRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "scope": {
         *         "project_ids": [],
         *         "allowed_roots": [],
         *         "allowed_actions": [],
         *         "data_egress_rules": []
         *       },
         *       "budget": {
         *         "calls": null,
         *         "tokens": null,
         *         "money": null,
         *         "currency": null
         *       },
         *       "stop_conditions": [
         *         "Fixture only"
         *       ],
         *       "expires_at": "2026-10-09T00:00:00Z"
         *     }
         */
        ApproveProposalRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            scope: components["schemas"]["ScopeV1"];
            budget: components["schemas"]["BudgetV1"];
            stop_conditions: string[];
            /** Format: date-time */
            expires_at: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "reason": "Fixture only"
         *     }
         */
        RevokeApprovalRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            reason: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "reason": "Fixture only"
         *     }
         */
        MissionControlRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            reason: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "context_packet_id": "fixture-packet"
         *     }
         */
        ClaimWorkItemRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            context_packet_id: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "attempt_id": "fixture-attempt",
         *       "lease_token": "fixture-not-valid",
         *       "fence": 1,
         *       "object_ref": "fixture-object",
         *       "content_digest": "fixture-digest",
         *       "snapshot_ref": "fixture-snapshot",
         *       "checks_ref": "fixture-checks"
         *     }
         */
        SubmitArtifactRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            attempt_id: string;
            lease_token: string;
            fence: number;
            object_ref: string;
            content_digest: string;
            snapshot_ref: string;
            checks_ref: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "reason": "Fixture only"
         *     }
         */
        AcceptArtifactRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            reason: string;
        };
        /**
         * @description Versioned SSE data envelope. Known SPEC12 payloads are discriminated by context-prefixed type; actor is injected by trusted publisher. S0 does not stream events.
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "attention.opportunity_admitted",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "opportunity_id": "fixture-opportunity",
         *         "revision": 1,
         *         "admission_id": "fixture-admission",
         *         "actor": "fixture-user",
         *         "policy_version": "fixture-policy"
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "workspace.proposal_approved",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "proposal_id": "fixture-proposal",
         *         "revision": 1,
         *         "grant_id": "fixture-grant",
         *         "epoch": 1
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "workspace.approval_revoked",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "grant_id": "fixture-grant",
         *         "new_epoch": 2,
         *         "reason": "Fixture only"
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "workspace.context_packet_built",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "packet_id": "fixture-packet",
         *         "snapshot_digest": "fixture-snapshot",
         *         "predecessor_id": null
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "swarm.work_item_claimed",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "work_item_id": "fixture-work",
         *         "attempt_id": "fixture-attempt",
         *         "holder_id": "fixture-holder",
         *         "fence": 1
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "swarm.artifact_submitted",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "artifact_id": "fixture-artifact",
         *         "attempt_id": "fixture-attempt",
         *         "snapshot": {
         *           "environment_id": "fixture",
         *           "repo_id": "fixture-repo",
         *           "worktree_id": "fixture-worktree",
         *           "head": "fixture-head",
         *           "index_digest": "fixture-index",
         *           "working_tree_digest": "fixture-content"
         *         },
         *         "checks_ref": "fixture-checks"
         *       }
         *     }
         * @example {
         *       "event_id": "fixture-event",
         *       "type": "swarm.artifact_accepted",
         *       "schema_version": 1,
         *       "aggregate_id": "fixture-opportunity",
         *       "aggregate_version": 1,
         *       "occurred_at": "2026-10-09T00:00:00Z",
         *       "correlation_id": "fixture-request",
         *       "causation_id": null,
         *       "payload": {
         *         "artifact_id": "fixture-artifact",
         *         "acceptance_id": "fixture-acceptance",
         *         "actor": "fixture-user",
         *         "experience_id": "fixture-experience"
         *       }
         *     }
         */
        EventV1: {
            event_id: string;
            type: string;
            /** @constant */
            schema_version: 1;
            aggregate_id: string;
            aggregate_version: number;
            /** Format: date-time */
            occurred_at: string;
            correlation_id: string;
            causation_id: string | null;
            payload: {
                [key: string]: unknown;
            };
        } & ({
            /** @constant */
            type: "attention.opportunity_admitted";
            payload: components["schemas"]["OpportunityAdmittedPayloadV1"];
        } | {
            /** @constant */
            type: "workspace.proposal_approved";
            payload: components["schemas"]["ProposalApprovedPayloadV1"];
        } | {
            /** @constant */
            type: "workspace.approval_revoked";
            payload: components["schemas"]["ApprovalRevokedPayloadV1"];
        } | {
            /** @constant */
            type: "workspace.context_packet_built";
            payload: components["schemas"]["ContextPacketBuiltPayloadV1"];
        } | {
            /** @constant */
            type: "swarm.work_item_claimed";
            payload: components["schemas"]["WorkItemClaimedPayloadV1"];
        } | {
            /** @constant */
            type: "swarm.artifact_submitted";
            payload: components["schemas"]["ArtifactSubmittedPayloadV1"];
        } | {
            /** @constant */
            type: "swarm.artifact_accepted";
            payload: components["schemas"]["ArtifactAcceptedPayloadV1"];
        });
        /**
         * @example {
         *       "environment_id": "fixture",
         *       "repo_id": "fixture-repo",
         *       "worktree_id": "fixture-worktree",
         *       "head": "fixture-head",
         *       "index_digest": "fixture-index",
         *       "working_tree_digest": "fixture-content"
         *     }
         */
        ProjectSnapshotV1: {
            environment_id: string;
            repo_id: string;
            worktree_id: string;
            head: string;
            index_digest: string;
            working_tree_digest: string;
        };
        /**
         * @example {
         *       "opportunity_id": "fixture-opportunity",
         *       "revision": 1,
         *       "admission_id": "fixture-admission",
         *       "actor": "fixture-user",
         *       "policy_version": "fixture-policy"
         *     }
         */
        OpportunityAdmittedPayloadV1: {
            opportunity_id: string;
            revision: number;
            admission_id: string;
            actor: string;
            policy_version: string;
        };
        /**
         * @example {
         *       "proposal_id": "fixture-proposal",
         *       "revision": 1,
         *       "grant_id": "fixture-grant",
         *       "epoch": 1
         *     }
         */
        ProposalApprovedPayloadV1: {
            proposal_id: string;
            revision: number;
            grant_id: string;
            epoch: number;
        };
        /**
         * @example {
         *       "grant_id": "fixture-grant",
         *       "new_epoch": 2,
         *       "reason": "Fixture only"
         *     }
         */
        ApprovalRevokedPayloadV1: {
            grant_id: string;
            new_epoch: number;
            reason: string;
        };
        /**
         * @example {
         *       "packet_id": "fixture-packet",
         *       "snapshot_digest": "fixture-snapshot",
         *       "predecessor_id": null
         *     }
         */
        ContextPacketBuiltPayloadV1: {
            packet_id: string;
            snapshot_digest: string;
            predecessor_id: string | null;
        };
        /**
         * @example {
         *       "work_item_id": "fixture-work",
         *       "attempt_id": "fixture-attempt",
         *       "holder_id": "fixture-holder",
         *       "fence": 1
         *     }
         */
        WorkItemClaimedPayloadV1: {
            work_item_id: string;
            attempt_id: string;
            holder_id: string;
            fence: number;
        };
        /**
         * @example {
         *       "artifact_id": "fixture-artifact",
         *       "attempt_id": "fixture-attempt",
         *       "snapshot": {
         *         "environment_id": "fixture",
         *         "repo_id": "fixture-repo",
         *         "worktree_id": "fixture-worktree",
         *         "head": "fixture-head",
         *         "index_digest": "fixture-index",
         *         "working_tree_digest": "fixture-content"
         *       },
         *       "checks_ref": "fixture-checks"
         *     }
         */
        ArtifactSubmittedPayloadV1: {
            artifact_id: string;
            attempt_id: string;
            snapshot: components["schemas"]["ProjectSnapshotV1"];
            checks_ref: string;
        };
        /**
         * @example {
         *       "artifact_id": "fixture-artifact",
         *       "acceptance_id": "fixture-acceptance",
         *       "actor": "fixture-user",
         *       "experience_id": "fixture-experience"
         *     }
         */
        ArtifactAcceptedPayloadV1: {
            artifact_id: string;
            acceptance_id: string;
            actor: string;
            experience_id: string;
        };
        ProvenanceV1: {
            processor: string;
            version: string;
            /** @description Actual processing mode reported by the adapter; manual, official_atom_and_pdf, original_export or summary_only. Never infer automatic processing from a record. */
            mode: string;
            source: string;
            /** @description Actual model reported by the processor, or explicit unknown; never inferred from client name. */
            model?: string;
        };
        MaterialRevisionV1: {
            material_id: string;
            revision: number;
            source_key: string;
            source_locator: string;
            content_digest: string;
            object_ref: string;
            source_spans: string[];
            provenance: components["schemas"]["ProvenanceV1"];
            /** Format: date-time */
            created_at: string;
            summary?: string;
            attachments?: components["schemas"]["AttachmentRefV1"][];
        };
        UsageV1: {
            id: string;
            material_id: string;
            actor_id: string;
            /** @enum {string} */
            actor_kind: "human" | "agent";
            action: string;
            /** Format: date-time */
            occurred_at: string;
        };
        DistillationV1: {
            /** @description Explicitly selected immutable prior records used by this processing operation. */
            prior_distillation_ids?: string[];
            id: string;
            input_refs: components["schemas"]["SourceRefV1"][];
            /** @enum {string} */
            stage: "content" | "topic" | "project";
            output_ref: string;
            output_text: string;
            /** @enum {string} */
            status: "succeeded";
            next_question: string | null;
            question: string;
            processing_config: string;
            related_refs: components["schemas"]["SourceRefV1"][];
            related_ideas: string[];
            conflicts: string[];
            pending_questions: string[];
            goal_refs: string[];
            existing_assets: string[];
            expected_improvement: string;
            minimum_artifact: string;
            missing_evidence: string[];
            provenance: components["schemas"]["ProvenanceV1"];
            reuse_key: string;
            /** Format: date-time */
            created_at: string;
            candidate_suggestion?: components["schemas"]["CandidateSuggestionV1"] | null;
        };
        RecordDistillationRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            input_refs: components["schemas"]["SourceRefV1"][];
            /** @enum {string} */
            stage: "content" | "topic" | "project";
            output_text: string;
            next_question?: string | null;
            question: string;
            processing_config: string;
            related_refs?: components["schemas"]["SourceRefV1"][];
            related_ideas?: string[];
            conflicts?: string[];
            pending_questions?: string[];
            goal_refs?: string[];
            existing_assets?: string[];
            expected_improvement?: string;
            minimum_artifact?: string;
            missing_evidence?: string[];
        };
        DistillationResultV1: {
            /** @constant */
            schema_version: 1;
            distillation: components["schemas"]["DistillationV1"];
            reused: boolean;
        };
        MaterialDetailV1: {
            /** @constant */
            schema_version: 1;
            material: components["schemas"]["MaterialV1"];
            revisions: components["schemas"]["MaterialRevisionV1"][];
            distillations: components["schemas"]["DistillationV1"][];
            uses: components["schemas"]["UsageV1"][];
        };
        ContentV1: {
            /** @constant */
            schema_version: 1;
            material_id: string;
            revision: number;
            content_digest: string;
            text: string;
            provenance: components["schemas"]["ProvenanceV1"];
        };
        ReviewV1: {
            id: string;
            opportunity_id: string;
            revision: number;
            /** @enum {string} */
            feedback: "adopt" | "later" | "reject" | "revise" | "already_solved";
            reason: string;
            dimensions?: components["schemas"]["DimensionsV1"];
            actor_id: string;
            /** Format: date-time */
            created_at: string;
        };
        OpportunityDetailV1: {
            /** @constant */
            schema_version: 1;
            opportunity: components["schemas"]["OpportunityV1"];
            revisions: components["schemas"]["OpportunityV1"][];
            reviews: components["schemas"]["ReviewV1"][];
        };
        OpportunityRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            title: string;
            evidence_refs: components["schemas"]["SourceRefV1"][];
            goal_refs?: string[];
            dimensions: components["schemas"]["DimensionsV1"];
            next_step: string;
            missing_evidence?: string[];
            purpose: string;
            distillation_ids: string[];
        };
        RecordUseRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            action: "reread" | "annotate" | "pin" | "adopt" | "mention" | "project_reuse";
        };
        JobV1: {
            /** @constant */
            schema_version: 1;
            job_id: string;
            /** @enum {string} */
            status: "queued" | "running" | "succeeded" | "failed" | "cancelled";
            kind: string;
            version: number;
            dedupe_key: string;
            operation_id: string;
            material_id: string | null;
            material_revision: number | null;
            error: components["schemas"]["ServiceErrorV1"] | null;
            attempts: number;
            max_attempts: number;
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            updated_at: string;
            /** Format: date-time */
            deadline_at: string;
            cancel_requested: boolean;
            external_started: boolean;
            delivery_unknown: boolean;
            distillation_id?: string | null;
        };
        ServiceErrorV1: {
            /** @enum {string} */
            code: "validation_failed" | "unsupported_capability" | "version_conflict" | "context_stale" | "index_stale" | "scope_denied" | "approval_required" | "approval_revoked" | "lease_lost" | "budget_exhausted" | "provider_unavailable" | "evidence_missing" | "delivery_unknown" | "not_found" | "internal_error";
            message: string;
            request_id: string;
            retryable: boolean;
            required_action: string;
        };
        JobCommandV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
        };
        LocalSessionV1: {
            /** @constant */
            schema_version: 1;
            actor_id: string;
            /** @enum {string} */
            actor_kind: "human";
            csrf_token: string;
        };
        JobListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["JobV1"][];
            next_cursor: string | null;
        };
        DistillationListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["DistillationV1"][];
            next_cursor: string | null;
        };
        AttachmentRefV1: {
            name: string;
            media_type: string;
            object_ref: string;
            source_locator: string;
        };
        UpdateMaterialRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            lifecycle: "active" | "archived" | "withdrawn";
            collection_reason: string | null;
            pinned: boolean;
        };
        MaterialDomainV1: {
            id: string;
            version: number;
            title: string;
            description: string;
            /** Format: date-time */
            created_at: string;
        };
        MaterialDomainRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            title: string;
            description: string;
        };
        MaterialDomainResultV1: {
            /** @constant */
            schema_version: 1;
            domain: components["schemas"]["MaterialDomainV1"];
        };
        SetMaterialDomainsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            domain_ids: string[];
        };
        ProjectSpaceV1: {
            id: string;
            version: number;
            title: string;
            material_refs: components["schemas"]["SourceRefV1"][];
            /** Format: date-time */
            created_at: string;
        };
        ProjectSpaceRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            title: string;
        };
        ProjectSpaceResultV1: {
            /** @constant */
            schema_version: 1;
            space: components["schemas"]["ProjectSpaceV1"];
        };
        ReferenceMaterialRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            material_id: string;
            revision: number;
        };
        RemoveMaterialReferenceRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            material_id: string;
        };
        MaterialDomainListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["MaterialDomainV1"][];
            next_cursor: string | null;
        };
        ProjectSpaceListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["ProjectSpaceV1"][];
            next_cursor: string | null;
        };
        RequestDistillationRequestV1: {
            /** @description Human selected project whose current fixed references and model consent are rechecked before processing. */
            project_id?: string;
            /** @description Exact CLI permitted by that project; selected project processing never silently falls back to a global provider. */
            cli?: string;
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            input_refs: components["schemas"]["SourceRefV1"][];
            /** @enum {string} */
            stage: "content" | "topic" | "project";
            question: string;
            processing_config: string;
            prior_distillation_ids?: string[];
        };
        RankingProfileV1: {
            id: string;
            version: number;
            enabled: boolean;
            weights: {
                goal_progress: number;
                current_interest: number;
                project_improvement: number;
                originality: number;
            };
            /** Format: date-time */
            created_at: string;
        };
        RankingProfileDetailV1: {
            /** @constant */
            schema_version: 1;
            configured: boolean;
            profile: components["schemas"]["RankingProfileV1"] | null;
            versions: components["schemas"]["RankingProfileV1"][];
        };
        UpdateRankingProfileRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            enabled: boolean;
            weights: {
                goal_progress: number;
                current_interest: number;
                project_improvement: number;
                originality: number;
            };
        };
        CandidateSuggestionV1: {
            title: string;
            evidence_refs: components["schemas"]["SourceRefV1"][];
            goal_refs: string[];
            dimensions: components["schemas"]["DimensionsV1"];
            next_step: string;
            missing_evidence: string[];
            purpose: string;
        };
        DistillerStatusV1: {
            /** @constant */
            schema_version: 1;
            available: boolean;
            processor: string;
            configuration_id: string | null;
            model: string | null;
            reason: string;
            required_action: string;
            allowed_source_keys: string[];
        };
        ListingMetadataV1: {
            external_id: string;
            locator: string;
            title: string;
            description: string;
            author: string;
            cover: string;
            published_at: number;
            provider_status: number | null;
            unavailable_reason: string;
        };
        SourceRecommendationV1: {
            /** @enum {string} */
            status: "pending" | "running" | "succeeded" | "failed" | "unknown" | "unavailable";
            text: string;
            reason: string;
            metadata_revision: number;
            provenance: components["schemas"]["ProvenanceV1"];
            configuration_id: string;
            error: components["schemas"]["ServiceErrorV1"] | null;
        };
        TrackingSourceV1: {
            /**
             * @description Omitted legacy input means public; browser_selected requires explicit human selected owner and folder identity and configured browser bridge.
             * @enum {string}
             */
            access_mode?: "public" | "browser_selected";
            id: string;
            version: number;
            /** @enum {string} */
            platform: "bilibili" | "douyin";
            /** @enum {string} */
            source_kind: "uploads" | "favorites";
            external_id: string;
            owner_id: string;
            locator: string;
            title: string;
            status: string;
            last_success_at: string | null;
            last_error: components["schemas"]["ServiceErrorV1"] | null;
            next_cursor: string | null;
            has_more: boolean;
            warnings: string[];
        };
        SourceItemV1: {
            source_id: string;
            external_id: string;
            revision: number;
            metadata: components["schemas"]["ListingMetadataV1"];
            stale: boolean;
            selected: boolean;
            import_job_id: string | null;
            material_id: string | null;
            recommendation: components["schemas"]["SourceRecommendationV1"] | null;
        };
        TrackingSourceResultV1: {
            /** @constant */
            schema_version: 1;
            source: components["schemas"]["TrackingSourceV1"];
            items: components["schemas"]["SourceItemV1"][];
            next_cursor: string | null;
            has_more: boolean;
            warnings: string[];
            jobs: components["schemas"]["ImportJobV1"][];
        };
        TrackingSourceListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["TrackingSourceV1"][];
            next_cursor: string | null;
        };
        BindTrackingSourceRequestV1: {
            /**
             * @description Omitted legacy input means public; browser_selected requires explicit human selected owner and folder identity and configured browser bridge.
             * @enum {string}
             */
            access_mode?: "public" | "browser_selected";
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            platform: "bilibili" | "douyin";
            /** @enum {string} */
            source_kind: "uploads" | "favorites";
            external_id: string;
            owner_id?: string;
            locator: string;
            title?: string;
        };
        SyncTrackingSourceRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @default 100 */
            limit: number;
            cursor?: string;
        };
        SourceItemSelectionV1: {
            external_id: string;
            revision: number;
        };
        RecommendSourceItemsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            project_id: string;
            cli: string;
            items: components["schemas"]["SourceItemSelectionV1"][];
        };
        SelectSourceItemsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            items: components["schemas"]["SourceItemSelectionV1"][];
            collection_reason?: string | null;
        };
        LocalProjectSettingsV1: {
            revision: number;
            allow_directory: boolean;
            allowed_subdirs: string[];
            expand_references: boolean;
            allowed_actions: string[];
            allowed_tools: string[];
            external_model_cli: string;
            allow_agent_control: boolean;
            history_roots: {
                [key: string]: string;
            };
        };
        LocalProjectV1: {
            id: string;
            name: string;
            root: string;
            space_id: string;
            settings: components["schemas"]["LocalProjectSettingsV1"];
            /** Format: date-time */
            created_at: string;
        };
        LocalProjectResultV1: {
            /** @constant */
            schema_version: 1;
            project: components["schemas"]["LocalProjectV1"];
        };
        LocalProjectListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["LocalProjectV1"][];
            next_cursor: string | null;
        };
        LocalProjectCandidateV1: {
            root: string;
            name: string;
        };
        LocalProjectCandidateListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["LocalProjectCandidateV1"][];
            next_cursor: string | null;
        };
        RegisterLocalProjectRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            name: string;
            root: string;
            space_id: string;
        };
        DiscoverLocalProjectsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            root: string;
        };
        LocalProjectSettingsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            expected_revision: number;
            settings: components["schemas"]["LocalProjectSettingsV1"];
        };
        LocalProjectGrantV1: {
            project_id: string;
            agent_id: string;
            actions: string[];
            revoked_at: string | null;
            expires_at: string | null;
        };
        LocalProjectGrantResultV1: {
            /** @constant */
            schema_version: 1;
            grant: components["schemas"]["LocalProjectGrantV1"];
        };
        GrantLocalProjectAgentRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            agent_id: string;
            actions: string[];
        };
        RevokeLocalProjectAgentRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            agent_id: string;
        };
        FixedReferenceV1: {
            material_id: string;
            revision: number;
        };
        LocalContextMaterialV1: {
            reference: components["schemas"]["FixedReferenceV1"];
            title: string;
            text: string;
            expanded?: boolean;
        };
        LocalContextFileV1: {
            /** @description Actual fixed file snapshot version returned by the project reader. */
            version?: string;
            path: string;
            text: string;
        };
        LocalContextRequestV1: {
            references: components["schemas"]["FixedReferenceV1"][];
            files: string[];
        };
        ReadLocalProjectContextRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            references?: components["schemas"]["FixedReferenceV1"][];
            files?: string[];
        };
        LocalContextPacketV1: {
            id: string;
            /** @constant */
            schema_version: 1;
            project_id: string;
            settings_revision: number;
            mode: string;
            materials: components["schemas"]["LocalContextMaterialV1"][];
            files: components["schemas"]["LocalContextFileV1"][];
            /**
             * Format: date-time
             * @description Time the controller captured this permission scope packet. In observed_history mode this is not a native historical message or session timestamp.
             */
            created_at: string;
        };
        NativeSessionV1: {
            operations?: {
                [key: string]: components["schemas"]["NativeOperationV1"];
            } | null;
            id: string;
            project_id: string;
            cli: string;
            native_id: string;
            version: string;
            mode: string;
            status: string;
            stop_confirmed: boolean;
            context_packet: components["schemas"]["LocalContextPacketV1"];
            source_session_id: string;
            /**
             * Format: date-time
             * @description Native session observation time; null when historical time is unknown. Context packet creation time is not native history evidence.
             */
            updated_at: string | null;
            limitations: string[];
            last_operation_id: string;
            pending_operation: string;
            /**
             * @description Unstarted only when launch was confirmed not to have created a process; never a substitute for unknown ownership.
             * @enum {string}
             */
            ownership: "owned" | "external_observed" | "unstarted";
        };
        NativeSessionResultV1: {
            /** @constant */
            schema_version: 1;
            session: components["schemas"]["NativeSessionV1"];
        };
        NativeSessionListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["NativeSessionV1"][];
            next_cursor: string | null;
        };
        NativeEventV1: {
            sequence: number;
            kind: string;
            text: string;
        };
        NativeObservationV1: {
            status: string;
            events: components["schemas"]["NativeEventV1"][];
            stop_confirmed: boolean;
            output_truncated: boolean;
        };
        NativeObservationResultV1: {
            /** @constant */
            schema_version: 1;
            observation: components["schemas"]["NativeObservationV1"];
        };
        NativeSessionRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            cli: string;
            message: string;
            context: components["schemas"]["LocalContextRequestV1"];
            deadline_seconds?: number;
            /** @enum {string} */
            mode?: "native" | "context_handoff";
            source_session_id?: string;
        };
        NativeMessageRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            message: string;
        };
        SourceCollectionV1: {
            /** @enum {string} */
            access_mode?: "public" | "browser_selected";
            external_id: string;
            owner_id: string;
            title: string;
            locator: string;
            item_count: number | null;
        };
        SourceCollectionListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["SourceCollectionV1"][];
            next_cursor: string | null;
        };
        ProjectAgentTokenV1: {
            token: string;
            project_id: string;
            agent_id: string;
            /** Format: date-time */
            expires_at: string;
        };
        ProjectAgentTokenResultV1: {
            /** @constant */
            schema_version: 1;
            credential: components["schemas"]["ProjectAgentTokenV1"];
        };
        IssueProjectAgentTokenRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            agent_id: string;
        };
        DiscoverNativeSessionsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            cli: string;
        };
        NativeContextListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["NativeEventV1"][];
            next_cursor: string | null;
        };
        NativeOperationV1: {
            action: string;
            /** @enum {string} */
            status: "pending" | "accepted" | "failed" | "unknown";
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            finished_at: string | null;
        };
        RegisteredProjectGitV1: {
            /** @enum {string} */
            status: "known" | "unknown" | "unavailable";
            reason: string;
            branch: string | null;
            head: string | null;
            /** Format: date-time */
            last_commit_at: string | null;
            dirty: boolean | null;
        };
        RegisteredProjectActivityV1: {
            /** @enum {string} */
            status: "known" | "unknown" | "unavailable";
            source: string;
            created_with_cli: string | null;
            /** Format: date-time */
            last_activity_at: string | null;
        };
        RegisteredProjectV1: {
            /** @description Stable aggregate project identity; grants no authority. */
            project_key?: string;
            contributors?: components["schemas"]["ProjectContributorV1"][];
            root: string;
            name: string;
            repo_id: string;
            /** @enum {string} */
            source: "orca_registered" | "subproject" | "git_worktree" | "native_project_metadata" | "explicit_directory";
            parent_root: string;
            git: components["schemas"]["RegisteredProjectGitV1"];
            activity: components["schemas"]["RegisteredProjectActivityV1"];
            limitations: string[];
        };
        RegisteredProjectDiscoveryFailureV1: {
            root: string;
            reason: string;
        };
        /**
         * @description Cached observations grant no permissions. Stale results retain the previous observation time and identify the current discovery failures.
         * @example {
         *       "status": "unknown",
         *       "observed_at": null,
         *       "projects": [],
         *       "failures": []
         *     }
         */
        RegisteredProjectSnapshotV1: {
            board?: components["schemas"]["ProjectSummaryV1"][];
            sources?: components["schemas"]["ProjectSourceObservationV1"][];
            /** @enum {string} */
            status: "unknown" | "complete" | "partial" | "stale";
            /** Format: date-time */
            observed_at: string | null;
            projects: components["schemas"]["RegisteredProjectV1"][];
            failures: components["schemas"]["RegisteredProjectDiscoveryFailureV1"][];
        };
        RegisteredProjectDiscoveryResultV1: {
            /** @constant */
            schema_version: 1;
            snapshot: components["schemas"]["RegisteredProjectSnapshotV1"];
        };
        RefreshRegisteredProjectsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
        };
        ProjectContributorV1: {
            cli: string;
            source: string;
            root: string;
            /** @description Observed native identity only, never a control permission. */
            session_id: string;
            /** Format: date-time */
            observed_at: string | null;
            /**
             * Format: date-time
             * @description Native header creation time; never current activity or progress.
             */
            created_at: string | null;
            /**
             * Format: date-time
             * @description An actual latest activity observation, otherwise null. Native creation headers do not establish activity.
             */
            activity_at: string | null;
        };
        ProjectSourceObservationV1: {
            cli: string;
            source: string;
            status: string;
            reason: string;
            /** @description Filesystem entries examined within the bounded known source. */
            entries_examined: number;
            /** @description Native metadata headers examined; not a session or project total. */
            headers_examined?: number;
            /** @description Distinct validated project roots identified in this source. */
            matched_roots?: number;
            /** @description Retained samples after per-root/client bounds; not total sessions. */
            retained_associations?: number;
            /** @description Unique valid metadata matches; distinct from retained samples. */
            matched_headers: number;
        };
        ProjectHumanMetadataV1: {
            notes: string;
            review: string;
            group: string;
            intent: string;
            archived: boolean;
            revision: number;
            /** Format: date-time */
            updated_at: string | null;
        };
        ProjectSummaryV1: {
            id: string;
            name: string;
            roots: string[];
            observations: components["schemas"]["RegisteredProjectV1"][];
            contributors: components["schemas"]["ProjectContributorV1"][];
            /** Format: date-time */
            last_activity_at: string | null;
            limitations: string[];
            human: components["schemas"]["ProjectHumanMetadataV1"];
        };
        ProjectSummaryResultV1: {
            /** @constant */
            schema_version: 1;
            project: components["schemas"]["ProjectSummaryV1"];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "manual-note",
         *       "expected_version": 1,
         *       "notes": "Human-owned notes",
         *       "review": "",
         *       "group": "",
         *       "intent": "",
         *       "archived": false,
         *       "revision": 0
         *     }
         */
        SetRegisteredProjectMetadataRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @description Human text, at most 65536 UTF-8 bytes. */
            notes: string;
            /** @description Human text, at most 65536 UTF-8 bytes. */
            review: string;
            /** @description Human group label, at most 200 UTF-8 bytes. */
            group: string;
            /** @description Human intent text, at most 4096 UTF-8 bytes. */
            intent: string;
            archived: boolean;
            /** @description Expected human metadata revision; zero means no prior annotation. */
            revision: number;
        };
        GitHubMetadataV1: {
            /** @enum {string} */
            metadata_source?: "github_rest" | "github_public_html" | "unknown";
            /** @description These metadata fields are unobserved; consumers must not display their scalar defaults as known facts. */
            unknown_fields?: string[];
            github_id: number;
            full_name: string;
            html_url: string;
            clone_url: string;
            description: string;
            default_branch: string;
            language: string;
            stars: number;
            archived: boolean;
            fork: boolean;
            /** Format: date-time */
            updated_at: string;
            /** Format: date-time */
            pushed_at: string;
        };
        /**
         * @description Anonymous public metadata and durable managed-clone observations. Empty root/head/project_id before explicit human placement; unknown or failed clone effects are never inferred successful. No grants, sessions or model permissions are created.
         * @example {
         *       "id": "github-1",
         *       "revision": 1,
         *       "metadata_revision": 1,
         *       "metadata": {
         *         "github_id": 1,
         *         "full_name": "example/repository",
         *         "html_url": "https://github.com/example/repository",
         *         "clone_url": "https://github.com/example/repository.git",
         *         "description": "Contract fixture only",
         *         "default_branch": "main",
         *         "language": "Go",
         *         "stars": 0,
         *         "archived": false,
         *         "fork": false,
         *         "updated_at": "2026-10-10T00:00:00Z",
         *         "pushed_at": "2026-10-10T00:00:00Z"
         *       },
         *       "sync_status": "synced",
         *       "synced_at": "2026-10-10T00:00:00Z",
         *       "space_id": "",
         *       "project_id": "",
         *       "clone_status": "not_placed",
         *       "root": "",
         *       "head": "",
         *       "clone_attempts": 0,
         *       "last_error": ""
         *     }
         */
        GitHubRepositoryV1: {
            id: string;
            revision: number;
            metadata_revision: number;
            metadata: components["schemas"]["GitHubMetadataV1"];
            /** @enum {string} */
            sync_status: "synced" | "stale" | "unknown";
            /** Format: date-time */
            synced_at: string;
            space_id: string;
            project_id: string;
            /** @enum {string} */
            clone_status: "not_placed" | "cloning" | "ready" | "failed" | "interrupted" | "unknown";
            root: string;
            head: string;
            clone_attempts: number;
            last_error: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "items": [
         *         {
         *           "id": "github-1",
         *           "revision": 1,
         *           "metadata_revision": 1,
         *           "metadata": {
         *             "github_id": 1,
         *             "full_name": "example/repository",
         *             "html_url": "https://github.com/example/repository",
         *             "clone_url": "https://github.com/example/repository.git",
         *             "description": "Contract fixture only",
         *             "default_branch": "main",
         *             "language": "Go",
         *             "stars": 0,
         *             "archived": false,
         *             "fork": false,
         *             "updated_at": "2026-10-10T00:00:00Z",
         *             "pushed_at": "2026-10-10T00:00:00Z"
         *           },
         *           "sync_status": "synced",
         *           "synced_at": "2026-10-10T00:00:00Z",
         *           "space_id": "",
         *           "project_id": "",
         *           "clone_status": "not_placed",
         *           "root": "",
         *           "head": "",
         *           "clone_attempts": 0,
         *           "last_error": ""
         *         }
         *       ],
         *       "next_cursor": null
         *     }
         */
        GitHubRepositoryListV1: {
            /** @constant */
            schema_version: 1;
            items: components["schemas"]["GitHubRepositoryV1"][];
            next_cursor: string | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "repository": {
         *         "id": "github-1",
         *         "revision": 1,
         *         "metadata_revision": 1,
         *         "metadata": {
         *           "github_id": 1,
         *           "full_name": "example/repository",
         *           "html_url": "https://github.com/example/repository",
         *           "clone_url": "https://github.com/example/repository.git",
         *           "description": "Contract fixture only",
         *           "default_branch": "main",
         *           "language": "Go",
         *           "stars": 0,
         *           "archived": false,
         *           "fork": false,
         *           "updated_at": "2026-10-10T00:00:00Z",
         *           "pushed_at": "2026-10-10T00:00:00Z"
         *         },
         *         "sync_status": "synced",
         *         "synced_at": "2026-10-10T00:00:00Z",
         *         "space_id": "",
         *         "project_id": "",
         *         "clone_status": "not_placed",
         *         "root": "",
         *         "head": "",
         *         "clone_attempts": 0,
         *         "last_error": ""
         *       }
         *     }
         */
        GitHubRepositoryResultV1: {
            /** @constant */
            schema_version: 1;
            repository: components["schemas"]["GitHubRepositoryV1"];
        };
        SyncGitHubRepositoriesRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @description Explicit public GitHub account, owner/repository or https://github.com URL. No credential or private scope. */
            input: string;
        };
        PlaceGitHubRepositoryRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @description Existing top-level development ProjectSpace selected by the human; expected_version is the cached repository revision. */
            space_id: string;
        };
        DiscoverSourceCollectionsRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            platform: "bilibili" | "douyin";
            owner_id: string;
            /**
             * @description Omitted legacy input means public; browser_selected requires explicit human selected owner and folder identity and configured browser bridge.
             * @enum {string}
             */
            access_mode?: "public" | "browser_selected";
        };
        PaperSearchHitV1: {
            /** @description Canonical import identity (doi:.../arxiv:... or canonical URL). The human selects a hit and imports it through ImportMaterial with this key for dedup. */
            source_key: string;
            /**
             * @description Official metadata index the hit came from.
             * @enum {string}
             */
            provider: "crossref" | "europepmc" | "arxiv";
            title: string;
            authors: string[];
            year: number;
            venue?: string;
            arxiv_id?: string;
            doi?: string;
            locator: string;
            abstract?: string;
            /**
             * @description Raw availability vocabulary. Metadata indexes always report abstract_only; full-text availability is resolved on import, never at search time.
             * @enum {string}
             */
            content_state: "readable_fulltext" | "abstract_only" | "paywall" | "restricted" | "unknown";
            /** @description Candidate public PDF links surfaced for human selection; a PDF is never presented as metadata full text. */
            pdf_urls: string[];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "query": "attention is all you need",
         *       "items": [],
         *       "next_cursor": null,
         *       "has_more": false,
         *       "warnings": []
         *     }
         */
        PaperSearchResultV1: {
            /** @constant */
            schema_version: 1;
            query: string;
            items: components["schemas"]["PaperSearchHitV1"][];
            next_cursor: string | null;
            has_more: boolean;
            warnings: string[];
        };
        ProgressEvidenceV1: {
            source_path: string;
            kind: string;
            version: string;
            excerpt?: string;
        };
        ProjectProgressV1: {
            project_id: string;
            /** @description Advisory stage label. The wire reports "unknown" for an unobserved project instead of inventing a stage. */
            status: string;
            summary?: string;
            percent?: number | null;
            /** @description Null until a human writes or an Agent inference concludes. */
            source: ("human" | "agent_inferred") | null;
            evidence: components["schemas"]["ProgressEvidenceV1"][];
            native_id?: string;
            model?: string | null;
            /** @description Null when no observation has been recorded; never a fabricated zero timestamp. */
            observed_at: string | null;
            warning?: string;
            revision: number;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "progress": {
         *         "project_id": "project-example",
         *         "status": "unknown",
         *         "source": null,
         *         "evidence": [],
         *         "observed_at": null,
         *         "revision": 0
         *       }
         *     }
         */
        ProjectProgressResultV1: {
            /** @constant */
            schema_version: 1;
            progress: components["schemas"]["ProjectProgressV1"];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "example-request",
         *       "expected_version": 1,
         *       "status": "in progress"
         *     }
         */
        SetProgressRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            status: string;
            summary?: string;
            percent?: number | null;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "example-request",
         *       "expected_version": 1,
         *       "files": [
         *         "TASK.md",
         *         "STATUS.md"
         *       ]
         *     }
         */
        InferProgressRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            files: string[];
        };
        /**
         * @example {
         *       "event_id": "e-1",
         *       "ts": 1700000000000
         *     }
         */
        EventRefV1: {
            event_id: string;
            ts: number;
        };
        KnowledgeEntryV1: {
            id: string;
            topic: string;
            fluid: string;
            crystallized: Record<string, never>;
            sources: components["schemas"]["EventRefV1"][];
            /** @enum {string} */
            status: "inbox" | "confirmed";
        };
        ExperienceV1: {
            id: string;
            pattern: string;
            /** @enum {string} */
            verdict: "avoid" | "prefer" | "condition";
            condition: string;
            evidence: components["schemas"]["EventRefV1"][];
            /** @enum {string} */
            status: "quarantine" | "approved";
        };
        ConflictV1: {
            id: string;
            topic: string;
            /** @enum {string} */
            kind: "attention_node" | "knowledge_entry";
            left_ref: string;
            right_ref: string;
            detail: string;
            /** @enum {string} */
            status: "pending" | "resolved_keep_left" | "resolved_keep_right" | "resolved_merge" | "dismissed";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "example-request",
         *       "expected_version": 1,
         *       "owner_project": "fixture-project"
         *     }
         */
        ConfirmKnowledgeRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            owner_project: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "example-request",
         *       "expected_version": 1,
         *       "owner_project": "fixture-project"
         *     }
         */
        ApproveExperienceRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            owner_project: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "example-request",
         *       "expected_version": 1,
         *       "decision": "merge"
         *     }
         */
        AdjudicateConflictRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            /** @enum {string} */
            decision: "keep_left" | "keep_right" | "merge" | "dismiss";
        };
        ConfirmKnowledgeResultV1: {
            /** @constant */
            schema_version: 1;
            entry: components["schemas"]["KnowledgeEntryV1"];
            version: number;
            owner_project: string;
        };
        ApproveExperienceResultV1: {
            /** @constant */
            schema_version: 1;
            experience: components["schemas"]["ExperienceV1"];
            version: number;
            owner_project: string;
        };
        AdjudicateConflictResultV1: {
            /** @constant */
            schema_version: 1;
            conflict: components["schemas"]["ConflictV1"];
            version: number;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "owner_project": "fixture-project",
         *       "target_project": "fixture-other"
         *     }
         */
        ShareMemoryRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            owner_project: string;
            target_project: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "owner_project": "fixture-project"
         *     }
         */
        DemoteMemoryRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            owner_project: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "project_id": "fixture-project"
         *     }
         */
        ReferenceKnowledgeRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            expected_version: number;
            project_id: string;
        };
        MemoryShareAuditV1: {
            owner_project: string;
            target_project: string;
            actor_id: string;
            /** @constant */
            actor_kind: "human";
            ts: number;
            request_id: string;
        };
        MemoryActionAuditV1: {
            action: string;
            actor_id: string;
            /** @constant */
            actor_kind: "human";
            ts: number;
            request_id: string;
            version: number;
        };
        KnowledgeVersionV1: {
            version: number;
            content: unknown;
            recorded_at: number;
        };
        StoredKnowledgeV1: {
            entry: components["schemas"]["KnowledgeEntryV1"];
            owner_project: string;
            version: number;
            created_at: number;
            updated_at: number;
            consolidated_at: number | null;
            crystallized_versions: components["schemas"]["KnowledgeVersionV1"][];
            shares: components["schemas"]["MemoryShareAuditV1"][];
            actions: components["schemas"]["MemoryActionAuditV1"][];
        };
        StoredExperienceV1: {
            experience: components["schemas"]["ExperienceV1"];
            owner_project: string;
            origin_event_id: string;
            version: number;
            created_at: number;
            updated_at: number;
            shares: components["schemas"]["MemoryShareAuditV1"][];
            actions: components["schemas"]["MemoryActionAuditV1"][];
        };
        KnowledgeReadResultV1: {
            /** @constant */
            schema_version: 1;
            knowledge: components["schemas"]["StoredKnowledgeV1"];
        };
        ExperienceReadResultV1: {
            /** @constant */
            schema_version: 1;
            experience: components["schemas"]["StoredExperienceV1"];
        };
        WorkFileFactV1: {
            source_path: string;
            /** Format: date-time */
            mtime: string | null;
            file_count: number;
            verified: boolean;
            reason?: string;
        };
        WorkHandoffV1: {
            content: string;
            /** Format: date-time */
            updated_at: string | null;
            /** @enum {string} */
            source: "" | "human" | "generated";
        };
        WorkV1: {
            id: string;
            platform: string;
            name: string;
            root: string;
            /** @enum {string} */
            kind: "folder" | "file";
            /** @enum {string} */
            source: "scan" | "json" | "registered";
            /** @enum {string} */
            attitude: "" | "continue" | "park" | "done";
            group: string;
            hidden: boolean;
            handoff: components["schemas"]["WorkHandoffV1"];
            files: components["schemas"]["WorkFileFactV1"];
            /** Format: date-time */
            updated_at: string | null;
            revision: number;
        };
        WorkDiscoveryFailureV1: {
            root: string;
            reason: string;
        };
        WorkPlatformStatV1: {
            platform: string;
            works: number;
            folders: number;
            handed_over: number;
            recent: number;
        };
        WorksSnapshotV1: {
            works: components["schemas"]["WorkV1"][];
            platforms: components["schemas"]["WorkPlatformStatV1"][];
            /** @enum {string} */
            status: "unknown" | "complete" | "partial" | "stale";
            /** Format: date-time */
            observed_at: string | null;
            failures: components["schemas"]["WorkDiscoveryFailureV1"][];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "snapshot": {
         *         "works": [],
         *         "platforms": [],
         *         "status": "unknown",
         *         "observed_at": null,
         *         "failures": []
         *       }
         *     }
         */
        WorksSnapshotResultV1: {
            /** @constant */
            schema_version: 1;
            snapshot: components["schemas"]["WorksSnapshotV1"];
        };
        WorkResultV1: {
            /** @constant */
            schema_version: 1;
            work: components["schemas"]["WorkV1"];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "roots": []
         *     }
         */
        WorksRefreshRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            roots?: string[];
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "content": "{\"works\":[]}"
         *     }
         */
        WorksDeliveryRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            content: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "expected_revision": 0,
         *       "attitude": "continue"
         *     }
         */
        WorkAttitudeRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            expected_revision: number;
            /** @enum {string} */
            attitude: "continue" | "park" | "done";
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "expected_revision": 0,
         *       "group": "research"
         *     }
         */
        WorkGroupRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            expected_revision: number;
            group: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "expected_revision": 0,
         *       "content": "fixture handoff"
         *     }
         */
        WorkHandoffRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            expected_revision: number;
            content: string;
        };
        /**
         * @example {
         *       "schema_version": 1,
         *       "request_id": "fixture-request",
         *       "expected_version": 1,
         *       "expected_revision": 0,
         *       "hidden": true
         *     }
         */
        WorkHiddenRequestV1: {
            /** @constant */
            schema_version: 1;
            request_id: string;
            /** @constant */
            expected_version: 1;
            expected_revision: number;
            hidden: boolean;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    listRelaySessionEvents: {
        parameters: {
            query?: {
                task_id?: string;
                start?: string;
                end?: string;
                page?: number;
                page_size?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Local pending admissions and explicit coverage, independent of cognitive snapshot values */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "events": [
                     *         {
                     *           "event_id": "ae6c15f2-394b-402c-9f53-b75fc2cd0783",
                     *           "kind": "edit_file",
                     *           "topic": "relay:cf1497b0-e092-43c0-80a9-33449bb57a40",
                     *           "tokens": null,
                     *           "scope": "public-project",
                     *           "ts": "2026-10-14T01:00:01Z"
                     *         }
                     *       ],
                     *       "coverage": [
                     *         {
                     *           "task_id": "cf1497b0-e092-43c0-80a9-33449bb57a40",
                     *           "terminal_id": "c48d6c05-3b58-4bdc-8a7b-46a471612a0b",
                     *           "worktree_id": "0623c675-ddb7-4d62-bd02-831f6cb6d51d",
                     *           "project_id": "public-project",
                     *           "status": "partial",
                     *           "gap": "prefix_unverified",
                     *           "lines": 2,
                     *           "projected": 2,
                     *           "duplicates": 0,
                     *           "unknown": 0,
                     *           "errors": 0,
                     *           "unknown_tokens": 2,
                     *           "remaining_bytes": 0,
                     *           "tail_bytes": 0,
                     *           "snapshot_complete": true,
                     *           "session_complete": false
                     *         }
                     *       ],
                     *       "total": 1,
                     *       "page": 1,
                     *       "page_size": 50
                     *     }
                     */
                    "application/json": components["schemas"]["RelaySessionEventsV1"];
                };
            };
            /** @description Guard, invalid query or unavailable read; no Agent action or automatic retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, invalid query or unavailable read; no Agent action or automatic retry */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, invalid query or unavailable read; no Agent action or automatic retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, invalid query or unavailable read; no Agent action or automatic retry */
            405: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, invalid query or unavailable read; no Agent action or automatic retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    reconcileRelayTaskEvents: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayEmptyRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayTaskResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRelayProjects: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayProjectsV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRelayAgents: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayAgentsV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRelayTasks: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayTasksV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getRelayTask: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayTaskResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRelayTaskEvents: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayEventsV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRelayTaskReceipts: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayReceiptsV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    createRelayIntent: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayIntentRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayTaskResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    confirmRelayIntent: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayConfirmRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayActionResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    sendRelayTask: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelaySendRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayActionResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    readRelayTask: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayEmptyRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayActionResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    stopRelayTask: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayEmptyRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayActionResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    resumeRelayTask: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayEmptyRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayActionResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    transitionRelayTask: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RelayTransitionRequestV1"];
            };
        };
        responses: {
            /** @description Derived relay state */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RelayTaskResultV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Guard, permission, conflict or unavailable result; no launch retry */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchHealth: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchHealthV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchToday: {
        parameters: {
            query?: {
                source?: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchTodayV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchFocus: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchFocusV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchEvidence: {
        parameters: {
            query?: {
                topic?: string;
                page?: number;
                page_size?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchEvidenceV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchAgents: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchAgentsV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchEvents: {
        parameters: {
            query?: {
                topic?: string;
                kind?: string;
                source?: string;
                start?: string;
                end?: string;
                page?: number;
                page_size?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchEventsV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getWorkbenchFilters: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Confirmed snapshot view */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkbenchFiltersV1"];
                };
            };
            /** @description Invalid filter or pagination */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local human session required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Published snapshot unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    discoverSourceCollections: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DiscoverSourceCollectionsRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SourceCollectionListV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listGitHubRepositories: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Durable cached observations; no automatic permission expansion */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["GitHubRepositoryListV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    syncGitHubRepositories: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SyncGitHubRepositoriesRequestV1"];
            };
        };
        responses: {
            /** @description Durable cached observations; no automatic permission expansion */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["GitHubRepositoryListV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    placeGitHubRepository: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["PlaceGitHubRepositoryRequestV1"];
            };
        };
        responses: {
            /** @description Durable cached observations; no automatic permission expansion */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["GitHubRepositoryResultV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error; never replay unknown effects automatically */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    probeNativeCLI: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DiscoverNativeSessionsRequestV1"];
            };
        };
        responses: {
            /** @description Current scoped service result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeSessionResultV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listSourceCollections: {
        parameters: {
            query: {
                platform: string;
                owner_id: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current scoped service result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SourceCollectionListV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    issueProjectAgentToken: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["IssueProjectAgentTokenRequestV1"];
            };
        };
        responses: {
            /** @description Current scoped service result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectAgentTokenResultV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    discoverNativeSessions: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DiscoverNativeSessionsRequestV1"];
            };
        };
        responses: {
            /** @description Current scoped service result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeSessionListV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    readNativeContext: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
                session_id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current scoped service result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeContextListV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable error; unknown effects are not replayed */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listTrackingSources: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceListV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    bindTrackingSource: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["BindTrackingSourceRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getTrackingSource: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    syncTrackingSource: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SyncTrackingSourceRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    recommendSourceItems: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RecommendSourceItemsRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    selectSourceItems: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SelectSourceItemsRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TrackingSourceResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listRegisteredProjects: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Persisted observations with explicit unknown, partial or stale status */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RegisteredProjectDiscoveryResultV1"];
                };
            };
            /** @description Invalid command */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human session or CSRF required; Agent credentials rejected */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Cache persistence failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Discovery service is not assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Discovery source unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    refreshRegisteredProjects: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RefreshRegisteredProjectsRequestV1"];
            };
        };
        responses: {
            /** @description Persisted observations with explicit unknown, partial or stale status */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RegisteredProjectDiscoveryResultV1"];
                };
            };
            /** @description Invalid command */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human session or CSRF required; Agent credentials rejected */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Cache persistence failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Discovery service is not assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Discovery source unavailable */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setRegisteredProjectMetadata: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                project_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SetRegisteredProjectMetadataRequestV1"];
            };
        };
        responses: {
            /** @description Updated project summary with durable human metadata */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSummaryResultV1"];
                };
            };
            /** @description Invalid command or unsupported fields */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human session and CSRF required; Agent credentials rejected */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Project is absent from the cached registered observations */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Manual metadata revision changed; reload before saving */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Persistence failed */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Metadata service is not assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listLocalProjects: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectListV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    registerLocalProject: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RegisterLocalProjectRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    discoverLocalProjects: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DiscoverLocalProjectsRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectCandidateListV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setLocalProjectSettings: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["LocalProjectSettingsRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    grantLocalProjectAgent: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GrantLocalProjectAgentRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectGrantResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    revokeLocalProjectAgent: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RevokeLocalProjectAgentRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalProjectGrantResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    readLocalProjectContext: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ReadLocalProjectContextRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalContextPacketV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listNativeSessions: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeSessionListV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    startNativeSession: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["NativeSessionRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeSessionResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    resumeNativeSession: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
                session_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["NativeSessionRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeSessionResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    sendNativeMessage: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
                session_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["NativeMessageRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeObservationResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    stopNativeSession: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token"?: string;
            };
            path: {
                id: string;
                session_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["JobCommandV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeObservationResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    observeNativeSession: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
                session_id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NativeObservationResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    searchPapers: {
        parameters: {
            query: {
                q: string;
                /** @description Provider defaults to crossref when omitted. */
                provider?: "crossref" | "europepmc" | "arxiv";
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Ephemeral metadata results; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PaperSearchResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getProjectProgress: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectProgressResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setProjectProgress: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SetProgressRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectProgressResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    inferProjectProgress: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["InferProgressRequestV1"];
            };
        };
        responses: {
            /** @description Current service result; unknown and failure states remain explicit */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectProgressResultV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Actionable structured error; no automatic replay of unknown effects */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listLocalAgents: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Cached inventory; an empty array means no observed entries */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalAgentListV1"];
                };
            };
            /** @description Human session required; Agent bearer grants no access */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Inventory query failed; do not substitute an empty array */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Local Agent inventory is not assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "local Agent inventory is not assembled",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "start_a_server_with_local_agent_inventory"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getHealth: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "status": "ok",
                     *       "service": "astrocyte"
                     *     }
                     */
                    "application/json": components["schemas"]["HealthV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getFoundation: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "stage": "S0",
                     *       "capabilities": {
                     *         "imports": false,
                     *         "approvals": false,
                     *         "execution": false,
                     *         "native_resume": false,
                     *         "handoff": false
                     *       },
                     *       "storage": {
                     *         "engine": "sqlite",
                     *         "schema_version": 1
                     *       },
                     *       "fixture": false
                     *     }
                     */
                    "application/json": components["schemas"]["FoundationV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listMaterials: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["MaterialListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listOpportunitys: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["OpportunityListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    createOpportunity: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["OpportunityRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpportunityDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listProjects: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["ProjectListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    registerProject: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "name": "Fixture only",
                 *       "environment_id": "fixture",
                 *       "root_path": "/fixture"
                 *     }
                 */
                "application/json": components["schemas"]["RegisterProjectRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listProposals: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["ProposalListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listSessions: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["SessionListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listMissions: {
        parameters: {
            query?: {
                /** @description Opaque continuation cursor */
                cursor?: string;
                /** @description Maximum returned items */
                limit?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "items": [],
                     *       "next_cursor": null
                     *     }
                     */
                    "application/json": components["schemas"]["MissionListV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getMission: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description S0 real read view; never populated from fixtures */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "mission": {
                     *         "id": "fixture-mission",
                     *         "version": 1,
                     *         "grant_id": "fixture-grant",
                     *         "goal": "Fixture only",
                     *         "status": "pending"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["MissionResponseV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    importMaterial: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "source_locator": "arxiv:fixture",
                 *       "source_key": "fixture-source",
                 *       "kind": "paper",
                 *       "content_digest": "fixture-digest",
                 *       "export_text": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["ImportMaterialRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "job_id": "fixture-job",
                     *       "status": "queued"
                     *     }
                     */
                    "application/json": components["schemas"]["ImportJobV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    reviewOpportunity: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "feedback": "later",
                 *       "reason": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["ReviewOpportunityRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    admitOpportunity: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "reason": "Fixture only",
                 *       "policy_version": null
                 *     }
                 */
                "application/json": components["schemas"]["AdmitOpportunityRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    resumeSession: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "context_packet_id": null
                 *     }
                 */
                "application/json": components["schemas"]["ResumeSessionRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "operation_id": "fixture-operation",
                     *       "status": "queued"
                     *     }
                     */
                    "application/json": components["schemas"]["OperationV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    handoffSession: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "context_packet_id": "fixture-packet",
                 *       "target_adapter": "fixture",
                 *       "target_project_id": "fixture-project"
                 *     }
                 */
                "application/json": components["schemas"]["HandoffSessionRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "operation_id": "fixture-operation",
                     *       "status": "queued"
                     *     }
                     */
                    "application/json": components["schemas"]["OperationV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    submitProposal: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1
                 *     }
                 */
                "application/json": components["schemas"]["SubmitProposalRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    approveProposal: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "scope": {
                 *         "project_ids": [],
                 *         "allowed_roots": [],
                 *         "allowed_actions": [],
                 *         "data_egress_rules": []
                 *       },
                 *       "budget": {
                 *         "calls": null,
                 *         "tokens": null,
                 *         "money": null,
                 *         "currency": null
                 *       },
                 *       "stop_conditions": [
                 *         "Fixture only"
                 *       ],
                 *       "expires_at": "2026-10-09T00:00:00Z"
                 *     }
                 */
                "application/json": components["schemas"]["ApproveProposalRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "grant_id": "fixture-grant",
                     *       "proposal_id": "fixture-proposal",
                     *       "proposal_revision": 1,
                     *       "epoch": 1,
                     *       "status": "active"
                     *     }
                     */
                    "application/json": components["schemas"]["ApprovalV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    revokeApproval: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "reason": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["RevokeApprovalRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "grant_id": "fixture-grant",
                     *       "proposal_id": "fixture-proposal",
                     *       "proposal_revision": 1,
                     *       "epoch": 1,
                     *       "status": "active"
                     *     }
                     */
                    "application/json": components["schemas"]["ApprovalV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    pauseMission: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "reason": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["MissionControlRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    cancelMission: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "reason": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["MissionControlRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "id": "fixture-resource",
                     *       "version": 1
                     *     }
                     */
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    claimWorkItem: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "context_packet_id": "fixture-packet"
                 *     }
                 */
                "application/json": components["schemas"]["ClaimWorkItemRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "work_item_id": "fixture-work",
                     *       "attempt_id": "fixture-attempt",
                     *       "holder_id": "fixture-holder",
                     *       "fence": 1,
                     *       "lease_token": "fixture-not-valid",
                     *       "expires_at": "2026-10-09T00:00:00Z"
                     *     }
                     */
                    "application/json": components["schemas"]["LeaseV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    submitArtifact: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "attempt_id": "fixture-attempt",
                 *       "lease_token": "fixture-not-valid",
                 *       "fence": 1,
                 *       "object_ref": "fixture-object",
                 *       "content_digest": "fixture-digest",
                 *       "snapshot_ref": "fixture-snapshot",
                 *       "checks_ref": "fixture-checks"
                 *     }
                 */
                "application/json": components["schemas"]["SubmitArtifactRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "artifact_id": "fixture-artifact",
                     *       "status": "submitted"
                     *     }
                     */
                    "application/json": components["schemas"]["ArtifactResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    acceptArtifact: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
            };
            path: {
                /** @description Registered resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                /**
                 * @example {
                 *       "schema_version": 1,
                 *       "request_id": "fixture-request",
                 *       "expected_version": 1,
                 *       "reason": "Fixture only"
                 *     }
                 */
                "application/json": components["schemas"]["AcceptArtifactRequestV1"];
            };
        };
        responses: {
            /** @description Future capability contract; NOT implemented in S0 */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "acceptance_id": "fixture-acceptance",
                     *       "artifact_id": "fixture-artifact",
                     *       "experience_id": "fixture-experience"
                     *     }
                     */
                    "application/json": components["schemas"]["AcceptanceV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "validation_failed",
                     *         "message": "Invalid request",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope or user-control access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "not_found",
                     *         "message": "Resource not found",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "version_conflict",
                     *         "message": "Resource version changed",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    streamEvents: {
        parameters: {
            query?: never;
            header?: {
                /** @description Opaque resumption cursor */
                "Last-Event-ID"?: string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Future SSE, each data frame contains EventV1 JSON */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example id: fixture-event
                     *     data: {"event_id":"fixture-event","type":"attention.opportunity_admitted","schema_version":1,"aggregate_id":"fixture-opportunity","aggregate_version":1,"occurred_at":"2026-10-09T00:00:00Z","correlation_id":"fixture-request","causation_id":null,"payload":{"opportunity_id":"fixture-opportunity","revision":1,"admission_id":"fixture-admission","actor":"fixture-user","policy_version":"fixture-policy"}}
                     */
                    "text/event-stream": string;
                };
            };
            /** @description Scope access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "scope_denied",
                     *         "message": "Scope access denied",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "internal_error",
                     *         "message": "Unexpected service failure",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in S0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in S0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current S0 capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getLocalSession: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LocalSessionV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getMaterial: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    updateMaterial: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateMaterialRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getMaterialContent: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
                revision: number;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ContentV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    recordMaterialUse: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RecordUseRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["VersionResultV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listDistillations: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DistillationListV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    recordDistillation: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RecordDistillationRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DistillationResultV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getOpportunity: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpportunityDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    reviseOpportunity: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["OpportunityRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OpportunityDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listJobs: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["JobListV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getJob: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["JobV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    retryJob: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["JobCommandV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["JobV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    cancelJob: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["JobCommandV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["JobV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getMaterialAttachment: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
                revision: number;
                name: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Original source bytes; no filesystem paths accepted */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/octet-stream": string;
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listMaterialDomains: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDomainListV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    createMaterialDomain: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["MaterialDomainRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDomainResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    reviseMaterialDomain: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["MaterialDomainRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDomainResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setMaterialDomains: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SetMaterialDomainsRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MaterialDetailV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listProjectSpaces: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSpaceListV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    createProjectSpace: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ProjectSpaceRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSpaceResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getProjectSpace: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSpaceResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    referenceMaterial: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ReferenceMaterialRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSpaceResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    removeMaterialReference: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RemoveMaterialReferenceRequestV1"];
            };
        };
        responses: {
            /** @description Human classification result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProjectSpaceResultV1"];
                };
            };
            /** @description Structured error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    requestDistillation: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RequestDistillationRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            202: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ImportJobV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Native processor or read isolation unsupported */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getRankingProfile: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RankingProfileDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Profile adapter not connected */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    updateRankingProfile: {
        parameters: {
            query?: never;
            header: {
                "Idempotency-Key": string;
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateRankingProfileRequestV1"];
            };
        };
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RankingProfileDetailV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Profile adapter not connected */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    getDistillerStatus: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Attention result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DistillerStatusV1"];
                };
            };
            /** @description Structured service error */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Structured service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    confirmKnowledge: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description KnowledgeEntry identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ConfirmKnowledgeRequestV1"];
            };
        };
        responses: {
            /** @description Confirmed entry (future DTO only in M0) */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "entry": {
                     *         "id": "k-1",
                     *         "topic": "attention-engine",
                     *         "fluid": "evolving summary",
                     *         "crystallized": {
                     *           "fact": "true"
                     *         },
                     *         "sources": [
                     *           {
                     *             "event_id": "e-1",
                     *             "ts": 1700000000000
                     *           }
                     *         ],
                     *         "status": "confirmed"
                     *       },
                     *       "version": 2,
                     *       "owner_project": "fixture-project"
                     *     }
                     */
                    "application/json": components["schemas"]["ConfirmKnowledgeResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version or status changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in M0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in M0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    approveExperience: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Experience identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ApproveExperienceRequestV1"];
            };
        };
        responses: {
            /** @description Approved experience (future DTO only in M0) */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "experience": {
                     *         "id": "x-1",
                     *         "pattern": "centralized scheduler",
                     *         "verdict": "avoid",
                     *         "condition": ">100 agents",
                     *         "evidence": [
                     *           {
                     *             "event_id": "e-9",
                     *             "ts": 1700000000000
                     *           }
                     *         ],
                     *         "status": "approved"
                     *       },
                     *       "version": 2,
                     *       "owner_project": "fixture-project"
                     *     }
                     */
                    "application/json": components["schemas"]["ApproveExperienceResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version or status changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in M0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in M0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    adjudicateConflict: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key; never automatic retry */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Conflict identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AdjudicateConflictRequestV1"];
            };
        };
        responses: {
            /** @description Adjudicated conflict (future DTO only in M0) */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "conflict": {
                     *         "id": "c-1",
                     *         "topic": "scheduler",
                     *         "kind": "knowledge_entry",
                     *         "left_ref": "k-1",
                     *         "right_ref": "k-2",
                     *         "detail": "contradictory capacity facts",
                     *         "status": "resolved_merge"
                     *       },
                     *       "version": 2
                     *     }
                     */
                    "application/json": components["schemas"]["AdjudicateConflictResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Scope or user-control access denied */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource version or status changed */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Unexpected service failure */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Capability is not implemented in M0 */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Capability is not implemented in M0",
                     *         "request_id": "example-request",
                     *         "retryable": false,
                     *         "required_action": "Read the current capabilities; choose an implemented operation"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    readEngineknowledge: {
        parameters: {
            query: {
                /** @description Project scope for persisted visibility filtering */
                project_id: string;
            };
            header?: never;
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["KnowledgeReadResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    sharesEngineknowledge: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ShareMemoryRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ConfirmKnowledgeResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    demoteEngineknowledge: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DemoteMemoryRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ConfirmKnowledgeResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    readEngineexperiences: {
        parameters: {
            query: {
                /** @description Project scope for persisted visibility filtering */
                project_id: string;
            };
            header?: never;
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ExperienceReadResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    sharesEngineexperiences: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ShareMemoryRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ApproveExperienceResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    demoteEngineexperiences: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DemoteMemoryRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ApproveExperienceResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    referenceEngineKnowledge: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ReferenceKnowledgeRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ConfirmKnowledgeResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    listWorks: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorksSnapshotResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    refreshWorks: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorksRefreshRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorksSnapshotResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    importWorksDelivery: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorksDeliveryRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorksSnapshotResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setWorkAttitude: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                work_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorkAttitudeRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setWorkGroup: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                work_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorkGroupRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setWorkHandoff: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                work_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorkHandoffRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
    setWorkHidden: {
        parameters: {
            query?: never;
            header: {
                /** @description Caller-scoped command deduplication key */
                "Idempotency-Key": string;
                /** @description Session-bound CSRF token */
                "X-CSRF-Token": string;
            };
            path: {
                /** @description Stored resource identifier */
                work_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WorkHiddenRequestV1"];
            };
        };
        responses: {
            /** @description Persisted result */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WorkResultV1"];
                };
            };
            /** @description Invalid request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Human authority or scope required */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Resource not found */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Version conflict or uncertain delivery */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
            /** @description Service has not been assembled */
            501: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    /**
                     * @example {
                     *       "schema_version": 1,
                     *       "error": {
                     *         "code": "unsupported_capability",
                     *         "message": "Service has not been assembled",
                     *         "request_id": "fixture-request",
                     *         "retryable": false,
                     *         "required_action": "configure_supported_service"
                     *       }
                     *     }
                     */
                    "application/json": components["schemas"]["ErrorV1"];
                };
            };
        };
    };
}
