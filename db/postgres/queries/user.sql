-- name: GetUserByID :one
SELECT * FROM "user"
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM "user"
WHERE LOWER(user_name) = LOWER($1) LIMIT 1;

-- name: FindFirstAdmin :one
SELECT * FROM "user"
WHERE is_admin = TRUE
ORDER BY created_at ASC
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM "user"
ORDER BY user_name ASC;

-- name: CountUsers :one
SELECT count(*) FROM "user";

-- name: UpsertUser :one
INSERT INTO "user" (
    id, user_name, name, email, password, is_admin, token_epoch, scrobble_filter, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id) DO UPDATE SET
    user_name = EXCLUDED.user_name,
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    password = CASE WHEN EXCLUDED.password != '' THEN EXCLUDED.password ELSE "user".password END,
    is_admin = EXCLUDED.is_admin,
    token_epoch = EXCLUDED.token_epoch,
    scrobble_filter = EXCLUDED.scrobble_filter,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: UpdateLastLoginAt :exec
UPDATE "user"
SET last_login_at = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateLastAccessAt :exec
UPDATE "user"
SET last_access_at = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;

-- name: GetUserLibraries :many
SELECT l.* FROM library l
JOIN user_library ul ON l.id = ul.library_id
WHERE ul.user_id = $1
ORDER BY l.name ASC;

-- name: AddUserLibrary :exec
INSERT INTO user_library (user_id, library_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ClearUserLibraries :exec
DELETE FROM user_library
WHERE user_id = $1;
