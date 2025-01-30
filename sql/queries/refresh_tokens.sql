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
        ?1,
        ?2,
        ?3,
        NULL,
        ?4,
        ?5
    );
-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = ?1
WHERE user_id = ?2
    AND device_info_id = ?3
    AND revoked_at IS NULL;