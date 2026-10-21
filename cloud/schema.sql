-- Sloth Mode account/device metadata and *explicitly opted-in* cloud task tracking.
-- Private local notes, prompts and model output are NEVER stored in these tables.
CREATE TABLE IF NOT EXISTS device_links (
 account_id TEXT NOT NULL,
 device_id TEXT NOT NULL,
 label TEXT NOT NULL,
 token_hash TEXT NOT NULL UNIQUE,
 created_at INTEGER NOT NULL,
 last_seen_at INTEGER,
 revoked_at INTEGER,
 PRIMARY KEY (account_id,device_id)
);
CREATE INDEX IF NOT EXISTS idx_sloth_devices_account ON device_links(account_id,revoked_at);
CREATE TABLE IF NOT EXISTS enrollment_codes (
 code_hash TEXT PRIMARY KEY,
 account_id TEXT NOT NULL,
 expires_at INTEGER NOT NULL,
 consumed_at INTEGER,
 issued_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sloth_enroll_account ON enrollment_codes(account_id,expires_at);

-- Cloud tasks are added only from the authenticated account page.
-- Account IDs are server-derived; browsers cannot pick another account ID.
CREATE TABLE IF NOT EXISTS account_tasks (
 account_id TEXT NOT NULL,
 task_id TEXT NOT NULL,
 task_text TEXT NOT NULL CHECK(length(task_text) BETWEEN 1 AND 280),
 done INTEGER NOT NULL DEFAULT 0 CHECK(done IN (0,1)),
 created_at INTEGER NOT NULL,
 updated_at INTEGER NOT NULL,
 PRIMARY KEY(account_id,task_id)
);
CREATE INDEX IF NOT EXISTS idx_sloth_tasks_account ON account_tasks(account_id,created_at DESC);
