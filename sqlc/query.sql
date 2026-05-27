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

-- name: InsertBlobRef :exec
INSERT INTO blob_refs (
    namespace,
    subject,
    id,
    meta_tag,
    blob_id
) VALUES (
    sqlc.arg(namespace),
    sqlc.arg(subject),
    sqlc.arg(id),
    sqlc.arg(meta_tag),
    sqlc.arg(blob_id)
);

-- name: GetBlobValuesByScope :many
SELECT DISTINCT b.blob_id, b.blob_value
FROM blob_refs br
JOIN blobs b ON b.blob_id = br.blob_id
WHERE br.namespace = sqlc.arg(namespace)
  AND br.subject = sqlc.arg(subject);

-- name: GetBlobValuesByRef :many
SELECT DISTINCT b.blob_id, b.blob_value
FROM blob_refs br
JOIN blobs b ON b.blob_id = br.blob_id
WHERE br.namespace = sqlc.arg(namespace)
  AND br.subject = sqlc.arg(subject)
  AND br.id = sqlc.arg(id)
  AND br.meta_tag = sqlc.arg(meta_tag);
