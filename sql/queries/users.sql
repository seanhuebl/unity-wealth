-- name: CreateUser :exec
INSERT INTO users (id, email, hashed_password)
VALUES (gen_random_uuid(), ?1, ?2);