-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetSecret :one
SELECT * FROM secret
WHERE id = $1 
LIMIT 1;

-- name: ListSecret :many
SELECT * FROM secret
ORDER BY created_at desc;

-- name: CreateSecret :one
INSERT INTO secret (
    encrypted_password,
    nonce
) VALUES (
  $1, $2
)
RETURNING *;

-- name: DeleteSecret :exec
delete from secret
WHERE id = $1;

-- name: PurgeSecret :exec
DELETE FROM secret
WHERE created_at < NOW() - INTERVAL '6 hours';