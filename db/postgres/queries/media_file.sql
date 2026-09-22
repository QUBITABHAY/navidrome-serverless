-- name: GetMediaFileByID :one
SELECT * FROM media_file
WHERE id = $1 LIMIT 1;

-- name: GetMediaFileByPath :one
SELECT * FROM media_file
WHERE path = $1 LIMIT 1;

-- name: ListMediaFilesByAlbum :many
SELECT * FROM media_file
WHERE album_id = $1 AND missing = FALSE
ORDER BY disc_number ASC, track_number ASC;

-- name: ListMediaFilesByArtist :many
SELECT * FROM media_file
WHERE (artist_id = $1 OR album_artist_id = $1) AND missing = FALSE
ORDER BY year DESC, album ASC, disc_number ASC, track_number ASC;

-- name: CountMediaFiles :one
SELECT count(*) FROM media_file
WHERE missing = FALSE;

-- name: UpsertMediaFile :one
INSERT INTO media_file (
    id, pid, library_id, folder_id, path, title, album, artist, artist_id,
    album_artist, album_artist_id, album_id, has_cover_art, track_number, disc_number,
    disc_subtitle, year, date, original_year, original_date, release_year, release_date,
    size, suffix, duration, bit_rate, sample_rate, bit_depth, channels, codec,
    probe_data, genre, sort_title, sort_album_name, sort_artist_name, sort_album_artist_name,
    order_title, order_album_name, order_artist_name, order_album_artist_name,
    compilation, comment, lyrics, bpm, explicit_status, catalog_num,
    mbz_recording_id, mbz_release_track_id, mbz_album_id, mbz_release_group_id,
    mbz_artist_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment,
    rg_album_gain, rg_album_peak, rg_track_gain, rg_track_peak,
    tags, participants, missing, birth_time, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22,
    $23, $24, $25, $26, $27, $28, $29, $30,
    $31, $32, $33, $34, $35, $36,
    $37, $38, $39, $40,
    $41, $42, $43, $44, $45, $46,
    $47, $48, $49, $50,
    $51, $52, $53, $54,
    $55, $56, $57, $58,
    $59, $60, $61, $62, $63, $64
)
ON CONFLICT (id) DO UPDATE SET
    pid = EXCLUDED.pid,
    library_id = EXCLUDED.library_id,
    folder_id = EXCLUDED.folder_id,
    path = EXCLUDED.path,
    title = EXCLUDED.title,
    album = EXCLUDED.album,
    artist = EXCLUDED.artist,
    artist_id = EXCLUDED.artist_id,
    album_artist = EXCLUDED.album_artist,
    album_artist_id = EXCLUDED.album_artist_id,
    album_id = EXCLUDED.album_id,
    has_cover_art = EXCLUDED.has_cover_art,
    track_number = EXCLUDED.track_number,
    disc_number = EXCLUDED.disc_number,
    disc_subtitle = EXCLUDED.disc_subtitle,
    year = EXCLUDED.year,
    date = EXCLUDED.date,
    original_year = EXCLUDED.original_year,
    original_date = EXCLUDED.original_date,
    release_year = EXCLUDED.release_year,
    release_date = EXCLUDED.release_date,
    size = EXCLUDED.size,
    suffix = EXCLUDED.suffix,
    duration = EXCLUDED.duration,
    bit_rate = EXCLUDED.bit_rate,
    sample_rate = EXCLUDED.sample_rate,
    bit_depth = EXCLUDED.bit_depth,
    channels = EXCLUDED.channels,
    codec = EXCLUDED.codec,
    probe_data = EXCLUDED.probe_data,
    genre = EXCLUDED.genre,
    sort_title = EXCLUDED.sort_title,
    sort_album_name = EXCLUDED.sort_album_name,
    sort_artist_name = EXCLUDED.sort_artist_name,
    sort_album_artist_name = EXCLUDED.sort_album_artist_name,
    order_title = EXCLUDED.order_title,
    order_album_name = EXCLUDED.order_album_name,
    order_artist_name = EXCLUDED.order_artist_name,
    order_album_artist_name = EXCLUDED.order_album_artist_name,
    compilation = EXCLUDED.compilation,
    comment = EXCLUDED.comment,
    lyrics = EXCLUDED.lyrics,
    bpm = EXCLUDED.bpm,
    explicit_status = EXCLUDED.explicit_status,
    catalog_num = EXCLUDED.catalog_num,
    mbz_recording_id = EXCLUDED.mbz_recording_id,
    mbz_release_track_id = EXCLUDED.mbz_release_track_id,
    mbz_album_id = EXCLUDED.mbz_album_id,
    mbz_release_group_id = EXCLUDED.mbz_release_group_id,
    mbz_artist_id = EXCLUDED.mbz_artist_id,
    mbz_album_artist_id = EXCLUDED.mbz_album_artist_id,
    mbz_album_type = EXCLUDED.mbz_album_type,
    mbz_album_comment = EXCLUDED.mbz_album_comment,
    rg_album_gain = EXCLUDED.rg_album_gain,
    rg_album_peak = EXCLUDED.rg_album_peak,
    rg_track_gain = EXCLUDED.rg_track_gain,
    rg_track_peak = EXCLUDED.rg_track_peak,
    tags = EXCLUDED.tags,
    participants = EXCLUDED.participants,
    missing = EXCLUDED.missing,
    birth_time = EXCLUDED.birth_time,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteMediaFile :exec
DELETE FROM media_file
WHERE id = $1;

-- name: AddMediaFileArtist :exec
INSERT INTO media_file_artists (media_file_id, artist_id, role, sub_role)
VALUES ($1, $2, $3, $4)
ON CONFLICT (artist_id, media_file_id, role, sub_role) DO NOTHING;

-- name: SearchMediaFiles :many
SELECT * FROM media_file
WHERE missing = FALSE
  AND (title ILIKE '%' || $1 || '%' OR artist ILIKE '%' || $1 || '%' OR album ILIKE '%' || $1 || '%')
ORDER BY title ASC
LIMIT $2 OFFSET $3;
