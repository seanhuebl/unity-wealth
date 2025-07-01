-- name: CreateUser :exec
INSERT INTO users (id, email, hashed_password)
VALUES ($1, $2, $3);
-- name: GetUserByEmail :one
SELECT id,
    hashed_password
FROM users
WHERE email = $1;