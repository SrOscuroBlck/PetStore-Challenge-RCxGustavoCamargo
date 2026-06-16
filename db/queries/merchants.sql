-- name: CreateMerchant :exec
INSERT INTO merchants (id, email_hash, email_encrypted, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMerchantByEmailHash :one
SELECT * FROM merchants WHERE email_hash = $1;

-- name: GetMerchantByID :one
SELECT * FROM merchants WHERE id = $1;
