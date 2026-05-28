-- name: UpsertBlob :one
INSERT INTO blobs (
    blob_key,
    blob_value,
    updated_at
) VALUES (
    sqlc.arg(blob_key),
    sqlc.arg(blob_value),
    sqlc.arg(updated_at)
)
ON CONFLICT(blob_key) DO UPDATE SET
    blob_value = excluded.blob_value,
    updated_at = excluded.updated_at
RETURNING blob_id;

-- name: UpdateBlobValue :exec
UPDATE blobs
SET blob_key = sqlc.arg(blob_key),
    blob_value = sqlc.arg(blob_value),
    updated_at = sqlc.arg(updated_at)
WHERE blob_id = sqlc.arg(blob_id);

-- name: GetBlobValuesByFilter :many
SELECT DISTINCT b.blob_id, b.blob_key, b.blob_value
FROM blob_entries br
JOIN blobs b ON b.blob_id = br.blob_id
WHERE (sqlc.narg(namespace) IS NULL OR br.namespace = sqlc.narg(namespace))
  AND (sqlc.narg(subject) IS NULL OR br.subject = sqlc.narg(subject))
  AND (sqlc.narg(id) IS NULL OR br.id = sqlc.narg(id))
  AND (sqlc.narg(meta_tag) IS NULL OR br.meta_tag = sqlc.narg(meta_tag));

-- name: GetBlobIDByBlobKey :one
SELECT b.blob_id
FROM (SELECT sqlc.arg(blob_key) AS blob_key) AS k
LEFT JOIN blobs b ON b.blob_key = k.blob_key;

-- name: CountBlobEntriesByFilter :one
-- sqlc.narg values are nullable filter inputs, not nullable blob_entries columns.
-- A nil filter field becomes `NULL IS NULL OR column = NULL`, which is TRUE,
-- so that field's predicate is ignored while non-nil fields still filter normally.
SELECT COUNT(DISTINCT blob_id)
FROM blob_entries
WHERE (sqlc.narg(namespace) IS NULL OR namespace = sqlc.narg(namespace))
  AND (sqlc.narg(subject) IS NULL OR subject = sqlc.narg(subject))
  AND (sqlc.narg(id) IS NULL OR id = sqlc.narg(id))
  AND (sqlc.narg(meta_tag) IS NULL OR meta_tag = sqlc.narg(meta_tag));

-- name: DeleteBlobsByFilter :exec
DELETE FROM blobs
WHERE blob_id IN (
    SELECT DISTINCT blob_id
    FROM blob_entries
    WHERE (sqlc.narg(namespace) IS NULL OR namespace = sqlc.narg(namespace))
      AND (sqlc.narg(subject) IS NULL OR subject = sqlc.narg(subject))
      AND (sqlc.narg(id) IS NULL OR id = sqlc.narg(id))
      AND (sqlc.narg(meta_tag) IS NULL OR meta_tag = sqlc.narg(meta_tag))
);

-- name: DeleteBlobByBlobKey :exec
DELETE FROM blobs
WHERE blob_key = sqlc.arg(blob_key);
