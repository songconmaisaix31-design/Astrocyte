-- 新业务来源元数据与既有经验同事务保存；既有引擎表和迁移保持原结构。
CREATE TABLE experience_completion_sources (
    experience_id TEXT PRIMARY KEY REFERENCES attention_engine_experiences(id),
    source_kind TEXT NOT NULL CHECK (source_kind IN ('swarm_artifact_accepted', 'relay_task_completed')),
    event_id TEXT NOT NULL CHECK (length(trim(event_id)) > 0),
    owner_project TEXT NOT NULL CHECK (length(trim(owner_project)) > 0),
    local_project_id TEXT NOT NULL CHECK (length(trim(local_project_id)) > 0),
    data TEXT NOT NULL CHECK (json_valid(data)),
    UNIQUE (owner_project, event_id)
);

INSERT INTO _migrations(id) VALUES ('023_experience_completion_sources');
