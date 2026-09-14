-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE id=$1;

-- name: GetUserByName :one
SELECT * FROM users WHERE name=$1;

-- name: CountUsersByName :one
SELECT count(*) as users_count FROM users WHERE name=$1;

-- name: ResetUsers :exec
TRUNCATE TABLE users CASCADE;