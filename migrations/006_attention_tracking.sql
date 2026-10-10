-- Public metadata bindings and revisions; selection reuses the existing jobs.
CREATE TABLE attention_tracking_sources (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    platform TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    external_id TEXT NOT NULL,
    data TEXT NOT NULL,
    UNIQUE(platform, source_kind, external_id)
);
CREATE TABLE attention_source_items (
    source_id TEXT NOT NULL REFERENCES attention_tracking_sources(id),
    external_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision > 0),
    data TEXT NOT NULL,
    PRIMARY KEY(source_id, external_id)
);
UPDATE foundation_schema SET version = 4 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('006_attention_tracking');
