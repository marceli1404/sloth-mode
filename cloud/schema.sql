-- Sloth Mode 0.11 account metadata. NEVER stores task text or Gemma prompts.
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
