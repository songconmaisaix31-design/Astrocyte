// Generated from contracts/openapi.yaml. Run pnpm generate; do not edit.
export interface paths {
    "/source-collections": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * list Source Collections
         * @description Public favorite folders for one explicit owner, metadata only; no auto binding or cookies.
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
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
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
             * @description summarize_url extracts a selected public video URL without export_text or local_file_ref. summarize, summarize_json and summarize_markdown import existing exports. Extraction does not imply model distillation or Agent authorization.
             * @enum {string}
             */
            adapter?: "arxiv" | "summarize_url" | "summarize" | "summarize_json" | "summarize_markdown" | "manual";
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
        };
        LocalContextFileV1: {
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
            /** Format: date-time */
            created_at: string;
        };
        NativeSessionV1: {
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
            /** Format: date-time */
            updated_at: string;
            limitations: string[];
            last_operation_id: string;
            pending_operation: string;
            /** @enum {string} */
            ownership: "owned" | "external_observed";
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
            mode?: "cooperative" | "context_handoff";
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
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
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
}
