CREATE TABLE attention_openviking_submissions (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    request_id TEXT NOT NULL UNIQUE,
    space_id TEXT NOT NULL REFERENCES attention_project_spaces(id),
    material_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision > 0),
    base_url TEXT NOT NULL,
    account TEXT NOT NULL,
    user_id TEXT NOT NULL,
    space_version INTEGER NOT NULL CHECK(space_version > 0),
    destination TEXT NOT NULL,
    status TEXT NOT NULL,
    data TEXT NOT NULL CHECK(json_valid(data) AND json_type(data) = 'object'
        AND json_extract(data, '$.id') IS id
        AND json_extract(data, '$.version') IS version
        AND json_extract(data, '$.request_id') IS request_id
        AND json_extract(data, '$.space_id') IS space_id
        AND json_extract(data, '$.material_id') IS material_id
        AND json_extract(data, '$.revision') IS revision
        AND json_extract(data, '$.base_url') IS base_url
        AND json_extract(data, '$.account') IS account
        AND json_extract(data, '$.user') IS user_id
        AND json_extract(data, '$.space_version') IS space_version
        AND json_extract(data, '$.destination') IS destination
        AND json_extract(data, '$.status') IS status)
);
CREATE UNIQUE INDEX attention_openviking_active_source
ON attention_openviking_submissions(space_id, material_id, revision, base_url, account, user_id, destination)
WHERE status <> 'failed';
