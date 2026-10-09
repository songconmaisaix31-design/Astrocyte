// Generated from contracts/openapi.yaml. Run pnpm generate; do not edit.
export interface paths {
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 human write. Requires local session and matching CSRF token. Import queues durable work; reviews do not approve or execute a Mission.
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
         * @description S1 human write. Requires local session and matching CSRF token. Import queues durable work; reviews do not approve or execute a Mission.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
         */
        get: operations["getMaterial"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /**
         * updateMaterial
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
         */
        get: operations["listDistillations"];
        put?: never;
        /**
         * recordDistillation
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         * @description S1 Attention service. Human writes require local session cookie and CSRF token; Agent bearer credentials are read-only. Reads do not increase human attention.
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
         */
        FoundationV1: {
            /** @constant */
            schema_version: 1;
            /** @enum {string} */
            stage: "S0";
            /** @description Availability of HTTP operations in S0, not evidence about installed local agent native abilities. */
            capabilities: {
                /** @constant */
                imports: false;
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
            /** @enum {string} */
            adapter?: "arxiv" | "summarize" | "manual";
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
            /** @enum {string} */
            mode: "manual" | "arxiv" | "summarize_export" | "summarize_extract";
            source: string;
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
            action: "reread" | "annotate" | "pin" | "adopt";
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
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
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
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
        };
    };
}
