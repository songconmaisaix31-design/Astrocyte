-- S1 Attention persistence. Aggregate versions are distinct from immutable revisions.
CREATE TABLE attention_materials (
    id TEXT PRIMARY KEY, source_key TEXT NOT NULL UNIQUE,
    version INTEGER NOT NULL CHECK(version > 0), data TEXT NOT NULL
);
CREATE TABLE attention_material_revisions (
    material_id TEXT NOT NULL REFERENCES attention_materials(id),
    revision INTEGER NOT NULL CHECK(revision > 0), content_digest TEXT NOT NULL, data TEXT NOT NULL,
    PRIMARY KEY(material_id, revision), UNIQUE(material_id, content_digest)
);
CREATE TABLE attention_distillations (
    id TEXT PRIMARY KEY, reuse_key TEXT NOT NULL UNIQUE, data TEXT NOT NULL
);
CREATE TABLE attention_opportunities (
    id TEXT PRIMARY KEY, version INTEGER NOT NULL CHECK(version > 0), data TEXT NOT NULL
);
CREATE TABLE attention_opportunity_revisions (
    opportunity_id TEXT NOT NULL REFERENCES attention_opportunities(id),
    revision INTEGER NOT NULL CHECK(revision > 0), data TEXT NOT NULL,
    PRIMARY KEY(opportunity_id, revision)
);
CREATE TABLE attention_jobs (
    id TEXT PRIMARY KEY, dedupe_key TEXT NOT NULL UNIQUE,
    version INTEGER NOT NULL CHECK(version > 0),
    status TEXT NOT NULL CHECK(status IN ('queued','running','succeeded','failed','cancelled')),
    data TEXT NOT NULL, payload TEXT NOT NULL, caller TEXT NOT NULL
);
CREATE INDEX attention_jobs_status ON attention_jobs(status);
CREATE TABLE attention_receipts (
    caller TEXT NOT NULL, command TEXT NOT NULL, key TEXT NOT NULL,
    digest TEXT NOT NULL, result TEXT NOT NULL, PRIMARY KEY(caller, command, key)
);
CREATE TABLE attention_outbox (
    id TEXT PRIMARY KEY, type TEXT NOT NULL, aggregate_id TEXT NOT NULL,
    aggregate_version INTEGER NOT NULL, occurred_at TEXT NOT NULL,
    correlation_id TEXT NOT NULL, causation_id TEXT NOT NULL, payload TEXT NOT NULL
);
INSERT OR IGNORE INTO _migrations(id) VALUES ('003_attention');
INSERT INTO foundation_schema(context, version) VALUES ('attention', 1);
