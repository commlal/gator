-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    TO_TIMESTAMP($4,'Dy, DD Mon YYYY HH24:MI:SS TMZ'),
    $5
)
RETURNING *;

-- name: PostInDatabase :one
SELECT * FROM posts
WHERE url = $1;

-- name: GetPostsForUser :many
SELECT posts.title AS title, posts.description AS body, posts.published_at AS publication_date FROM posts
INNER JOIN feed_follows
ON posts.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY published_at 
LIMIT $2;