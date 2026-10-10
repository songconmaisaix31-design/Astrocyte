-- Public metadata cache and explicit human placement/managed-clone observations.
-- No credentials or implicit project grants are stored here.
CREATE TABLE local_github_repositories (
    id TEXT PRIMARY KEY,
    revision INTEGER NOT NULL CHECK (revision >= 1),
    data TEXT NOT NULL
);
UPDATE foundation_schema SET version = 4 WHERE context = 'workspace';
INSERT OR IGNORE INTO _migrations(id) VALUES ('009_github_repositories');
