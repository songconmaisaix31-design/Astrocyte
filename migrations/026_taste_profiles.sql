-- 完整品味配置只保存人工输入；启动和升级均不生成默认配置。
CREATE TABLE attention_taste_profiles (
    id TEXT PRIMARY KEY CHECK(id = 'attention'),
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL CHECK(json_valid(data) AND json_type(data) = 'object'
        AND json_extract(data, '$.id') IS 'attention'
        AND json_extract(data, '$.version') IS version)
);
CREATE TABLE attention_taste_profile_revisions (
    profile_id TEXT NOT NULL REFERENCES attention_taste_profiles(id),
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL CHECK(json_valid(data) AND json_type(data) = 'object'
        AND json_extract(data, '$.id') IS profile_id
        AND json_extract(data, '$.version') IS version),
    PRIMARY KEY(profile_id, version)
);
