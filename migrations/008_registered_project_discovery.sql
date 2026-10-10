-- Cached observations of already registered project roots; never permissions.
-- One bounded snapshot is replaced atomically by the discovery repository.
CREATE TABLE local_project_discovery (
    id TEXT PRIMARY KEY CHECK(id = 'registered'),
    data TEXT NOT NULL
);
UPDATE foundation_schema SET version = 3 WHERE context = 'workspace';
INSERT OR IGNORE INTO _migrations(id) VALUES ('008_registered_project_discovery');
