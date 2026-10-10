-- Metadata-only catalog observations gathered by explicit human discovery.
CREATE TABLE attention_source_catalogs (
    platform TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    access_mode TEXT NOT NULL CHECK(access_mode IN ('public','browser_selected')),
    data TEXT NOT NULL,
    PRIMARY KEY(platform,owner_id,access_mode)
);
UPDATE foundation_schema SET version = 6 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('011_source_catalogs');
