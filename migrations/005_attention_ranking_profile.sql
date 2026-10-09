-- User-owned, single Attention candidate ranking profile. No default composite
-- weights are inserted. Revisions remain immutable and settings grant no access.
CREATE TABLE attention_ranking_profiles (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL
);
CREATE TABLE attention_ranking_profile_revisions (
    profile_id TEXT NOT NULL REFERENCES attention_ranking_profiles(id),
    version INTEGER NOT NULL CHECK(version > 0),
    data TEXT NOT NULL,
    PRIMARY KEY(profile_id, version)
);
UPDATE foundation_schema SET version = 3 WHERE context = 'attention';
INSERT OR IGNORE INTO _migrations(id) VALUES ('005_attention_ranking_profile');
