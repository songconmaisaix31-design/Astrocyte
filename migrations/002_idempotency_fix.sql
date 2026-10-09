-- Migration 002: Fix foundation_idempotency table per SPEC §12.6
-- Uses caller+command_type+idempotency_key composite PK instead of single key.

-- Drop the incorrectly-keyed table (was unused in S0, safe to recreate).
DROP TABLE IF EXISTS foundation_idempotency;

-- Recreate with correct composite primary key per SPEC §12.6:
-- "业务命令按调用者、用例、Idempotency-Key 去重"
CREATE TABLE IF NOT EXISTS foundation_idempotency (
    caller          TEXT NOT NULL,
    command_type    TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_digest  TEXT NOT NULL,
    result_code     TEXT NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (caller, command_type, idempotency_key)
);

-- Record this migration
INSERT OR IGNORE INTO _migrations (id) VALUES ('002_idempotency_fix');
