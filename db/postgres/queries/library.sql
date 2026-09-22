-- name: GetLibraryByID :one
SELECT * FROM library
WHERE id = $1 LIMIT 1;

-- name: GetLibraryPath :one
SELECT path FROM library
WHERE id = $1 LIMIT 1;

-- name: ListLibraries :many
SELECT * FROM library
ORDER BY name ASC;

-- name: CountLibraries :one
SELECT count(*) FROM library;

-- name: UpsertLibrary :one
INSERT INTO library (
    id, name, path, remote_path, default_new_users, created_at, updated_at
) VALUES (
    CASE WHEN @id::int > 0 THEN @id::int ELSE nextval('library_id_seq') END,
    @name::varchar,
    @path::text,
    @remote_path::text,
    @default_new_users::boolean,
    @created_at::timestamptz,
    @updated_at::timestamptz
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    path = EXCLUDED.path,
    remote_path = EXCLUDED.remote_path,
    default_new_users = EXCLUDED.default_new_users,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteLibrary :exec
DELETE FROM library
WHERE id = $1;

-- name: GetUsersWithLibraryAccess :many
SELECT u.* FROM "user" u
JOIN user_library ul ON u.id = ul.user_id
WHERE ul.library_id = $1
UNION
SELECT * FROM "user" WHERE is_admin = TRUE
ORDER BY user_name ASC;

-- name: ScanBegin :exec
UPDATE library
SET full_scan_in_progress = $2, last_scan_started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ScanEnd :exec
UPDATE library
SET full_scan_in_progress = FALSE, last_scan_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ScanInProgress :one
SELECT EXISTS (
    SELECT 1 FROM library WHERE full_scan_in_progress = TRUE
) AS in_progress;

-- name: RefreshLibraryStats :exec
UPDATE library l SET
    total_songs = COALESCE((SELECT count(*) FROM media_file WHERE library_id = l.id AND missing = FALSE), 0),
    total_albums = COALESCE((SELECT count(*) FROM album WHERE library_id = l.id AND missing = FALSE), 0),
    total_artists = COALESCE((SELECT count(DISTINCT mfa.artist_id) FROM media_file_artists mfa JOIN media_file mf ON mfa.media_file_id = mf.id WHERE mf.library_id = l.id AND mf.missing = FALSE), 0),
    total_folders = COALESCE((SELECT count(*) FROM folder WHERE library_id = l.id AND missing = FALSE), 0),
    total_files = COALESCE((SELECT count(*) FROM media_file WHERE library_id = l.id), 0),
    total_missing_files = COALESCE((SELECT count(*) FROM media_file WHERE library_id = l.id AND missing = TRUE), 0),
    total_size = COALESCE((SELECT sum(size) FROM media_file WHERE library_id = l.id AND missing = FALSE), 0),
    total_duration = COALESCE((SELECT sum(duration) FROM media_file WHERE library_id = l.id AND missing = FALSE), 0),
    updated_at = CURRENT_TIMESTAMP
WHERE l.id = $1;
