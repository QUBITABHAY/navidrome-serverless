-- name: GetAnnotation :one
SELECT * FROM annotation
WHERE user_id = $1 AND item_id = $2 AND item_type = $3 LIMIT 1;

-- name: SetStar :exec
INSERT INTO annotation (
    ann_id, user_id, item_id, item_type, starred, starred_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET
    starred = EXCLUDED.starred,
    starred_at = EXCLUDED.starred_at;

-- name: SetRating :exec
INSERT INTO annotation (
    ann_id, user_id, item_id, item_type, rating, rated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET
    rating = EXCLUDED.rating,
    rated_at = EXCLUDED.rated_at;

-- name: IncPlayCount :exec
INSERT INTO annotation (
    ann_id, user_id, item_id, item_type, play_count, play_date
) VALUES (
    $1, $2, $3, $4, 1, $5
)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET
    play_count = COALESCE(annotation.play_count, 0) + 1,
    play_date = EXCLUDED.play_date;
