ALTER TABLE attention_engine_knowledge ADD COLUMN consolidated_at INTEGER;
INSERT OR IGNORE INTO _migrations(id) VALUES ('018_attention_engine_consolidation_time');
