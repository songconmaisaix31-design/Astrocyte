-- Swarm 任务、工作项、租约、成果与事件去重存储。
CREATE TABLE swarm_missions (
    id TEXT PRIMARY KEY,
    grant_id TEXT NOT NULL UNIQUE,
    project_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL
);

CREATE TABLE swarm_work_items (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES swarm_missions(id),
    version INTEGER NOT NULL,
    fence INTEGER NOT NULL DEFAULT 0,
    data TEXT NOT NULL
);

CREATE TABLE swarm_leases (
    work_item_id TEXT PRIMARY KEY REFERENCES swarm_work_items(id),
    mission_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    holder_id TEXT NOT NULL,
    fence INTEGER NOT NULL,
    operation_id TEXT NOT NULL UNIQUE,
    data TEXT NOT NULL
);

CREATE TABLE swarm_artifacts (
    id TEXT PRIMARY KEY,
    work_item_id TEXT NOT NULL REFERENCES swarm_work_items(id),
    version INTEGER NOT NULL,
    data TEXT NOT NULL
);

CREATE TABLE swarm_consumed_events (
    consumer_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    PRIMARY KEY (consumer_id, event_id)
);

INSERT INTO foundation_schema(context, version) VALUES ('swarm', 1);
INSERT INTO _migrations(id) VALUES ('020_swarm');
