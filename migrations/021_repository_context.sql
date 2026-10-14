-- 仅保存仓库上下文索引的版本与自有路径元数据。
CREATE TABLE repository_context_indexes (
    project_id TEXT NOT NULL,
    worktree_id TEXT NOT NULL,
    backend TEXT NOT NULL,
    generation TEXT NOT NULL,
    metadata_json TEXT NOT NULL,
    PRIMARY KEY (project_id, worktree_id, backend)
);

UPDATE foundation_schema SET version = 8 WHERE context = 'workspace';
INSERT INTO _migrations(id) VALUES ('021_repository_context');
