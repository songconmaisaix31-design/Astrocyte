CREATE TABLE attention_engine_knowledge_references (
 entry_id TEXT NOT NULL REFERENCES attention_engine_knowledge(id),
 project_id TEXT NOT NULL CHECK(project_id!=''),
 knowledge_version INTEGER NOT NULL CHECK(knowledge_version>=1),
 crystallized_version INTEGER,
 actor_id TEXT NOT NULL CHECK(actor_id!=''),
 actor_kind TEXT NOT NULL CHECK(actor_kind='human'),
 ts INTEGER NOT NULL,
 request_id TEXT NOT NULL,
 PRIMARY KEY(entry_id,request_id),
 FOREIGN KEY(entry_id,crystallized_version) REFERENCES attention_engine_knowledge_versions(entry_id,version)
);
CREATE TRIGGER attention_engine_knowledge_references_update BEFORE UPDATE ON attention_engine_knowledge_references
BEGIN SELECT RAISE(ABORT,'knowledge references are append-only'); END;
CREATE TRIGGER attention_engine_knowledge_references_delete BEFORE DELETE ON attention_engine_knowledge_references
BEGIN SELECT RAISE(ABORT,'knowledge references are append-only'); END;
INSERT OR IGNORE INTO _migrations(id) VALUES ('017_attention_engine_references');
