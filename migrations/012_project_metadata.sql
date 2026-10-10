-- Human annotations stay independent of replaceable discovery observations.
-- A discovered project ID is not a space binding or an Agent permission.
CREATE TABLE local_project_metadata (
    project_id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision >= 1)
);
UPDATE foundation_schema SET version = 5 WHERE context = 'workspace';
INSERT OR IGNORE INTO _migrations(id) VALUES ('012_project_metadata');
