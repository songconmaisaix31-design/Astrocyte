-- 委托规则由实际用户配置，判断绑定保留规则和 Opportunity 固定版本。
CREATE TABLE judgment_delegation_rules (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision >= 1),
    state TEXT NOT NULL CHECK(state IN ('disabled','active','paused','revoked')),
    data TEXT NOT NULL CHECK(json_valid(data)),
    UNIQUE(project_id, workflow_id)
);

CREATE TABLE judgment_admission_bindings (
    result_id TEXT PRIMARY KEY REFERENCES judgment_results(id),
    rule_id TEXT NOT NULL REFERENCES judgment_delegation_rules(id),
    rule_revision INTEGER NOT NULL CHECK(rule_revision >= 1),
    project_id TEXT NOT NULL,
    workflow_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    opportunity_id TEXT NOT NULL,
    opportunity_revision INTEGER NOT NULL CHECK(opportunity_revision >= 1),
    data TEXT NOT NULL CHECK(json_valid(data)),
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_admission_bindings_opportunity
    ON judgment_admission_bindings(project_id, opportunity_id, opportunity_revision);

UPDATE foundation_schema SET version=2 WHERE context='judgment';
INSERT INTO _migrations(id) VALUES ('025_judgment_delegation');
