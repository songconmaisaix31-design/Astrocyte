-- Workspace 固定目标与提案版本、人类批准、上下文包和会话绑定。
CREATE TABLE workspace_goals (
    id TEXT PRIMARY KEY,
    revision INTEGER NOT NULL
);

CREATE TABLE workspace_goal_revisions (
    id TEXT NOT NULL REFERENCES workspace_goals(id),
    revision INTEGER NOT NULL,
    data TEXT NOT NULL,
    PRIMARY KEY (id, revision)
);

CREATE TABLE workspace_proposals (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL,
    revision INTEGER NOT NULL,
    opportunity_id TEXT,
    opportunity_revision INTEGER NOT NULL,
    data TEXT NOT NULL,
    UNIQUE (opportunity_id, opportunity_revision)
);

CREATE TABLE workspace_proposal_revisions (
    id TEXT NOT NULL REFERENCES workspace_proposals(id),
    revision INTEGER NOT NULL,
    data TEXT NOT NULL,
    PRIMARY KEY (id, revision)
);

CREATE TABLE workspace_approvals (
    id TEXT PRIMARY KEY,
    epoch INTEGER NOT NULL,
    proposal_id TEXT NOT NULL,
    proposal_revision INTEGER NOT NULL,
    data TEXT NOT NULL,
    UNIQUE (proposal_id, proposal_revision),
    FOREIGN KEY (proposal_id, proposal_revision)
        REFERENCES workspace_proposal_revisions(id, revision)
);

ALTER TABLE foundation_idempotency ADD COLUMN result_json TEXT NOT NULL DEFAULT 'null';

CREATE TABLE workspace_context_packets (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES local_agent_projects(id),
    predecessor_id TEXT REFERENCES workspace_context_packets(id),
    data TEXT NOT NULL,
    object_ref TEXT NOT NULL
);

CREATE TABLE workspace_session_bindings (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL,
    project_id TEXT NOT NULL REFERENCES local_agent_projects(id),
    session_id TEXT NOT NULL REFERENCES local_agent_sessions(id),
    grant_id TEXT NOT NULL REFERENCES workspace_approvals(id),
    packet_id TEXT NOT NULL REFERENCES workspace_context_packets(id),
    data TEXT NOT NULL,
    UNIQUE (project_id, session_id)
);

UPDATE foundation_schema SET version = 7 WHERE context = 'workspace';
INSERT INTO _migrations(id) VALUES ('019_workspace_approvals');
