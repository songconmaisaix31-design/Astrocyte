-- 实际判断结果、人工反馈与逐题型策略；版本和权限由业务服务核实。
CREATE TABLE judgment_results (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_kind TEXT NOT NULL CHECK(actor_kind IN ('human','agent')),
    result_version INTEGER NOT NULL CHECK(result_version = 1),
    cache_key TEXT NOT NULL,
    cacheable INTEGER NOT NULL CHECK(cacheable IN (0,1)),
    request_json TEXT NOT NULL CHECK(json_valid(request_json)),
    result_json TEXT NOT NULL CHECK(json_valid(result_json)),
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_results_cache ON judgment_results(cache_key, cacheable, created_at DESC);

CREATE TABLE judgment_feedback (
    id TEXT PRIMARY KEY,
    result_id TEXT NOT NULL REFERENCES judgment_results(id),
    result_version INTEGER NOT NULL CHECK(result_version = 1),
    actor_id TEXT NOT NULL,
    actor_kind TEXT NOT NULL CHECK(actor_kind = 'human'),
    kind TEXT NOT NULL CHECK(kind IN ('adopt','later','reject','revise','already_solved')),
    reason TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX judgment_feedback_result ON judgment_feedback(result_id, created_at);

CREATE TABLE judgment_policies (
    project_id TEXT NOT NULL,
    question_type TEXT NOT NULL CHECK(question_type IN ('choice','score','truth')),
    version INTEGER NOT NULL CHECK(version >= 1),
    data TEXT NOT NULL CHECK(json_valid(data)),
    PRIMARY KEY(project_id, question_type)
);

INSERT INTO foundation_schema(context, version) VALUES ('judgment', 1);
INSERT INTO _migrations(id) VALUES ('024_judgment');
