-- Cached model-inferred progress observations from approved fixed TASK/STATUS
-- files. Never authoritative project state, a grant or a human approval.
CREATE TABLE local_project_progress (
    project_id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision >= 1)
);
UPDATE foundation_schema SET version = 6 WHERE context = 'workspace';
INSERT OR IGNORE INTO _migrations(id) VALUES ('013_project_progress');
