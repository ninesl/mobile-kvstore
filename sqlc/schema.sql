PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
-- PRAGMA busy_timeout = 5000;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS blobs (
    blob_id INTEGER PRIMARY KEY AUTOINCREMENT,
    blob_key TEXT NOT NULL,
    blob_value BLOB NOT NULL,
    updated_at INTEGER NOT NULL,

    UNIQUE (blob_key)
);

CREATE TABLE IF NOT EXISTS blob_entries (
    namespace TEXT NOT NULL,
    subject INTEGER NOT NULL,
    id INTEGER NOT NULL,
    meta_tag TEXT NOT NULL,
    blob_id INTEGER NOT NULL,

    PRIMARY KEY (namespace, subject, id, meta_tag, blob_id),

    FOREIGN KEY (blob_id)
        REFERENCES blobs(blob_id)
        ON DELETE CASCADE
);
