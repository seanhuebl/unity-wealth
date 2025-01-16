-- name: CreateUser :exec
INSERT INTO users (id, email, hashed_password)
VALUES (gen_random_uuid(), ?1, ?2);
-- name: GetUserByEmail :one
SELECT id,
    hashed_password
FROM users
WHERE email = ?1;