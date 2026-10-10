-- Rebuild the two related tables atomically with foreign_keys still enabled.
-- Preserve all existing IDs, versions, selected items and JSON observations.
CREATE TABLE tracking_source_items_upgrade AS SELECT * FROM attention_source_items;
CREATE TABLE tracking_sources_upgrade AS SELECT * FROM attention_tracking_sources;
DROP TABLE attention_source_items;
DROP TABLE attention_tracking_sources;
CREATE TABLE attention_tracking_sources (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    platform TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    external_id TEXT NOT NULL,
    owner_id TEXT NOT NULL DEFAULT '',
    access_mode TEXT NOT NULL DEFAULT 'public' CHECK(access_mode IN ('public','browser_selected')),
    data TEXT NOT NULL,
    UNIQUE(platform, source_kind, owner_id, external_id, access_mode)
);
INSERT INTO attention_tracking_sources(id,version,platform,source_kind,external_id,owner_id,access_mode,data)
SELECT id,version,platform,source_kind,external_id,
       COALESCE(json_extract(data,'$.owner_id'),''),
       COALESCE(NULLIF(json_extract(data,'$.access_mode'),''),'public'),data
FROM tracking_sources_upgrade;
CREATE TABLE attention_source_items (
    source_id TEXT NOT NULL REFERENCES attention_tracking_sources(id),
    external_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision > 0),
    data TEXT NOT NULL,
    PRIMARY KEY(source_id, external_id)
);
INSERT INTO attention_source_items SELECT * FROM tracking_source_items_upgrade;
DROP TABLE tracking_source_items_upgrade;
DROP TABLE tracking_sources_upgrade;
UPDATE foundation_schema SET version = 5 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('010_tracking_source_identity');
