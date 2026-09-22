-- name: GetArtistByID :one
SELECT * FROM artist
WHERE id = $1 LIMIT 1;

-- name: ListArtists :many
SELECT * FROM artist
WHERE missing = FALSE
ORDER BY sort_artist_name ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: CountArtists :one
SELECT count(*) FROM artist
WHERE missing = FALSE;

-- name: UpsertArtist :one
INSERT INTO artist (
    id, name, sort_artist_name, order_artist_name, mbz_artist_id,
    biography, small_image_url, medium_image_url, large_image_url,
    external_url, similar_artists, external_info_updated_at, missing,
    uploaded_image, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13,
    $14, $15, $16
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    sort_artist_name = EXCLUDED.sort_artist_name,
    order_artist_name = EXCLUDED.order_artist_name,
    mbz_artist_id = EXCLUDED.mbz_artist_id,
    biography = EXCLUDED.biography,
    small_image_url = EXCLUDED.small_image_url,
    medium_image_url = EXCLUDED.medium_image_url,
    large_image_url = EXCLUDED.large_image_url,
    external_url = EXCLUDED.external_url,
    similar_artists = EXCLUDED.similar_artists,
    external_info_updated_at = EXCLUDED.external_info_updated_at,
    missing = EXCLUDED.missing,
    uploaded_image = EXCLUDED.uploaded_image,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeleteArtist :exec
DELETE FROM artist
WHERE id = $1;

-- name: AddLibraryArtist :exec
INSERT INTO library_artist (library_id, artist_id, stats)
VALUES ($1, $2, $3)
ON CONFLICT (library_id, artist_id) DO UPDATE SET
    stats = EXCLUDED.stats;
