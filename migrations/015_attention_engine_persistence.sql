-- Additive upgrade from M0. Legacy rows have no asserted origin/raw bytes or
-- project owner; stores must not treat those unknown fields as verified scope.
ALTER TABLE attention_engine_events ADD COLUMN origin_key TEXT;
ALTER TABLE attention_engine_events ADD COLUMN received_at INTEGER;
ALTER TABLE attention_engine_events ADD COLUMN raw_bytes BLOB CHECK(raw_bytes IS NULL OR typeof(raw_bytes)='blob');
ALTER TABLE attention_engine_events ADD COLUMN observation TEXT;
ALTER TABLE attention_engine_events ADD COLUMN mapping_status TEXT NOT NULL DEFAULT 'mapped' CHECK(mapping_status IN ('pending','mapped'));
CREATE UNIQUE INDEX attention_engine_events_origin ON attention_engine_events(origin_key) WHERE origin_key IS NOT NULL;
CREATE UNIQUE INDEX attention_engine_signals_receipt ON attention_engine_signals(event_id,topic,class);
CREATE TRIGGER attention_engine_signal_mapped BEFORE INSERT ON attention_engine_signals
WHEN (SELECT mapping_status FROM attention_engine_events WHERE id=NEW.event_id) != 'mapped'
BEGIN SELECT RAISE(ABORT,'pending observation cannot produce signal'); END;

ALTER TABLE attention_engine_nodes ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0;
UPDATE attention_engine_nodes SET updated_at=last_active;
ALTER TABLE attention_engine_nodes ADD COLUMN lifecycle TEXT NOT NULL DEFAULT 'active';
ALTER TABLE attention_engine_nodes ADD COLUMN version INTEGER NOT NULL DEFAULT 1 CHECK(version>=1);

ALTER TABLE attention_engine_knowledge ADD COLUMN owner_project TEXT NOT NULL DEFAULT '';
ALTER TABLE attention_engine_knowledge ADD COLUMN version INTEGER NOT NULL DEFAULT 1 CHECK(version>=1);
ALTER TABLE attention_engine_knowledge ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE attention_engine_knowledge ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0;
CREATE INDEX attention_engine_knowledge_owner ON attention_engine_knowledge(owner_project,status);
CREATE TABLE attention_engine_knowledge_versions (
 entry_id TEXT NOT NULL REFERENCES attention_engine_knowledge(id),
 version INTEGER NOT NULL CHECK(version>=1),
 content TEXT NOT NULL,
 recorded_at INTEGER NOT NULL,
 PRIMARY KEY(entry_id,version)
);
CREATE TRIGGER attention_engine_knowledge_versions_update BEFORE UPDATE ON attention_engine_knowledge_versions
BEGIN SELECT RAISE(ABORT,'crystallized versions are append-only'); END;
CREATE TRIGGER attention_engine_knowledge_versions_delete BEFORE DELETE ON attention_engine_knowledge_versions
BEGIN SELECT RAISE(ABORT,'crystallized versions are append-only'); END;
CREATE TABLE attention_engine_knowledge_shares (
 entry_id TEXT NOT NULL REFERENCES attention_engine_knowledge(id),
 owner_project TEXT NOT NULL CHECK(owner_project!=''),
 target_project TEXT NOT NULL CHECK(target_project!='' AND target_project!=owner_project),
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 PRIMARY KEY(entry_id,target_project)
);
CREATE TABLE attention_engine_knowledge_actions (
 entry_id TEXT NOT NULL REFERENCES attention_engine_knowledge(id),
 action TEXT NOT NULL CHECK(action IN ('confirm','demote','share','reference')),
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 version INTEGER NOT NULL,
 PRIMARY KEY(entry_id,request_id)
);

ALTER TABLE attention_engine_experiences ADD COLUMN owner_project TEXT NOT NULL DEFAULT '';
ALTER TABLE attention_engine_experiences ADD COLUMN origin_event_id TEXT REFERENCES attention_engine_events(id);
ALTER TABLE attention_engine_experiences ADD COLUMN version INTEGER NOT NULL DEFAULT 1 CHECK(version>=1);
ALTER TABLE attention_engine_experiences ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE attention_engine_experiences ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX attention_engine_experiences_origin ON attention_engine_experiences(owner_project,origin_event_id) WHERE origin_event_id IS NOT NULL;
CREATE INDEX attention_engine_experiences_owner ON attention_engine_experiences(owner_project,status);
CREATE TABLE attention_engine_experience_shares (
 experience_id TEXT NOT NULL REFERENCES attention_engine_experiences(id),
 owner_project TEXT NOT NULL CHECK(owner_project!=''),
 target_project TEXT NOT NULL CHECK(target_project!='' AND target_project!=owner_project),
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 PRIMARY KEY(experience_id,target_project)
);
CREATE TABLE attention_engine_experience_actions (
 experience_id TEXT NOT NULL REFERENCES attention_engine_experiences(id),
 action TEXT NOT NULL CHECK(action IN ('approve','demote','share')),
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 version INTEGER NOT NULL,
 PRIMARY KEY(experience_id,request_id)
);

CREATE TABLE attention_engine_sleep_runs (
 id TEXT PRIMARY KEY,
 day TEXT NOT NULL UNIQUE,
 started_at INTEGER NOT NULL,
 completed_at INTEGER NOT NULL,
 data TEXT NOT NULL
);
CREATE TABLE attention_engine_sleep_metrics (
 run_id TEXT PRIMARY KEY REFERENCES attention_engine_sleep_runs(id),
 key_facts_before INTEGER NOT NULL CHECK(key_facts_before>=0),
 key_facts_retained INTEGER NOT NULL CHECK(key_facts_retained>=0 AND key_facts_retained<=key_facts_before),
 bytes_before INTEGER NOT NULL CHECK(bytes_before>=0),
 bytes_after INTEGER NOT NULL CHECK(bytes_after>=0),
 retention_precision REAL NOT NULL,
 storage_reduction REAL NOT NULL
);
CREATE TABLE attention_engine_sleep_candidates (
 run_id TEXT NOT NULL REFERENCES attention_engine_sleep_runs(id),
 candidate_id TEXT NOT NULL,
 kind TEXT NOT NULL,
 data TEXT NOT NULL,
 PRIMARY KEY(run_id,candidate_id)
);
CREATE TABLE attention_engine_sleep_conflicts (
 run_id TEXT NOT NULL REFERENCES attention_engine_sleep_runs(id),
 conflict_id TEXT NOT NULL REFERENCES attention_engine_conflicts(id),
 PRIMARY KEY(run_id,conflict_id)
);
ALTER TABLE attention_engine_conflicts ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE attention_engine_conflicts ADD COLUMN version INTEGER NOT NULL DEFAULT 1 CHECK(version>=1);
CREATE TABLE attention_engine_conflict_actions (
 conflict_id TEXT NOT NULL REFERENCES attention_engine_conflicts(id),
 action TEXT NOT NULL,
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 version INTEGER NOT NULL,
 PRIMARY KEY(conflict_id,request_id)
);
UPDATE foundation_schema SET version=8 WHERE context='attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('015_attention_engine_persistence');
