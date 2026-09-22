-- name: GetPlaylistByID :one
SELECT * FROM playlist
WHERE id = $1 LIMIT 1;

-- name: ListPlaylists :many
SELECT * FROM playlist
WHERE public = TRUE OR owner_id = $1
ORDER BY name ASC;

-- name: UpsertPlaylist :one
INSERT INTO playlist (
    id, name, comment, duration, size, song_count, owner_id,
    public, path, sync, uploaded_image, external_image_url, imported_hash,
    rules, evaluated_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $15, $16, $17
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    comment = EXCLUDED.comment,
    duration = EXCLUDED.duration,
    size = EXCLUDED.size,
    song_count = EXCLUDED.song_count,
    owner_id = EXCLUDED.owner_id,
    public = EXCLUDED.public,
    path = EXCLUDED.path,
    sync = EXCLUDED.sync,
    uploaded_image = EXCLUDED.uploaded_image,
    external_image_url = EXCLUDED.external_image_url,
    imported_hash = EXCLUDED.imported_hash,
    rules = EXCLUDED.rules,
    evaluated_at = EXCLUDED.evaluated_at,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeletePlaylist :exec
DELETE FROM playlist
WHERE id = $1;

-- name: GetPlaylistTracks :many
SELECT mf.* FROM media_file mf
JOIN playlist_tracks pt ON mf.id = pt.media_file_id
WHERE pt.playlist_id = $1
ORDER BY pt.id ASC;

-- name: AddPlaylistTrack :exec
INSERT INTO playlist_tracks (id, playlist_id, media_file_id)
VALUES ($1, $2, $3)
ON CONFLICT (playlist_id, id) DO UPDATE SET
    media_file_id = EXCLUDED.media_file_id;

-- name: ClearPlaylistTracks :exec
DELETE FROM playlist_tracks
WHERE playlist_id = $1;
