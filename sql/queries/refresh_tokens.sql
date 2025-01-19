-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
        id,
        token,
        expires_at,
        revoked_at,
        user_id,
    )
VALUES (
        gen_random_uuid(),
        ?1,
        ?2,
        DATETIME('now', '+60 days'),
        NULL,
        $3
    );