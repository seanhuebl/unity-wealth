-- name: GetDeviceInfoByUser :one
SELECT id
FROM device_info_logs
WHERE user_id = ?1
    AND device_type = ?2
    AND browser = ?3
    AND browser_version = ?4
    AND os = ?5
    AND os_version = ?6
LIMIT 1;
-- name: CreateDeviceInfo :one
INSERT INTO device_info_logs (
        id,
        user_id,
        device_type,
        browser,
        browser_version,
        os,
        os_version
    )
VALUES (
        ?1,
        ?2,
        ?3,
        ?4,
        ?5,
        ?6,
        ?7
    )
RETURNING id;