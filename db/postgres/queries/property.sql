-- name: GetProperty :one
SELECT value FROM property
WHERE id = $1 LIMIT 1;

-- name: PutProperty :exec
INSERT INTO property (id, value)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
    value = EXCLUDED.value;

-- name: DeleteProperty :exec
DELETE FROM property
WHERE id = $1;
