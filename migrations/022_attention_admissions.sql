-- 准入是固定候选版本的独立业务记录，执行批准继续由 Workspace 管理。
CREATE TABLE attention_admissions (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL REFERENCES attention_opportunities(id),
    opportunity_revision INTEGER NOT NULL CHECK (opportunity_revision > 0),
    version INTEGER NOT NULL CHECK (version > 0),
    admitted_event_id TEXT NOT NULL UNIQUE,
    owner_namespace TEXT NOT NULL,
    data TEXT NOT NULL,
    UNIQUE (opportunity_id, opportunity_revision)
);

INSERT INTO _migrations(id) VALUES ('022_attention_admissions');
