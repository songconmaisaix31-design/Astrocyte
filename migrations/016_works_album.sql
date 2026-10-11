CREATE TABLE works_album (
 key TEXT PRIMARY KEY,
 data TEXT NOT NULL
);
CREATE TABLE works_metadata (
 work_id TEXT PRIMARY KEY,
 data TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision>=1)
);
CREATE TABLE works_requests (
 caller_id TEXT NOT NULL,
 command TEXT NOT NULL,
 request_key TEXT NOT NULL,
 input BLOB NOT NULL CHECK(typeof(input)='blob'),
 status TEXT NOT NULL CHECK(status IN ('pending','accepted','unknown','rejected')),
 result BLOB CHECK(result IS NULL OR typeof(result)='blob'),
 error BLOB CHECK(error IS NULL OR typeof(error)='blob'),
 PRIMARY KEY(caller_id,command,request_key)
);
UPDATE foundation_schema SET version=7 WHERE context='workspace';
INSERT OR IGNORE INTO _migrations(id) VALUES ('016_works_album');
