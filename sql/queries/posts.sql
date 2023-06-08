-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsByUser :many
WITH user_id_from_api_key AS (
    SELECT id
    FROM users
    WHERE api_key = $1
),
user_feed_follows AS (
    SELECT ff.feed_id
    FROM feed_follows ff
    JOIN user_id_from_api_key u ON ff.user_id = u.user_id
)
SELECT f.id
FROM feeds f
JOIN user_feed_follows uff ON f.feed_id = uff.feed_id
ORDER BY f.last_fetched_at
LIMIT $2;