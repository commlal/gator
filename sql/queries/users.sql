-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    gen_random_uuid(), 
    NOW(),
    NOW(),
    $1
)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;

-- name: PurgeUsers :exec
DELETE FROM users;

-- name: GetAllUsers :many
SELECT name FROM users;

-- name: GetUserByID :one
SELECT name FROM users
WHERE id = $1;