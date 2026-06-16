-- name: CreateStore :exec
INSERT INTO stores (id, merchant_id, name, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetStoreByMerchantID :one
SELECT * FROM stores WHERE merchant_id = $1;

-- name: GetStoreByID :one
SELECT * FROM stores WHERE id = $1;
