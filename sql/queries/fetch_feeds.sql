-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds
WHERE last_fetched_at IS NULL or last_fetched_at < NOW() - INTERVAL '60 minutes'
ORDER BY last_fetched_at
LIMIT $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2, updated_at = $3
WHERE id = $1;