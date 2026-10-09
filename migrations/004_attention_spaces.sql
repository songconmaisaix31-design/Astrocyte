-- Human classification and explicit project-space references. Neither table
-- grants execution authority or default Agent access.
CREATE TABLE attention_domains (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL
);
CREATE TABLE attention_project_spaces (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL
);
UPDATE foundation_schema SET version = 2 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('004_attention_spaces');
