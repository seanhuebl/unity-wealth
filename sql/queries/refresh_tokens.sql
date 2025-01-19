-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
        id,
        token_hash,
        expires_at,
        revoked_at,
        user_id,
        device_info_id
    )
VALUES (
        gen_random_uuid(),
        ?1,
        DATETIME('now', '+60 days'),
        NULL,
        ?2,
        ?3
    );
-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = ?1
    AND device_info_id = ?2
    AND revoked_at IS NULL;