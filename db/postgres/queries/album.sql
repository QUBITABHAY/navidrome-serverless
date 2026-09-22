-- name: GetAlbumByID :one
SELECT * FROM album
WHERE id = $1 LIMIT 1;

-- name: ListAlbums :many
SELECT * FROM album
WHERE missing = FALSE
ORDER BY sort_album_name ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: ListAlbumsByArtist :many
SELECT * FROM album
WHERE (artist_id = $1 OR album_artist_id = $1) AND missing = FALSE
ORDER BY min_year DESC, name ASC;

-- name: CountAlbums :one
SELECT count(*) FROM album
WHERE missing = FALSE;

-- name: UpsertAlbum :one
INSERT INTO album (
    id, library_id, name, artist_id, artist, album_artist, album_artist_id,
    embed_art_path, cover_art_path, cover_art_id, max_year, min_year, date,
    max_original_year, min_original_year, original_date, release_date,
    compilation, comment, song_count, duration, size, discs,
    sort_album_name, sort_album_artist_name, order_album_name, order_album_artist_name,
    catalog_num, mbz_album_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment,
    mbz_release_group_id, folder_ids, explicit_status, rg_album_gain, rg_album_peak,
    description, small_image_url, medium_image_url, large_image_url, external_url,
    external_info_updated_at, genre, tags, missing, imported_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $15, $16, $17,
    $18, $19, $20, $21, $22, $23,
    $24, $25, $26, $27,
    $28, $29, $30, $31, $32,
    $33, $34, $35, $36, $37,
    $38, $39, $40, $41, $42,
    $43, $44, $45, $46, $47, $48, $49
)
ON CONFLICT (id) DO UPDATE SET
    library_id = EXCLUDED.library_id,
    name = EXCLUDED.name,
    artist_id = EXCLUDED.artist_id,
    artist = EXCLUDED.artist,
    album_artist = EXCLUDED.album_artist,
    album_artist_id = EXCLUDED.album_artist_id,
    embed_art_path = EXCLUDED.embed_art_path,
    cover_art_path = EXCLUDED.cover_art_path,
    cover_art_id = EXCLUDED.cover_art_id,
    max_year = EXCLUDED.max_year,
    min_year = EXCLUDED.min_year,
    date = EXCLUDED.date,
    max_original_year = EXCLUDED.max_original_year,
    min_original_year = EXCLUDED.min_original_year,
    original_date = EXCLUDED.original_date,
    release_date = EXCLUDED.release_date,
    compilation = EXCLUDED.compilation,
    comment = EXCLUDED.comment,
    song_count = EXCLUDED.song_count,
    duration = EXCLUDED.duration,
    size = EXCLUDED.size,
    discs = EXCLUDED.discs,
    sort_album_name = EXCLUDED.sort_album_name,
    sort_album_artist_name = EXCLUDED.sort_album_artist_name,
    order_album_name = EXCLUDED.order_album_name,
    order_album_artist_name = EXCLUDED.order_album_artist_name,
    catalog_num = EXCLUDED.catalog_num,
    mbz_album_id = EXCLUDED.mbz_album_id,
    mbz_album_artist_id = EXCLUDED.mbz_album_artist_id,
    mbz_album_type = EXCLUDED.mbz_album_type,
    mbz_album_comment = EXCLUDED.mbz_album_comment,
    mbz_release_group_id = EXCLUDED.mbz_release_group_id,
    folder_ids = EXCLUDED.folder_ids,
    explicit_status = EXCLUDED.explicit_status,
    rg_album_gain = EXCLUDED.rg_album_gain,
    rg_album_peak = EXCLUDED.rg_album_peak,
    description = EXCLUDED.description,
    small_image_url = EXCLUDED.small_image_url,
    medium_image_url = EXCLUDED.medium_image_url,
    large_image_url = EXCLUDED.large_image_url,
    external_url = EXCLUDED.external_url,
    external_info_updated_at = EXCLUDED.external_info_updated_at,
    genre = EXCLUDED.genre,
    tags = EXCLUDED.tags,
    missing = EXCLUDED.missing,
    imported_at = EXCLUDED.imported_at,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteAlbum :exec
DELETE FROM album
WHERE id = $1;

-- name: AddAlbumArtist :exec
INSERT INTO album_artists (album_id, artist_id, role, sub_role)
VALUES ($1, $2, $3, $4)
ON CONFLICT (artist_id, album_id, role, sub_role) DO NOTHING;

-- name: GetAlbumArtists :many
SELECT aa.*, a.name as artist_name FROM album_artists aa
JOIN artist a ON aa.artist_id = a.id
WHERE aa.album_id = $1;
