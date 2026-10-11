-- 014: Attention Engine cognitive-core tables (task book §4). M0 contract lock:
-- schema only. The per-domain stores (event/signal/engine/knowledge/experience)
-- are written by later tracks against these tables. Nothing here grants
-- authority, imports credentials or performs model processing.

CREATE TABLE attention_engine_events (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    ts INTEGER NOT NULL,
    kind TEXT NOT NULL,
    topic TEXT NOT NULL DEFAULT '',
    data TEXT NOT NULL
);
CREATE INDEX attention_engine_events_topic ON attention_engine_events(topic);
CREATE INDEX attention_engine_events_source_kind ON attention_engine_events(source, kind);

CREATE TABLE attention_engine_signals (
    event_id TEXT NOT NULL REFERENCES attention_engine_events(id),
    topic TEXT NOT NULL,
    class TEXT NOT NULL,
    weight REAL NOT NULL,
    ts INTEGER NOT NULL,
    data TEXT NOT NULL
);
CREATE INDEX attention_engine_signals_topic ON attention_engine_signals(topic);

CREATE TABLE attention_engine_nodes (
    id TEXT PRIMARY KEY,
    topic TEXT NOT NULL UNIQUE,
    attention REAL NOT NULL,
    momentum REAL NOT NULL,
    confidence REAL NOT NULL,
    last_active INTEGER NOT NULL,
    data TEXT NOT NULL
);

CREATE TABLE attention_engine_experiences (
    id TEXT PRIMARY KEY,
    pattern TEXT NOT NULL,
    verdict TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('quarantine','approved')),
    data TEXT NOT NULL
);
CREATE INDEX attention_engine_experiences_status ON attention_engine_experiences(status);

CREATE TABLE attention_engine_knowledge (
    id TEXT PRIMARY KEY,
    topic TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('inbox','confirmed')),
    data TEXT NOT NULL
);
CREATE INDEX attention_engine_knowledge_status ON attention_engine_knowledge(status);

CREATE TABLE attention_engine_conflicts (
    id TEXT PRIMARY KEY,
    topic TEXT NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending','resolved_keep_left','resolved_keep_right','resolved_merge','dismissed')),
    data TEXT NOT NULL
);
CREATE INDEX attention_engine_conflicts_status ON attention_engine_conflicts(status);

UPDATE foundation_schema SET version = 7 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('014_attention_engine');
